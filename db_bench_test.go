package blueconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkFixPath(b *testing.B) {
	paths := []string{
		"/apps/web/frontend",
		"//multiple///slashes//path",
		"root/already/prefixed",
		"/",
		"simple",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fixpath(paths[i%len(paths)])
	}
}

func BenchmarkParsePath(b *testing.B) {
	path := "root/apps/web/frontend/port/8080"

	b.Run("offset_0", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			parsePath(path, 0)
		}
	})

	b.Run("offset_1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			parsePath(path, 1)
		}
	})

	b.Run("offset_2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			parsePath(path, 2)
		}
	})
}

func BenchmarkCreatePath(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/bench/path%d", i)
		tr.CreatePath(path)
	}
}

func BenchmarkSetValue(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	tr.CreatePath("/bench/config")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%100)
		tr.SetValue(fmt.Sprintf("/bench/config/%s", key), "value")
	}
}

func BenchmarkGetValue(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	tr.CreatePath("/bench/config")
	for i := 0; i < 100; i++ {
		tr.SetValue(fmt.Sprintf("/bench/config/key%d", i), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%100)
		tr.GetValue(fmt.Sprintf("/bench/config/%s", key))
	}
}

func BenchmarkSetValues(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	tr.CreatePath("/bench/config")

	values := map[string]interface{}{
		"host":     "localhost",
		"port":     "8080",
		"timeout":  "30",
		"maxconns": "100",
		"debug":    "true",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.SetValues("/bench/config", values)
	}
}

func BenchmarkGetAllPropsWithValues(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	tr.CreatePath("/bench/config")
	values := map[string]interface{}{
		"key1":  "value1",
		"key2":  "value2",
		"key3":  "value3",
		"key4":  "value4",
		"key5":  "value5",
		"key6":  "value6",
		"key7":  "value7",
		"key8":  "value8",
		"key9":  "value9",
		"key10": "value10",
	}
	tr.SetValues("/bench/config", values)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.GetAllPropsWithValues("/bench/config")
	}
}

func BenchmarkGetNodesInPath(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	// Create multiple nodes
	for i := 0; i < 10; i++ {
		tr.CreatePath(fmt.Sprintf("/bench/nodes/node%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.GetNodesInPath("/bench/nodes")
	}
}

func BenchmarkDeleteValue(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	tr.CreatePath("/bench/config")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		key := fmt.Sprintf("key%d", i)
		tr.SetValue(fmt.Sprintf("/bench/config/%s", key), "value")
		b.StartTimer()

		tr.DeleteValue("/bench/config", key)
	}
}

func BenchmarkHasValue(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	tr.CreatePath("/bench/config")
	tr.SetValue("/bench/config/exists", "yes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.HasValue("/bench/config", "exists")
	}
}

// ============================================================================
// Complex Workflow Benchmarks
// ============================================================================

func BenchmarkFullWorkflow(b *testing.B) {
	b.Run("create_set_get_delete", func(b *testing.B) {
		tr, _ := createTestTree(&testing.T{})
		defer tr.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path := fmt.Sprintf("/bench/workflow%d", i%100)
			tr.CreatePath(path)
			tr.SetValue(fmt.Sprintf("%s/key", path), "value")
			tr.GetValue(fmt.Sprintf("%s/key", path))
			tr.DeleteValue(path, "key")
		}
	})
}

func BenchmarkHierarchicalOperations(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create hierarchical structure
		basePath := fmt.Sprintf("/bench/app%d", i%10)
		tr.CreatePath(fmt.Sprintf("%s/service1/config", basePath))
		tr.CreatePath(fmt.Sprintf("%s/service2/config", basePath))

		// Set values
		tr.SetValue(fmt.Sprintf("%s/service1/config/port", basePath), "8080")
		tr.SetValue(fmt.Sprintf("%s/service2/config/port", basePath), "8081")

		// Navigate hierarchy
		tr.GetNodesInPath(basePath)
		tr.GetValue(fmt.Sprintf("%s/service1/config/port", basePath))
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	// Setup data
	tr.CreatePath("/bench/concurrent")
	for i := 0; i < 100; i++ {
		tr.SetValue(fmt.Sprintf("/bench/concurrent/key%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%100)
			tr.GetValue(fmt.Sprintf("/bench/concurrent/%s", key))
			i++
		}
	})
}

func BenchmarkConcurrentWrites(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	tr.CreatePath("/bench/concurrent")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i)
			tr.SetValue(fmt.Sprintf("/bench/concurrent/%s", key), "value")
			i++
		}
	})
}

func BenchmarkMixedOperations(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	tr.CreatePath("/bench/mixed")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		switch i % 4 {
		case 0:
			tr.SetValue(fmt.Sprintf("/bench/mixed/key%d", i%100), "value")
		case 1:
			tr.GetValue(fmt.Sprintf("/bench/mixed/key%d", i%100))
		case 2:
			tr.GetAllProps("/bench/mixed")
		case 3:
			tr.HasValue("/bench/mixed", fmt.Sprintf("key%d", i%100))
		}
	}
}

