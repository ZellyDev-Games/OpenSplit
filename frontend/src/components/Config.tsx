import { useEffect, useState } from "react";

import { Dispatch } from "../../wailsjs/go/dispatcher/Service";
import { EventsOn, WindowSetSize } from "../../wailsjs/runtime";
import { Command } from "../App";
import {ConfigPayload, KeyInfo} from "../models/configPayload";

export type ConfigParams = {
    configPayload: ConfigPayload;
};

const RECORDING_ARMED = 10;

export default function Config({ configPayload }: ConfigParams) {
    const [recording, setRecording] = useState(false);
    const [config, setConfig] = useState<ConfigPayload>(configPayload);

    useEffect(() => {
        WindowSetSize(700, 800);
        return EventsOn("config:update", (newConfigPayload: ConfigPayload) => {
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

    const getHotkeyName = (ki: KeyInfo): string => {
        console.log(ki)
        let ret =  (ki.modifiers !== null && ki.modifiers.length > 0) && ki.modifier_locale_names.join(" + ") + " + " || ""
        ret += ki.locale_name && ki.locale_name || ""
        if (ret === "") {
            return "No Hotkey Assigned"
        } else {
            return ret
        }
    }

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
                        {getHotkeyName(config.key_config[command[0]])}
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
