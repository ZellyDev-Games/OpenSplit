import { ContextMenu } from "../ContextMenu";
import SplitList from "./SplitList";
import Timer from "./Timer";
import React, { useEffect } from "react";
import { MenuItem, useContextMenu } from "../../hooks/useContextMenu";
import { useNavigate } from "react-router";
import { Quit, WindowGetPosition, WindowSetPosition } from "../../../wailsjs/runtime";
import { CloseSplitFile, LoadSplitFile } from "../../../wailsjs/go/session/Service";

export default function Splitter() {
    const navigate = useNavigate();
    const contextMenu = useContextMenu();
    const contextMenuItems: MenuItem[] = [
        {
            label: "Edit Splits",
            onClick: async () => {
                await updateWindowPos();
                navigate("/edit");
            },
        },
        {
            label: "Load Splits",
            onClick: async () => {
                await LoadSplitFile();
            },
        },
        {
            label: "Close Split File",
            onClick: async () => {
                await CloseSplitFile();
            },
        },
        {
            type: "separator",
        },
        {
            label: "Save Session",
            onClick: () => {},
        },
        {
            type: "separator",
        },
        {
            label: "Exit OpenSplit",
            onClick: () => Quit(),
        },
    ];

    const updateWindowPos = async () => {
        const pos = await WindowGetPosition();
        localStorage.setItem("window-pos-x", String(pos.x));
        localStorage.setItem("window-pos-y", String(pos.y));
        console.log(`set window position to ${pos.x} - ${pos.y}`);
    };

    const applyWindowPos = () => {
        const x = Number(localStorage.getItem("window-pos-x"));
        const y = Number(localStorage.getItem("window-pos-y"));
        console.log(`applying window position ${x} - ${y}`);
        WindowSetPosition(x == 0 ? 200 : x, y == 0 ? 200 : y);
    };

    useEffect(() => {
        applyWindowPos();
    }, []);

    return (
        <div {...contextMenu.bind} className="splitter">
            <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
            <SplitList />
            <Timer />
        </div>
    );
}
