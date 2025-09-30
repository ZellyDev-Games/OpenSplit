import { useEffect, useState } from "react";

import {EventsOn, WindowSetSize} from "../../wailsjs/runtime";
import {ConfigPayload} from "../models/configPayload";
import {Dispatch} from "../../wailsjs/go/dispatcher/Service";
import {Command} from "../App";

export type ConfigParams = {
    configPayload: ConfigPayload;
};

const RECORDING_ARMED = 10

export default function Config({ configPayload }: ConfigParams) {
    const [recording, setRecording] = useState(false);
    const [config, setConfig] = useState<ConfigPayload>(configPayload);

    useEffect(() => {
        WindowSetSize(700, 800);
        return EventsOn("config:update", (newConfigPayload) => {
            console.log("received update from backend", newConfigPayload);
            setConfig(newConfigPayload);
            setRecording(false);
        })
    }, []);

    const armHotkey = async (command: Command) => {
        const reply = await Dispatch(command, null)
        if (reply.code == RECORDING_ARMED) {
            console.log("backend confirms recording is armed")
            setRecording(true);
        }
    }

    return (
        <div className="container form-container">
            <h2>OpenSplit Configuration</h2>
            <div className="options">
                <h3>Hotkeys</h3>
                    <div className="row">
                        <div className="hotkeyContainer">
                            <p className="hotkeyID">Split: </p>
                            <p className="hotkeyValue">
                                {(config.key_config[Command.SPLIT].key_code && config.key_config[Command.SPLIT].locale_name) || "No Key Assignment"}
                            </p>
                            <button disabled={recording} onClick={() => armHotkey(Command.SPLIT)}>Record Hotkey</button>
                        </div>
                    </div>
            </div>
            <div className="actions">
                <button onClick={() => Dispatch(Command.SUBMIT, JSON.stringify(config))}>Save</button>
                <button onClick={() => Dispatch(Command.CANCEL, null)}>Cancel</button>
            </div>
        </div>
    );
}
