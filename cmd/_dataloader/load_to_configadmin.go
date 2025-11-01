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
	log.Println("=== BlueConfig - Load ZipCodes to ConfigAdmin ===")

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

	// Create store in configadmin stores directory
	storesDir := "../configadmin/stores"
	if err := os.MkdirAll(storesDir, 0755); err != nil {
		log.Fatalf("Failed to create stores directory: %v", err)
	}

	storePath := filepath.Join(storesDir, "zipcodes.db")
	log.Printf("Creating/Opening database: %s", storePath)

	tree, err := blueconfig.NewOrOpenTree(blueconfig.TreeOptions{
		StorageLocationOnDisk: storePath,
		Port:                  0,
		Token:                 "",
	})
	if err != nil {
		log.Fatalf("Failed to create/open zipcode store: %v", err)
	}
	defer tree.Close()

	// Initialize store metadata for ConfigAdmin
	log.Println("Setting up store metadata...")
	err = tree.CreateNodeWithProps("root/__storeinfo", map[string]interface{}{
		"__title":      "US Zip Codes",
		"__icon":       "fa-map-marker-alt",
		"__color":      "#e74c3c",
		"name":         "zipcodes",
		"displayName":  "US Zip Codes",
		"description":  "US Cities and Zip Code database with geographic information",
		"createdAt":    time.Now().Format(time.RFC3339),
		"lastAccessed": time.Now().Format(time.RFC3339),
		"recordCount":  fmt.Sprintf("%d", len(cities)),
	})
	if err != nil {
		log.Fatalf("Failed to create store metadata: %v", err)
	}

	// Load zipcodes using batch method for best performance
	log.Println("Loading zipcodes (this may take a moment)...")
	startTime := time.Now()

	// Strategy: Load in batches to balance memory and performance
	batchSize := 1000
	totalBatches := (len(cities) + batchSize - 1) / batchSize

	tree.CreatePath("root/zipcodes")

	for batchNum := 0; batchNum < totalBatches; batchNum++ {
		start := batchNum * batchSize
		end := start + batchSize
		if end > len(cities) {
			end = len(cities)
		}

		batch := cities[start:end]
		nodes := make(map[string]map[string]interface{})

		for _, city := range batch {
			zipPath := fmt.Sprintf("root/zipcodes/%d", city.ZipCode)
			nodes[zipPath] = map[string]interface{}{
				"zip_code":  city.ZipCode,
				"latitude":  city.Latitude,
				"longitude": city.Longitude,
				"city":      city.City,
				"state":     city.State,
				"county":    city.County,
			}
		}

		if err := tree.BatchCreateNodes(nodes); err != nil {
			log.Fatalf("Failed to load batch %d: %v", batchNum+1, err)
		}

		log.Printf("  Progress: %d/%d records (%.1f%%)", end, len(cities), float64(end)/float64(len(cities))*100)
	}

	duration := time.Since(startTime)

	// Create some useful indexes/views
	log.Println("\nCreating state index...")
	stateMap := make(map[string]int)
	for _, city := range cities {
		if city.State != "" {
			stateMap[city.State]++
		}
	}

	tree.CreatePath("root/__indexes/by_state")
	stateNodes := make(map[string]map[string]interface{})
	for state, count := range stateMap {
		statePath := fmt.Sprintf("root/__indexes/by_state/%s", state)
		stateNodes[statePath] = map[string]interface{}{
			"state":      state,
			"zip_count":  count,
			"index_type": "state",
		}
	}
	tree.BatchCreateNodes(stateNodes)

	log.Println("\n=== Loading Complete ===")
	log.Printf("Total records: %d", len(cities))
	log.Printf("Unique states: %d", len(stateMap))
	log.Printf("Database location: %s", storePath)
	log.Printf("Load time: %v", duration)
	log.Printf("Average: %.2f records/sec", float64(len(cities))/duration.Seconds())
	log.Println("\nStore is now available in ConfigAdmin!")
	log.Println("\nData structure:")
	log.Println("  - root/zipcodes/{zipcode} - All zipcode records")
	log.Println("  - root/__indexes/by_state/{state} - State statistics")
	log.Println("\nYou can now:")
	log.Println("  1. Start ConfigAdmin: cd ../configadmin && ./start.sh")
	log.Println("  2. Select 'US Zip Codes' store from the UI")
	log.Println("  3. Browse and query zipcode data")
}
