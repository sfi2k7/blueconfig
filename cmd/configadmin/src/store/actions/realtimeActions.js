import { createAction } from "@reduxjs/toolkit";

// WebSocket connection actions
export const wsConnect = createAction("realtime/connect");
export const wsDisconnect = createAction("realtime/disconnect");
export const wsConnected = createAction("realtime/connected");
export const wsDisconnected = createAction("realtime/disconnected");
export const wsError = createAction("realtime/error");

// Real-time event actions
export const jobPopped = createAction(
  "realtime/jobPopped",
  (jobID, channel) => ({
    payload: { jobID, channel, timestamp: Date.now() },
  }),
);

export const jobRouted = createAction("realtime/jobRouted", (jobID, to) => ({
  payload: { jobID, to, timestamp: Date.now() },
}));

// Activity history actions
export const addActivity = createAction("realtime/addActivity");
export const clearActivities = createAction("realtime/clearActivities");
