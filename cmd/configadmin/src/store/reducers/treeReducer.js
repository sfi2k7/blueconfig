import { createReducer } from "@reduxjs/toolkit";
import {
  expandNode,
  collapseNode,
  setCurrentPath,
  clearNodeData,
  setNodeChildren,
  updateSpecialProperty,
  fetchNodeChildren,
  createNode,
  deleteNode,
  clearAllTreeState,
} from "../actions/treeActions";

// Initial state for tree
const initialState = {
  // nodes structure: { 'root': { expanded: true, children: ['users', 'config'], loading: false, loaded: false } }
  nodes: {
    root: {
      expanded: true,
      children: [],
      loading: false,
      loaded: false,
    },
  },
  // Special properties for each node: { 'root': { __title: 'BlueConfig', __icon: 'fa-database', __color: '#3498db' } }
  specialProperties: {},
  currentPath: "root",
  expandedPaths: ["root"],
  loading: false,
  error: null,
};

// Helper to get parent path
const getParentPath = (path) => {
  if (path === "root" || path === "") return null;
  const segments = path.split("/");
  segments.pop();
  return segments.length > 0 ? segments.join("/") : "root";
};

// Helper to check if path is expanded
const isPathExpanded = (state, path) => {
  return state.expandedPaths.includes(path);
};

// Helper to add to expanded paths
const addToExpandedPaths = (state, path) => {
  if (!state.expandedPaths.includes(path)) {
    state.expandedPaths.push(path);
  }
};

// Helper to remove from expanded paths and all children
const removeFromExpandedPaths = (state, path) => {
  state.expandedPaths = state.expandedPaths.filter((p) => {
    return p !== path && !p.startsWith(path + "/");
  });
};

