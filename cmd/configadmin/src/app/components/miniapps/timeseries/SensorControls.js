import React from 'react';

const SensorControls = ({ timeRange, interval, onTimeRangeChange, onIntervalChange, onRefresh, onCreateSensor }) => {
  return (
    <div className="sensor-controls card mb-3">
      <div className="card-body">
        <div className="row align-items-end">
          <div className="col-md-3">
            <label className="form-label small">Time Range</label>
            <select
              className="form-select form-select-sm"
              value={timeRange}
              onChange={(e) => onTimeRangeChange(e.target.value)}
            >
              <option value="5m">Last 5 minutes</option>
              <option value="15m">Last 15 minutes</option>
              <option value="1h">Last 1 hour</option>
              <option value="6h">Last 6 hours</option>
              <option value="12h">Last 12 hours</option>
              <option value="24h">Last 24 hours</option>
            </select>
          </div>

          <div className="col-md-3">
            <label className="form-label small">Interval</label>
            <select
              className="form-select form-select-sm"
              value={interval}
              onChange={(e) => onIntervalChange(e.target.value)}
            >
              <option value="10s">10 seconds</option>
              <option value="30s">30 seconds</option>
              <option value="1m">1 minute</option>
              <option value="5m">5 minutes</option>
              <option value="15m">15 minutes</option>
              <option value="1h">1 hour</option>
            </select>
          </div>

          <div className="col-md-6 text-end">
            <button className="btn btn-primary btn-sm me-2" onClick={onRefresh}>
              <i className="fas fa-sync me-1"></i>
              Refresh
            </button>
            <button className="btn btn-success btn-sm" onClick={onCreateSensor}>
              <i className="fas fa-plus me-1"></i>
              Create Sensor
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SensorControls;
