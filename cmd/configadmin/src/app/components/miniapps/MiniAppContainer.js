import React from 'react';
import DefaultProperties from './DefaultProperties';

// Mini-app registry
const MINIAPP_REGISTRY = {
  'default': DefaultProperties,
  'timeseries': React.lazy(() => import('./timeseries')),
};

const MiniAppContainer = ({ nodeType, currentPath, properties }) => {
  // Determine which mini-app to load based on node type
  const appType = nodeType || 'default';
  const MiniAppComponent = MINIAPP_REGISTRY[appType] || MINIAPP_REGISTRY.default;

  return (
    <div className="miniapp-container h-100">
      <React.Suspense fallback={
        <div className="d-flex justify-content-center align-items-center h-100">
          <div className="spinner-border text-primary" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
        </div>
      }>
        <MiniAppComponent
          currentPath={currentPath}
          properties={properties}
        />
      </React.Suspense>
    </div>
  );
};

export default MiniAppContainer;
