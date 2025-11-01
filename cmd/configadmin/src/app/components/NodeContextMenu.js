import React, { useState } from "react";
import { useDispatch } from "react-redux";
import {
  createNode,
  deleteNode,
  setProperty,
} from "../../store/actions/treeActions";

const NodeContextMenu = ({ path, nodeName, onClose, position }) => {
  const dispatch = useDispatch();

  // Local state for modals
  const [showAddChild, setShowAddChild] = useState(false);
  const [showSetIcon, setShowSetIcon] = useState(false);
  const [showSetTitle, setShowSetTitle] = useState(false);
  const [showSetColor, setShowSetColor] = useState(false);

  // Form state
  const [childName, setChildName] = useState("");
  const [iconValue, setIconValue] = useState("");
  const [titleValue, setTitleValue] = useState("");
  const [colorValue, setColorValue] = useState("#3498db");

  // Common Font Awesome icons
  const commonIcons = [
    "fa-folder",
    "fa-user",
    "fa-users",
    "fa-database",
    "fa-server",
    "fa-cog",
    "fa-file",
    "fa-lock",
    "fa-key",
    "fa-envelope",
    "fa-bell",
    "fa-calendar",
    "fa-chart-bar",
    "fa-home",
    "fa-briefcase",
    "fa-cloud",
    "fa-gear",
    "fa-star",
    "fa-heart",
    "fa-bookmark",
    "fa-flag",
    "fa-tag",
    "fa-shield",
  ];

  // Common colors
  const commonColors = [
    { name: "Blue", value: "#3498db" },
    { name: "Green", value: "#2ecc71" },
    { name: "Red", value: "#e74c3c" },
    { name: "Orange", value: "#f39c12" },
    { name: "Purple", value: "#9b59b6" },
    { name: "Teal", value: "#1abc9c" },
    { name: "Gray", value: "#95a5a6" },
    { name: "Dark", value: "#34495e" },
    { name: "Pink", value: "#e91e63" },
    { name: "Cyan", value: "#00bcd4" },
    { name: "Lime", value: "#cddc39" },
    { name: "Amber", value: "#ffc107" },
    { name: "Indigo", value: "#3f51b5" },
    { name: "Brown", value: "#795548" },
  ];

  const handleAddChild = async () => {
    if (!childName.trim()) {
      alert("Child node name cannot be empty");
      return;
    }

    const childPath =
      path === "root" ? `root/${childName}` : `${path}/${childName}`;
    await dispatch(createNode(childPath));
    setChildName("");
    setShowAddChild(false);
    onClose();
  };

  const handleDeleteNode = async () => {
    if (path === "root") {
      alert("Cannot delete root node");
      return;
    }

    if (!window.confirm(`Delete node "${nodeName}" and all its children?`)) {
      return;
    }

    await dispatch(deleteNode(path));
    onClose();
  };

  const handleSetIcon = async () => {
    if (!iconValue.trim()) {
      alert("Icon value cannot be empty");
      return;
    }

    await dispatch(
      setProperty({
        path: path,
        key: "__icon",
        value: iconValue,
      }),
    );
    setIconValue("");
    setShowSetIcon(false);
    onClose();
  };

  const handleSetTitle = async () => {
    if (!titleValue.trim()) {
      alert("Title cannot be empty");
      return;
    }

    await dispatch(
      setProperty({
        path: path,
        key: "__title",
        value: titleValue,
      }),
    );
    setTitleValue("");
    setShowSetTitle(false);
    onClose();
  };

  const handleSetColor = async () => {
    await dispatch(
      setProperty({
        path: path,
        key: "__color",
        value: colorValue,
      }),
    );
    setColorValue("#3498db");
    setShowSetColor(false);
    onClose();
  };

  return (
    <>
      {/* Context Menu Backdrop */}
      <div
        className="context-menu-backdrop"
        onClick={onClose}
        style={{
          position: "fixed",
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          zIndex: 1000,
        }}
      />

      {/* Context Menu */}
      <div
        className="context-menu"
        style={{
          position: "fixed",
          top: position.y,
          left: position.x,
          zIndex: 1001,
        }}
      >
        <div className="context-menu-header p-2">
          <small className="text-muted">
            <i className="fas fa-folder me-1"></i>
            {nodeName}
          </small>
        </div>

        <div className="context-menu-body">
          {/* Add Child Node */}
          <button
            className="context-menu-item"
            onClick={() => setShowAddChild(true)}
          >
            <i className="fas fa-plus-circle me-2 text-success"></i>
            Add Child Node
          </button>

          <div className="dropdown-divider"></div>

          {/* Set Special Properties */}
          <button
            className="context-menu-item"
            onClick={() => setShowSetTitle(true)}
          >
            <i className="fas fa-heading me-2 text-primary"></i>
            Set Title
          </button>

          <button
            className="context-menu-item"
            onClick={() => setShowSetIcon(true)}
          >
            <i className="fas fa-icons me-2 text-info"></i>
            Set Icon
          </button>

          <button
            className="context-menu-item"
            onClick={() => setShowSetColor(true)}
          >
            <i className="fas fa-palette me-2 text-warning"></i>
            Set Color
          </button>

          {path !== "root" && (
            <>
              <div className="dropdown-divider"></div>

              {/* Delete Node */}
              <button
                className="context-menu-item text-danger"
                onClick={handleDeleteNode}
              >
                <i className="fas fa-trash me-2"></i>
                Delete Node
              </button>
            </>
          )}
        </div>
      </div>

      {/* Modal: Add Child Node */}
      {showAddChild && (
        <div className="modal-backdrop-custom">
          <div className="modal-dialog-custom">
            <div className="modal-content">
              <div className="modal-header">
                <h6 className="modal-title">
                  <i className="fas fa-plus-circle me-2"></i>
                  Add Child Node
                </h6>
                <button
                  type="button"
                  className="btn-close"
                  onClick={() => {
                    setShowAddChild(false);
                    setChildName("");
                  }}
                ></button>
              </div>
              <div className="modal-body">
                <p className="text-muted small mb-2">
                  Parent: <code>{path}</code>
                </p>
                <label className="form-label">Child Node Name</label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="e.g., users, config, settings"
                  value={childName}
                  onChange={(e) => setChildName(e.target.value)}
                  onKeyPress={(e) => e.key === "Enter" && handleAddChild()}
                  autoFocus
                />
              </div>
              <div className="modal-footer">
                <button
                  className="btn btn-secondary btn-sm"
                  onClick={() => {
                    setShowAddChild(false);
                    setChildName("");
                  }}
                >
                  Cancel
                </button>
                <button
                  className="btn btn-success btn-sm"
                  onClick={handleAddChild}
                  disabled={!childName.trim()}
                >
                  <i className="fas fa-plus me-1"></i>
                  Create
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Modal: Set Icon */}
      {showSetIcon && (
        <div className="modal-backdrop-custom">
          <div className="modal-dialog-custom">
            <div className="modal-content">
              <div className="modal-header">
                <h6 className="modal-title">
                  <i className="fas fa-icons me-2"></i>
                  Set Node Icon
                </h6>
                <button
                  type="button"
                  className="btn-close"
                  onClick={() => {
                    setShowSetIcon(false);
                    setIconValue("");
                  }}
                ></button>
              </div>
              <div className="modal-body">
                <label className="form-label">Font Awesome Icon Class</label>
                <input
                  type="text"
                  className="form-control mb-3"
                  placeholder="e.g., fa-user, fa-folder, fa-database"
                  value={iconValue}
                  onChange={(e) => setIconValue(e.target.value)}
                  onKeyPress={(e) => e.key === "Enter" && handleSetIcon()}
                />

                <label className="form-label">Common Icons</label>
                <div className="icon-grid">
                  {commonIcons.map((icon) => (
                    <button
                      key={icon}
                      className={`icon-choice ${iconValue === icon ? "selected" : ""}`}
                      onClick={() => setIconValue(icon)}
                      title={icon}
                    >
                      <i className={`fas ${icon}`}></i>
                    </button>
                  ))}
                </div>
              </div>
              <div className="modal-footer">
                <button
                  className="btn btn-secondary btn-sm"
                  onClick={() => {
                    setShowSetIcon(false);
                    setIconValue("");
                  }}
                >
                  Cancel
                </button>
                <button
                  className="btn btn-primary btn-sm"
                  onClick={handleSetIcon}
                  disabled={!iconValue.trim()}
                >
                  <i className="fas fa-save me-1"></i>
                  Save
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Modal: Set Title */}
      {showSetTitle && (
        <div className="modal-backdrop-custom">
          <div className="modal-dialog-custom">
            <div className="modal-content">
              <div className="modal-header">
                <h6 className="modal-title">
                  <i className="fas fa-heading me-2"></i>
                  Set Node Title
                </h6>
                <button
                  type="button"
                  className="btn-close"
                  onClick={() => {
                    setShowSetTitle(false);
                    setTitleValue("");
                  }}
                ></button>
              </div>
              <div className="modal-body">
                <p className="text-muted small mb-2">
                  Node: <code>{nodeName}</code>
                </p>
                <label className="form-label">Display Title</label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="e.g., My Users, Database Config"
                  value={titleValue}
                  onChange={(e) => setTitleValue(e.target.value)}
                  onKeyPress={(e) => e.key === "Enter" && handleSetTitle()}
                  autoFocus
                />
                <small className="text-muted">
                  This title will be displayed instead of the node name
                </small>
              </div>
              <div className="modal-footer">
                <button
                  className="btn btn-secondary btn-sm"
                  onClick={() => {
                    setShowSetTitle(false);
                    setTitleValue("");
                  }}
                >
                  Cancel
                </button>
                <button
                  className="btn btn-primary btn-sm"
                  onClick={handleSetTitle}
                  disabled={!titleValue.trim()}
                >
                  <i className="fas fa-save me-1"></i>
                  Save
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Modal: Set Color */}
      {showSetColor && (
        <div className="modal-backdrop-custom">
          <div className="modal-dialog-custom">
            <div className="modal-content">
              <div className="modal-header">
                <h6 className="modal-title">
                  <i className="fas fa-palette me-2"></i>
                  Set Node Color
                </h6>
                <button
                  type="button"
                  className="btn-close"
                  onClick={() => {
                    setShowSetColor(false);
                    setColorValue("#3498db");
                  }}
                ></button>
              </div>
              <div className="modal-body">
                <label className="form-label">Color Value</label>
                <div className="input-group mb-3">
                  <input
                    type="color"
                    className="form-control form-control-color"
                    value={colorValue}
                    onChange={(e) => setColorValue(e.target.value)}
                  />
                  <input
                    type="text"
                    className="form-control"
                    value={colorValue}
                    onChange={(e) => setColorValue(e.target.value)}
                  />
                </div>

                <label className="form-label">Common Colors</label>
                <div className="color-grid">
                  {commonColors.map((color) => (
                    <button
                      key={color.value}
                      className={`color-choice ${colorValue === color.value ? "selected" : ""}`}
                      onClick={() => setColorValue(color.value)}
                      title={color.name}
                      style={{ backgroundColor: color.value }}
                    >
                      {colorValue === color.value && (
                        <i className="fas fa-check text-white"></i>
                      )}
                    </button>
                  ))}
                </div>
              </div>
              <div className="modal-footer">
                <button
                  className="btn btn-secondary btn-sm"
                  onClick={() => {
                    setShowSetColor(false);
                    setColorValue("#3498db");
                  }}
                >
                  Cancel
                </button>
                <button
                  className="btn btn-primary btn-sm"
                  onClick={handleSetColor}
                >
                  <i className="fas fa-save me-1"></i>
                  Save
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default NodeContextMenu;
