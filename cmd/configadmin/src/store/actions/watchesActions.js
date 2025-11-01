import { createAction, createAsyncThunk } from "@reduxjs/toolkit";
import { watchesApi } from "../../api/api";

// Async thunks
export const fetchWatches = createAsyncThunk(
  "watches/fetchWatches",
  async (_, { rejectWithValue }) => {
    try {
      const response = await watchesApi.getAll();
      return response.data.result || [];
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch watches"
      );
    }
  }
);

export const toggleWatchStatus = createAsyncThunk(
  "watches/toggleStatus",
  async ({ watch, status }, { rejectWithValue }) => {
    try {
      const response = await watchesApi.toggleStatus(watch, status);
      if (response.data.result === "ok") {
        return { watch, status };
      }
      return rejectWithValue("Failed to toggle watch status");
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to toggle watch status"
      );
    }
  }
);

export const deleteWatch = createAsyncThunk(
  "watches/delete",
  async (watch, { rejectWithValue }) => {
    try {
      const response = await watchesApi.delete(watch);
      if (response.data.result === "ok") {
        return watch;
      }
      return rejectWithValue(response.data.error || "Failed to delete watch");
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to delete watch"
      );
    }
  }
);

// Synchronous actions
export const setWatchesLoading = createAction("watches/setLoading");
export const setWatchesError = createAction("watches/setError");
export const clearWatchesError = createAction("watches/clearError");
