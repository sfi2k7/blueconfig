package blueconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

// ============================================================================
// HTTP API Tests for Timeseries
// ============================================================================

func setupTestHTTPServer(t *testing.T) (*Tree, string, string) {
	tmpfile := fmt.Sprintf("/tmp/test_http_ts_%d.db", time.Now().UnixNano())
	port := 9000 + rand.Intn(1000) // Random port between 9000-9999
	tr, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: tmpfile,
		Port:                  port,
		Token:                 "test-token",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Initialize timeseries
	err = tr.InitTimeseries()
	if err != nil {
		t.Fatal(err)
	}

	// Start server in background
	go tr.Serve()
	time.Sleep(200 * time.Millisecond) // Wait for server to start

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	return tr, tmpfile, baseURL
}

func makeRequest(t *testing.T, method, url string, body interface{}, token string) *http.Response {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if token != "" {
		q := req.URL.Query()
		q.Add("token", token)
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

func parseResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	defer resp.Body.Close()

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return result
}

// ============================================================================
// Init and Sensor Management Tests
// ============================================================================

func TestHTTP_InitTimeseries(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	url := baseURL + "/timeseries/init?token=test-token"
	resp := makeRequest(t, "POST", url, nil, "")

	result := parseResponse(t, resp)
	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}
	if result["result"] != true {
		t.Error("Expected result to be true")
	}
}

func TestHTTP_CreateSensor(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	url := baseURL + "/timeseries/sensors/create"
	body := map[string]string{
		"name":        "cpu_usage",
		"sensor_type": "gauge",
		"unit":        "percent",
		"retention":   "24h",
	}

	resp := makeRequest(t, "POST", url, body, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}
	if result["result"] != true {
		t.Error("Expected result to be true")
	}

	// Verify sensor was created
	sensors, _ := tr.ListSensors()
	if len(sensors) != 1 || sensors[0] != "cpu_usage" {
		t.Errorf("Sensor not created properly, got %v", sensors)
	}
}

