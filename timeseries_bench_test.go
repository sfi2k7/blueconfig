package blueconfig

import (
	"fmt"
	"testing"
	"time"
)

// ============================================================================
// Benchmark Tests - Sensor Management
// ============================================================================

func BenchmarkInitTimeseries(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tmpfile := fmt.Sprintf("/tmp/bench_init_%d.db", time.Now().UnixNano())
		tr, _ := NewOrOpenTree(TreeOptions{
			StorageLocationOnDisk: tmpfile,
		})
		b.StartTimer()

		tr.InitTimeseries()

		b.StopTimer()
		tr.Close()
	}
}

func BenchmarkCreateSensor(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sensorName := fmt.Sprintf("sensor_%d", i)
		tr.CreateSensor(sensorName, SensorTypeGauge, "percent", "24h")
	}
}

func BenchmarkListSensors(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Create 100 sensors
	for i := 0; i < 100; i++ {
		tr.CreateSensor(fmt.Sprintf("sensor_%d", i), SensorTypeGauge, "unit", "24h")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.ListSensors()
	}
}

func BenchmarkGetSensorInfo(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "test_sensor"
	tr.CreateSensor(sensorName, SensorTypeGauge, "unit", "24h")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.GetSensorInfo(sensorName)
	}
}

func BenchmarkDeleteSensor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tr, tmpfile := createTimeseriesTree(&testing.T{})
		sensorName := "temp_sensor"
		tr.CreateSensor(sensorName, SensorTypeGauge, "unit", "1h")

		// Add some events
		for j := 0; j < 100; j++ {
			event := Event{Value: float64(j), TS: time.Now().Unix()}
			tr.PostEvent(sensorName, event)
		}
		b.StartTimer()

		tr.DeleteSensor(sensorName)

		b.StopTimer()
		cleanupTimeseriesTree(tr, tmpfile)
	}
}

// ============================================================================
// Benchmark Tests - Event Operations
// ============================================================================

func BenchmarkPostEvent(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "cpu_usage"
	tr.CreateSensor(sensorName, SensorTypeGauge, "percent", "24h")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := Event{
			Value: 50.0 + float64(i%50),
			TS:    time.Now().Unix(),
		}
		tr.PostEvent(sensorName, event)
	}
}

func BenchmarkPostEventWithAutoID(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "memory"
	tr.CreateSensor(sensorName, SensorTypeGauge, "bytes", "1h")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := Event{
			Value: float64(1024000 + i),
			TS:    time.Now().Unix(),
		}
		tr.PostEvent(sensorName, event)
	}
}

func BenchmarkGetAllEvents(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "temperature"
	tr.CreateSensor(sensorName, SensorTypeGauge, "celsius", "24h")

	// Add 1000 events
	baseTime := time.Now().Unix() - 3600
	for i := 0; i < 1000; i++ {
		event := Event{
			ID:    fmt.Sprintf("event_%d", i),
			Value: 20.0 + float64(i%10),
			TS:    baseTime + int64(i),
		}
		tr.PostEvent(sensorName, event)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.GetAllEvents(sensorName)
	}
}

func BenchmarkGetLatestValue(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "disk_usage"
	tr.CreateSensor(sensorName, SensorTypeGauge, "percent", "24h")

	// Add events
	now := time.Now().Unix()
	for i := 0; i < 100; i++ {
		event := Event{
			Value: float64(i),
			TS:    now - int64(100-i),
		}
		tr.PostEvent(sensorName, event)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.GetLatestValue(sensorName)
	}
}

func BenchmarkDeleteOldEvents(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "network"
	tr.CreateSensor(sensorName, SensorTypeCounter, "bytes", "1h")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Add mix of old and new events
		now := time.Now().Unix()
		for j := 0; j < 50; j++ {
			event := Event{
				Value: float64(j),
				TS:    now - 7200, // 2 hours old
			}
			tr.PostEvent(sensorName, event)
		}
		for j := 0; j < 50; j++ {
			event := Event{
				Value: float64(j),
				TS:    now - 1800, // 30 min old
			}
			tr.PostEvent(sensorName, event)
		}
		b.StartTimer()

		tr.DeleteOldEvents(sensorName)
	}
}

