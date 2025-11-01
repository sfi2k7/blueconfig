import { createAction, createAsyncThunk } from "@reduxjs/toolkit";
import { channelsApi } from "../../api/api";

// Async thunks
export const fetchChannels = createAsyncThunk(
  "channels/fetchChannels",
  async (_, { rejectWithValue }) => {
    try {
      const response = await channelsApi.getAll();
      return response.data.result || [];
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch channels"
      );
    }
  }
);

export const toggleChannelStatus = createAsyncThunk(
  "channels/toggleStatus",
  async ({ channel, status }, { rejectWithValue }) => {
    try {
      const response = await channelsApi.toggleStatus(channel, status);
      if (response.data.result === "ok") {
        return { channel, status };
      }
      return rejectWithValue("Failed to toggle channel status");
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to toggle channel status"
      );
    }
  }
);

export const deleteChannel = createAsyncThunk(
  "channels/delete",
  async (channel, { rejectWithValue }) => {
    try {
      const response = await channelsApi.delete(channel);
      if (response.data.result === "ok") {
        return channel;
      }
      return rejectWithValue(response.data.error || "Failed to delete channel");
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to delete channel"
      );
    }
  }
);

// Synchronous actions
export const setChannelsLoading = createAction("channels/setLoading");
export const setChannelsError = createAction("channels/setError");
export const clearChannelsError = createAction("channels/clearError");
