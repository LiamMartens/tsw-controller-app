import "./style.css";
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { StatusBar } from "@capacitor/status-bar";

StatusBar.setOverlaysWebView({ overlay: false });

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
