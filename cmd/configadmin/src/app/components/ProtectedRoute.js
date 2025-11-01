import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import api from "../../api/api";

const ProtectedRoute = ({ children }) => {
  const navigate = useNavigate();
  const [checking, setChecking] = useState(true);
  const [hasStore, setHasStore] = useState(false);

  useEffect(() => {
    checkStoreSelection();
  }, []);

  const checkStoreSelection = async () => {
    try {
      const response = await api.get("/current-store");
      if (response.data.success && response.data.data) {
        // Store is selected
        setHasStore(true);
      } else {
        // No store selected, redirect to landing
        navigate("/landing");
      }
    } catch (err) {
      // Error or no store, redirect to landing
      navigate("/landing");
    } finally {
      setChecking(false);
    }
  };

  if (checking) {
    return (
      <div
        className="d-flex align-items-center justify-content-center"
        style={{ height: "100vh" }}
      >
        <div className="text-center">
          <div className="spinner-border text-primary mb-3" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
          <p className="text-muted">Checking store selection...</p>
        </div>
      </div>
    );
  }

  return hasStore ? children : null;
};

export default ProtectedRoute;
