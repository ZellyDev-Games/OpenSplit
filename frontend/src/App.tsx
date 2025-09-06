import Timer from "./components/Timer";
import {MenuItem, useContextMenu} from "./hooks/useContextMenu";
import {ContextMenu} from "./components/ContextMenu";
import {Quit, Size, WindowGetSize, WindowSetSize} from "../wailsjs/runtime";
import {Route, Routes, useNavigate} from "react-router";
import React, {useCallback, useEffect, useMemo, useState} from "react";
import SplitEditor from "./components/SplitEditor";
import {setActiveSkin} from "./skinLoader";
import {LoadSplitFile} from "../wailsjs/go/session/Service";
import SplitList from "./components/SplitList";
import {session} from "../wailsjs/go/models";
import SplitFilePayload = session.SplitFilePayload;
import useWindowResize from "./hooks/useWindowResize";

function App() {
    const navigate = useNavigate()
    const contextMenu = useContextMenu()
    const [loadedSplitFile, setLoadedSplitFile] = useState<SplitFilePayload | null>(null)

    const contextMenuItems: MenuItem[] = [
        {
            label: "Reload skins", onClick: () => {setActiveSkin("default")}
        },
        {
            label: "Edit Splits", onClick: () =>  {
                navigate("/edit")
            },
        },{
            label: "Load Splits", onClick: async () => {
                await loadSplitFile()
            },
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

    const loadSplitFile = async() => {
        LoadSplitFile().then(data => setLoadedSplitFile(data)).catch(e => console.log(e));
    }

    return (
        <div { ...contextMenu.bind } id="App" className="app">
                <Routes>
                    <Route path="/" element={
                        <div className="splitter">
                            <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
                            <Timer />
                            <SplitList splitFile={loadedSplitFile} />
                        </div>
                    }/>
                    <Route path="/edit" element={<SplitEditor loadedSplitFile={loadedSplitFile} />}/>
                </Routes>
        </div>
    )
}

export default App
