import React, { useState } from "react";
import Tree from "../components/Tree";
import RightPanel from "../components/RightPanel";

const HomePage = () => {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  const toggleSidebar = () => {
    setSidebarCollapsed(!sidebarCollapsed);
  };

  return (
    <div className="d-flex h-100 position-relative">
      {/* Left Side - Tree (Collapsible Sidebar) */}
      <div
        className="bg-light border-end"
        style={{
          width: sidebarCollapsed ? "0px" : "300px",
          minWidth: sidebarCollapsed ? "0px" : "300px",
          maxWidth: sidebarCollapsed ? "0px" : "300px",
          height: "calc(100vh - 56px)",
          overflowY: "auto",
          overflowX: "hidden",
          padding: sidebarCollapsed ? "0" : "1rem",
          transition: "all 0.3s ease-in-out",
        }}
      >
        {!sidebarCollapsed && <Tree />}
      </div>

      {/* Toggle Button */}
      <button
        className="btn btn-primary btn-sm shadow-sm position-absolute"
        onClick={toggleSidebar}
        title={sidebarCollapsed ? "Expand Sidebar" : "Collapse Sidebar"}
        style={{
          left: sidebarCollapsed ? "10px" : "290px",
          top: "50%",
          transform: "translateY(-50%)",
          zIndex: 1050,
          width: "40px",
          height: "40px",
          borderRadius: "50%",
          padding: "0",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          border: "2px solid #fff",
          transition: "left 0.3s ease-in-out",
        }}
      >
        <i
          className={`fas ${
            sidebarCollapsed ? "fa-chevron-right" : "fa-chevron-left"
          }`}
          style={{ fontSize: "14px" }}
        ></i>
      </button>

      {/* Right Side - Content Area (Always Visible) */}
      <div
        className="flex-grow-1 p-3"
        style={{
          height: "calc(100vh - 56px)",
          overflowY: "auto",
          transition: "all 0.3s ease-in-out",
        }}
      >
        <RightPanel />
      </div>
    </div>
  );
};

export default HomePage;
