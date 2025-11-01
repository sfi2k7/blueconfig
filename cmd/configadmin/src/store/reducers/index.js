import { combineReducers } from "@reduxjs/toolkit";
import treeReducer from "./treeReducer";
import propertiesReducer from "./propertiesReducer";
import themeReducer from "./themeReducer";

const rootReducer = combineReducers({
  tree: treeReducer,
  properties: propertiesReducer,
  theme: themeReducer,
});

export default rootReducer;
