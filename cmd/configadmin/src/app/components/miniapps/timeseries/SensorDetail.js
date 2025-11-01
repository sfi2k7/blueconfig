import React, { useState } from "react";
import api from "../../../../api/api";

const SensorDetail = ({ sensor, onSensorDeleted, onRefresh }) => {
  const [deleting, setDeleting] = useState(false);
  const [cleaning, setCleaning] = useState(false);

  const handleDeleteSensor = async () => {
    if (
      !window.confirm(
        `Are you sure you want to delete sensor "${sensor.name}"?\n\nThis will permanently delete all associated events.`,
      )
    ) {
      return;
    }

    setDeleting(true);
    try {
      // Use node deletion API if we have a path
      if (sensor.path) {
        await api.delete(`/node/${sensor.path}`);
        alert(`Sensor "${sensor.name}" deleted successfully`);
      } else {
        // Fallback to old API
        const response = await api.delete(`/timeseries/sensors/${sensor.name}`);
        if (response.data.success) {
          alert(`Sensor "${sensor.name}" deleted successfully`);
        }
      }

      if (onSensorDeleted) {
        onSensorDeleted();
      }
    } catch (err) {
      alert(`Failed to delete sensor: ${err.message}`);
      console.error("Delete sensor failed:", err);
    } finally {
      setDeleting(false);
    }
  };

  const handleCleanup = async () => {
    const retention = sensor.retention || sensor.__retention || "24h";
    if (
      !window.confirm(
        `Clean up old events for sensor "${sensor.name}"?\n\nThis will delete events older than the retention period (${retention}).`,
      )
    ) {
      return;
    }

    setCleaning(true);
    try {
      // For node-based sensors, we'd need to implement cleanup logic
      // For now, just use the old API if available
      if (sensor.path) {
        alert(
          "Cleanup is not yet implemented for node-based sensors. You can manually delete event nodes if needed.",
        );
      } else {
        const response = await api.post(
          `/timeseries/sensors/${sensor.name}/cleanup`,
        );
        if (response.data.success) {
          const deleted = response.data.result?.deleted || 0;
          alert(`Successfully deleted ${deleted} old event(s)`);
        }
      }

      if (onRefresh) {
        onRefresh();
      }
    } catch (err) {
      alert(`Failed to cleanup events: ${err.message}`);
      console.error("Cleanup failed:", err);
    } finally {
      setCleaning(false);
    }
  };

  const formatTimestamp = (ts) => {
    if (!ts) return "N/A";
    const date = new Date(parseInt(ts) * 1000);
    return date.toLocaleString();
  };

  // Get the correct property values (support both formats)
  const sensorType =
    sensor.type || sensor.__sensorType || sensor.__sensor_type || "N/A";
  const unit = sensor.unit || sensor.__unit || "N/A";
  const retention = sensor.retention || sensor.__retention || "N/A";
  const eventCount = sensor.event_count || sensor.__event_count || "0";
  const created = sensor.created || sensor.__created;
  const lastUpdated = sensor.lastupdated || sensor.__lastupdated;

  return (
    <div className="sensor-detail card mb-3">
      <div className="card-header">
        <h6 className="mb-0">
          <i className="fas fa-info-circle me-2"></i>
          Sensor Details
        </h6>
      </div>
      <div className="card-body">
        <dl className="row mb-0">
          <dt className="col-sm-5">Name:</dt>
          <dd className="col-sm-7">{sensor.name}</dd>

          <dt className="col-sm-5">Type:</dt>
          <dd className="col-sm-7">
            <span
              className={`badge ${sensorType === "gauge" ? "bg-info" : "bg-primary"}`}
            >
              {sensorType}
            </span>
          </dd>

          <dt className="col-sm-5">Unit:</dt>
          <dd className="col-sm-7">{unit}</dd>

          <dt className="col-sm-5">Retention:</dt>
          <dd className="col-sm-7">{retention}</dd>

          {sensor.path && (
            <>
              <dt className="col-sm-5">Path:</dt>
              <dd className="col-sm-7">
                <small className="text-muted font-monospace">
                  {sensor.path}
                </small>
              </dd>
            </>
          )}

          <dt className="col-sm-5">Event Count:</dt>
          <dd className="col-sm-7">
            <span className="badge bg-secondary">{eventCount}</span>
          </dd>

          {created && (
            <>
              <dt className="col-sm-5">Created:</dt>
              <dd className="col-sm-7">
                <small className="text-muted">{formatTimestamp(created)}</small>
              </dd>
            </>
          )}

          {lastUpdated && (
            <>
              <dt className="col-sm-5">Last Updated:</dt>
              <dd className="col-sm-7">
                <small className="text-muted">
                  {formatTimestamp(lastUpdated)}
                </small>
              </dd>
            </>
          )}
        </dl>

        <hr />

        <div className="d-grid gap-2">
          <button
            className="btn btn-warning btn-sm"
            onClick={handleCleanup}
            disabled={cleaning || deleting}
          >
            {cleaning ? (
              <>
                <span
                  className="spinner-border spinner-border-sm me-2"
                  role="status"
                ></span>
                Cleaning...
              </>
            ) : (
              <>
                <i className="fas fa-broom me-2"></i>
                Cleanup Old Events
              </>
            )}
          </button>

          <button
            className="btn btn-danger btn-sm"
            onClick={handleDeleteSensor}
            disabled={deleting || cleaning}
          >
            {deleting ? (
              <>
                <span
                  className="spinner-border spinner-border-sm me-2"
                  role="status"
                ></span>
                Deleting...
              </>
            ) : (
              <>
                <i className="fas fa-trash me-2"></i>
                Delete Sensor
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};

export default SensorDetail;
