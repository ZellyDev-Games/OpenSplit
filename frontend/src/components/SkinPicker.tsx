import { useEffect, useState } from "react";

import { GetAvailableSkins } from "../../wailsjs/go/skin/Service";
import { OpenSkinsFolder } from "../../wailsjs/go/sysopen/Service";
import { setActiveSkin } from "../skinLoader";

export default function SkinPicker() {
    const [active, setActive] = useState<string>("default");
    const [available, setAvailable] = useState<string[]>([]);

    useEffect(() => {
        setTimeout(async () => {
            const skins = await GetAvailableSkins();
            setAvailable(skins);
        }, 1000);
    }, []);

    return (
        <div style={{ display: "flex", gap: 8 }}>
            {available &&
                available.map((name) => (
                    <button
                        key={name}
                        onClick={() => {
                            setActiveSkin(name);
                            setActive(name);
                        }}
                    >
                        {name}
                        {active === name ? " ✓" : ""}
                    </button>
                ))}
            <button onClick={OpenSkinsFolder}>Open Skins Folder</button>
        </div>
    );
}
