package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sfi2k7/blueconfig"
)

// CityRecord represents a US city zipcode entry
type CityRecord struct {
	ZipCode   int     `json:"zip_code"`
	Latitude  float64 `json:"latitude,string"`
	Longitude float64 `json:"longitude,string"`
	City      string  `json:"city"`
	State     string  `json:"state"`
	County    string  `json:"county"`
}

// UnmarshalJSON handles empty string values for latitude and longitude
func (c *CityRecord) UnmarshalJSON(data []byte) error {
	type Alias CityRecord
	aux := &struct {
		Latitude  interface{} `json:"latitude"`
		Longitude interface{} `json:"longitude"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle latitude
	switch v := aux.Latitude.(type) {
	case float64:
		c.Latitude = v
	case string:
		if v == "" {
			c.Latitude = 0.0
		} else {
			fmt.Sscanf(v, "%f", &c.Latitude)
		}
	}

	// Handle longitude
	switch v := aux.Longitude.(type) {
	case float64:
		c.Longitude = v
	case string:
		if v == "" {
			c.Longitude = 0.0
		} else {
			fmt.Sscanf(v, "%f", &c.Longitude)
		}
	}

	return nil
}

func main() {
	log.Println("=== BlueConfig ZipCode Data Loader ===")

	// Read the JSON file
	jsonPath := "../configadmin/USCities.json"
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("Failed to read USCities.json: %v", err)
	}

	var cities []CityRecord
	err = json.Unmarshal(data, &cities)
	if err != nil {
		log.Fatalf("Failed to parse USCities.json: %v", err)
	}

	log.Printf("Loaded %d city records from JSON file", len(cities))

	// Create stores directory if it doesn't exist
	storesDir := "./stores"
	if err := os.MkdirAll(storesDir, 0755); err != nil {
		log.Fatalf("Failed to create stores directory: %v", err)
	}

	// Create or open the zipcode store
	storePath := filepath.Join(storesDir, "zipcodes.db")
	log.Printf("Creating/Opening database: %s", storePath)

	tree, err := blueconfig.NewOrOpenTree(blueconfig.TreeOptions{
		StorageLocationOnDisk: storePath,
		Port:                  0, // No HTTP server
		Token:                 "",
	})
	if err != nil {
		log.Fatalf("Failed to create/open zipcode store: %v", err)
	}
	defer tree.Close()

	// Initialize store metadata
	log.Println("Initializing store metadata...")
	tree.CreatePath("root/__storeinfo")
	tree.SetValue("root/__storeinfo/__title", "ZipCodes")
	tree.SetValue("root/__storeinfo/__icon", "fa-map-marker")
	tree.SetValue("root/__storeinfo/__color", "#3498db")
	tree.SetValue("root/__storeinfo/name", "zipcodes")
	tree.SetValue("root/__storeinfo/displayName", "US Zip Codes")
	tree.SetValue("root/__storeinfo/description", "US Cities and Zip Code database")
	tree.SetValue("root/__storeinfo/createdAt", time.Now().Format(time.RFC3339))
	tree.SetValue("root/__storeinfo/recordCount", fmt.Sprintf("%d", len(cities)))

	// Load zipcodes in a flat structure
	log.Println("Loading zipcodes (flat structure)...")
	startTime := time.Now()
	if err := loadFlatZipCodes(tree, cities); err != nil {
		log.Fatalf("Failed to load flat zipcodes: %v", err)
	}
	flatDuration := time.Since(startTime)
	log.Printf("Flat structure loaded in %v", flatDuration)

	// Load zipcodes in a hierarchical structure (State -> City -> ZipCode)
	log.Println("Loading zipcodes (hierarchical by state)...")
	startTime = time.Now()
	if err := loadHierarchicalByState(tree, cities); err != nil {
		log.Fatalf("Failed to load hierarchical zipcodes: %v", err)
	}
	hierDuration := time.Since(startTime)
	log.Printf("Hierarchical structure loaded in %v", hierDuration)

	// Load zipcodes indexed by state only
	log.Println("Loading zipcodes (by state only)...")
	startTime = time.Now()
	if err := loadByStateOnly(tree, cities); err != nil {
		log.Fatalf("Failed to load by-state zipcodes: %v", err)
	}
	stateDuration := time.Since(startTime)
	log.Printf("By-state structure loaded in %v", stateDuration)

	log.Println("\n=== Loading Complete ===")
	log.Printf("Total records: %d", len(cities))
	log.Printf("Database location: %s", storePath)
	log.Printf("Flat load time: %v", flatDuration)
	log.Printf("Hierarchical load time: %v", hierDuration)
	log.Printf("By-state load time: %v", stateDuration)
	log.Println("\nData structures created:")
	log.Println("  - root/zipcodes/flat/{zipcode}")
	log.Println("  - root/zipcodes/by_state/{state}/{city}/{zipcode}")
	log.Println("  - root/zipcodes/by_state_flat/{state}/{zipcode}")
}

// loadFlatZipCodes loads all zipcodes in a flat structure
// Path: root/zipcodes/flat/{zipcode}
func loadFlatZipCodes(tree *blueconfig.Tree, cities []CityRecord) error {
	tree.CreatePath("root/zipcodes/flat")

	for i, city := range cities {
		zipPath := fmt.Sprintf("root/zipcodes/flat/%d", city.ZipCode)

		values := map[string]interface{}{
			"zip_code":  city.ZipCode,
			"latitude":  city.Latitude,
			"longitude": city.Longitude,
			"city":      city.City,
			"state":     city.State,
			"county":    city.County,
		}

		if err := tree.CreateNodeWithProps(zipPath, values); err != nil {
			return fmt.Errorf("failed to create node with props for %s: %v", zipPath, err)
		}

		if (i+1)%1000 == 0 {
			log.Printf("  Loaded %d/%d records...", i+1, len(cities))
		}
	}

	return nil
}

// loadHierarchicalByState loads zipcodes in a hierarchical structure
// Path: root/zipcodes/by_state/{state}/{city}/{zipcode}
func loadHierarchicalByState(tree *blueconfig.Tree, cities []CityRecord) error {
	tree.CreatePath("root/zipcodes/by_state")

	// Group cities by state and city
	stateMap := make(map[string]map[string][]CityRecord)
	for _, city := range cities {
		if city.State == "" || city.City == "" {
			continue
		}

		if stateMap[city.State] == nil {
			stateMap[city.State] = make(map[string][]CityRecord)
		}
		stateMap[city.State][city.City] = append(stateMap[city.State][city.City], city)
	}

	totalLoaded := 0
	for state, cityMap := range stateMap {
		statePath := fmt.Sprintf("root/zipcodes/by_state/%s", state)
		tree.CreatePath(statePath)

		for cityName, cityRecords := range cityMap {
			// Sanitize city name for use as a path component
			sanitizedCity := sanitizeName(cityName)
			cityPath := fmt.Sprintf("%s/%s", statePath, sanitizedCity)
			tree.CreatePath(cityPath)

			for _, city := range cityRecords {
				zipPath := fmt.Sprintf("%s/%d", cityPath, city.ZipCode)

				values := map[string]interface{}{
					"zip_code":  city.ZipCode,
					"latitude":  city.Latitude,
					"longitude": city.Longitude,
					"city":      city.City,
					"state":     city.State,
					"county":    city.County,
				}

				if err := tree.CreateNodeWithProps(zipPath, values); err != nil {
					return fmt.Errorf("failed to create node with props for %s: %v", zipPath, err)
				}

				totalLoaded++
				if totalLoaded%1000 == 0 {
					log.Printf("  Loaded %d/%d records...", totalLoaded, len(cities))
				}
			}
		}
	}

	return nil
}

// loadByStateOnly loads zipcodes grouped only by state (no city hierarchy)
// Path: root/zipcodes/by_state_flat/{state}/{zipcode}
func loadByStateOnly(tree *blueconfig.Tree, cities []CityRecord) error {
	tree.CreatePath("root/zipcodes/by_state_flat")

	for i, city := range cities {
		if city.State == "" {
			continue
		}

		statePath := fmt.Sprintf("root/zipcodes/by_state_flat/%s", city.State)
		tree.CreatePath(statePath)

		zipPath := fmt.Sprintf("%s/%d", statePath, city.ZipCode)

		values := map[string]interface{}{
			"zip_code":  city.ZipCode,
			"latitude":  city.Latitude,
			"longitude": city.Longitude,
			"city":      city.City,
			"state":     city.State,
			"county":    city.County,
		}

		if err := tree.CreateNodeWithProps(zipPath, values); err != nil {
			return fmt.Errorf("failed to create node with props for %s: %v", zipPath, err)
		}

		if (i+1)%1000 == 0 {
			log.Printf("  Loaded %d/%d records...", i+1, len(cities))
		}
	}

	return nil
}

// sanitizeName removes or replaces invalid characters from names
func sanitizeName(name string) string {
	// Replace problematic characters
	replacer := map[rune]string{
		'/':  "_",
		'\\': "_",
		':':  "_",
		'*':  "_",
		'?':  "_",
		'"':  "_",
		'<':  "_",
		'>':  "_",
		'|':  "_",
		' ':  "_",
		'.':  "_",
	}

	result := ""
	for _, ch := range name {
		if replacement, found := replacer[ch]; found {
			result += replacement
		} else {
			result += string(ch)
		}
	}

	return result
}