// ============================================================================
// Memory Allocation Benchmarks
// ============================================================================

func BenchmarkMemoryAllocation(b *testing.B) {
	tr, _ := createTestTree(&testing.T{})
	defer tr.Close()

	tr.CreatePath("/bench/memory")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tr.SetValue(fmt.Sprintf("/bench/memory/key%d", i%100), "value")
		tr.GetValue(fmt.Sprintf("/bench/memory/key%d", i%100))
	}
}

// ============================================================================
// US Cities Zipcode Data Loading Benchmark
// ============================================================================

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

func BenchmarkLoadUSCities(b *testing.B) {
	// Read the JSON file once
	data, err := os.ReadFile("cmd/configadmin/USCities.json")
	if err != nil {
		b.Fatalf("Failed to read USCities.json: %v", err)
	}

	var allCities []CityRecord
	err = json.Unmarshal(data, &allCities)
	if err != nil {
		b.Fatalf("Failed to parse USCities.json: %v", err)
	}

	b.Logf("Loaded %d city records from JSON", len(allCities))

	// Use a subset for load_all test (first 1000 records)
	cities := allCities
	if len(cities) > 1000 {
		cities = cities[:1000]
	}

	b.Run("load_1000_cities", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			tr, _ := createTestTree(&testing.T{})
			tr.CreatePath("/zipcodes")
			b.StartTimer()

			// Load cities using batch method
			for _, city := range cities {
				zipPath := fmt.Sprintf("/zipcodes/%d", city.ZipCode)

				values := map[string]interface{}{
					"latitude":  city.Latitude,
					"longitude": city.Longitude,
					"city":      city.City,
					"state":     city.State,
					"county":    city.County,
				}

				tr.CreateNodeWithProps(zipPath, values)
			}

			b.StopTimer()
			tr.Close()
		}
	})

	b.Run("load_and_query_100", func(b *testing.B) {
		// Use smaller subset for load and query
		querySet := cities
		if len(querySet) > 100 {
			querySet = querySet[:100]
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			tr, _ := createTestTree(&testing.T{})
			tr.CreatePath("/zipcodes")

			// Load cities using batch method
			for _, city := range querySet {
				zipPath := fmt.Sprintf("/zipcodes/%d", city.ZipCode)

				values := map[string]interface{}{
					"latitude":  city.Latitude,
					"longitude": city.Longitude,
					"city":      city.City,
					"state":     city.State,
					"county":    city.County,
				}

				tr.CreateNodeWithProps(zipPath, values)
			}
			b.StartTimer()

			// Query zipcodes
			for j := 0; j < 50; j++ {
				city := querySet[j%len(querySet)]
				zipPath := fmt.Sprintf("/zipcodes/%d", city.ZipCode)
				tr.GetAllPropsWithValues(zipPath)
			}

			b.StopTimer()
			tr.Close()
		}
	})

	b.Run("incremental_load", func(b *testing.B) {
		tr, _ := createTestTree(&testing.T{})
		defer tr.Close()

		tr.CreatePath("/zipcodes")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			city := cities[i%len(cities)]
			zipPath := fmt.Sprintf("/zipcodes/%d", city.ZipCode)

			values := map[string]interface{}{
				"latitude":  city.Latitude,
				"longitude": city.Longitude,
				"city":      city.City,
				"state":     city.State,
				"county":    city.County,
			}

			tr.CreateNodeWithProps(zipPath, values)
		}
	})

	b.Run("query_existing", func(b *testing.B) {
		// Pre-load data
		tr, _ := createTestTree(&testing.T{})
		defer tr.Close()

		tr.CreatePath("/zipcodes")
		for _, city := range cities {
			zipPath := fmt.Sprintf("/zipcodes/%d", city.ZipCode)

			values := map[string]interface{}{
				"latitude":  city.Latitude,
				"longitude": city.Longitude,
				"city":      city.City,
				"state":     city.State,
				"county":    city.County,
			}

			tr.CreateNodeWithProps(zipPath, values)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			city := cities[i%len(cities)]
			zipPath := fmt.Sprintf("/zipcodes/%d", city.ZipCode)
			tr.GetAllPropsWithValues(zipPath)
		}
	})

	b.Run("batch_create_all", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			tr, _ := createTestTree(&testing.T{})
			tr.CreatePath("/zipcodes")

			// Prepare batch data
			nodes := make(map[string]map[string]interface{})
			for _, city := range cities {
				zipPath := fmt.Sprintf("/zipcodes/%d", city.ZipCode)
				nodes[zipPath] = map[string]interface{}{
					"latitude":  city.Latitude,
					"longitude": city.Longitude,
					"city":      city.City,
					"state":     city.State,
					"county":    city.County,
				}
			}
			b.StartTimer()

			// Load all cities in one transaction
			tr.BatchCreateNodes(nodes)

			b.StopTimer()
			tr.Close()
		}
	})
}
