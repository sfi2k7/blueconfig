import React from 'react';

const getSensorIcon = (sensorType) => {
  if (sensorType === 'gauge') return 'tachometer-alt';
  if (sensorType === 'counter') return 'sort-numeric-up';
  return 'chart-line';
};

const SensorList = ({ sensors, selectedSensor, onSelectSensor }) => {
  return (
    <div className="sensor-list card">
      <div className="card-header">
        <h6 className="mb-0">
          <i className="fas fa-list me-2"></i>
          Sensors ({sensors.length})
        </h6>
      </div>
      <div className="list-group list-group-flush" style={{ maxHeight: '600px', overflowY: 'auto' }}>
        {sensors.map((sensor) => (
          <button
            key={sensor.name}
            className={`list-group-item list-group-item-action ${
              selectedSensor?.name === sensor.name ? 'active' : ''
            }`}
            onClick={() => onSelectSensor(sensor)}
          >
            <div className="d-flex justify-content-between align-items-start">
              <div className="flex-grow-1">
                <div className="fw-bold">
                  <i className={`fas fa-${getSensorIcon(sensor.__sensor_type)} me-2`}></i>
                  {sensor.name}
                </div>
                <div className="small mt-1">
                  {sensor.__sensor_type && (
                    <span className="badge bg-secondary me-1">{sensor.__sensor_type}</span>
                  )}
                  {sensor.__retention && (
                    <span className="badge bg-info me-1">{sensor.__retention}</span>
                  )}
                  {sensor.__unit && (
                    <span className="text-muted">{sensor.__unit}</span>
                  )}
                </div>
              </div>
              {sensor.__event_count && (
                <span className="badge bg-primary rounded-pill">{sensor.__event_count}</span>
              )}
            </div>
          </button>
        ))}
      </div>
    </div>
  );
};

export default SensorList;
