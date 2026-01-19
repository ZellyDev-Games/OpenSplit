import React, { useEffect } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { WindowSetPosition, WindowSetSize } from "../../../wailsjs/runtime";
import { Command } from "../../App";
import { MenuItem, useContextMenu } from "../../hooks/useContextMenu";
import { ConfigPayload } from "../../models/configPayload";
import SessionPayload from "../../models/sessionPayload";
import { ContextMenu } from "../ContextMenu";
import SegmentList from "./SegmentList";
import Timer from "./Timer";

export enum CompareAgainst {
    Best = "best",
    Average = "average",
}

export type Comparison = CompareAgainst.Best | CompareAgainst.Average;

type SplitterParams = {
    sessionPayload: SessionPayload;
    configPayload: ConfigPayload;
};

export default function Splitter({ sessionPayload, configPayload }: SplitterParams) {
    const contextMenu = useContextMenu();
    const [contextMenuItems, setContextMenuItems] = React.useState<MenuItem[]>([]);
    const [comparison, setComparison] = React.useState<Comparison>(CompareAgainst.Average);
    const [globalHotkeys, setGlobalHotkeys] = React.useState<boolean>(configPayload.global_hotkeys_active);

    useEffect(() => {
        (async () => {
            setContextMenuItems(await buildContextMenu());
        })();
    }, [globalHotkeys]);

    useEffect(() => {
        (async () => {
            if (sessionPayload.loaded_split_file) {
                WindowSetSize(
                    sessionPayload.loaded_split_file.window_width,
                    sessionPayload.loaded_split_file.window_height,
                );

                WindowSetPosition(sessionPayload.loaded_split_file.window_x, sessionPayload.loaded_split_file.window_y);
            }
        })();
    }, [sessionPayload.loaded_split_file?.id]);

    const buildContextMenu = async (): Promise<MenuItem[]> => {
        const contextMenuItems: MenuItem[] = [];
        contextMenuItems.push({
            label: (globalHotkeys ? "✓ " : "") + "Global Hotkeys",
            onClick: async () => {
                Dispatch(Command.TOGGLEGLOBAL, null).then((r) => {
                    if (r.code == 0) {
                        setGlobalHotkeys(r.message === "true");
                    }
                });
            },
        });

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
            label: "Compare Against Average",
            onClick: () => {
                setComparison(CompareAgainst.Average);
            },
        });

        contextMenuItems.push({
            label: "Compare Against Best",
            onClick: () => {
                setComparison(CompareAgainst.Best);
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
            <SegmentList sessionPayload={sessionPayload} comparison={comparison} />
            <Timer offset={(sessionPayload.loaded_split_file?.offset || 0) * -1} />
        </div>
    );
}
