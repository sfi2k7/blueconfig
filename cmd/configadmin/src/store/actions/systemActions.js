import { createAction, createAsyncThunk } from "@reduxjs/toolkit";
import { systemApi } from "../../api/api";

// Async thunks
export const fetchSystemInfo = createAsyncThunk(
  "system/fetchInfo",
  async (_, { rejectWithValue }) => {
    try {
      const response = await systemApi.getInfo();
      return response.data.result || {};
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch system info",
      );
    }
  },
);

export const fetchSystemStats = createAsyncThunk(
  "system/fetchStats",
  async (_, { rejectWithValue }) => {
    try {
      const response = await systemApi.getStats();
      return response.data.result || {};
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch system stats",
      );
    }
  },
);

export const toggleSystemStatus = createAsyncThunk(
  "system/toggleStatus",
  async (status, { rejectWithValue }) => {
    try {
      const response = await systemApi.toggleStatus(status);
      if (response.data.result === "ok") {
        return status;
      }
      return rejectWithValue("Failed to toggle system status");
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to toggle system status",
      );
    }
  },
);

// Synchronous actions
export const setSystemLoading = createAction("system/setLoading");
export const setSystemError = createAction("system/setError");
export const clearSystemError = createAction("system/clearError");