// Tree reducer using createReducer
const treeReducer = createReducer(initialState, (builder) => {
  builder
    // ========================================================================
    // Expand/Collapse Actions
    // ========================================================================
    .addCase(expandNode, (state, action) => {
      const path = action.payload;

      // Initialize node if doesn't exist
      if (!state.nodes[path]) {
        state.nodes[path] = {
          expanded: false,
          children: [],
          loading: false,
          loaded: false,
        };
      }

      state.nodes[path].expanded = true;
      addToExpandedPaths(state, path);
    })

    .addCase(collapseNode, (state, action) => {
      const path = action.payload;

      if (state.nodes[path]) {
        state.nodes[path].expanded = false;
        removeFromExpandedPaths(state, path);

        // Clear child node data to save memory
        const childrenToRemove = [];
        Object.keys(state.nodes).forEach((nodePath) => {
          if (nodePath.startsWith(path + "/")) {
            childrenToRemove.push(nodePath);
          }
        });

        // Remove nested children but keep immediate children names
        childrenToRemove.forEach((childPath) => {
          delete state.nodes[childPath];
        });
      }
    })

    // ========================================================================
    // Current Path Management
    // ========================================================================
    .addCase(setCurrentPath, (state, action) => {
      state.currentPath = action.payload;
    })

    // ========================================================================
    // Node Data Management
    // ========================================================================
    .addCase(clearNodeData, (state, action) => {
      const path = action.payload;
      if (state.nodes[path]) {
        state.nodes[path].children = [];
        state.nodes[path].loaded = false;
      }
    })

    .addCase(setNodeChildren, (state, action) => {
      const { path, children } = action.payload;
      if (!state.nodes[path]) {
        state.nodes[path] = {
          expanded: false,
          children: [],
          loading: false,
          loaded: false,
        };
      }
      state.nodes[path].children = children;
      state.nodes[path].loaded = true;
    })

    // ========================================================================
    // Fetch Node Children (Async Thunk)
    // ========================================================================
    .addCase(fetchNodeChildren.pending, (state, action) => {
      const path = action.meta.arg;
      if (!state.nodes[path]) {
        state.nodes[path] = {
          expanded: false,
          children: [],
          loading: false,
          loaded: false,
        };
      }
      state.nodes[path].loading = true;
      state.error = null;
    })

    .addCase(fetchNodeChildren.fulfilled, (state, action) => {
      const { path, children, hasChildren, specialProperties } = action.payload;

      if (state.nodes[path]) {
        state.nodes[path].children = children || [];
        state.nodes[path].hasChildren = hasChildren;
        state.nodes[path].loading = false;
        state.nodes[path].loaded = true;
        state.nodes[path].expanded = true;
      }

      // Store special properties if provided
      if (specialProperties) {
        Object.entries(specialProperties).forEach(([childPath, props]) => {
          if (props && Object.keys(props).length > 0) {
            state.specialProperties[childPath] = props;
          }
        });
      }

      addToExpandedPaths(state, path);
      state.error = null;
    })

    .addCase(fetchNodeChildren.rejected, (state, action) => {
      const path = action.meta.arg;
      if (state.nodes[path]) {
        state.nodes[path].loading = false;
      }
      state.error = action.payload || "Failed to fetch node children";
    })

    // ========================================================================
    // Create Node (Async Thunk)
    // ========================================================================
    .addCase(createNode.pending, (state) => {
      state.loading = true;
      state.error = null;
    })

    .addCase(createNode.fulfilled, (state, action) => {
      const { path, name } = action.payload;
      const parentPath = getParentPath(path);

      // Initialize the new node
      state.nodes[path] = {
        expanded: false,
        children: [],
        loading: false,
        loaded: false,
      };

      // Add to parent's children list if parent exists
      if (parentPath && state.nodes[parentPath]) {
        if (!state.nodes[parentPath].children.includes(name)) {
          state.nodes[parentPath].children.push(name);
        }
      }

      state.loading = false;
      state.error = null;
    })

    .addCase(createNode.rejected, (state, action) => {
      state.loading = false;
      state.error = action.payload || "Failed to create node";
    })

    // ========================================================================
    // Delete Node (Async Thunk)
    // ========================================================================
    .addCase(deleteNode.pending, (state) => {
      state.loading = true;
      state.error = null;
    })

    .addCase(deleteNode.fulfilled, (state, action) => {
      const { path } = action.payload;
      const parentPath = getParentPath(path);
      const nodeName = path.split("/").pop();

      // Remove from parent's children
      if (parentPath && state.nodes[parentPath]) {
        state.nodes[parentPath].children = state.nodes[
          parentPath
        ].children.filter((child) => child !== nodeName);
      }

      // Remove node and all its children from state
      const pathsToRemove = [path];
      Object.keys(state.nodes).forEach((nodePath) => {
        if (nodePath.startsWith(path + "/")) {
          pathsToRemove.push(nodePath);
        }
      });

      pathsToRemove.forEach((p) => {
        delete state.nodes[p];
      });

      // Remove from expanded paths
      removeFromExpandedPaths(state, path);

      // If current path was deleted, move to parent
      if (
        state.currentPath === path ||
        state.currentPath.startsWith(path + "/")
      ) {
        state.currentPath = parentPath || "root";
      }

      state.loading = false;
      state.error = null;
    })

    .addCase(deleteNode.rejected, (state, action) => {
      state.loading = false;
      state.error = action.payload || "Failed to delete node";
    })

    // ========================================================================
    // Update Special Property (Sync)
    // ========================================================================
    .addCase(updateSpecialProperty, (state, action) => {
      const { path, key, value } = action.payload;

      // Initialize special properties for this path if needed
      if (!state.specialProperties[path]) {
        state.specialProperties[path] = {};
      }

      // If value is null, delete the property
      if (value === null || value === undefined) {
        delete state.specialProperties[path][key];
        // Clean up empty objects
        if (Object.keys(state.specialProperties[path]).length === 0) {
          delete state.specialProperties[path];
        }
      } else {
        // Set the special property
        state.specialProperties[path][key] = value;
      }
    })

    // ========================================================================
    // Clear All Tree State
    // ========================================================================
    .addCase(clearAllTreeState, (state) => {
      // Reset to initial state
      state.nodes = {
        root: {
          expanded: true,
          children: [],
          loading: false,
          loaded: false,
        },
      };
      state.specialProperties = {};
      state.currentPath = "root";
      state.expandedPaths = ["root"];
      state.loading = false;
      state.error = null;
    });
});

export default treeReducer;
