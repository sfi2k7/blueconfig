import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import {
  setEditMode,
  fetchProperties,
  setProperty,
  deleteProperty,
  navigateToPath,
} from "../../../store/actions/treeActions";

// Helper to format path for display (hide "root" prefix)
const formatPathForDisplay = (path) => {
  if (!path) return "";
  if (path === "root") return "/";
  return path.replace(/^root\/?/, "/");
};

const Properties = () => {
  const dispatch = useDispatch();

  // Single useSelector calls for state
  const treeState = useSelector((state) => state.tree);
  const propertiesState = useSelector((state) => state.properties);

  // Extract needed values
  const { currentPath } = treeState;
  const { properties, editMode, isDirty, loading, saving, error } =
    propertiesState;

  // Local state for editing
  const [editingKey, setEditingKey] = useState(null);
  const [editingValue, setEditingValue] = useState("");
  const [newKey, setNewKey] = useState("");
  const [newValue, setNewValue] = useState("");
  const [showAddForm, setShowAddForm] = useState(false);
  const [pathInput, setPathInput] = useState("");

  // Update path input when current path changes
  useEffect(() => {
    if (currentPath) {
      setPathInput(formatPathForDisplay(currentPath));
    }
  }, [currentPath]);

  // Reset local state when edit mode changes
  useEffect(() => {
    if (!editMode) {
      setEditingKey(null);
      setEditingValue("");
      setShowAddForm(false);
      setNewKey("");
      setNewValue("");
    }
  }, [editMode]);

  // Content type detection
  const detectContentType = (value) => {
    if (typeof value !== "string") return "text";
    const trimmed = value.trim();

    if (trimmed.startsWith("{") || trimmed.startsWith("[")) {
      try {
        JSON.parse(trimmed);
        return "json";
      } catch {
        return "text";
      }
    }
    if (
      trimmed.startsWith("#") ||
      trimmed.includes("\n##") ||
      trimmed.includes("\n###")
    ) {
      return "markdown";
    }
    if (trimmed.startsWith("<!DOCTYPE") || trimmed.startsWith("<html")) {
      return "html";
    }
    if (trimmed.startsWith("data:image/")) {
      return "image";
    }
    if (
      trimmed.length > 100 &&
      (trimmed.includes("\n") || trimmed.includes("<"))
    ) {
      return "blob";
    }
    return "text";
  };

  // Format value for display
  const formatValue = (value, contentType) => {
    if (contentType === "json") {
      try {
        return JSON.stringify(JSON.parse(value), null, 2);
      } catch {
        return value;
      }
    }
    return value;
  };

  // Toggle edit mode
  const handleToggleEditMode = () => {
    if (editMode && isDirty) {
      if (!window.confirm("You have unsaved changes. Discard them?")) {
        return;
      }
    }
    dispatch(setEditMode(!editMode));
  };

  // Start editing a property
  const handleEditProperty = (key, value) => {
    setEditingKey(key);
    setEditingValue(value);
  };

  // Save edited property
  const handleSaveProperty = async (key) => {
    if (editingValue === properties[key]) {
      setEditingKey(null);
      return;
    }

    await dispatch(
      setProperty({
        path: currentPath,
        key: key,
        value: editingValue,
      }),
    );

    setEditingKey(null);
    setEditingValue("");
  };

  // Cancel editing
  const handleCancelEdit = () => {
    setEditingKey(null);
    setEditingValue("");
  };

  // Delete property
  const handleDeleteProperty = async (key) => {
    if (!window.confirm(`Delete property "${key}"?`)) {
      return;
    }

    await dispatch(
      deleteProperty({
        path: currentPath,
        key: key,
      }),
    );
  };

  // Add new property
  const handleAddProperty = async () => {
    if (!newKey.trim()) {
      alert("Property key cannot be empty");
      return;
    }

    if (properties.hasOwnProperty(newKey)) {
      alert("Property key already exists");
      return;
    }

    await dispatch(
      setProperty({
        path: currentPath,
        key: newKey,
        value: newValue,
      }),
    );

    setNewKey("");
    setNewValue("");
    setShowAddForm(false);
  };

  // Navigate to path
  const handleGoToPath = async () => {
    if (!pathInput) return;

    // Convert display path back to internal path
    let internalPath = pathInput;
    if (pathInput === "/") {
      internalPath = "root";
    } else if (pathInput.startsWith("/")) {
      internalPath = "root" + pathInput;
    }

    if (internalPath !== currentPath) {
      // Use navigateToPath to expand tree and set current path
      await dispatch(navigateToPath(internalPath));
      // Fetch properties for the target path
      dispatch(fetchProperties(internalPath));
    }
  };

  // Separate special and regular properties
  const specialProps = {};
  const regularProps = {};

  Object.keys(properties).forEach((key) => {
    if (key.startsWith("__")) {
      specialProps[key] = properties[key];
    } else {
      regularProps[key] = properties[key];
    }
  });

  return (
    <div className="properties-container">
      {/* Path Navigation */}
      <div className="properties-header">
        <div className="input-group mb-3">
          <span className="input-group-text">
            <i className="fas fa-route"></i>
          </span>
          <input
            type="text"
            className="form-control"
            placeholder="Enter path (e.g., /users/john or / for root)"
            value={pathInput}
            onChange={(e) => setPathInput(e.target.value)}
            onKeyPress={(e) => e.key === "Enter" && handleGoToPath()}
          />
          <button
            className="btn btn-primary"
            onClick={handleGoToPath}
            disabled={
              !pathInput || pathInput === formatPathForDisplay(currentPath)
            }
          >
            <i className="fas fa-arrow-right"></i> Go
          </button>
        </div>

        <div className="d-flex justify-content-between align-items-center mb-3">
          <div>
            <h5 className="mb-0">
              <i className="fas fa-list me-2"></i>
              Properties
              {currentPath && (
                <small className="text-muted ms-2">
                  ({formatPathForDisplay(currentPath)})
                </small>
              )}
            </h5>
          </div>
          <div>
            <button
              className={`btn btn-sm ${editMode ? "btn-success" : "btn-outline-secondary"}`}
              onClick={handleToggleEditMode}
              disabled={loading}
            >
              <i
                className={`fas ${editMode ? "fa-lock-open" : "fa-lock"} me-1`}
              ></i>
              {editMode ? "Read/Write" : "Read Only"}
            </button>
          </div>
        </div>
      </div>

      {/* Loading/Error States */}
      {loading && (
        <div className="text-center py-4">
          <div className="spinner-border text-primary" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
        </div>
      )}

      {error && (
        <div className="alert alert-danger" role="alert">
          <i className="fas fa-exclamation-triangle me-2"></i>
          {error}
        </div>
      )}

      {!loading && currentPath && (
        <div className="properties-body">
          {/* Special Properties Section */}
          {Object.keys(specialProps).length > 0 && (
            <div className="mb-4">
              <h6 className="text-muted mb-2">
                <i className="fas fa-star me-1"></i>
                Special Properties
              </h6>
              <div className="table-responsive">
                <table className="table table-sm table-bordered">
                  <thead className="table-light">
                    <tr>
                      <th style={{ width: "30%" }}>Key</th>
                      <th style={{ width: "50%" }}>Value</th>
                      {editMode && <th style={{ width: "20%" }}>Actions</th>}
                    </tr>
                  </thead>
                  <tbody>
                    {Object.entries(specialProps).map(([key, value]) => (
                      <tr key={key}>
                        <td>
                          <code className="text-primary">{key}</code>
                          {key === "__password" && (
                            <small className="d-block text-warning">
                              <i className="fas fa-lock me-1"></i>
                              Store Password
                            </small>
                          )}
                        </td>
                        <td>
                          {editMode && editingKey === key ? (
                            <input
                              type={key === "__password" ? "password" : "text"}
                              className="form-control form-control-sm"
                              value={editingValue}
                              onChange={(e) => setEditingValue(e.target.value)}
                              autoFocus
                            />
                          ) : (
                            <span className="font-monospace">
                              {key === "__password" ? "••••••••" : value}
                            </span>
                          )}
                        </td>
                        {editMode && (
                          <td>
                            {editingKey === key ? (
                              <>
                                <button
                                  className="btn btn-sm btn-success me-1"
                                  onClick={() => handleSaveProperty(key)}
                                  disabled={saving}
                                >
                                  <i className="fas fa-save"></i>
                                </button>
                                <button
                                  className="btn btn-sm btn-secondary"
                                  onClick={handleCancelEdit}
                                >
                                  <i className="fas fa-times"></i>
                                </button>
                              </>
                            ) : (
                              <>
                                <button
                                  className="btn btn-sm btn-warning me-1"
                                  onClick={() => handleEditProperty(key, value)}
                                >
                                  <i className="fas fa-edit"></i>
                                </button>
                                <button
                                  className="btn btn-sm btn-danger"
                                  onClick={() => handleDeleteProperty(key)}
                                  disabled={saving}
                                >
                                  <i className="fas fa-trash"></i>
                                </button>
                              </>
                            )}
                          </td>
                        )}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Regular Properties Section */}
          <div className="mb-3">
            <h6 className="text-muted mb-2">
              <i className="fas fa-database me-1"></i>
              Properties
            </h6>
            {Object.keys(regularProps).length === 0 ? (
              <div className="alert alert-info">
                <i className="fas fa-info-circle me-2"></i>
                No properties found for this node.
              </div>
            ) : (
              <div className="table-responsive">
                <table className="table table-sm table-bordered table-hover">
                  <thead className="table-light">
                    <tr>
                      <th style={{ width: "25%" }}>Key</th>
                      <th style={{ width: "10%" }}>Type</th>
                      <th style={{ width: "45%" }}>Value</th>
                      {editMode && <th style={{ width: "20%" }}>Actions</th>}
                    </tr>
                  </thead>
                  <tbody>
                    {Object.entries(regularProps).map(([key, value]) => {
                      const contentType = detectContentType(value);
                      const displayValue = formatValue(value, contentType);

                      return (
                        <tr key={key}>
                          <td>
                            <code>{key}</code>
                          </td>
                          <td>
                            <span
                              className={`badge bg-${
                                contentType === "json"
                                  ? "info"
                                  : contentType === "markdown"
                                    ? "warning"
                                    : "secondary"
                              }`}
                            >
                              {contentType}
                            </span>
                          </td>
                          <td>
                            {editMode && editingKey === key ? (
                              <textarea
                                className="form-control form-control-sm font-monospace"
                                rows={contentType === "text" ? 1 : 5}
                                value={editingValue}
                                onChange={(e) =>
                                  setEditingValue(e.target.value)
                                }
                                autoFocus
                              />
                            ) : (
                              <pre
                                className="mb-0"
                                style={{
                                  maxHeight: "150px",
                                  overflow: "auto",
                                  fontSize: "0.85rem",
                                  whiteSpace:
                                    contentType === "text"
                                      ? "nowrap"
                                      : "pre-wrap",
                                }}
                              >
                                {displayValue}
                              </pre>
                            )}
                          </td>
                          {editMode && (
                            <td>
                              {editingKey === key ? (
                                <>
                                  <button
                                    className="btn btn-sm btn-success me-1"
                                    onClick={() => handleSaveProperty(key)}
                                    disabled={saving}
                                  >
                                    <i className="fas fa-save"></i>
                                  </button>
                                  <button
                                    className="btn btn-sm btn-secondary"
                                    onClick={handleCancelEdit}
                                  >
                                    <i className="fas fa-times"></i>
                                  </button>
                                </>
                              ) : (
                                <>
                                  <button
                                    className="btn btn-sm btn-warning me-1"
                                    onClick={() =>
                                      handleEditProperty(key, value)
                                    }
                                  >
                                    <i className="fas fa-edit"></i>
                                  </button>
                                  <button
                                    className="btn btn-sm btn-danger"
                                    onClick={() => handleDeleteProperty(key)}
                                    disabled={saving}
                                  >
                                    <i className="fas fa-trash"></i>
                                  </button>
                                </>
                              )}
                            </td>
                          )}
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {/* Add Property Form */}
          {editMode && (
            <div className="add-property-section">
              {!showAddForm ? (
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => setShowAddForm(true)}
                >
                  <i className="fas fa-plus me-1"></i>
                  Add Property
                </button>
              ) : (
                <div className="card">
                  <div className="card-body">
                    <h6 className="card-title">Add New Property</h6>
                    <div className="mb-2">
                      <label className="form-label">Key</label>
                      <input
                        type="text"
                        className="form-control form-control-sm"
                        placeholder="Property key"
                        value={newKey}
                        onChange={(e) => setNewKey(e.target.value)}
                      />
                    </div>
                    <div className="mb-3">
                      <label className="form-label">Value</label>
                      <textarea
                        className="form-control form-control-sm"
                        rows="3"
                        placeholder="Property value"
                        value={newValue}
                        onChange={(e) => setNewValue(e.target.value)}
                      />
                    </div>
                    <div className="d-flex gap-2">
                      <button
                        className="btn btn-success btn-sm"
                        onClick={handleAddProperty}
                        disabled={!newKey.trim() || saving}
                      >
                        <i className="fas fa-save me-1"></i>
                        Save
                      </button>
                      <button
                        className="btn btn-secondary btn-sm"
                        onClick={() => {
                          setShowAddForm(false);
                          setNewKey("");
                          setNewValue("");
                        }}
                      >
                        <i className="fas fa-times me-1"></i>
                        Cancel
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Saving Indicator */}
          {saving && (
            <div className="alert alert-info mt-3">
              <i className="fas fa-spinner fa-spin me-2"></i>
              Saving changes...
            </div>
          )}
        </div>
      )}

      {!loading && !currentPath && (
        <div className="alert alert-secondary text-center">
          <i className="fas fa-arrow-left me-2"></i>
          Select a node from the tree to view its properties
        </div>
      )}
    </div>
  );
};

export default Properties;
