import React from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Link,
  useNavigate,
} from "react-router-dom";
import { Provider, useDispatch, useSelector } from "react-redux";
import store from "./store";
import HomePage from "./app/pages/HomePage";
import LandingPage from "./app/pages/LandingPage";
import ProtectedRoute from "./app/components/ProtectedRoute";
import { toggleTheme } from "./store/reducers/themeReducer";
import {
  clearAllTreeState,
  clearAllPropertiesState,
} from "./store/actions/treeActions";
import api from "./api/api";

// Navigation Component
const Navigation = ({ showLogout = true }) => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const themeState = useSelector((state) => state.theme);
  const isDark = themeState.mode === "dark";

  const handleThemeToggle = () => {
    dispatch(toggleTheme());
  };

  const handleLogout = async () => {
    try {
      // Clear Redux state
      dispatch(clearAllTreeState());
      dispatch(clearAllPropertiesState());

      // Clear cookie
      await api.post("/logout");
      navigate("/landing");
    } catch (err) {
      console.error("Logout failed:", err);
      // Clear state even on error
      dispatch(clearAllTreeState());
      dispatch(clearAllPropertiesState());
      navigate("/landing");
    }
  };

  return (
    <nav className="navbar navbar-expand-lg navbar-dark bg-dark">
      <div className="container-fluid">
        <Link className="navbar-brand" to="/">
          <i className="fas fa-database me-2"></i>
          BlueConfig Admin
        </Link>
        <button
          className="navbar-toggler"
          type="button"
          data-bs-toggle="collapse"
          data-bs-target="#navbarNav"
          aria-controls="navbarNav"
          aria-expanded="false"
          aria-label="Toggle navigation"
        >
          <span className="navbar-toggler-icon"></span>
        </button>
        <div className="collapse navbar-collapse" id="navbarNav">
          <ul className="navbar-nav ms-auto">
            <li className="nav-item">
              <Link className="nav-link" to="/">
                <i className="fas fa-home me-1"></i>
                Browser
              </Link>
            </li>
            {showLogout && (
              <li className="nav-item">
                <button
                  className="nav-link"
                  onClick={handleLogout}
                  style={{ background: "none", border: "none" }}
                >
                  <i className="fas fa-sign-out-alt me-1"></i>
                  Logout
                </button>
              </li>
            )}
            <li className="nav-item">
              <button
                className="theme-toggle-btn nav-link"
                onClick={handleThemeToggle}
                title={isDark ? "Switch to Light Mode" : "Switch to Dark Mode"}
              >
                <i className={`fas ${isDark ? "fa-sun" : "fa-moon"}`}></i>
              </button>
            </li>
          </ul>
        </div>
      </div>
    </nav>
  );
};

// Main App Component
const App = () => {
  return (
    <Provider store={store}>
      <Router>
        <Routes>
          <Route path="/landing" element={<LandingPage />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <div className="d-flex flex-column vh-100">
                  <Navigation showLogout={true} />
                  <div className="flex-grow-1 overflow-hidden">
                    <HomePage />
                  </div>
                </div>
              </ProtectedRoute>
            }
          />
          <Route
            path="/:path/*"
            element={
              <ProtectedRoute>
                <div className="d-flex flex-column vh-100">
                  <Navigation showLogout={true} />
                  <div className="flex-grow-1 overflow-hidden">
                    <HomePage />
                  </div>
                </div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </Router>
    </Provider>
  );
};

export default App;
