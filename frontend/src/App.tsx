import Timer from "./components/Timer";
import {MenuItem, useContextMenu} from "./hooks/useContextMenu.js";
import {ContextMenu} from "./components/ContextMenu";
import {Quit} from "../wailsjs/runtime";
import {Route, Routes, useNavigate} from "react-router";
import React, {useEffect} from "react";
import SplitEditor from "./components/SplitEditor";
import {setActiveSkin} from "./skinLoader";

function App() {
    const navigate = useNavigate()
    const contextMenu = useContextMenu()
    const contextMenuItems: MenuItem[] = [
        {
            label: "Reload skins", onClick: () => {setActiveSkin("default")}
        },
        {
            label: "Edit Splits", onClick: () => navigate("/edit"),
        },{
            label: "Load Splits", onClick: () => console.log("Load Splits"),
        },{
            type: "separator",
        }, {
            label: "Save Session", onClick: () => console.log("Save Session")
        }, {
            type: "separator",
        }, {
            label: "Exit OpenSplit", onClick: () => Quit(),
        }
    ]

    useEffect(() => {
        setActiveSkin("default")
    })

    return (
        <div { ...contextMenu.bind } id="App" className="app">
                <Routes>
                    <Route path="/" element={
                        <>
                            <Timer />
                            <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
                        </>
                    }/>
                    <Route path="/edit" element={<SplitEditor />}/>
                </Routes>
        </div>
    )
}

export default App
