import React, { useEffect } from "react";

import { Dispatch } from "../../../wailsjs/go/statemachine/Service";
import { EventsOn } from "../../../wailsjs/runtime";
import { Command } from "../../App";
import { MenuItem, useContextMenu } from "../../hooks/useContextMenu";
import SessionPayload from "../../models/sessionPayload";
import { ContextMenu } from "../ContextMenu";
import SplitList from "./SplitList";
import Timer from "./Timer";

type SplitterParams = {
    sessionPayload: SessionPayload;
};

export default function Splitter({ sessionPayload }: SplitterParams) {
    const contextMenu = useContextMenu();
    const [contextMenuItems, setContextMenuItems] = React.useState<MenuItem[]>([]);
    const [session, setSession] = React.useState<SessionPayload>(sessionPayload);

    // Subscribe to session updates from the backend
    useEffect(() => {
        return EventsOn("session:update", (payload: SessionPayload) => setSession(payload));
    }, []);

    useEffect(() => {
        (async () => {
            setContextMenuItems(await buildContextMenu());
        })();
    }, [sessionPayload]);

    const buildContextMenu = async (): Promise<MenuItem[]> => {
        const contextMenuItems: MenuItem[] = [];
        contextMenuItems.push({
            label: "Edit Split File",
            onClick: async () => {
                await Dispatch(Command.EDIT, null);
            },
        });

        contextMenuItems.push({
            label: "Save",
            onClick: async () => {
                await Dispatch(Command.SAVE, null);
            },
        });

        contextMenuItems.push({ type: "separator" });

        contextMenuItems.push({
            label: "Close Split File",
            onClick: () => {
                Dispatch(Command.CLOSE, null);
            },
        });

        contextMenuItems.push({
            label: "Exit OpenSplit",
            onClick: async () => Dispatch(Command.QUIT, null),
        });

        return contextMenuItems;
    };

    return (
        <div {...contextMenu.bind} className="splitter">
            <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
            <SplitList sessionPayload={session} />
            <Timer />
        </div>
    );
}
