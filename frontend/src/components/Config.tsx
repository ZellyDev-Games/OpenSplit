import {useEffect, useState} from "react";
import {KeyConfig} from "../models/keyInfo";
import {WindowSetSize} from "../../wailsjs/runtime";

export type ConfigParams = {
    hotkeys: KeyConfig[]
}

export default function Config({hotkeys}: ConfigParams) {
    const [hotkeyConfig, setHotkeyConfig] = useState<KeyConfig[]>(hotkeys);

    useEffect(() => {
        WindowSetSize(700, 800)
    }, [])

    return(
        <div className="container form-container">
            <h2>OpenSplit Configuration</h2>
            <div className="options">
                <h3>Hotkeys</h3>
                {
                    hotkeyConfig.map((config) => (
                        <div className="row">
                            <div className="hotkeyContainer">
                                <p className="hotkeyID">{config.command}: </p>
                                <p className="hotkeyValue">{config.info.code && config.info.name || "No Key Assignment"}</p>
                                <button>Record Hotkey</button>
                            </div>
                        </div>
                    ))
                }
            </div>
            <div className="actions">
                <button>Save</button>
                <button>Cancel</button>
            </div>
        </div>
    )
}
