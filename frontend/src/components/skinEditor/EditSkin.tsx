import { useState } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { Command } from "../../models/command";

type Props = {
    skins: string[];
};

export default function EditSkin({ skins }: Props) {
    const [selectedSkin, setSelectedSkin] = useState("");

    async function submit() {
        if (!selectedSkin) {
            return;
        }

        await Dispatch(
            Command.SUBMIT,
            JSON.stringify({
                name: selectedSkin,
            }),
        );
    }

    async function cancel() {
        await Dispatch(Command.CANCEL, null);
    }

    return (
        <div className="skin-selector">
            <h2>Edit Skin</h2>

            <label>Select Skin</label>

            <select value={selectedSkin} onChange={(e) => setSelectedSkin(e.target.value)}>
                <option value="">Select a skin...</option>

                {skins.map((skin) => (
                    <option key={skin} value={skin}>
                        {skin}
                    </option>
                ))}
            </select>

            <div className="skin-selector-buttons">
                <button disabled={!selectedSkin} onClick={submit}>
                    Edit Skin
                </button>

                <button onClick={cancel}>Cancel</button>
            </div>
        </div>
    );
}
