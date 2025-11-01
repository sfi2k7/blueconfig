package blueconfig

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// Constants
// ============================================================================

const (
	TimeseriesBasePath = "root/timeseries"
	SensorsPath        = "root/timeseries/Sensors"
)

const (
	SensorTypeGauge   = "gauge"
	SensorTypeCounter = "counter"
)

const (
	OutputTypeMin   = "min"
	OutputTypeMax   = "max"
	OutputTypeAvg   = "avg"
	OutputTypeSum   = "sum"
	OutputTypeCount = "count"
	OutputTypeGraph = "graph"
)

// ============================================================================
// Types
// ============================================================================

type Event struct {
	ID    string  `json:"id"`
	Value float64 `json:"value"`
	TS    int64   `json:"ts"`
}

type WindowType string

const (
	TumblingWindow WindowType = "tumbling"
	SlidingWindow  WindowType = "sliding"
)

type SensorQuery struct {
	WindowType WindowType
	WindowSize string // "5m", "1h", "24h"
	Scale      string // "30s", "1m", "15m"
	OutputType string // "min", "max", "avg", "sum", "count", "graph"
}

type DataPoint struct {
	Timestamp int64
	Value     float64
	Min       float64
	Max       float64
	Avg       float64
	Sum       float64
	Count     int
}

type SensorResult struct {
	SensorName string
	SensorType string
	Points     []DataPoint
}

// ============================================================================
// Initialization
// ============================================================================

// InitTimeseries initializes the timeseries structure with metadata
func (t *Tree) InitTimeseries() error {
	// Create base path
	err := t.CreatePath(TimeseriesBasePath)
	if err != nil {
		return err
	}

	// Set timeseries type marker
	err = t.SetValue(TimeseriesBasePath+"/__type", "timeseries")
	if err != nil {
		return err
	}

	// Create Sensors path with metadata
	err = t.CreatePath(SensorsPath)
	if err != nil {
		return err
	}

	// Set Sensors metadata
	now := strconv.FormatInt(time.Now().Unix(), 10)
	metadata := map[string]interface{}{
		"__type":         "timeseries",
		"__lastupdated":  now,
		"__lastcleaned":  now,
		"__sensor_count": "0",
	}

	return t.SetValues(SensorsPath, metadata)
}

// ============================================================================
// Sensor Management
// ============================================================================

// CreateSensor creates a timeseries sensor node
func (t *Tree) CreateSensor(sensorName string, sensorType string, unit string, retention string) error {
	// Validate sensor type
	if sensorType != SensorTypeGauge && sensorType != SensorTypeCounter {
		return errors.New("invalid sensor type: must be 'gauge' or 'counter'")
	}

	// Check if parent has timeseries type
	parentType, err := t.GetValue(SensorsPath + "/__type")
	if err != nil || parentType != "timeseries" {
		return errors.New("parent path is not a timeseries endpoint")
	}

	// Create sensor path
	sensorPath := SensorsPath + "/" + sensorName

	// Set sensor properties
	now := strconv.FormatInt(time.Now().Unix(), 10)
	props := map[string]interface{}{
		"__type":        "sensor",
		"__sensor_type": sensorType,
		"__unit":        unit,
		"__retention":   retention,
		"__created":     now,
		"__lastupdated": now,
		"__event_count": "0",
	}

	err = t.CreateNodeWithProps(sensorPath, props)
	if err != nil {
		return err
	}

	// Update Sensors metadata
	t.updateSensorsMetadata()

	return nil
}

