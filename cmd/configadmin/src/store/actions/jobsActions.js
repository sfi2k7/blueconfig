import { createAction, createAsyncThunk } from "@reduxjs/toolkit";
import { jobsApi } from "../../api/api";

// Async thunks
export const fetchAllJobs = createAsyncThunk(
  "jobs/fetchAll",
  async (_, { rejectWithValue }) => {
    try {
      const response = await jobsApi.getAll();
      return response.data.result || [];
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch jobs"
      );
    }
  }
);

export const searchJobs = createAsyncThunk(
  "jobs/search",
  async (pattern, { rejectWithValue }) => {
    try {
      const response = await jobsApi.search(pattern);
      return response.data.result || [];
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to search jobs"
      );
    }
  }
);

export const fetchJobsByChannel = createAsyncThunk(
  "jobs/fetchByChannel",
  async (channel, { rejectWithValue }) => {
    try {
      const response = await jobsApi.getByChannel(channel);
      return response.data.result || [];
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch jobs by channel"
      );
    }
  }
);

export const fetchJobById = createAsyncThunk(
  "jobs/fetchById",
  async (id, { rejectWithValue }) => {
    try {
      const response = await jobsApi.getById(id);
      return response.data.result || null;
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.error || "Failed to fetch job"
      );
    }
  }
);

// Synchronous actions
export const setJobsLoading = createAction("jobs/setLoading");
export const setJobsError = createAction("jobs/setError");
export const clearJobsError = createAction("jobs/clearError");
export const setSelectedJob = createAction("jobs/setSelectedJob");
export const clearSelectedJob = createAction("jobs/clearSelectedJob");
export const setSearchPattern = createAction("jobs/setSearchPattern");
export const setSelectedChannel = createAction("jobs/setSelectedChannel");
export const setViewMode = createAction("jobs/setViewMode");
