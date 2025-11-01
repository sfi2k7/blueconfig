import React, { useEffect, useState } from "react";
import api from "../../../../api/api";

const EventLog = ({ sensor, limit = 10 }) => {
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (sensor) {
      loadEvents();

      // Auto-refresh every 5 seconds
      const intervalId = setInterval(loadEvents, 5000);
      return () => clearInterval(intervalId);
    }
  }, [sensor]);

  const loadEvents = async () => {
    if (!sensor) return;

    setLoading(true);
    setError(null);
    try {
      // Load events from sensor node's children
      if (sensor.path) {
        const response = await api.get(`/node/${sensor.path}`);

        if (response.data.success) {
          const children = response.data.data.children || [];

          // Load properties for each child (event)
          const eventPromises = children
            .slice(0, limit * 2)
            .map(async (childName) => {
              try {
                const eventPath = `${sensor.path}/${childName}`;
                const propsResponse = await api.get(`/properties/${eventPath}`);

                if (propsResponse.data.success) {
                  const props = propsResponse.data.data.properties || {};
                  return {
                    ts: props.timestamp || props.ts || Date.now() / 1000,
                    value: props.value || props.val || 0,
                    name: childName,
                  };
                }
              } catch (err) {
                return null;
              }
            });

          const allEvents = (await Promise.all(eventPromises))
            .filter((e) => e !== null)
            .sort((a, b) => b.ts - a.ts)
            .slice(0, limit);

          setEvents(allEvents);
        }
      } else {
        // Fallback to old API
        const response = await api.get(
          `/timeseries/sensors/${sensor.name}/events`,
        );

        if (response.data.success) {
          const allEvents = response.data.result || [];

          // Sort by timestamp descending and take latest
          const latest = allEvents.sort((a, b) => b.ts - a.ts).slice(0, limit);

          setEvents(latest);
        }
      }
    } catch (err) {
      // Don't show error for no events
      console.warn("No events available:", err);
      setEvents([]);
    } finally {
      setLoading(false);
    }
  };

  const formatTimestamp = (ts) => {
    const date = new Date(ts * 1000);
    return date.toLocaleString();
  };

  return (
    <div className="event-log card mb-3">
      <div className="card-header d-flex justify-content-between align-items-center">
        <h6 className="mb-0">
          <i className="fas fa-stream me-2"></i>
          Recent Events
        </h6>
        {loading && (
          <div
            className="spinner-border spinner-border-sm text-primary"
            role="status"
          >
            <span className="visually-hidden">Loading...</span>
          </div>
        )}
      </div>
      <div className="card-body p-0">
        {error && (
          <div className="alert alert-danger m-2 mb-0">
            <i className="fas fa-exclamation-triangle me-2"></i>
            {error}
          </div>
        )}

        {events.length === 0 && !loading ? (
          <div className="text-center text-muted py-3">
            <i className="fas fa-inbox me-2"></i>
            No events yet. Post events to see them here.
          </div>
        ) : (
          <div
            className="table-responsive"
            style={{ maxHeight: "300px", overflowY: "auto" }}
          >
            <table className="table table-sm table-hover mb-0">
              <thead className="table-light sticky-top">
                <tr>
                  <th>Time</th>
                  <th className="text-end">Value</th>
                </tr>
              </thead>
              <tbody>
                {events.map((event, idx) => (
                  <tr key={idx}>
                    <td className="small">{formatTimestamp(event.ts)}</td>
                    <td className="text-end">
                      <strong>{event.value}</strong>
                      {(sensor.unit || sensor.__unit) && (
                        <span className="text-muted ms-1">
                          {sensor.unit || sensor.__unit}
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default EventLog;
