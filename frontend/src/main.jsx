import "./styles/index.css";

import React from "react";
import { createRoot } from "react-dom/client";
import { HashRouter } from "react-router";

import App from "./App";

const container = document.getElementById("root");

const root = createRoot(container);

root.render(
    <React.StrictMode>
        <HashRouter>
            <App />
        </HashRouter>
    </React.StrictMode>,
);
