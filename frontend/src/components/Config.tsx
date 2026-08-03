import { useEffect, useState } from "react";

import { Dispatch, OpenSkinsFolder, OpenSplitFileFolder } from "../../wailsjs/go/dispatcher/Service";
import { GetAvailableSkins, SetSkin } from "../../wailsjs/go/skin/Service";
import { EventsOn } from "../../wailsjs/runtime";
import { Command } from "../models/command";
import { ConfigPayload, KeyInfo } from "../models/configPayload";
import { log } from "../utils/logger";

export type ConfigParams = {
    configPayload: ConfigPayload;
};

const RECORDING_ARMED = 10;
const DEFAULT_ROLLING_AVG = 20;

export default function Config({ configPayload }: ConfigParams) {
    const [recording, setRecording] = useState(false);
    const [config, setConfig] = useState<ConfigPayload>(configPayload);
    const [availableSkins, setAvailableSkins] = useState<Array<string>>([]);
    const [rollingAvg, setRollingAvg] = useState<number>(configPayload.rolling_average_runs ?? DEFAULT_ROLLING_AVG);

    useEffect(() => {
        const loadSkins = async () => {
            const skins = await GetAvailableSkins();
            log.debug("Loaded skins", skins);
            setAvailableSkins(skins);
        };

        void loadSkins();
    }, []);

    useEffect(() => {
        const offHotkey = EventsOn("config:hotkey-recorded", (evt: { command: Command; key_info: KeyInfo }) => {
            setRecording(false);

            setConfig((prev) => ({
                ...prev,
                key_config: {
                    ...(prev.key_config ?? {}),
                    [evt.command]: evt.key_info,
                },
            }));
        });

        return () => {
            offHotkey();
        };
    }, []);

    useEffect(() => {
        setRollingAvg(config.rolling_average_runs ?? DEFAULT_ROLLING_AVG);
    }, [config]);

    useEffect(() => {
        void SetSkin(config.selected_skin, true);
    }, [config.selected_skin]);

    const hasHotkey = (ki?: KeyInfo) => {
        return !!ki && !!ki.locale_name;
    };

    // backend records the next keypress
    const armHotkey = async (command: Command) => {
        const reply = await Dispatch(command, null);
        if (reply.code == RECORDING_ARMED) {
            log.debug("Hotkey recording armed");
            setRecording(true);
        }
    };

    const clearHotkey = (command: Command) => {
        setConfig((prev) => {
            const keyConfig = { ...(prev.key_config ?? {}) };
            delete keyConfig[command];

            return {
                ...prev,
                key_config: keyConfig,
            };
        });
    };

    const getHotkeyName = (ki?: KeyInfo): string => {
        if (!ki) {
            return "No Hotkey Assigned";
        }

        const modifiers = ki.modifier_locale_names?.length > 0 ? ki.modifier_locale_names.join(" + ") + " + " : "";

        const key = ki.locale_name ?? "";

        const ret = modifiers + key;

        return ret.trim() === "" ? "No Hotkey Assigned" : ret;
    };

    // generated dynamically
    const displayHotkeyRows = () => {
        const commands: [Command, string][] = [
            [Command.SPLIT, "Split"],
            [Command.UNDO, "Undo Split"],
            [Command.SKIP, "Skip Split"],
            [Command.PAUSE, "Pause Run"],
            [Command.RESET, "Reset Run"],
            [Command.COMPARISON_LEFT, "Previous Comparison"],
            [Command.COMPARISON_RIGHT, "Next Comparison"],
        ];

        return commands.map(([command, label]) => (
            <div className="hotkey-row" key={command}>
                <div className="row hotkeyContainer">
                    <p className="hotkeyID">{label}:</p>
                    <p className="hotkeyValue">{getHotkeyName(config.key_config?.[command])}</p>

                    <button disabled={recording} onClick={() => armHotkey(command)}>
                        {recording ? "Recording" : "Record Hotkey"}
                    </button>

                    <button
                        disabled={recording || !hasHotkey(config.key_config?.[command])}
                        onClick={() => clearHotkey(command)}
                    >
                        Clear
                    </button>
                </div>
            </div>
        ));
    };

    return (
        <div className="container form-container options-container">
            <h2>OpenSplit Configuration</h2>
            <div id="options">
                <h3>Hotkeys</h3>
                {displayHotkeyRows()}
            </div>

            <hr />

            <div className="config-section config-skins">
                <h3>Active Skin</h3>
                <select
                    className="config-skin-select"
                    id="selectedSkin"
                    value={config.selected_skin}
                    onChange={(e) =>
                        setConfig((prev) => ({
                            ...prev,
                            selected_skin: e.target.value,
                        }))
                    }
                >
                    {availableSkins.map((s) => (
                        <option value={s} key={s}>
                            {s}
                        </option>
                    ))}
                </select>
            </div>

            <hr />

            <div id="rollingAvg">
                <h3>Rolling Average Window</h3>

                <select
                    value={rollingAvg}
                    onChange={(e) => {
                        const v = parseInt(e.target.value);

                        setRollingAvg(v);

                        setConfig((prev) => ({
                            ...prev,
                            rolling_average_runs: v,
                        }));
                    }}
                >
                    {[5, 10, 20, 50, 100].map((n) => (
                        <option key={n} value={n}>
                            Last {n} Runs
                        </option>
                    ))}
                </select>
            </div>

            <hr />

            <div id="directories">
                <h3>Directories</h3>
                <button onClick={OpenSplitFileFolder}>Open Splitfile Folder</button>
                <button onClick={OpenSkinsFolder}>Open Skins Folder</button>
            </div>

            <hr />

            <div className="actions">
                <button onClick={() => Dispatch(Command.SUBMIT, JSON.stringify(config))}>Save</button>
                <button onClick={() => Dispatch(Command.CANCEL, null)}>Cancel</button>
            </div>
        </div>
    );
}
