import Timer from "./components/Timer";
import {MenuItem, useContextMenu} from "./hooks/useContextMenu";
import {ContextMenu} from "./components/ContextMenu";
import {Quit} from "../wailsjs/runtime";
import {Route, Routes, useNavigate} from "react-router";
import React, {useState} from "react";
import SplitEditor from "./components/SplitEditor";
import {LoadSplitFile} from "../wailsjs/go/session/Service";
import SplitList, {CompareAgainst} from "./components/SplitList";
import {session} from "../wailsjs/go/models";

function App() {
    const navigate = useNavigate()
    const contextMenu = useContextMenu()
    const [loadedSplitFile, setLoadedSplitFile] = useState<session.SplitFilePayload | null>(null)
    const [compareAgainst, setCompareAgainst] = useState<CompareAgainst>("best")

    const contextMenuItems: MenuItem[] = [
        {
            label: "Close SplitFile", onClick: () => {setLoadedSplitFile(null)}
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
        LoadSplitFile().then(data => {
            console.log(data);
            if(data === null) return;
            setLoadedSplitFile(data);
        }).catch(e => console.log(e));
    }

    return (
        <div { ...contextMenu.bind } id="App" className="app">
            <Routes>
                <Route path="/" element={
                    <div className="splitter">
                        <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
                        <SplitList compareAgainst={compareAgainst} splitFile={loadedSplitFile} />
                        <Timer />
                    </div>
                }/>
                <Route path="/edit" element={<SplitEditor setLoadedSplitFile={setLoadedSplitFile} loadedSplitFile={loadedSplitFile} />}/>
            </Routes>
        </div>
    )
}

export default App