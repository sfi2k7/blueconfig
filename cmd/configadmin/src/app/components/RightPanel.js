import React, { useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { fetchProperties } from "../../store/actions/treeActions";
import MiniAppContainer from "./miniapps/MiniAppContainer";

const RightPanel = () => {
  const dispatch = useDispatch();
  const currentPath = useSelector((state) => state.tree.currentPath);
  const properties = useSelector((state) => state.properties.properties);
  const loading = useSelector((state) => state.properties.loading);

  // Fetch properties when current path changes
  useEffect(() => {
    if (currentPath) {
      dispatch(fetchProperties(currentPath));
    }
  }, [currentPath, dispatch]);

  // Determine node type from properties
  const nodeType = properties?.__type || "default";

  if (!currentPath) {
    return (
      <div className="d-flex justify-content-center align-items-center h-100">
        <div className="alert alert-secondary">
          <i className="fas fa-arrow-left me-2"></i>
          Select a node from the tree to view its content
        </div>
      </div>
    );
  }

  if (loading && !properties) {
    return (
      <div className="d-flex justify-content-center align-items-center h-100">
        <div className="spinner-border text-primary" role="status">
          <span className="visually-hidden">Loading...</span>
        </div>
      </div>
    );
  }

  return (
    <MiniAppContainer
      nodeType={nodeType}
      currentPath={currentPath}
      properties={properties}
    />
  );
};

export default RightPanel;
