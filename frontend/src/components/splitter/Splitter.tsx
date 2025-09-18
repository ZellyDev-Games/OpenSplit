import React, { useEffect } from "react";
import { useNavigate } from "react-router";

import { session } from "../../../wailsjs/go/models";
import { CloseSplitFile, GetLoadedSplitFile, LoadSplitFile, SaveSplitFile } from "../../../wailsjs/go/session/Service";
import { EventsOn, Quit } from "../../../wailsjs/runtime";
import { MenuItem, useContextMenu } from "../../hooks/useContextMenu";
import useWindowResize from "../../hooks/useWindowResize";
import { ContextMenu } from "../ContextMenu";
import SplitList from "./SplitList";
import Timer from "./Timer";
import Welcome from "./Welcome";
import ServicePayload = session.ServicePayload;
import SplitFilePayload = session.SplitFilePayload;

export default function Splitter() {
    const navigate = useNavigate();
    const contextMenu = useContextMenu();
    const [contextMenuItems, setContextMenuItems] = React.useState<MenuItem[]>([]);
    const [splitFile, setSplitFile] = React.useState<SplitFilePayload | null>(null);
    const [setWindowPosition, getPageSize] = useWindowResize("splitter");

    useEffect(() => {
        (async () => {
            setContextMenuItems(await buildContextMenu());
        })();

        return EventsOn("session:update", async (session: ServicePayload) => {
            setContextMenuItems(await buildContextMenu());
            setSplitFile(session.split_file ?? null);
        });
    }, []);

    const buildContextMenu = async (): Promise<MenuItem[]> => {
        const contextMenuItems: MenuItem[] = [];
        const splitFile = await GetLoadedSplitFile();
        contextMenuItems.push({
            label: splitFile ? "Edit Split File" : "Create Split File",
            onClick: async () => {
                navigate("/edit");
            },
        });

        if (splitFile) {
            contextMenuItems.push({
                label: "Save",
                onClick: async () => {
                    const [w, h] = getPageSize("splitter");
                    const [x, y] = await setWindowPosition("splitter");
                    await SaveSplitFile(w, h, x, y).catch((err) => console.log(err));
                },
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

    if (!splitFile) {
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
