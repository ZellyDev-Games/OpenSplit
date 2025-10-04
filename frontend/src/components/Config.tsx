import { useEffect, useState } from "react";

import { Dispatch } from "../../wailsjs/go/dispatcher/Service";
import { EventsOn, WindowSetSize } from "../../wailsjs/runtime";
import { Command } from "../App";
import { ConfigPayload } from "../models/configPayload";

export type ConfigParams = {
    configPayload: ConfigPayload;
};

const RECORDING_ARMED = 10;

export default function Config({ configPayload }: ConfigParams) {
    const [recording, setRecording] = useState(false);
    const [config, setConfig] = useState<ConfigPayload>(configPayload);

    useEffect(() => {
        WindowSetSize(700, 800);
        return EventsOn("config:update", (newConfigPayload) => {
            console.log("received update from backend", newConfigPayload);
            setConfig(newConfigPayload);
            setRecording(false);
        });
    }, []);

    const armHotkey = async (command: Command) => {
        const reply = await Dispatch(command, null);
        if (reply.code == RECORDING_ARMED) {
            console.log("backend confirms recording is armed");
            setRecording(true);
        }
    };

    const displayHotkeyRows = () => {
        const commands: [Command, string][] = [
            [Command.SPLIT, "Split"],
            [Command.UNDO, "Undo Split"],
            [Command.SKIP, "Skip Split"],
            [Command.PAUSE, "Pause Run"],
            [Command.RESET, "Reset Run"],
        ];

        return commands.map((command: [Command, string]) => (
            <div className="row" key={command[0]}>
                <div className="hotkeyContainer">
                    <p className="hotkeyID">{command[1]}: </p>
                    <p className="hotkeyValue">
                        {(config.key_config[command[0]].key_code && config.key_config[command[0]].locale_name) ||
                            "No Key Assignment"}
                    </p>
                    <button disabled={recording} onClick={() => armHotkey(command[0])}>
                        {(recording && "Recording") || "Record Hotkey"}
                    </button>
                </div>
            </div>
        ));
    };

    return (
        <div className="container form-container">
            <h2>OpenSplit Configuration</h2>
            <div className="options">
                <h3>Hotkeys</h3>
                {displayHotkeyRows()}
            </div>
            <div className="actions">
                <button onClick={() => Dispatch(Command.SUBMIT, JSON.stringify(config))}>Save</button>
                <button onClick={() => Dispatch(Command.CANCEL, null)}>Cancel</button>
            </div>
        </div>
    );
}