// PostEvent adds a new event to a sensor
func (t *Tree) PostEvent(sensorName string, event Event) error {
	sensorPath := SensorsPath + "/" + sensorName

	// Check if sensor exists and is valid
	sensorType, err := t.GetValue(sensorPath + "/__type")
	if err != nil {
		return errors.New("sensor does not exist")
	}
	if sensorType != "sensor" {
		return errors.New("path is not a sensor")
	}

	// Generate event ID if not provided
	eventID := event.ID
	if eventID == "" {
		eventID = fmt.Sprintf("%d_%d", event.TS, time.Now().UnixNano())
	}

	// Create event node
	eventPath := sensorPath + "/" + eventID
	eventProps := map[string]interface{}{
		"value": strconv.FormatFloat(event.Value, 'f', -1, 64),
		"ts":    strconv.FormatInt(event.TS, 10),
	}

	err = t.CreateNodeWithProps(eventPath, eventProps)
	if err != nil {
		return err
	}

	// Update sensor metadata
	now := strconv.FormatInt(time.Now().Unix(), 10)
	t.SetValue(sensorPath+"/__lastupdated", now)

	// Update event count
	eventCount, _ := t.GetValue(sensorPath + "/__event_count")
	count, _ := strconv.Atoi(eventCount)
	count++
	t.SetValue(sensorPath+"/__event_count", strconv.Itoa(count))

	// Update Sensors metadata
	t.SetValue(SensorsPath+"/__lastupdated", now)

	return nil
}

// ListSensors returns all sensors
func (t *Tree) ListSensors() ([]string, error) {
	sensors, err := t.GetNodesInPath(SensorsPath)
	if err != nil {
		return nil, err
	}
	return sensors, nil
}

// GetSensorInfo returns sensor metadata
func (t *Tree) GetSensorInfo(sensorName string) (map[string]string, error) {
	sensorPath := SensorsPath + "/" + sensorName
	return t.GetAllPropsWithValues(sensorPath)
}

// DeleteSensor deletes a sensor and all its events
func (t *Tree) DeleteSensor(sensorName string) error {
	sensorPath := SensorsPath + "/" + sensorName

	// Delete with force to remove all events
	err := t.DeleteNode(sensorPath, true)
	if err != nil {
		return err
	}

	// Update Sensors metadata
	t.updateSensorsMetadata()

	return nil
}

// ============================================================================
// Event Management
// ============================================================================

