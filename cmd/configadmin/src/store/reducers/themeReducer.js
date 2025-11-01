import { createReducer, createAction } from "@reduxjs/toolkit";

// Actions
export const toggleTheme = createAction("theme/toggle");
export const setTheme = createAction("theme/set");

// Initial state - check localStorage for saved preference
const getInitialTheme = () => {
  const saved = localStorage.getItem("blueconfig-theme");
  if (saved) return saved;

  // Check system preference
  if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
    return "dark";
  }

  return "light";
};

const initialState = {
  mode: getInitialTheme(), // 'light' or 'dark'
};

// Theme reducer
const themeReducer = createReducer(initialState, (builder) => {
  builder
    .addCase(toggleTheme, (state) => {
      state.mode = state.mode === "light" ? "dark" : "light";
      localStorage.setItem("blueconfig-theme", state.mode);
      applyThemeToDocument(state.mode);
    })
    .addCase(setTheme, (state, action) => {
      state.mode = action.payload;
      localStorage.setItem("blueconfig-theme", state.mode);
      applyThemeToDocument(state.mode);
    });
});

// Helper to apply theme classes to document
const applyThemeToDocument = (mode) => {
  if (mode === "dark") {
    document.documentElement.setAttribute("data-theme", "dark");
    document.body.classList.add("dark-theme");
    document.body.classList.remove("light-theme");
  } else {
    document.documentElement.setAttribute("data-theme", "light");
    document.body.classList.add("light-theme");
    document.body.classList.remove("dark-theme");
  }
};

// Initialize theme on load
applyThemeToDocument(getInitialTheme());

export default themeReducer;
