import Timer from "./components/Timer";
import {MenuItem, useContextMenu} from "./hooks/useContextMenu.js";
import {ContextMenu} from "./components/ContextMenu";
import {Quit} from "../wailsjs/runtime";

function App() {
    const contextMenu = useContextMenu()
    const contextMenuItems: MenuItem[] = [
        {
            label: "Edit Splits", onClick: () => console.log("Edit Splits"),
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

    return (
        <div { ...contextMenu.bind } id="App" className="app">
            <Timer />
            <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
        </div>
    )
}

export default App
