package blueconfig

import (
	"fmt"
	"log"
	"time"
)

// ExampleTree_InitTimeseries demonstrates initializing the timeseries system
func ExampleTree_InitTimeseries() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/example_timeseries.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	// Initialize timeseries structure
	err = tree.InitTimeseries()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Timeseries initialized")
	// Output:
	// Timeseries initialized
}

// ExampleTree_CreateSensor demonstrates creating a CPU usage sensor
func ExampleTree_CreateSensor() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/example_sensor.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	tree.InitTimeseries()

	// Create a CPU usage sensor (gauge type)
	err = tree.CreateSensor("server1_cpu", SensorTypeGauge, "percent", "24h")
	if err != nil {
		log.Fatal(err)
	}

	// Create a network traffic sensor (counter type)
	err = tree.CreateSensor("server1_network_rx", SensorTypeCounter, "bytes", "24h")
	if err != nil {
		log.Fatal(err)
	}

	sensors, _ := tree.ListSensors()
	fmt.Printf("Created %d sensors\n", len(sensors))

	// Output:
	// Created 2 sensors
}

// ExampleTree_PostEvent demonstrates posting sensor events
func ExampleTree_PostEvent() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/example_events.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	tree.InitTimeseries()
	tree.CreateSensor("cpu_usage", SensorTypeGauge, "percent", "24h")

	// Post CPU usage event
	event := Event{
		Value: 45.5,
		TS:    time.Now().Unix(),
	}

	err = tree.PostEvent("cpu_usage", event)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Event posted")
	// Output:
	// Event posted
}

// ExampleTree_GetSensorData demonstrates querying sensor data
func ExampleTree_GetSensorData() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/example_query.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	tree.InitTimeseries()
	tree.CreateSensor("temperature", SensorTypeGauge, "celsius", "24h")

	// Post some temperature readings
	now := time.Now().Unix()
	for i := 0; i < 10; i++ {
		event := Event{
			Value: 20.0 + float64(i),
			TS:    now - int64((10-i)*60), // Past 10 minutes
		}
		tree.PostEvent("temperature", event)
	}

	// Query last 5 minutes with 1 minute intervals
	query := SensorQuery{
		WindowType: TumblingWindow,
		WindowSize: "5m",
		Scale:      "1m",
		OutputType: OutputTypeGraph,
	}

	result, err := tree.GetSensorData("temperature", query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sensor: %s, Points: %d\n", result.SensorName, len(result.Points))
	// Output:
	// Sensor: temperature, Points: 5
}

// Example_serverMonitoring demonstrates monitoring a server
func Example_serverMonitoring() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/server_monitoring.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	tree.InitTimeseries()

	// Create sensors for server monitoring
	sensors := map[string]struct {
		sensorType string
		unit       string
	}{
		"server1_cpu_usage":    {SensorTypeGauge, "percent"},
		"server1_memory_used":  {SensorTypeGauge, "bytes"},
		"server1_disk_used":    {SensorTypeGauge, "bytes"},
		"server1_load_average": {SensorTypeGauge, "float"},
	}

	for name, config := range sensors {
		tree.CreateSensor(name, config.sensorType, config.unit, "24h")
	}

	// Post metrics
	now := time.Now().Unix()
	tree.PostEvent("server1_cpu_usage", Event{Value: 45.5, TS: now})
	tree.PostEvent("server1_memory_used", Event{Value: 8589934592, TS: now}) // 8GB
	tree.PostEvent("server1_disk_used", Event{Value: 107374182400, TS: now}) // 100GB
	tree.PostEvent("server1_load_average", Event{Value: 2.5, TS: now})

	list, _ := tree.ListSensors()
	fmt.Printf("Monitoring %d metrics\n", len(list))

	// Output:
	// Monitoring 4 metrics
}

// Example_iotSensors demonstrates IoT temperature monitoring
func Example_iotSensors() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/iot_sensors.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	tree.InitTimeseries()

	// Create sensors for 3 IoT devices
	for i := 1; i <= 3; i++ {
		sensorName := fmt.Sprintf("iot_device_%d_temp", i)
		tree.CreateSensor(sensorName, SensorTypeGauge, "celsius", "24h")
	}

	// Post temperature readings
	now := time.Now().Unix()
	tree.PostEvent("iot_device_1_temp", Event{Value: 22.5, TS: now})
	tree.PostEvent("iot_device_2_temp", Event{Value: 23.1, TS: now})
	tree.PostEvent("iot_device_3_temp", Event{Value: 21.8, TS: now})

	// Get latest temperature from device 1
	latest, _ := tree.GetLatestValue("iot_device_1_temp")
	fmt.Printf("Device 1 temp: %.1f°C\n", latest.Value)

	// Output:
	// Device 1 temp: 22.5°C
}

// ExampleTree_DeleteOldEvents demonstrates cleanup of old data
func ExampleTree_DeleteOldEvents() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/cleanup_example.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	tree.InitTimeseries()
	tree.CreateSensor("disk_io", SensorTypeCounter, "bytes", "1h")

	now := time.Now().Unix()

	// Add old events (2 hours ago)
	for i := 0; i < 5; i++ {
		event := Event{
			Value: float64(i * 1000),
			TS:    now - 7200,
		}
		tree.PostEvent("disk_io", event)
	}

	// Add recent events (30 minutes ago)
	for i := 0; i < 3; i++ {
		event := Event{
			Value: float64(i * 2000),
			TS:    now - 1800,
		}
		tree.PostEvent("disk_io", event)
	}

	// Delete events older than retention (1h)
	deleted, _ := tree.DeleteOldEvents("disk_io")
	fmt.Printf("Deleted %d old events\n", deleted)

	// Output:
	// Deleted 5 old events
}

// Example_multiServerWorkflow demonstrates complete workflow for multiple servers
func Example_multiServerWorkflow() {
	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: "/tmp/multi_server.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tree.Close()

	tree.InitTimeseries()

	// Setup sensors for 3 servers
	servers := []string{"web1", "web2", "db1"}
	metrics := []string{"cpu", "memory", "disk"}

	for _, server := range servers {
		for _, metric := range metrics {
			sensorName := fmt.Sprintf("%s_%s", server, metric)
			tree.CreateSensor(sensorName, SensorTypeGauge, "percent", "24h")
		}
	}

	// Simulate metric collection
	now := time.Now().Unix()
	for _, server := range servers {
		tree.PostEvent(fmt.Sprintf("%s_cpu", server), Event{Value: 45.0, TS: now})
		tree.PostEvent(fmt.Sprintf("%s_memory", server), Event{Value: 60.0, TS: now})
		tree.PostEvent(fmt.Sprintf("%s_disk", server), Event{Value: 70.0, TS: now})
	}

	allSensors, _ := tree.ListSensors()
	fmt.Printf("Total sensors: %d\n", len(allSensors))

	// Output:
	// Total sensors: 9
}
