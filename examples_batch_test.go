package blueconfig

import (
	"fmt"
	"log"
)

// ExampleTree_CreateNodeWithProps demonstrates creating a node with properties in a single operation
func ExampleTree_CreateNodeWithProps() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/example_create_node.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	// Create a user node with all properties in one transaction
	err = tree.CreateNodeWithProps("/users/john", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
		"role":  "admin",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the properties
	props, _ := tree.GetAllPropsWithValues("/users/john")
	fmt.Printf("User: %s, Email: %s\n", props["name"], props["email"])

	// Output:
	// User: John Doe, Email: john@example.com
}

// ExampleTree_SetNodeWithProps demonstrates updating/creating a node with properties
func ExampleTree_SetNodeWithProps() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/example_set_node.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	// SetNodeWithProps creates the path if it doesn't exist and sets properties
	err = tree.SetNodeWithProps("/config/database", map[string]interface{}{
		"host":     "localhost",
		"port":     5432,
		"database": "myapp",
		"ssl":      true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Read a specific value
	port, _ := tree.GetValue("/config/database/port")
	fmt.Printf("Database port: %s\n", port)

	// Output:
	// Database port: 5432
}

// ExampleTree_BatchCreateNodes demonstrates creating multiple nodes in a single transaction
func ExampleTree_BatchCreateNodes() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/example_batch.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	// Create multiple users in a single transaction - much faster!
	users := map[string]map[string]interface{}{
		"/users/alice": {
			"name":  "Alice Smith",
			"email": "alice@example.com",
			"role":  "developer",
		},
		"/users/bob": {
			"name":  "Bob Johnson",
			"email": "bob@example.com",
			"role":  "designer",
		},
		"/users/carol": {
			"name":  "Carol Williams",
			"email": "carol@example.com",
			"role":  "manager",
		},
	}

	err = tree.BatchCreateNodes(users)
	if err != nil {
		log.Fatal(err)
	}

	// List all users
	userList, _ := tree.GetNodesInPath("/users")
	fmt.Printf("Created %d users\n", len(userList))

	// Output:
	// Created 3 users
}

// ExampleTree_BatchCreateNodes_zipCodes demonstrates loading large datasets efficiently
func ExampleTree_BatchCreateNodes_zipCodes() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/zipcodes.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	// Simulate loading zipcode data
	zipcodes := map[string]map[string]interface{}{
		"/zipcodes/10001": {
			"city":      "New York",
			"state":     "NY",
			"latitude":  40.7506,
			"longitude": -73.9971,
		},
		"/zipcodes/90001": {
			"city":      "Los Angeles",
			"state":     "CA",
			"latitude":  33.9731,
			"longitude": -118.2479,
		},
		"/zipcodes/60601": {
			"city":      "Chicago",
			"state":     "IL",
			"latitude":  41.8857,
			"longitude": -87.6189,
		},
	}

	// Load all zipcodes in one transaction
	err = tree.BatchCreateNodes(zipcodes)
	if err != nil {
		log.Fatal(err)
	}

	// Query a specific zipcode
	props, _ := tree.GetAllPropsWithValues("/zipcodes/10001")
	fmt.Printf("%s, %s\n", props["city"], props["state"])

	// Output:
	// New York, NY
}

// ExampleTree_performance_comparison demonstrates the performance difference
func ExampleTree_performance_comparison() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/performance.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	// OLD WAY (slower - multiple transactions):
	// for i := 0; i < 1000; i++ {
	//     path := fmt.Sprintf("/items/%d", i)
	//     tree.CreatePath(path)           // Transaction 1
	//     tree.SetValues(path, props)     // Transaction 2
	// }

	// NEW WAY (faster - single operation per node):
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/items/%d", i)
		tree.CreateNodeWithProps(path, map[string]interface{}{
			"name":  fmt.Sprintf("Item %d", i),
			"price": i * 10,
		})
	}

	// BEST WAY (fastest - all nodes in one transaction):
	items := make(map[string]map[string]interface{})
	for i := 100; i < 200; i++ {
		path := fmt.Sprintf("/items/%d", i)
		items[path] = map[string]interface{}{
			"name":  fmt.Sprintf("Item %d", i),
			"price": i * 10,
		}
	}
	tree.BatchCreateNodes(items)

	count, _ := tree.GetNodesInPath("/items")
	fmt.Printf("Total items: %d\n", len(count))

	// Output:
	// Total items: 200
}
