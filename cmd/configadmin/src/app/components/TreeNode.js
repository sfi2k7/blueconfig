import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import {
  expandNode,
  collapseNode,
  fetchNodeChildren,
  setCurrentPath,
} from "../../store/actions/treeActions";
import NodeContextMenu from "./NodeContextMenu";

const TreeNode = ({ path, name, level = 0 }) => {
  const dispatch = useDispatch();

  // Single useSelector call for all tree state
  const treeState = useSelector((state) => state.tree);

  // Extract needed values from state
  const nodeData = treeState.nodes[path];
  const currentPath = treeState.currentPath;
  const specialProps = treeState.specialProperties[path] || {};

  // Context menu state
  const [contextMenu, setContextMenu] = useState(null);

  // Node state
  const isExpanded = nodeData?.expanded || false;
  const isLoading = nodeData?.loading || false;
  const children = nodeData?.children || [];
  const isLoaded = nodeData?.loaded || false;
  const isActive = currentPath === path;

  // Get special properties for display
  const getDisplayName = () => {
    if (specialProps.__title) {
      return specialProps.__title;
    }
    return name;
  };

  const getNodeColor = () => {
    if (specialProps.__color) {
      return specialProps.__color;
    }
    return null;
  };

  const getNodeIcon = () => {
    if (specialProps.__icon) {
      return specialProps.__icon;
    }
    return null;
  };

  // Handle expand/collapse toggle
  const handleToggle = async (e) => {
    e.stopPropagation();

    if (isExpanded) {
      // Collapse the node
      dispatch(collapseNode(path));
    } else {
      // Expand the node
      dispatch(expandNode(path));

      // Fetch children if not already loaded
      if (!isLoaded) {
        dispatch(fetchNodeChildren(path));
      }
    }
  };

  // Handle node click (select node to view properties)
  const handleNodeClick = () => {
    dispatch(setCurrentPath(path));
  };

  // Handle right-click for context menu
  const handleContextMenu = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
    });
  };

  // Close context menu
  const closeContextMenu = () => {
    setContextMenu(null);
  };

  // Calculate indentation
  const indentStyle = {
    paddingLeft: `${level * 20}px`,
  };

  const displayName = getDisplayName();
  const nodeColor = getNodeColor();
  const nodeIcon = getNodeIcon();

  return (
    <div className="tree-node">
      <div
        className={`tree-node-content ${isActive ? "active" : ""}`}
        style={indentStyle}
        onClick={handleNodeClick}
        onContextMenu={handleContextMenu}
      >
        {/* Expand/Collapse Icon */}
        <span className="tree-toggle" onClick={handleToggle}>
          {isLoading ? (
            <i className="fas fa-spinner fa-spin"></i>
          ) : children.length > 0 || !isLoaded ? (
            <i
              className={`fas ${isExpanded ? "fa-minus-square" : "fa-plus-square"}`}
            ></i>
          ) : (
            <i className="fas fa-square" style={{ opacity: 0.3 }}></i>
          )}
        </span>

        {/* Node Icon */}
        <span className="tree-icon">
          {nodeIcon ? (
            <i
              className={`fas ${nodeIcon}`}
              style={{ color: nodeColor || "#6c757d" }}
            ></i>
          ) : (
            <i
              className="fas fa-folder"
              style={{ color: nodeColor || "#ffc107" }}
            ></i>
          )}
        </span>

        {/* Node Name */}
        <span className="tree-label" style={{ color: nodeColor || "inherit" }}>
          {displayName}
        </span>
      </div>

      {/* Render children if expanded */}
      {isExpanded && children.length > 0 && (
        <div className="tree-children">
          {children.map((childName) => {
            const childPath =
              path === "root" ? `root/${childName}` : `${path}/${childName}`;
            return (
              <TreeNode
                key={childPath}
                path={childPath}
                name={childName}
                level={level + 1}
              />
            );
          })}
        </div>
      )}

      {/* Context Menu */}
      {contextMenu && (
        <NodeContextMenu
          path={path}
          nodeName={displayName}
          onClose={closeContextMenu}
          position={contextMenu}
        />
      )}
    </div>
  );
};

export default TreeNode;
