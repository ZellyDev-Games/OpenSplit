import { SetStateAction, useEffect, useState } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { Command } from "../../models/command";
import SessionPayload from "../../models/sessionPayload";
import { log } from "../../utils/logger";
import { MenuItem } from "../useContextMenu";
import { CompareAgainst, Comparison } from "./useComparison";

export type SplitterLayout = "vertical" | "horizontal";

type UseSplitterMenuParams = {
    disableContextMenu: boolean;
    globalHotkeysInitial: boolean;
    comparison: Comparison;
    setComparison: React.Dispatch<SetStateAction<Comparison>>;
    sessionPayload: SessionPayload;
    initialLayout: SplitterLayout;
};

export function useSplitterMenu({
    disableContextMenu,
    globalHotkeysInitial,
    comparison,
    setComparison,
    sessionPayload,
    initialLayout,
}: UseSplitterMenuParams) {
    const [items, setItems] = useState<MenuItem[]>([]);
    const [globalHotkeys, setGlobalHotkeys] = useState(globalHotkeysInitial);
    const [layout, setLayout] = useState<SplitterLayout>(initialLayout);

    useEffect(() => {
        const savedLayout = sessionPayload.loaded_split_file?.layout;

        if (savedLayout === "vertical" || savedLayout === "horizontal") {
            setLayout(savedLayout);
            return;
        }

        setLayout(initialLayout);
    }, [initialLayout, sessionPayload.loaded_split_file?.layout]);

    useEffect(() => {
        if (disableContextMenu) {
            return;
        }

        (async () => {
            setItems(await buildContextMenu());
        })();
    }, [disableContextMenu, globalHotkeys, comparison, layout, sessionPayload.loaded_split_file?.wr?.show]);

    const setSplitterLayout = async (nextLayout: SplitterLayout) => {
        if (layout === nextLayout) {
            return;
        }

        log.debug("Splitter layout changed", nextLayout);

        const result = await Dispatch(Command.SETLAYOUT, nextLayout);

        if (result.code !== 0) {
            log.error("Failed to save splitter layout", result.message);
            return;
        }

        setLayout(nextLayout);
    };

    const buildContextMenu = async (): Promise<MenuItem[]> => {
        const contextMenuItems: MenuItem[] = [];

        contextMenuItems.push({
            label: (globalHotkeys ? "✓ " : "") + "Global Hotkeys",
            onClick: async () => {
                Dispatch(Command.TOGGLEGLOBAL, null).then((r) => {
                    if (r.code === 0) {
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

        contextMenuItems.push({
            type: "separator",
        });

        /*
         * Layout
         */

        contextMenuItems.push({
            label: (layout === "vertical" ? "✓ " : "") + "Vertical Layout",
            onClick: async () => {
                await setSplitterLayout("vertical");
            },
        });

        contextMenuItems.push({
            label: (layout === "horizontal" ? "✓ " : "") + "Horizontal Layout",
            onClick: async () => {
                await setSplitterLayout("horizontal");
            },
        });

        contextMenuItems.push({
            type: "separator",
        });

        /*
         * World record
         */

        contextMenuItems.push({
            label: ((sessionPayload.loaded_split_file?.wr?.show ?? false) ? "✓ " : "") + "Display World Record",
            onClick: async () => {
                await Dispatch(Command.TOGGLEWR, null);
            },
        });

        contextMenuItems.push({
            type: "separator",
        });

        /*
         * Comparison
         */

        contextMenuItems.push({
            label: (comparison === CompareAgainst.Average ? "✓ " : "") + "Compare Against Average",
            onClick: () => {
                log.debug("Comparison mode changed", comparison);
                setComparison(CompareAgainst.Average);
            },
        });

        contextMenuItems.push({
            label: (comparison === CompareAgainst.Best ? "✓ " : "") + "Compare Against Best Run",
            onClick: () => {
                log.debug("Comparison mode changed", comparison);
                setComparison(CompareAgainst.Best);
            },
        });

        contextMenuItems.push({
            label: (comparison === CompareAgainst.SumOfBest ? "✓ " : "") + "Compare Against Sum of Best Segments",
            onClick: () => {
                log.debug("Comparison mode changed", comparison);
                setComparison(CompareAgainst.SumOfBest);
            },
        });

        contextMenuItems.push({
            type: "separator",
        });

        /*
         * Close / exit
         */

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

    return {
        items,
        globalHotkeys,
        layout,
    };
}
