import React, { useEffect, useState } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import api from "../../../../api/api";

const SensorChart = ({ sensor, timeRange, interval }) => {
  const [chartData, setChartData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (sensor && sensor.name) {
      loadChartData();
    }
  }, [sensor, timeRange, interval]);

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp * 1000);

    // Parse interval to determine format
    const intervalMatch = interval.match(/^(\d+)([smh])$/);
    const intervalSeconds = intervalMatch
      ? parseInt(intervalMatch[1]) *
        ({ s: 1, m: 60, h: 3600 }[intervalMatch[2]] || 60)
      : 60;

    if (intervalSeconds < 60) {
      // Show seconds for sub-minute intervals
      return date.toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      });
    } else if (intervalSeconds < 3600) {
      // Show HH:MM for minute intervals
      return date.toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      });
    } else {
      // Show date for hour+ intervals
      return date.toLocaleDateString([], {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    }
  };

  const loadChartData = async () => {
    setLoading(true);
    setError(null);

    try {
      // Use the backend GetSensorData API
      const response = await api.get(
        `/timeseries/sensors/${sensor.name}/data`,
        {
          params: {
            window_size: timeRange,
            scale: interval,
            window_type: "tumbling",
            output_type: "graph",
          },
        },
      );

      if (response.data.success) {
        const result = response.data.data.result;
        const points = result.Points || [];

        console.log(`Loaded ${points.length} data points for ${sensor.name}`);

        const formattedData = points.map((p) => ({
          timestamp: formatTimestamp(p.Timestamp),
          rawTimestamp: p.Timestamp,
          value:
            p.Count > 0 && p.Avg !== undefined
              ? parseFloat(p.Avg.toFixed(2))
              : null,
          min:
            p.Count > 0 && p.Min !== undefined
              ? parseFloat(p.Min.toFixed(2))
              : null,
          max:
            p.Count > 0 && p.Max !== undefined
              ? parseFloat(p.Max.toFixed(2))
              : null,
          count: p.Count || 0,
        }));

        setChartData(formattedData);
      } else {
        setError(response.data.error || "Failed to load chart data");
      }
    } catch (err) {
      console.error("Failed to load chart data:", err);
      setError(err.response?.data?.error || err.message);
      setChartData([]);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="sensor-chart card mb-3">
        <div className="card-body text-center py-5">
          <div className="spinner-border text-primary" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
          <p className="mt-2">Loading chart data...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="sensor-chart card mb-3">
        <div className="card-body">
          <div className="alert alert-danger mb-0">
            <i className="fas fa-exclamation-triangle me-2"></i>
            Failed to load chart: {error}
          </div>
        </div>
      </div>
    );
  }

  if (chartData.length === 0) {
    return (
      <div className="sensor-chart card mb-3">
        <div className="card-header">
          <h6 className="mb-0">
            <i className="fas fa-chart-line me-2"></i>
            {sensor.name}
            {(sensor.unit || sensor.__unit) && (
              <span className="text-muted ms-2">
                ({sensor.unit || sensor.__unit})
              </span>
            )}
          </h6>
        </div>
        <div className="card-body">
          <div className="alert alert-info mb-0">
            <i className="fas fa-info-circle me-2"></i>
            No event data available in the selected time range ({timeRange}).
            Try a larger time window or post events to this sensor.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="sensor-chart card mb-3">
      <div className="card-header">
        <h6 className="mb-0">
          <i className="fas fa-chart-line me-2"></i>
          {sensor.name}
          {(sensor.unit || sensor.__unit) && (
            <span className="text-muted ms-2">
              ({sensor.unit || sensor.__unit})
            </span>
          )}
          <span className="text-muted ms-2 small">
            ({chartData.filter((p) => p.count > 0).length}/{chartData.length}{" "}
            buckets with data • {timeRange} range • {interval} interval)
          </span>
        </h6>
      </div>
      <div className="card-body">
        <ResponsiveContainer width="100%" height={300}>
          <LineChart
            data={chartData}
            margin={{ top: 5, right: 30, left: 20, bottom: 60 }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis
              dataKey="timestamp"
              tick={{ fontSize: 11 }}
              angle={-45}
              textAnchor="end"
              height={80}
              interval="preserveStartEnd"
            />
            <YAxis tick={{ fontSize: 12 }} />
            <Tooltip
              contentStyle={{
                backgroundColor: "#fff",
                border: "1px solid #ccc",
              }}
              formatter={(value, name) => {
                if (name === "Value") return [value, "Avg"];
                return [value, name];
              }}
            />
            <Legend />
            <Line
              type="monotone"
              dataKey="value"
              stroke="#0d6efd"
              strokeWidth={2}
              name="Value"
              dot={{ r: 3 }}
              activeDot={{ r: 5 }}
              connectNulls
            />
            {chartData[0]?.min !== undefined && (
              <Line
                type="monotone"
                dataKey="min"
                stroke="#198754"
                strokeWidth={1}
                strokeDasharray="3 3"
                name="Min"
                dot={false}
                connectNulls
              />
            )}
            {chartData[0]?.max !== undefined && (
              <Line
                type="monotone"
                dataKey="max"
                stroke="#dc3545"
                strokeWidth={1}
                strokeDasharray="3 3"
                name="Max"
                dot={false}
                connectNulls
              />
            )}
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};

export default SensorChart;
