import {ContextMenu} from "../ContextMenu";
import SplitList from "./SplitList";
import Timer from "./Timer";
import React from "react";
import {MenuItem, useContextMenu} from "../../hooks/useContextMenu";
import {useNavigate} from "react-router";
import {Quit} from "../../../wailsjs/runtime";
import {CloseSplitFile, LoadSplitFile} from "../../../wailsjs/go/session/Service";

export default function Splitter() {
    const navigate = useNavigate();
    const contextMenu = useContextMenu()
    const contextMenuItems: MenuItem[] = [
        {
            label: "Edit Splits", onClick: () =>  {
                navigate("/edit")
            },
        },{
            label: "Load Splits", onClick: async () => {
                await LoadSplitFile()
            }
        },{
            label: "Close Split File", onClick: async () => {
                await CloseSplitFile()
            }
        },{
            type: "separator",
        }, {
            label: "Save Session", onClick: () => {}
        }, {
            type: "separator",
        }, {
            label: "Exit OpenSplit", onClick: () => Quit(),
        }
    ]
    return (
        <div { ...contextMenu.bind } className="splitter">
            <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
            <SplitList />
            <Timer />
        </div>)
}