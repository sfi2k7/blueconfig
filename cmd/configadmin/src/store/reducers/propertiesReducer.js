import { createReducer } from "@reduxjs/toolkit";
import {
  setEditMode,
  clearProperties,
  setPropertyDirty,
  fetchProperties,
  setProperty,
  deleteProperty,
  clearAllPropertiesState,
} from "../actions/treeActions";

// Initial state for properties
const initialState = {
  currentNodePath: null,
  properties: {},
  originalProperties: {}, // Track original state for dirty checking
  editMode: false,
  isDirty: false,
  dirtyKeys: [], // Array instead of Set for Redux serialization
  loading: false,
  saving: false,
  error: null,
};

// Helper to check if properties are dirty
const checkIfDirty = (original, current) => {
  const originalKeys = Object.keys(original);
  const currentKeys = Object.keys(current);

  // Check if keys changed
  if (originalKeys.length !== currentKeys.length) return true;

  // Check if values changed
  for (const key of currentKeys) {
    if (original[key] !== current[key]) return true;
  }

  return false;
};

// Helper to get special properties (prefixed with __)
const getSpecialProperties = (properties) => {
  const special = {};
  Object.keys(properties).forEach((key) => {
    if (key.startsWith("__")) {
      special[key] = properties[key];
    }
  });
  return special;
};

// Helper to get regular properties (not prefixed with __)
const getRegularProperties = (properties) => {
  const regular = {};
  Object.keys(properties).forEach((key) => {
    if (!key.startsWith("__")) {
      regular[key] = properties[key];
    }
  });
  return regular;
};

// Properties reducer using createReducer
const propertiesReducer = createReducer(initialState, (builder) => {
  builder
    // ========================================================================
    // Edit Mode Management
    // ========================================================================
    .addCase(setEditMode, (state, action) => {
      const enabled = action.payload;

      state.editMode = enabled;

      // If exiting edit mode, reset to original properties
      if (!enabled && state.isDirty) {
        state.properties = { ...state.originalProperties };
        state.isDirty = false;
        state.dirtyKeys = [];
      }
    })

    // ========================================================================
    // Clear Properties
    // ========================================================================
    .addCase(clearProperties, (state) => {
      state.currentNodePath = null;
      state.properties = {};
      state.originalProperties = {};
      state.editMode = false;
      state.isDirty = false;
      state.dirtyKeys = [];
      state.error = null;
    })

    // ========================================================================
    // Set Property Dirty State
    // ========================================================================
    .addCase(setPropertyDirty, (state, action) => {
      const { key, isDirty } = action.payload;

      if (isDirty) {
        if (!state.dirtyKeys.includes(key)) {
          state.dirtyKeys.push(key);
        }
      } else {
        state.dirtyKeys = state.dirtyKeys.filter((k) => k !== key);
      }

      // Update overall dirty state
      state.isDirty = checkIfDirty(state.originalProperties, state.properties);
    })

    // ========================================================================
    // Fetch Properties (Async Thunk)
    // ========================================================================
    .addCase(fetchProperties.pending, (state) => {
      state.loading = true;
      state.error = null;
    })

    .addCase(fetchProperties.fulfilled, (state, action) => {
      const { path, properties } = action.payload;

      state.currentNodePath = path;
      state.properties = { ...properties };
      state.originalProperties = { ...properties };
      state.isDirty = false;
      state.dirtyKeys = [];
      state.loading = false;
      state.error = null;

      // Exit edit mode when loading new node
      state.editMode = false;
    })

    .addCase(fetchProperties.rejected, (state, action) => {
      state.loading = false;
      state.error = action.payload || "Failed to fetch properties";
    })

    // ========================================================================
    // Set Property (Async Thunk)
    // ========================================================================
    .addCase(setProperty.pending, (state) => {
      state.saving = true;
      state.error = null;
    })

    .addCase(setProperty.fulfilled, (state, action) => {
      const { path, key, value } = action.payload;

      // Only update if this is the current node
      if (state.currentNodePath === path) {
        state.properties[key] = value;
        state.originalProperties[key] = value;

        // Remove from dirty keys
        state.dirtyKeys = state.dirtyKeys.filter((k) => k !== key);

        // Recalculate dirty state
        state.isDirty = checkIfDirty(
          state.originalProperties,
          state.properties,
        );
      }

      state.saving = false;
      state.error = null;
    })

    .addCase(setProperty.rejected, (state, action) => {
      state.saving = false;
      state.error = action.payload || "Failed to set property";
    })

    // ========================================================================
    // Delete Property (Async Thunk)
    // ========================================================================
    .addCase(deleteProperty.pending, (state) => {
      state.saving = true;
      state.error = null;
    })

    .addCase(deleteProperty.fulfilled, (state, action) => {
      const { path, key } = action.payload;

      // Only update if this is the current node
      if (state.currentNodePath === path) {
        delete state.properties[key];
        delete state.originalProperties[key];

        // Remove from dirty keys
        state.dirtyKeys = state.dirtyKeys.filter((k) => k !== key);

        // Recalculate dirty state
        state.isDirty = checkIfDirty(
          state.originalProperties,
          state.properties,
        );
      }

      state.saving = false;
      state.error = null;
    })

    .addCase(deleteProperty.rejected, (state, action) => {
      state.saving = false;
      state.error = action.payload || "Failed to delete property";
    })

    // ========================================================================
    // Clear All Properties State
    // ========================================================================
    .addCase(clearAllPropertiesState, (state) => {
      // Reset to initial state
      state.currentNodePath = null;
      state.properties = {};
      state.originalProperties = {};
      state.editMode = false;
      state.isDirty = false;
      state.dirtyKeys = [];
      state.loading = false;
      state.saving = false;
      state.error = null;
    });
});

export default propertiesReducer;
