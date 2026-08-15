import { useState } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { SplitterLayout } from "../../hooks/splitter/useSplitterMenu";
import { Command } from "../../models/command";

export default function NewSkin() {
    const [name, setName] = useState("");
    const [layout, setLayout] = useState<SplitterLayout>("vertical");

    async function create() {
        const skinName = name.trim();

        if (!skinName) {
            return;
        }

        await Dispatch(
            Command.SUBMIT,
            JSON.stringify({
                name: skinName,
                layout,
            }),
        );
    }

    async function cancel() {
        await Dispatch(Command.CANCEL, null);
    }

    return (
        <div className="skin-creator">
            <h2>New Skin</h2>

            <label>Skin Name</label>

            <input value={name} placeholder="my-skin" onChange={(e) => setName(e.target.value)} />

            <label>Open Editor As</label>

            <div className="radio-row">
                <label>
                    <input
                        type="radio"
                        name="new-skin-layout"
                        value="vertical"
                        checked={layout === "vertical"}
                        onChange={() => setLayout("vertical")}
                    />
                    Vertical
                </label>

                <label>
                    <input
                        type="radio"
                        name="new-skin-layout"
                        value="horizontal"
                        checked={layout === "horizontal"}
                        onChange={() => setLayout("horizontal")}
                    />
                    Horizontal
                </label>
            </div>

            <div className="skin-editor-buttons">
                <button disabled={!name.trim()} onClick={create}>
                    Create Skin
                </button>

                <button onClick={cancel}>Cancel</button>
            </div>
        </div>
    );
}
