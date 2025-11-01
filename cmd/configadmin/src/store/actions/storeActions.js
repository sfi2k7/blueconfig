import { createAction, createAsyncThunk } from "@reduxjs/toolkit";
import { storeApi } from "../../api/api";

// Async thunks
export const fetchStoreStatus = createAsyncThunk(
  "store/fetchStatus",
  async (_, { rejectWithValue }) => {
    try {
      const response = await storeApi.getStatus();
      return response.data.result || {};
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch store status",
      );
    }
  },
);

export const fetchStoreStats = createAsyncThunk(
  "store/fetchStats",
  async (_, { rejectWithValue }) => {
    try {
      const response = await storeApi.getStats();
      return response.data.result || {};
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch store stats",
      );
    }
  },
);

// Synchronous actions
export const setStoreLoading = createAction("store/setLoading");
export const setStoreError = createAction("store/setError");
export const clearStoreError = createAction("store/clearError");
