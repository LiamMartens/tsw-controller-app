import "./app.css";
import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";

const app = document.getElementById("app")!;

const root = ReactDOM.createRoot(app);

root.render(React.createElement(App));
