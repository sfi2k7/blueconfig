import { createAction, createAsyncThunk } from "@reduxjs/toolkit";
import { timeseriesApi } from "../../api/api";

// Async thunks
export const fetchTimeseriesData = createAsyncThunk(
  "timeseries/fetchData",
  async (window = "1h", { rejectWithValue }) => {
    try {
      const response = await timeseriesApi.getData(window);
      return {
        window,
        data: response.data.result || {},
      };
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch timeseries data",
      );
    }
  },
);

// Synchronous actions
export const setTimeseriesWindow = createAction("timeseries/setWindow");
export const setTimeseriesLoading = createAction("timeseries/setLoading");
export const setTimeseriesError = createAction("timeseries/setError");
export const clearTimeseriesError = createAction("timeseries/clearError");

// Real-time update action
export const timeseriesUpdate = createAction(
  "timeseries/update",
  (window, data) => ({
    payload: { window, data },
  }),
);