func TestHTTP_ListSensors(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	tr.CreateSensor("sensor1", SensorTypeGauge, "unit", "24h")
	tr.CreateSensor("sensor2", SensorTypeCounter, "unit", "24h")

	url := baseURL + "/timeseries/sensors"
	resp := makeRequest(t, "GET", url, nil, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	sensors := result["result"].([]interface{})
	if len(sensors) != 2 {
		t.Errorf("Expected 2 sensors, got %d", len(sensors))
	}
}

func TestHTTP_GetSensorInfo(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	tr.CreateSensor("test_sensor", SensorTypeGauge, "celsius", "24h")

	url := baseURL + "/timeseries/sensors/test_sensor/info"
	resp := makeRequest(t, "GET", url, nil, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	info := result["result"].(map[string]interface{})
	if info["__sensor_type"] != "gauge" {
		t.Errorf("Expected gauge, got %v", info["__sensor_type"])
	}
	if info["__unit"] != "celsius" {
		t.Errorf("Expected celsius, got %v", info["__unit"])
	}
}

func TestHTTP_DeleteSensor(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	tr.CreateSensor("temp_sensor", SensorTypeGauge, "unit", "1h")

	url := baseURL + "/timeseries/sensors/temp_sensor"
	resp := makeRequest(t, "DELETE", url, nil, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	// Verify sensor is deleted
	sensors, _ := tr.ListSensors()
	if len(sensors) != 0 {
		t.Error("Sensor was not deleted")
	}
}

// ============================================================================
// Event Management Tests
// ============================================================================

func TestHTTP_PostEvent(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	tr.CreateSensor("cpu_usage", SensorTypeGauge, "percent", "24h")

	url := baseURL + "/timeseries/sensors/cpu_usage/event"
	body := Event{
		ID:    "event1",
		Value: 45.5,
		TS:    time.Now().Unix(),
	}

	resp := makeRequest(t, "POST", url, body, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	// Verify event was posted
	events, _ := tr.GetAllEvents("cpu_usage")
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].Value != 45.5 {
		t.Errorf("Expected value 45.5, got %f", events[0].Value)
	}
}

func TestHTTP_GetAllEvents(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	tr.CreateSensor("temperature", SensorTypeGauge, "celsius", "24h")

	// Post some events
	now := time.Now().Unix()
	for i := 0; i < 5; i++ {
		tr.PostEvent("temperature", Event{
			ID:    fmt.Sprintf("event_%d", i),
			Value: 20.0 + float64(i),
			TS:    now + int64(i*60),
		})
	}

	url := baseURL + "/timeseries/sensors/temperature/events"
	resp := makeRequest(t, "GET", url, nil, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	events := result["result"].([]interface{})
	if len(events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(events))
	}
}

func TestHTTP_GetLatestValue(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	tr.CreateSensor("memory", SensorTypeGauge, "bytes", "24h")

	now := time.Now().Unix()
	tr.PostEvent("memory", Event{ID: "e1", Value: 100, TS: now - 300})
	tr.PostEvent("memory", Event{ID: "e2", Value: 200, TS: now - 200})
	tr.PostEvent("memory", Event{ID: "e3", Value: 300, TS: now}) // Latest

	url := baseURL + "/timeseries/sensors/memory/latest"
	resp := makeRequest(t, "GET", url, nil, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	latest := result["result"].(map[string]interface{})
	if latest["id"] != "e3" {
		t.Errorf("Expected latest event ID e3, got %v", latest["id"])
	}
	if latest["value"] != float64(300) {
		t.Errorf("Expected value 300, got %v", latest["value"])
	}
}

func TestHTTP_DeleteOldEvents(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	tr.CreateSensor("disk_io", SensorTypeCounter, "bytes", "1h")

	now := time.Now().Unix()
	// Add old events
	tr.PostEvent("disk_io", Event{ID: "old1", Value: 100, TS: now - 7200})
	tr.PostEvent("disk_io", Event{ID: "old2", Value: 200, TS: now - 7200})
	// Add recent events
	tr.PostEvent("disk_io", Event{ID: "recent1", Value: 300, TS: now - 1800})

	url := baseURL + "/timeseries/sensors/disk_io/cleanup"
	resp := makeRequest(t, "POST", url, nil, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	deletedMap := result["result"].(map[string]interface{})
	deleted := int(deletedMap["deleted"].(float64))
	if deleted != 2 {
		t.Errorf("Expected to delete 2 events, got %d", deleted)
	}
}

// ============================================================================
// Query Tests
// ============================================================================

func TestHTTP_GetSensorData(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	tr.CreateSensor("cpu", SensorTypeGauge, "percent", "24h")

	// Add events over 10 minutes
	now := time.Now().Unix()
	baseTime := now - 600
	for i := 0; i < 30; i++ {
		tr.PostEvent("cpu", Event{
			Value: 40.0 + float64(i%20),
			TS:    baseTime + int64(i*20),
		})
	}

	url := baseURL + "/timeseries/sensors/cpu/data?window_size=5m&scale=1m"
	resp := makeRequest(t, "GET", url, nil, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	data := result["result"].(map[string]interface{})
	if data["SensorName"] != "cpu" {
		t.Errorf("Expected sensor name cpu, got %v", data["SensorName"])
	}

	points := data["Points"].([]interface{})
	expectedPoints := 5 // 5m / 1m = 5 points
	if len(points) != expectedPoints {
		t.Errorf("Expected %d points, got %d", expectedPoints, len(points))
	}
}

func TestHTTP_GetSensorDataWithAllParams(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	tr.CreateSensor("network", SensorTypeCounter, "bytes", "24h")

	now := time.Now().Unix()
	baseTime := now - 3600
	for i := 0; i < 60; i++ {
		tr.PostEvent("network", Event{
			Value: float64(i * 1000),
			TS:    baseTime + int64(i*60),
		})
	}

	url := baseURL + "/timeseries/sensors/network/data?window_type=tumbling&window_size=1h&scale=10m&output_type=graph"
	resp := makeRequest(t, "GET", url, nil, "test-token")
	result := parseResponse(t, resp)

	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	data := result["result"].(map[string]interface{})
	points := data["Points"].([]interface{})
	expectedPoints := 6 // 1h / 10m = 6 points
	if len(points) != expectedPoints {
		t.Errorf("Expected %d points, got %d", expectedPoints, len(points))
	}
}

// ============================================================================
// Authentication Tests
// ============================================================================

func TestHTTP_AuthenticationRequired(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Try without token
	url := baseURL + "/timeseries/sensors"
	resp := makeRequest(t, "GET", url, nil, "")

	result := parseResponse(t, resp)
	if result["error"] == nil {
		t.Error("Expected error for missing token")
	}
}

func TestHTTP_InvalidToken(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	url := baseURL + "/timeseries/sensors"
	resp := makeRequest(t, "GET", url, nil, "wrong-token")

	result := parseResponse(t, resp)
	if result["error"] == nil {
		t.Error("Expected error for invalid token")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestHTTP_FullWorkflow(t *testing.T) {
	tr, tmpfile, baseURL := setupTestHTTPServer(t)
	defer cleanupTimeseriesTree(tr, tmpfile)

	token := "test-token"

	// 1. Initialize timeseries
	initURL := baseURL + "/timeseries/init"
	resp := makeRequest(t, "POST", initURL, nil, token)
	result := parseResponse(t, resp)
	if result["error"] != nil {
		t.Fatalf("Init failed: %v", result["error"])
	}

	// 2. Create sensor
	createURL := baseURL + "/timeseries/sensors/create"
	createBody := map[string]string{
		"name":        "server1_cpu",
		"sensor_type": "gauge",
		"unit":        "percent",
		"retention":   "24h",
	}
	resp = makeRequest(t, "POST", createURL, createBody, token)
	result = parseResponse(t, resp)
	if result["error"] != nil {
		t.Fatalf("Create sensor failed: %v", result["error"])
	}

	// 3. Post events
	now := time.Now().Unix()
	for i := 0; i < 10; i++ {
		eventURL := baseURL + "/timeseries/sensors/server1_cpu/event"
		event := Event{
			Value: 40.0 + float64(i),
			TS:    now - int64((10-i)*60),
		}
		resp = makeRequest(t, "POST", eventURL, event, token)
		result = parseResponse(t, resp)
		if result["error"] != nil {
			t.Fatalf("Post event failed: %v", result["error"])
		}
	}

	// 4. Query data
	queryURL := baseURL + "/timeseries/sensors/server1_cpu/data?window_size=5m&scale=1m"
	resp = makeRequest(t, "GET", queryURL, nil, token)
	result = parseResponse(t, resp)
	if result["error"] != nil {
		t.Fatalf("Query failed: %v", result["error"])
	}

	data := result["result"].(map[string]interface{})
	points := data["Points"].([]interface{})
	if len(points) != 5 {
		t.Errorf("Expected 5 points, got %d", len(points))
	}

	// 5. Get sensor info
	infoURL := baseURL + "/timeseries/sensors/server1_cpu/info"
	resp = makeRequest(t, "GET", infoURL, nil, token)
	result = parseResponse(t, resp)
	if result["error"] != nil {
		t.Fatalf("Get info failed: %v", result["error"])
	}

	// 6. List sensors
	listURL := baseURL + "/timeseries/sensors"
	resp = makeRequest(t, "GET", listURL, nil, token)
	result = parseResponse(t, resp)
	if result["error"] != nil {
		t.Fatalf("List sensors failed: %v", result["error"])
	}

	sensors := result["result"].([]interface{})
	if len(sensors) != 1 {
		t.Errorf("Expected 1 sensor, got %d", len(sensors))
	}
}
