import { createAction, createAsyncThunk } from "@reduxjs/toolkit";
import api from "../../api/api";

// ============================================================================
// Tree Actions
// ============================================================================

// Sync actions for tree state management
export const expandNode = createAction("tree/expandNode");
export const collapseNode = createAction("tree/collapseNode");
export const setCurrentPath = createAction("tree/setCurrentPath");
export const clearNodeData = createAction("tree/clearNodeData");
export const setNodeChildren = createAction("tree/setNodeChildren");
export const updateSpecialProperty = createAction("tree/updateSpecialProperty");
export const clearAllTreeState = createAction("tree/clearAllTreeState");

// Navigate to path and expand all parent nodes
export const navigateToPath = createAsyncThunk(
  "tree/navigateToPath",
  async (targetPath, { dispatch, getState }) => {
    try {
      // Normalize path
      let path = targetPath;
      if (path === "/" || path === "") {
        path = "root";
      } else if (path.startsWith("/")) {
        path = "root" + path;
      }

      // Get all segments in the path
      const segments = path.split("/").filter((s) => s);

      // Build array of paths to expand (all parents)
      const pathsToExpand = [];
      for (let i = 0; i < segments.length; i++) {
        const currentPath = segments.slice(0, i + 1).join("/");
        pathsToExpand.push(currentPath);
      }

      // Expand each path in sequence
      for (const pathToExpand of pathsToExpand) {
        const state = getState();
        const nodeData = state.tree.nodes[pathToExpand];

        // If node doesn't exist or isn't loaded, fetch children
        if (!nodeData || !nodeData.loaded) {
          dispatch(expandNode(pathToExpand));
          await dispatch(fetchNodeChildren(pathToExpand)).unwrap();
        } else if (!nodeData.expanded) {
          // Just expand if already loaded
          dispatch(expandNode(pathToExpand));
        }
      }

      // Set current path to the target
      dispatch(setCurrentPath(path));

      return { path };
    } catch (error) {
      throw error;
    }
  },
);

// Async thunk to fetch node children
export const fetchNodeChildren = createAsyncThunk(
  "tree/fetchNodeChildren",
  async (path, { rejectWithValue }) => {
    try {
      const response = await api.get(`/node/${path}`);
      if (response.data.success) {
        const children = response.data.data.children || [];

        // Fetch special properties for each child
        const specialProperties = {};
        await Promise.all(
          children.map(async (childName) => {
            try {
              const childPath =
                path === "root" ? `root/${childName}` : `${path}/${childName}`;
              const propsResponse = await api.get(`/properties/${childPath}`);

              if (propsResponse.data.success) {
                const props = propsResponse.data.data.properties || {};
                const specialProps = {};

                // Extract only special properties (__title, __icon, __color)
                if (props.__title) specialProps.__title = props.__title;
                if (props.__icon) specialProps.__icon = props.__icon;
                if (props.__color) specialProps.__color = props.__color;

                if (Object.keys(specialProps).length > 0) {
                  specialProperties[childPath] = specialProps;
                }
              }
            } catch (err) {
              // Silently ignore errors for individual children
              console.warn(`Failed to fetch properties for ${childName}:`, err);
            }
          }),
        );

        return {
          path,
          children,
          hasChildren: response.data.data.hasChildren,
          specialProperties,
        };
      }
      return rejectWithValue(response.data.error || "Failed to fetch children");
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || error.message);
    }
  },
);

// Async thunk to create a new node
export const createNode = createAsyncThunk(
  "tree/createNode",
  async (path, { rejectWithValue }) => {
    try {
      const response = await api.post(`/node/${path}`);
      if (response.data.success) {
        return {
          path,
          name: response.data.data.name,
        };
      }
      return rejectWithValue(response.data.error || "Failed to create node");
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || error.message);
    }
  },
);

// Async thunk to delete a node
export const deleteNode = createAsyncThunk(
  "tree/deleteNode",
  async (path, { rejectWithValue }) => {
    try {
      const response = await api.delete(`/node/${path}`);
      if (response.data.success) {
        return { path };
      }
      return rejectWithValue(response.data.error || "Failed to delete node");
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || error.message);
    }
  },
);

// ============================================================================
// Property Actions
// ============================================================================

// Sync actions for property state
export const setEditMode = createAction("properties/setEditMode");
export const clearProperties = createAction("properties/clearProperties");
export const setPropertyDirty = createAction("properties/setPropertyDirty");
export const clearAllPropertiesState = createAction(
  "properties/clearAllPropertiesState",
);

// Async thunk to fetch properties for a node
export const fetchProperties = createAsyncThunk(
  "properties/fetchProperties",
  async (path, { rejectWithValue }) => {
    try {
      const response = await api.get(`/properties/${path}`);
      if (response.data.success) {
        return {
          path: response.data.data.path,
          properties: response.data.data.properties || {},
        };
      }
      return rejectWithValue(
        response.data.error || "Failed to fetch properties",
      );
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || error.message);
    }
  },
);

// Async thunk to set a property
export const setProperty = createAsyncThunk(
  "properties/setProperty",
  async ({ path, key, value }, { rejectWithValue, dispatch }) => {
    try {
      const response = await api.post(`/properties/${path}`, { key, value });
      if (response.data.success) {
        // If it's a special property, update tree state as well
        if (key.startsWith("__")) {
          dispatch(updateSpecialProperty({ path, key, value }));
        }
        return { path, key, value };
      }
      return rejectWithValue(response.data.error || "Failed to set property");
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || error.message);
    }
  },
);

// Async thunk to delete a property
export const deleteProperty = createAsyncThunk(
  "properties/deleteProperty",
  async ({ path, key }, { rejectWithValue, dispatch }) => {
    try {
      const response = await api.delete(
        `/properties/${path}?key=${encodeURIComponent(key)}`,
      );
      if (response.data.success) {
        // If it's a special property, update tree state as well
        if (key.startsWith("__")) {
          dispatch(updateSpecialProperty({ path, key, value: null }));
        }
        return { path, key };
      }
      return rejectWithValue(
        response.data.error || "Failed to delete property",
      );
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || error.message);
    }
  },
);
