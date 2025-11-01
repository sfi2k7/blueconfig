import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useDispatch } from "react-redux";
import api from "../../api/api";
import {
  clearAllTreeState,
  clearAllPropertiesState,
} from "../../store/actions/treeActions";

const LandingPage = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const [stores, setStores] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newStore, setNewStore] = useState({
    name: "",
    displayName: "",
    description: "",
    password: "",
  });
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    loadStores();
  }, []);

  const loadStores = async () => {
    try {
      setLoading(true);
      const response = await api.get("/stores");
      if (response.data.success) {
        setStores(response.data.data || []);
      } else {
        setError(response.data.error);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectStore = async (storeName, storeHasPassword = false) => {
    let password = "";

    // If store has password, prompt for it
    if (storeHasPassword) {
      password = prompt(`Enter password for "${storeName}":`);
      if (password === null) {
        // User cancelled
        return;
      }
    }

    try {
      const response = await api.post(`/stores/${storeName}/select`, {
        password: password,
      });
      if (response.data.success) {
        // Clear all existing tree and properties state
        dispatch(clearAllTreeState());
        dispatch(clearAllPropertiesState());

        // Navigate to main app
        navigate("/");
      } else {
        alert(`Failed to select store: ${response.data.error}`);
      }
    } catch (err) {
      if (err.response && err.response.status === 401) {
        alert("Invalid password. Please try again.");
      } else {
        alert(`Error selecting store: ${err.message}`);
      }
    }
  };

  const handleCreateStore = async (e) => {
    e.preventDefault();

    if (!newStore.name.trim()) {
      alert("Store name is required");
      return;
    }

    try {
      setCreating(true);
      const response = await api.post("/stores", {
        name: newStore.name,
        displayName: newStore.displayName || newStore.name,
        description: newStore.description,
        password: newStore.password,
      });

      if (response.data.success) {
        // Reload stores list
        await loadStores();
        // Reset form
        setNewStore({
          name: "",
          displayName: "",
          description: "",
          password: "",
        });
        setShowCreateForm(false);
        // Auto-select the new store
        await handleSelectStore(newStore.name);
      } else {
        alert(`Failed to create store: ${response.data.error}`);
      }
    } catch (err) {
      alert(`Error creating store: ${err.message}`);
    } finally {
      setCreating(false);
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return "Unknown";
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return "Today";
    if (diffDays === 1) return "Yesterday";
    if (diffDays < 7) return `${diffDays} days ago`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
    if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
    return `${Math.floor(diffDays / 365)} years ago`;
  };

  return (
    <div className="landing-page">
      {/* Background */}
      <div className="landing-background"></div>

      {/* Content */}
      <div className="landing-content">
        <div className="landing-card">
          {/* Header */}
          <div className="landing-header">
            <div className="landing-logo">
              <i className="fas fa-database"></i>
            </div>
            <h1 className="landing-title">BlueConfig Admin</h1>
            <p className="landing-subtitle">Select Configuration Store</p>
          </div>

          {/* Loading State */}
          {loading && (
            <div className="text-center py-5">
              <div className="spinner-border text-primary" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
              <p className="mt-3 text-muted">Loading stores...</p>
            </div>
          )}

          {/* Error State */}
          {error && (
            <div className="alert alert-danger" role="alert">
              <i className="fas fa-exclamation-triangle me-2"></i>
              {error}
            </div>
          )}

          {/* Stores List */}
          {!loading && !error && (
            <>
              <div className="stores-list">
                {stores.length === 0 ? (
                  <div className="text-center py-4 text-muted">
                    <i className="fas fa-inbox fa-3x mb-3"></i>
                    <p>No stores available. Create one to get started.</p>
                  </div>
                ) : (
                  stores.map((store) => (
                    <div
                      key={store.name}
                      className="store-item"
                      onClick={() =>
                        handleSelectStore(store.name, store.hasPassword)
                      }
                    >
                      <div className="store-icon">
                        <i
                          className={`fas ${store.icon || "fa-database"}`}
                          style={{ color: store.color || "#3498db" }}
                        ></i>
                      </div>
                      <div className="store-info">
                        <div className="store-name">{store.displayName}</div>
                        {store.description && (
                          <div className="store-description">
                            {store.description}
                          </div>
                        )}
                        <div className="store-meta">
                          <small className="text-muted">
                            Created {formatDate(store.createdAt)}
                          </small>
                          {store.hasPassword && (
                            <small className="ms-2">
                              <i className="fas fa-lock"></i> Protected
                            </small>
                          )}
                        </div>
                      </div>
                      <div className="store-arrow">
                        <i className="fas fa-chevron-right"></i>
                      </div>
                    </div>
                  ))
                )}
              </div>

              {/* Create Store Button */}
              {!showCreateForm && (
                <div className="landing-footer">
                  <button
                    className="btn btn-outline-primary btn-lg w-100"
                    onClick={() => setShowCreateForm(true)}
                  >
                    <i className="fas fa-plus me-2"></i>
                    Create New Store
                  </button>
                </div>
              )}

              {/* Create Store Form */}
              {showCreateForm && (
                <div className="create-store-form">
                  <h5 className="mb-3">
                    <i className="fas fa-plus-circle me-2"></i>
                    Create New Store
                  </h5>
                  <form onSubmit={handleCreateStore}>
                    <div className="mb-3">
                      <label className="form-label">Store Name *</label>
                      <input
                        type="text"
                        className="form-control"
                        placeholder="e.g., production, development"
                        value={newStore.name}
                        onChange={(e) =>
                          setNewStore({ ...newStore, name: e.target.value })
                        }
                        required
                        autoFocus
                      />
                      <small className="text-muted">
                        Lowercase, no spaces (e.g., production)
                      </small>
                    </div>

                    <div className="mb-3">
                      <label className="form-label">Display Name</label>
                      <input
                        type="text"
                        className="form-control"
                        placeholder="e.g., Production Configuration"
                        value={newStore.displayName}
                        onChange={(e) =>
                          setNewStore({
                            ...newStore,
                            displayName: e.target.value,
                          })
                        }
                      />
                    </div>

                    <div className="mb-3">
                      <label className="form-label">Description</label>
                      <textarea
                        className="form-control"
                        rows="2"
                        placeholder="Optional description"
                        value={newStore.description}
                        onChange={(e) =>
                          setNewStore({
                            ...newStore,
                            description: e.target.value,
                          })
                        }
                      />
                    </div>

                    <div className="mb-3">
                      <label className="form-label">Password (Optional)</label>
                      <input
                        type="password"
                        className="form-control"
                        placeholder="Leave empty for no password"
                        value={newStore.password}
                        onChange={(e) =>
                          setNewStore({ ...newStore, password: e.target.value })
                        }
                      />
                      <small className="text-muted">
                        Password can be changed later in __storeinfo properties
                      </small>
                    </div>

                    <div className="d-flex gap-2">
                      <button
                        type="submit"
                        className="btn btn-primary flex-grow-1"
                        disabled={creating}
                      >
                        {creating ? (
                          <>
                            <span
                              className="spinner-border spinner-border-sm me-2"
                              role="status"
                            ></span>
                            Creating...
                          </>
                        ) : (
                          <>
                            <i className="fas fa-check me-2"></i>
                            Create Store
                          </>
                        )}
                      </button>
                      <button
                        type="button"
                        className="btn btn-secondary"
                        onClick={() => {
                          setShowCreateForm(false);
                          setNewStore({
                            name: "",
                            displayName: "",
                            description: "",
                            password: "",
                          });
                        }}
                        disabled={creating}
                      >
                        Cancel
                      </button>
                    </div>
                  </form>
                </div>
              )}
            </>
          )}
        </div>

        {/* Footer */}
        <div className="landing-page-footer">
          <small className="text-muted">
            BlueConfig Admin v1.0.0 &copy; 2024
          </small>
        </div>
      </div>
    </div>
  );
};

export default LandingPage;
