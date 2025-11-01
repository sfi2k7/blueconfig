import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { fetchNodeChildren } from "../../../../store/actions/treeActions";
import SensorList from "./SensorList";
import SensorChart from "./SensorChart";
import SensorControls from "./SensorControls";
import EventLog from "./EventLog";
import SensorDetail from "./SensorDetail";
import api from "../../../../api/api";

const TimeseriesDashboard = ({ currentPath, properties }) => {
  const dispatch = useDispatch();
  const [sensors, setSensors] = useState([]);
  const [selectedSensor, setSelectedSensor] = useState(null);
  const [timeRange, setTimeRange] = useState("1h");
  const [interval, setInterval] = useState("1m");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    loadSensors();
  }, [currentPath]);

  const loadSensors = async () => {
    setLoading(true);
    setError(null);
    try {
      // First, check if there's a "Sensors" child node
      const nodeResponse = await api.get(`/node/${currentPath}`);

      if (nodeResponse.data.success) {
        const children = nodeResponse.data.data.children || [];

        // Look for Sensors node
        const hasSensorsNode = children.includes("Sensors");

        let sensorList = [];

        if (hasSensorsNode) {
          // Load sensors from the Sensors child node
          const sensorsPath = `${currentPath}/Sensors`;
          const sensorsResponse = await api.get(`/node/${sensorsPath}`);

          if (sensorsResponse.data.success) {
            const sensorNames = sensorsResponse.data.data.children || [];

            // Load info for each sensor
            const sensorInfoPromises = sensorNames.map(async (name) => {
              try {
                const sensorPath = `${sensorsPath}/${name}`;
                const propsResponse = await api.get(
                  `/properties/${sensorPath}`,
                );

                if (propsResponse.data.success) {
                  const props = propsResponse.data.data.properties || {};
                  return {
                    name,
                    path: sensorPath,
                    type: props.__sensorType || props.type || "gauge",
                    unit: props.__unit || props.unit || "",
                    retention: props.__retention || props.retention || "24h",
                    description: props.__description || props.description || "",
                    ...props,
                  };
                }
              } catch (err) {
                console.warn(`Failed to load info for ${name}:`, err);
                return {
                  name,
                  path: `${sensorsPath}/${name}`,
                  type: "gauge",
                };
              }
            });

            sensorList = await Promise.all(sensorInfoPromises);
          }
        } else {
          // No Sensors node - treat direct children as potential sensors
          const sensorInfoPromises = children.map(async (name) => {
            try {
              const childPath = `${currentPath}/${name}`;
              const propsResponse = await api.get(`/properties/${childPath}`);

              if (propsResponse.data.success) {
                const props = propsResponse.data.data.properties || {};

                // Only include if it looks like a sensor (has sensor-like properties)
                if (
                  props.__sensorType ||
                  props.type ||
                  props.__unit ||
                  props.unit
                ) {
                  return {
                    name,
                    path: childPath,
                    type: props.__sensorType || props.type || "gauge",
                    unit: props.__unit || props.unit || "",
                    retention: props.__retention || props.retention || "24h",
                    description: props.__description || props.description || "",
                    ...props,
                  };
                }
              }
              return null;
            } catch (err) {
              console.warn(`Failed to load info for ${name}:`, err);
              return null;
            }
          });

          const results = await Promise.all(sensorInfoPromises);
          sensorList = results.filter((s) => s !== null);
        }

        setSensors(sensorList);

        // Auto-select first sensor if none selected
        if (!selectedSensor && sensorList.length > 0) {
          setSelectedSensor(sensorList[0]);
        } else if (selectedSensor) {
          // Update selected sensor with fresh data
          const updated = sensorList.find(
            (s) => s.name === selectedSensor.name,
          );
          if (updated) {
            setSelectedSensor(updated);
          } else {
            setSelectedSensor(sensorList[0] || null);
          }
        }
      }
    } catch (err) {
      setError(err.message);
      console.error("Failed to load sensors:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    loadSensors();
  };

  const handleCreateSensor = async () => {
    const sensorName = prompt("Enter sensor name:");
    if (!sensorName) return;

    const sensorType = prompt("Enter sensor type (gauge/counter):", "gauge");
    const unit = prompt("Enter unit (optional):") || "";
    const retention = prompt("Enter retention period (e.g., 24h, 7d):", "24h");

    try {
      setLoading(true);

      // Ensure Sensors node exists
      const sensorsPath = `${currentPath}/Sensors`;
      try {
        await api.post(`/node/${sensorsPath}`);
      } catch (e) {
        // May already exist, that's fine
      }

      // Create the sensor node
      const sensorPath = `${sensorsPath}/${sensorName}`;
      await api.post(`/node/${sensorPath}`);

      // Set sensor properties
      await api.post(`/properties/${sensorPath}`, {
        key: "__sensorType",
        value: sensorType,
      });

      if (unit) {
        await api.post(`/properties/${sensorPath}`, {
          key: "__unit",
          value: unit,
        });
      }

      if (retention) {
        await api.post(`/properties/${sensorPath}`, {
          key: "__retention",
          value: retention,
        });
      }

      // Mark as timeseries type
      await api.post(`/properties/${sensorPath}`, {
        key: "__type",
        value: "sensor",
      });

      // Refresh sensor list
      await loadSensors();

      alert("Sensor created successfully!");
    } catch (err) {
      alert(`Failed to create sensor: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleSensorDeleted = async (sensorToDelete) => {
    if (!sensorToDelete || !sensorToDelete.path) return;

    const confirmed = window.confirm(
      `Are you sure you want to delete sensor "${sensorToDelete.name}"?`,
    );

    if (!confirmed) return;

    try {
      await api.delete(`/node/${sensorToDelete.path}`);
      setSelectedSensor(null);
      loadSensors();
    } catch (err) {
      alert(`Failed to delete sensor: ${err.message}`);
    }
  };

  return (
    <div className="timeseries-dashboard p-3">
      <div className="mb-3">
        <div className="d-flex justify-content-between align-items-center">
          <div>
            <h4>
              <i className="fas fa-chart-line me-2"></i>
              Timeseries Dashboard
            </h4>
            <small className="text-muted">{currentPath}</small>
          </div>
        </div>
      </div>

      {error && (
        <div
          className="alert alert-danger alert-dismissible fade show"
          role="alert"
        >
          <i className="fas fa-exclamation-triangle me-2"></i>
          {error}
          <button
            type="button"
            className="btn-close"
            onClick={() => setError(null)}
          ></button>
        </div>
      )}

      {loading && sensors.length === 0 ? (
        <div className="text-center py-5">
          <div className="spinner-border text-primary" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
          <p className="mt-2">Loading sensors...</p>
        </div>
      ) : sensors.length === 0 ? (
        <div className="alert alert-info">
          <i className="fas fa-info-circle me-2"></i>
          No sensors found under this node.
          <button
            className="btn btn-sm btn-primary ms-3"
            onClick={handleCreateSensor}
          >
            <i className="fas fa-plus me-1"></i>
            Create First Sensor
          </button>
        </div>
      ) : (
        <div className="row">
          <div className="col-md-3">
            <SensorList
              sensors={sensors}
              selectedSensor={selectedSensor}
              onSelectSensor={setSelectedSensor}
            />
          </div>

          <div className="col-md-9">
            <SensorControls
              timeRange={timeRange}
              interval={interval}
              onTimeRangeChange={setTimeRange}
              onIntervalChange={setInterval}
              onRefresh={handleRefresh}
              onCreateSensor={handleCreateSensor}
            />

            {selectedSensor && (
              <>
                <SensorChart
                  sensor={selectedSensor}
                  timeRange={timeRange}
                  interval={interval}
                />

                <div className="row mt-3">
                  <div className="col-md-6">
                    <EventLog sensor={selectedSensor} limit={10} />
                  </div>
                  <div className="col-md-6">
                    <SensorDetail
                      sensor={selectedSensor}
                      onSensorDeleted={() =>
                        handleSensorDeleted(selectedSensor)
                      }
                      onRefresh={loadSensors}
                    />
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default TimeseriesDashboard;