// ============================================================================
// Benchmark Tests - Query Operations
// ============================================================================

func BenchmarkGetSensorData_5m_30s(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "cpu"
	tr.CreateSensor(sensorName, SensorTypeGauge, "percent", "24h")

	// Add events over 10 minutes
	now := time.Now().Unix()
	baseTime := now - 600
	for i := 0; i < 120; i++ {
		event := Event{
			Value: 40.0 + float64(i%20),
			TS:    baseTime + int64(i*5),
		}
		tr.PostEvent(sensorName, event)
	}

	query := SensorQuery{
		WindowType: TumblingWindow,
		WindowSize: "5m",
		Scale:      "30s",
		OutputType: OutputTypeGraph,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.GetSensorData(sensorName, query)
	}
}

func BenchmarkGetSensorData_1h_5m(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "memory"
	tr.CreateSensor(sensorName, SensorTypeGauge, "bytes", "24h")

	// Add events over 2 hours
	now := time.Now().Unix()
	baseTime := now - 7200
	for i := 0; i < 720; i++ {
		event := Event{
			Value: float64(1000000 + i*1000),
			TS:    baseTime + int64(i*10),
		}
		tr.PostEvent(sensorName, event)
	}

	query := SensorQuery{
		WindowType: TumblingWindow,
		WindowSize: "1h",
		Scale:      "5m",
		OutputType: OutputTypeGraph,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.GetSensorData(sensorName, query)
	}
}

func BenchmarkGetSensorData_24h_1h(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "disk"
	tr.CreateSensor(sensorName, SensorTypeGauge, "percent", "7d")

	// Add events over 48 hours
	now := time.Now().Unix()
	baseTime := now - 172800
	for i := 0; i < 2880; i++ {
		event := Event{
			Value: float64(50 + i%30),
			TS:    baseTime + int64(i*60),
		}
		tr.PostEvent(sensorName, event)
	}

	query := SensorQuery{
		WindowType: TumblingWindow,
		WindowSize: "24h",
		Scale:      "1h",
		OutputType: OutputTypeGraph,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.GetSensorData(sensorName, query)
	}
}

// ============================================================================
// Benchmark Tests - Real World Scenarios
// ============================================================================

func BenchmarkServerMonitoring_SingleServer(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Create sensors for one server
	sensors := []struct {
		name       string
		sensorType string
		unit       string
	}{
		{"cpu_usage", SensorTypeGauge, "percent"},
		{"memory_used", SensorTypeGauge, "bytes"},
		{"disk_used", SensorTypeGauge, "bytes"},
		{"network_rx", SensorTypeCounter, "bytes"},
		{"network_tx", SensorTypeCounter, "bytes"},
		{"load_avg", SensorTypeGauge, "float"},
	}

	for _, s := range sensors {
		tr.CreateSensor(s.name, s.sensorType, s.unit, "24h")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		now := time.Now().Unix()
		// Simulate posting metrics for all sensors
		for _, s := range sensors {
			event := Event{
				Value: 50.0 + float64(i%50),
				TS:    now,
			}
			tr.PostEvent(s.name, event)
		}
	}
}

