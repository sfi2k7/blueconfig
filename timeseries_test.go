package blueconfig

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// ============================================================================
// Helper Functions
// ============================================================================

func createTimeseriesTree(t *testing.T) (*Tree, string) {
	tmpfile := fmt.Sprintf("/tmp/test_timeseries_%d.db", time.Now().UnixNano())
	tr, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: tmpfile,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Initialize timeseries
	err = tr.InitTimeseries()
	if err != nil {
		t.Fatal(err)
	}

	return tr, tmpfile
}

func cleanupTimeseriesTree(tr *Tree, tmpfile string) {
	if tr != nil {
		tr.Close()
	}
	os.Remove(tmpfile)
}

// ============================================================================
// Initialization Tests
// ============================================================================

func TestInitTimeseries(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Check base path exists
	tsType, err := tr.GetValue(TimeseriesBasePath + "/__type")
	if err != nil {
		t.Fatalf("Failed to get timeseries type: %v", err)
	}
	if tsType != "timeseries" {
		t.Errorf("Expected timeseries type, got %s", tsType)
	}

	// Check Sensors path metadata
	sensorsType, err := tr.GetValue(SensorsPath + "/__type")
	if err != nil {
		t.Fatalf("Failed to get sensors type: %v", err)
	}
	if sensorsType != "timeseries" {
		t.Errorf("Expected timeseries type on Sensors, got %s", sensorsType)
	}

	// Check metadata exists
	metadata, err := tr.GetAllPropsWithValues(SensorsPath)
	if err != nil {
		t.Fatalf("Failed to get sensors metadata: %v", err)
	}

	expectedKeys := []string{"__type", "__lastupdated", "__lastcleaned", "__sensor_count"}
	for _, key := range expectedKeys {
		if _, exists := metadata[key]; !exists {
			t.Errorf("Missing metadata key: %s", key)
		}
	}
}

// ============================================================================
// Sensor Management Tests
// ============================================================================

func TestCreateSensor(t *testing.T) {
	tests := []struct {
		name       string
		sensorName string
		sensorType string
		unit       string
		retention  string
		expectErr  bool
	}{
		{"valid_gauge", "cpu_usage", SensorTypeGauge, "percent", "24h", false},
		{"valid_counter", "disk_writes", SensorTypeCounter, "bytes", "7d", false},
		{"invalid_type", "invalid", "unknown", "unit", "1h", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, tmpfile := createTimeseriesTree(t)
			defer cleanupTimeseriesTree(tr, tmpfile)

			err := tr.CreateSensor(tt.sensorName, tt.sensorType, tt.unit, tt.retention)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectErr {
				// Verify sensor was created
				sensorPath := SensorsPath + "/" + tt.sensorName
				sType, err := tr.GetValue(sensorPath + "/__type")
				if err != nil {
					t.Errorf("Failed to get sensor type: %v", err)
				}
				if sType != "sensor" {
					t.Errorf("Expected sensor type, got %s", sType)
				}

				// Verify sensor properties
				props, err := tr.GetAllPropsWithValues(sensorPath)
				if err != nil {
					t.Errorf("Failed to get sensor props: %v", err)
				}

				if props["__sensor_type"] != tt.sensorType {
					t.Errorf("Expected sensor_type %s, got %s", tt.sensorType, props["__sensor_type"])
				}
				if props["__unit"] != tt.unit {
					t.Errorf("Expected unit %s, got %s", tt.unit, props["__unit"])
				}
				if props["__retention"] != tt.retention {
					t.Errorf("Expected retention %s, got %s", tt.retention, props["__retention"])
				}
			}
		})
	}
}

func TestCreateSensorWithoutTimeseriesParent(t *testing.T) {
	tmpfile := fmt.Sprintf("/tmp/test_no_ts_%d.db", time.Now().UnixNano())
	tr, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: tmpfile,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Don't initialize timeseries
	err = tr.CreateSensor("test_sensor", SensorTypeGauge, "unit", "1h")
	if err == nil {
		t.Error("Expected error when creating sensor without timeseries parent")
	}
}

func TestListSensors(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Create multiple sensors
	sensors := []string{"cpu_usage", "memory_used", "disk_io"}
	for _, s := range sensors {
		err := tr.CreateSensor(s, SensorTypeGauge, "percent", "24h")
		if err != nil {
			t.Fatalf("Failed to create sensor %s: %v", s, err)
		}
	}

	// List sensors
	list, err := tr.ListSensors()
	if err != nil {
		t.Fatalf("Failed to list sensors: %v", err)
	}

	if len(list) != len(sensors) {
		t.Errorf("Expected %d sensors, got %d", len(sensors), len(list))
	}

	// Verify all sensors are in list
	sensorMap := make(map[string]bool)
	for _, s := range list {
		sensorMap[s] = true
	}
	for _, s := range sensors {
		if !sensorMap[s] {
			t.Errorf("Sensor %s not found in list", s)
		}
	}
}

