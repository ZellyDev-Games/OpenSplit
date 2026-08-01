import { useState } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { Command } from "../../models/command";

export default function NewSkin() {
    const [name, setName] = useState("");

    async function create() {
        const skinName = name.trim();

        if (!skinName) {
            return;
        }

        await Dispatch(
            Command.SUBMIT,
            JSON.stringify({
                name: skinName,
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

            <div className="skin-editor-buttons">
                <button disabled={!name.trim()} onClick={create}>
                    Create Skin
                </button>

                <button onClick={cancel}>Cancel</button>
            </div>
        </div>
    );
}