// GetAllEvents returns all events for a sensor
func (t *Tree) GetAllEvents(sensorName string) ([]Event, error) {
	sensorPath := SensorsPath + "/" + sensorName

	var events []Event
	err := t.ScanNodes(sensorPath, func(node NodeInfo) error {
		// Skip metadata properties (those starting with __)
		if strings.HasPrefix(node.Name, "__") {
			return nil
		}

		valueStr := node.Props["value"]
		tsStr := node.Props["ts"]

		if valueStr == "" || tsStr == "" {
			return nil
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil
		}

		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			return nil
		}

		events = append(events, Event{
			ID:    node.Name,
			Value: value,
			TS:    ts,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return events, nil
}

// DeleteOldEvents removes events older than retention period
func (t *Tree) DeleteOldEvents(sensorName string) (int, error) {
	sensorPath := SensorsPath + "/" + sensorName

	// Get retention period
	retention, err := t.GetValue(sensorPath + "/__retention")
	if err != nil {
		return 0, err
	}

	retentionSeconds, err := parseDuration(retention)
	if err != nil {
		return 0, err
	}

	cutoffTime := time.Now().Unix() - retentionSeconds

	// Get all events
	events, err := t.GetAllEvents(sensorName)
	if err != nil {
		return 0, err
	}

	deletedCount := 0
	for _, event := range events {
		if event.TS < cutoffTime {
			eventPath := sensorPath + "/" + event.ID
			err := t.DeleteNode(eventPath, false)
			if err == nil {
				deletedCount++
			}
		}
	}

	// Update metadata
	if deletedCount > 0 {
		now := strconv.FormatInt(time.Now().Unix(), 10)
		t.SetValue(sensorPath+"/__lastcleaned", now)
		t.SetValue(SensorsPath+"/__lastcleaned", now)

		// Update event count
		eventCount, _ := t.GetValue(sensorPath + "/__event_count")
		count, _ := strconv.Atoi(eventCount)
		count -= deletedCount
		if count < 0 {
			count = 0
		}
		t.SetValue(sensorPath+"/__event_count", strconv.Itoa(count))
	}

	return deletedCount, nil
}

// ============================================================================
// Query and Analysis
// ============================================================================

// GetSensorData queries sensor data with windowing
func (t *Tree) GetSensorData(sensorName string, query SensorQuery) (*SensorResult, error) {
	sensorPath := SensorsPath + "/" + sensorName

	// Get sensor type
	sensorType, err := t.GetValue(sensorPath + "/__sensor_type")
	if err != nil {
		return nil, errors.New("sensor not found")
	}

	// Parse window size and scale
	windowSeconds, err := parseDuration(query.WindowSize)
	if err != nil {
		return nil, fmt.Errorf("invalid window size: %v", err)
	}

	scaleSeconds, err := parseDuration(query.Scale)
	if err != nil {
		return nil, fmt.Errorf("invalid scale: %v", err)
	}

	// Calculate time range
	now := time.Now().Unix()
	startTime := now - windowSeconds

	// Get all events
	allEvents, err := t.GetAllEvents(sensorName)
	if err != nil {
		return nil, err
	}

	// Filter events within time range
	var events []Event
	for _, event := range allEvents {
		if event.TS >= startTime && event.TS <= now {
			events = append(events, event)
		}
	}

	// Group events into buckets
	buckets := make(map[int64][]float64)
	numBuckets := int(windowSeconds / scaleSeconds)

	for _, event := range events {
		bucketIndex := (event.TS - startTime) / scaleSeconds
		bucketTimestamp := startTime + (bucketIndex * scaleSeconds)
		buckets[bucketTimestamp] = append(buckets[bucketTimestamp], event.Value)
	}

	// Calculate data points
	var points []DataPoint
	for i := 0; i < numBuckets; i++ {
		bucketTimestamp := startTime + (int64(i) * scaleSeconds)
		values := buckets[bucketTimestamp]

		point := DataPoint{
			Timestamp: bucketTimestamp,
		}

		if len(values) > 0 {
			point.Count = len(values)
			point.Min = values[0]
			point.Max = values[0]
			point.Sum = 0

			for _, v := range values {
				if v < point.Min {
					point.Min = v
				}
				if v > point.Max {
					point.Max = v
				}
				point.Sum += v
			}

			point.Avg = point.Sum / float64(point.Count)
			point.Value = point.Avg // Default to average
		}

		points = append(points, point)
	}

	// Format result based on output type
	result := &SensorResult{
		SensorName: sensorName,
		SensorType: sensorType,
		Points:     points,
	}

	return result, nil
}

// GetLatestValue returns the most recent value for a sensor
func (t *Tree) GetLatestValue(sensorName string) (*Event, error) {
	events, err := t.GetAllEvents(sensorName)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, errors.New("no events found")
	}

	// Find latest event
	latest := events[0]
	for _, event := range events {
		if event.TS > latest.TS {
			latest = event
		}
	}

	return &latest, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// parseDuration converts duration strings to seconds
func parseDuration(duration string) (int64, error) {
	duration = strings.TrimSpace(duration)
	if len(duration) < 2 {
		return 0, errors.New("invalid duration format")
	}

	valueStr := duration[:len(duration)-1]
	unit := duration[len(duration)-1:]

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, err
	}

	switch unit {
	case "s":
		return value, nil
	case "m":
		return value * 60, nil
	case "h":
		return value * 3600, nil
	case "d":
		return value * 86400, nil
	default:
		return 0, errors.New("invalid duration unit: must be s, m, h, or d")
	}
}

// updateSensorsMetadata updates the Sensors node metadata
func (t *Tree) updateSensorsMetadata() {
	sensors, err := t.ListSensors()
	if err != nil {
		return
	}

	now := strconv.FormatInt(time.Now().Unix(), 10)
	t.SetValue(SensorsPath+"/__lastupdated", now)
	t.SetValue(SensorsPath+"/__sensor_count", strconv.Itoa(len(sensors)))
}
