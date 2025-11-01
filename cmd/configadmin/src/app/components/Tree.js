import React, { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import TreeNode from "./TreeNode";
import {
  fetchNodeChildren,
  updateSpecialProperty,
} from "../../store/actions/treeActions";
import api from "../../api/api";

const Tree = () => {
  const dispatch = useDispatch();

  // Single useSelector call for all tree state
  const treeState = useSelector((state) => state.tree);

  const { nodes, error, specialProperties } = treeState;
  const rootNode = nodes.root;

  // Load root children on mount
  useEffect(() => {
    if (!rootNode.loaded) {
      dispatch(fetchNodeChildren("root"));
    }
  }, [dispatch, rootNode.loaded]);

  // Load special properties for root node itself
  useEffect(() => {
    const loadRootSpecialProperties = async () => {
      if (!specialProperties.root) {
        try {
          const response = await api.get("/properties/root");
          if (response.data.success && response.data.data.properties) {
            const props = response.data.data.properties;

            // Update each special property individually
            if (props.__title) {
              dispatch(
                updateSpecialProperty({
                  path: "root",
                  key: "__title",
                  value: props.__title,
                }),
              );
            }
            if (props.__icon) {
              dispatch(
                updateSpecialProperty({
                  path: "root",
                  key: "__icon",
                  value: props.__icon,
                }),
              );
            }
            if (props.__color) {
              dispatch(
                updateSpecialProperty({
                  path: "root",
                  key: "__color",
                  value: props.__color,
                }),
              );
            }
          }
        } catch (err) {
          console.warn("Failed to fetch root special properties:", err);
        }
      }
    };
    loadRootSpecialProperties();
  }, [dispatch, specialProperties.root]);

  return (
    <div className="tree-container">
      <div className="tree-header">
        <h5>
          <i className="fas fa-sitemap me-2"></i>
          Configuration Tree
        </h5>
      </div>

      <div className="tree-body">
        {error && (
          <div className="alert alert-danger alert-sm" role="alert">
            <i className="fas fa-exclamation-triangle me-2"></i>
            {error}
          </div>
        )}

        <TreeNode path="root" name="root" level={0} />
      </div>
    </div>
  );
};

export default Tree;
