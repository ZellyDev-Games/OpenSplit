import { ContextMenu } from "../ContextMenu";
import SplitList from "./SplitList";
import Timer from "./Timer";
import React, { useEffect } from "react";
import { MenuItem, useContextMenu } from "../../hooks/useContextMenu";
import { useNavigate } from "react-router";
import { EventsOn, Quit, WindowGetPosition, WindowSetPosition } from "../../../wailsjs/runtime";
import { CloseSplitFile, GetLoadedSplitFile, LoadSplitFile, SaveSplitFile } from "../../../wailsjs/go/session/Service";
import { session } from "../../../wailsjs/go/models";
import ServicePayload = session.ServicePayload;
import Welcome from "./Welcome";
import useWindowResize from "../../hooks/useWindowResize";

export default function Splitter() {
    const navigate = useNavigate();
    useWindowResize("splitter");
    const contextMenu = useContextMenu();
    const [contextMenuItems, setContextMenuItems] = React.useState<MenuItem[]>([]);
    const [splitFileLoaded, setSplitFileLoaded] = React.useState(false);

    useEffect(() => {
        (async () => {
            setContextMenuItems(await buildContextMenu());
        })();

        return EventsOn("session:update", async (session: ServicePayload) => {
            setContextMenuItems(await buildContextMenu());
            setSplitFileLoaded(session.split_file !== null);
        });
    }, []);

    const buildContextMenu = async (): Promise<MenuItem[]> => {
        const contextMenuItems: MenuItem[] = [];
        const splitFile = await GetLoadedSplitFile();
        contextMenuItems.push({
            label: splitFile ? "Edit Split File" : "Create Split File",
            onClick: async () => {
                await updateWindowPos();
                navigate("/edit");
            },
        });

        if (splitFile) {
            contextMenuItems.push({
                label: "Save",
                onClick: async () => await SaveSplitFile().catch((err) => console.log(err)),
            });
        }

        contextMenuItems.push({
            label: "Open Split File",
            onClick: async () => await LoadSplitFile().catch((err) => console.log(err)),
        });

        contextMenuItems.push({ type: "separator" });

        if (splitFile) {
            contextMenuItems.push({
                label: "Close Split File",
                onClick: async () => await CloseSplitFile(),
            });
        }

        contextMenuItems.push({
            label: "Exit OpenSplit",
            onClick: async () => Quit(),
        });

        return contextMenuItems;
    };

    const updateWindowPos = async () => {
        const pos = await WindowGetPosition();
        localStorage.setItem("window-pos-x", String(pos.x));
        localStorage.setItem("window-pos-y", String(pos.y));
        console.log(`set window position to ${pos.x} - ${pos.y}`);
    };

    const applyWindowPos = () => {
        const x = Number(localStorage.getItem("window-pos-x"));
        const y = Number(localStorage.getItem("window-pos-y"));
        console.log(`applying window position ${x} - ${y}`);
        WindowSetPosition(x == 0 ? 200 : x, y == 0 ? 200 : y);
    };

    useEffect(() => {
        applyWindowPos();
    }, []);

    if (!splitFileLoaded) {
        return <Welcome />;
    }

    return (
        <div {...contextMenu.bind} className="splitter">
            <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
            <SplitList />
            <Timer />
        </div>
    );
}