func BenchmarkServerMonitoring_20Servers(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Create sensors for 20 servers
	serverCount := 20
	metricsPerServer := 6

	for s := 1; s <= serverCount; s++ {
		metrics := []string{"cpu_usage", "memory_used", "disk_used", "network_rx", "network_tx", "load_avg"}
		for _, m := range metrics {
			sensorName := fmt.Sprintf("server%d_%s", s, m)
			sensorType := SensorTypeGauge
			if m == "network_rx" || m == "network_tx" {
				sensorType = SensorTypeCounter
			}
			tr.CreateSensor(sensorName, sensorType, "unit", "24h")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		now := time.Now().Unix()
		// Post one metric update per server
		serverID := (i % serverCount) + 1
		metricID := i % metricsPerServer
		metrics := []string{"cpu_usage", "memory_used", "disk_used", "network_rx", "network_tx", "load_avg"}
		sensorName := fmt.Sprintf("server%d_%s", serverID, metrics[metricID])

		event := Event{
			Value: float64(i % 100),
			TS:    now,
		}
		tr.PostEvent(sensorName, event)
	}
}

func BenchmarkIOTSensors_100Devices(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Create 100 IoT sensors
	deviceCount := 100

	for d := 1; d <= deviceCount; d++ {
		sensorName := fmt.Sprintf("iot_device_%d_temperature", d)
		tr.CreateSensor(sensorName, SensorTypeGauge, "celsius", "24h")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deviceID := (i % deviceCount) + 1
		sensorName := fmt.Sprintf("iot_device_%d_temperature", deviceID)

		event := Event{
			Value: 20.0 + float64(i%10),
			TS:    time.Now().Unix(),
		}
		tr.PostEvent(sensorName, event)
	}
}

func BenchmarkFullWorkflow_OneHour(b *testing.B) {
	b.Run("create_post_query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			tr, tmpfile := createTimeseriesTree(&testing.T{})
			sensorName := "test_sensor"
			tr.CreateSensor(sensorName, SensorTypeGauge, "unit", "24h")

			// Simulate 1 hour of data (1 event per minute)
			now := time.Now().Unix()
			baseTime := now - 3600
			for j := 0; j < 60; j++ {
				event := Event{
					Value: float64(j),
					TS:    baseTime + int64(j*60),
				}
				tr.PostEvent(sensorName, event)
			}
			b.StartTimer()

			// Query the data
			query := SensorQuery{
				WindowType: TumblingWindow,
				WindowSize: "1h",
				Scale:      "5m",
				OutputType: OutputTypeGraph,
			}
			tr.GetSensorData(sensorName, query)

			b.StopTimer()
			cleanupTimeseriesTree(tr, tmpfile)
		}
	})
}

// ============================================================================
// Benchmark Tests - Concurrent Operations
// ============================================================================

func BenchmarkConcurrentPostEvents(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	// Create multiple sensors
	for i := 0; i < 10; i++ {
		sensorName := fmt.Sprintf("sensor_%d", i)
		tr.CreateSensor(sensorName, SensorTypeGauge, "unit", "24h")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			sensorID := i % 10
			sensorName := fmt.Sprintf("sensor_%d", sensorID)

			event := Event{
				Value: float64(i),
				TS:    time.Now().Unix(),
			}
			tr.PostEvent(sensorName, event)
			i++
		}
	})
}

func BenchmarkConcurrentQueries(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "shared_sensor"
	tr.CreateSensor(sensorName, SensorTypeGauge, "unit", "24h")

	// Add some data
	now := time.Now().Unix()
	baseTime := now - 3600
	for i := 0; i < 360; i++ {
		event := Event{
			Value: float64(i),
			TS:    baseTime + int64(i*10),
		}
		tr.PostEvent(sensorName, event)
	}

	query := SensorQuery{
		WindowType: TumblingWindow,
		WindowSize: "1h",
		Scale:      "5m",
		OutputType: OutputTypeGraph,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tr.GetSensorData(sensorName, query)
		}
	})
}

// ============================================================================
// Benchmark Tests - Helper Functions
// ============================================================================

func BenchmarkParseDuration(b *testing.B) {
	durations := []string{"30s", "5m", "1h", "24h", "7d"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseDuration(durations[i%len(durations)])
	}
}

func BenchmarkScanNodes(b *testing.B) {
	tr, tmpfile := createTimeseriesTree(&testing.T{})
	defer cleanupTimeseriesTree(tr, tmpfile)

	sensorName := "scan_test"
	tr.CreateSensor(sensorName, SensorTypeGauge, "unit", "1h")

	// Add 100 events
	now := time.Now().Unix()
	for i := 0; i < 100; i++ {
		event := Event{
			ID:    fmt.Sprintf("event_%d", i),
			Value: float64(i),
			TS:    now + int64(i),
		}
		tr.PostEvent(sensorName, event)
	}

	sensorPath := SensorsPath + "/" + sensorName

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.ScanNodes(sensorPath, func(node NodeInfo) error {
			return nil
		})
	}
}