func TestGetSensorInfo(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "test_sensor"
	err := tr.CreateSensor(sensorName, SensorTypeGauge, "celsius", "24h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	info, err := tr.GetSensorInfo(sensorName)
	if err != nil {
		t.Fatalf("Failed to get sensor info: %v", err)
	}

	if info["__sensor_type"] != SensorTypeGauge {
		t.Errorf("Expected gauge, got %s", info["__sensor_type"])
	}
	if info["__unit"] != "celsius" {
		t.Errorf("Expected celsius, got %s", info["__unit"])
	}
}

func TestDeleteSensor(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "temp_sensor"
	err := tr.CreateSensor(sensorName, SensorTypeGauge, "celsius", "1h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	// Add some events
	for i := 0; i < 5; i++ {
		event := Event{
			Value: 20.0 + float64(i),
			TS:    time.Now().Unix(),
		}
		tr.PostEvent(sensorName, event)
	}

	// Delete sensor
	err = tr.DeleteSensor(sensorName)
	if err != nil {
		t.Fatalf("Failed to delete sensor: %v", err)
	}

	// Verify sensor is gone
	_, err = tr.GetSensorInfo(sensorName)
	if err == nil {
		t.Error("Sensor still exists after deletion")
	}
}

// ============================================================================
// Event Management Tests
// ============================================================================

func TestPostEvent(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "cpu_usage"
	err := tr.CreateSensor(sensorName, SensorTypeGauge, "percent", "24h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	// Post event
	event := Event{
		ID:    "event1",
		Value: 45.5,
		TS:    time.Now().Unix(),
	}

	err = tr.PostEvent(sensorName, event)
	if err != nil {
		t.Fatalf("Failed to post event: %v", err)
	}

	// Verify event was created
	eventPath := SensorsPath + "/" + sensorName + "/" + event.ID
	value, err := tr.GetValue(eventPath + "/value")
	if err != nil {
		t.Fatalf("Failed to get event value: %v", err)
	}
	if value != "45.5" {
		t.Errorf("Expected value 45.5, got %s", value)
	}

	// Verify sensor metadata updated
	sensorPath := SensorsPath + "/" + sensorName
	eventCount, err := tr.GetValue(sensorPath + "/__event_count")
	if err != nil {
		t.Fatalf("Failed to get event count: %v", err)
	}
	if eventCount != "1" {
		t.Errorf("Expected event count 1, got %s", eventCount)
	}
}

func TestPostEventWithAutoID(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "memory"
	err := tr.CreateSensor(sensorName, SensorTypeGauge, "bytes", "1h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	// Post event without ID
	event := Event{
		Value: 1024000,
		TS:    time.Now().Unix(),
	}

	err = tr.PostEvent(sensorName, event)
	if err != nil {
		t.Fatalf("Failed to post event: %v", err)
	}

	// Verify event was created
	events, err := tr.GetAllEvents(sensorName)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

func TestPostEventToNonExistentSensor(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	event := Event{
		Value: 100,
		TS:    time.Now().Unix(),
	}

	err := tr.PostEvent("nonexistent", event)
	if err == nil {
		t.Error("Expected error when posting to non-existent sensor")
	}
}

func TestGetAllEvents(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "temperature"
	err := tr.CreateSensor(sensorName, SensorTypeGauge, "celsius", "24h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	// Post multiple events
	expectedEvents := 10
	baseTime := time.Now().Unix()
	for i := 0; i < expectedEvents; i++ {
		event := Event{
			ID:    fmt.Sprintf("event_%d", i),
			Value: 20.0 + float64(i),
			TS:    baseTime + int64(i*60),
		}
		err = tr.PostEvent(sensorName, event)
		if err != nil {
			t.Fatalf("Failed to post event %d: %v", i, err)
		}
	}

	// Get all events
	events, err := tr.GetAllEvents(sensorName)
	if err != nil {
		t.Fatalf("Failed to get all events: %v", err)
	}

	if len(events) != expectedEvents {
		t.Errorf("Expected %d events, got %d", expectedEvents, len(events))
	}

	// Verify events have correct structure
	for _, event := range events {
		if event.ID == "" {
			t.Error("Event ID is empty")
		}
		if event.Value == 0 {
			t.Error("Event value is zero")
		}
		if event.TS == 0 {
			t.Error("Event timestamp is zero")
		}
	}
}

func TestDeleteOldEvents(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "disk_io"
	err := tr.CreateSensor(sensorName, SensorTypeCounter, "bytes", "1h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	now := time.Now().Unix()
	twoHoursAgo := now - 7200
	fortyFiveMinAgo := now - 2700
	thirtyMinAgo := now - 1800

	// Post old events (outside retention)
	oldEvents := []Event{
		{ID: "old1", Value: 100, TS: twoHoursAgo},
		{ID: "old2", Value: 200, TS: twoHoursAgo + 60},
	}

	// Post recent events (within retention)
	recentEvents := []Event{
		{ID: "recent1", Value: 300, TS: fortyFiveMinAgo},
		{ID: "recent2", Value: 400, TS: thirtyMinAgo},
		{ID: "recent3", Value: 500, TS: now},
	}

	for _, event := range append(oldEvents, recentEvents...) {
		tr.PostEvent(sensorName, event)
	}

	// Delete old events
	deleted, err := tr.DeleteOldEvents(sensorName)
	if err != nil {
		t.Fatalf("Failed to delete old events: %v", err)
	}

	if deleted != len(oldEvents) {
		t.Errorf("Expected to delete %d events, deleted %d", len(oldEvents), deleted)
	}

	// Verify only recent events remain
	events, err := tr.GetAllEvents(sensorName)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) != len(recentEvents) {
		t.Errorf("Expected %d remaining events, got %d", len(recentEvents), len(events))
	}
}

// ============================================================================
// Query and Analysis Tests
// ============================================================================

func TestGetSensorData(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "cpu_usage"
	err := tr.CreateSensor(sensorName, SensorTypeGauge, "percent", "24h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	// Create events over 5 minutes
	now := time.Now().Unix()
	baseTime := now - 300 // 5 minutes ago

	for i := 0; i < 10; i++ {
		event := Event{
			ID:    fmt.Sprintf("event_%d", i),
			Value: 40.0 + float64(i),
			TS:    baseTime + int64(i*30), // Every 30 seconds
		}
		tr.PostEvent(sensorName, event)
	}

	// Query with 5m window and 1m scale
	query := SensorQuery{
		WindowType: TumblingWindow,
		WindowSize: "5m",
		Scale:      "1m",
		OutputType: OutputTypeGraph,
	}

	result, err := tr.GetSensorData(sensorName, query)
	if err != nil {
		t.Fatalf("Failed to get sensor data: %v", err)
	}

	if result.SensorName != sensorName {
		t.Errorf("Expected sensor name %s, got %s", sensorName, result.SensorName)
	}

	if result.SensorType != SensorTypeGauge {
		t.Errorf("Expected sensor type %s, got %s", SensorTypeGauge, result.SensorType)
	}

	expectedPoints := 5 // 5m / 1m = 5 points
	if len(result.Points) != expectedPoints {
		t.Errorf("Expected %d points, got %d", expectedPoints, len(result.Points))
	}

	// Verify data points have correct structure
	for i, point := range result.Points {
		if point.Timestamp == 0 {
			t.Errorf("Point %d has zero timestamp", i)
		}
		// Points with data should have count > 0
		if point.Count > 0 {
			if point.Min == 0 && point.Max == 0 && point.Avg == 0 {
				t.Errorf("Point %d has data but all aggregates are zero", i)
			}
		}
	}
}

func TestGetSensorDataWithDifferentScales(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "network_traffic"
	err := tr.CreateSensor(sensorName, SensorTypeCounter, "bytes", "1h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	// Create events over 10 minutes
	now := time.Now().Unix()
	baseTime := now - 600 // 10 minutes ago

	for i := 0; i < 60; i++ {
		event := Event{
			ID:    fmt.Sprintf("event_%d", i),
			Value: float64(i * 100),
			TS:    baseTime + int64(i*10), // Every 10 seconds
		}
		tr.PostEvent(sensorName, event)
	}

	tests := []struct {
		name           string
		windowSize     string
		scale          string
		expectedPoints int
	}{
		{"10m_30s", "10m", "30s", 20},
		{"10m_1m", "10m", "1m", 10},
		{"10m_2m", "10m", "2m", 5},
		{"5m_30s", "5m", "30s", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := SensorQuery{
				WindowType: TumblingWindow,
				WindowSize: tt.windowSize,
				Scale:      tt.scale,
				OutputType: OutputTypeGraph,
			}

			result, err := tr.GetSensorData(sensorName, query)
			if err != nil {
				t.Fatalf("Failed to get sensor data: %v", err)
			}

			if len(result.Points) != tt.expectedPoints {
				t.Errorf("Expected %d points, got %d", tt.expectedPoints, len(result.Points))
			}
		})
	}
}

func TestGetLatestValue(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "temperature"
	err := tr.CreateSensor(sensorName, SensorTypeGauge, "celsius", "1h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	// Post events with different timestamps
	now := time.Now().Unix()
	events := []Event{
		{ID: "event1", Value: 20.0, TS: now - 300},
		{ID: "event2", Value: 22.0, TS: now - 200},
		{ID: "event3", Value: 25.0, TS: now - 100},
		{ID: "event4", Value: 23.0, TS: now}, // Latest
	}

	for _, event := range events {
		tr.PostEvent(sensorName, event)
	}

	// Get latest value
	latest, err := tr.GetLatestValue(sensorName)
	if err != nil {
		t.Fatalf("Failed to get latest value: %v", err)
	}

	if latest.ID != "event4" {
		t.Errorf("Expected latest event ID event4, got %s", latest.ID)
	}
	if latest.Value != 23.0 {
		t.Errorf("Expected latest value 23.0, got %f", latest.Value)
	}
}

func TestGetLatestValueNoEvents(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "empty_sensor"
	err := tr.CreateSensor(sensorName, SensorTypeGauge, "unit", "1h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	_, err = tr.GetLatestValue(sensorName)
	if err == nil {
		t.Error("Expected error when getting latest value with no events")
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"30s", 30, false},
		{"5m", 300, false},
		{"1h", 3600, false},
		{"24h", 86400, false},
		{"7d", 604800, false},
		{"invalid", 0, true},
		{"10x", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDuration(tt.input)

			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.hasError && result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestFullWorkflow(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Create sensor
	sensorName := "server1_cpu"
	err := tr.CreateSensor(sensorName, SensorTypeGauge, "percent", "24h")
	if err != nil {
		t.Fatalf("Failed to create sensor: %v", err)
	}

	// Post events simulating 10 minutes of CPU usage
	now := time.Now().Unix()
	baseTime := now - 600

	for i := 0; i < 60; i++ {
		event := Event{
			Value: 30.0 + float64(i%30), // Oscillating between 30-60%
			TS:    baseTime + int64(i*10),
		}
		err = tr.PostEvent(sensorName, event)
		if err != nil {
			t.Fatalf("Failed to post event: %v", err)
		}
	}

	// Query data
	query := SensorQuery{
		WindowType: TumblingWindow,
		WindowSize: "10m",
		Scale:      "1m",
		OutputType: OutputTypeGraph,
	}

	result, err := tr.GetSensorData(sensorName, query)
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}

	if len(result.Points) != 10 {
		t.Errorf("Expected 10 points, got %d", len(result.Points))
	}

	// Get latest value
	latest, err := tr.GetLatestValue(sensorName)
	if err != nil {
		t.Fatalf("Failed to get latest value: %v", err)
	}

	if latest == nil {
		t.Fatal("Latest value is nil")
	}

	// List sensors
	sensors, err := tr.ListSensors()
	if err != nil {
		t.Fatalf("Failed to list sensors: %v", err)
	}

	if len(sensors) != 1 || sensors[0] != sensorName {
		t.Errorf("Expected sensor list to contain %s", sensorName)
	}
}

func TestMultipleSensorsWorkflow(t *testing.T) {
	tr, tmpfile := createTimeseriesTree(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Create multiple sensors for a server
	sensors := map[string]string{
		"cpu_usage":   SensorTypeGauge,
		"memory_used": SensorTypeGauge,
		"disk_io":     SensorTypeCounter,
		"network_rx":  SensorTypeCounter,
	}

	for name, stype := range sensors {
		err := tr.CreateSensor(name, stype, "unit", "24h")
		if err != nil {
			t.Fatalf("Failed to create sensor %s: %v", name, err)
		}
	}

	// Post events to each sensor
	now := time.Now().Unix()
	for name := range sensors {
		for i := 0; i < 10; i++ {
			event := Event{
				Value: float64(i * 10),
				TS:    now - int64(i*60),
			}
			tr.PostEvent(name, event)
		}
	}

	// Verify all sensors have events
	list, _ := tr.ListSensors()
	if len(list) != len(sensors) {
		t.Errorf("Expected %d sensors, got %d", len(sensors), len(list))
	}

	for name := range sensors {
		events, err := tr.GetAllEvents(name)
		if err != nil {
			t.Errorf("Failed to get events for %s: %v", name, err)
		}
		if len(events) != 10 {
			t.Errorf("Expected 10 events for %s, got %d", name, len(events))
		}
	}
}
