/**
 * Primary race timer window.
 *
 * Hosts:
 *  - Timer
 *  - Segment list
 *  - Context menu
 *  - Comparison mode
 */

import { SetStateAction, useEffect, useState } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { EventsOn } from "../../../wailsjs/runtime";
import { MenuItem, useContextMenu } from "../../hooks/useContextMenu";
import { Command } from "../../models/command";
import { ConfigPayload } from "../../models/configPayload";
import SessionPayload from "../../models/sessionPayload";
import { log } from "../../utils/logger";
import { ContextMenu } from "../ContextMenu";
import SegmentList from "./SegmentList";
import Timer from "./Timer";

export enum CompareAgainst {
    Best = "best",
    Average = "average",
    SumOfBest = "sumOfBest",
}

export type Comparison = CompareAgainst.Best | CompareAgainst.Average | CompareAgainst.SumOfBest;

const comparisons: Comparison[] = [CompareAgainst.Average, CompareAgainst.Best, CompareAgainst.SumOfBest];

type SplitterParams = {
    sessionPayload: SessionPayload;
    configPayload: ConfigPayload;
    disableContextMenu?: boolean;
    forceExpandAll?: boolean;

    comparison?: Comparison;
    onComparisonChange?: React.Dispatch<SetStateAction<Comparison>>;
};

export default function Splitter({
    sessionPayload,
    configPayload,
    disableContextMenu = false,
    forceExpandAll = false,
    comparison: controlledComparison,
    onComparisonChange,
}: SplitterParams) {
    const contextMenu = useContextMenu();
    const [contextMenuItems, setContextMenuItems] = useState<MenuItem[]>([]);
    const [globalHotkeys, setGlobalHotkeys] = useState<boolean>(configPayload.global_hotkeys_active);

    const [internalComparison, setInternalComparison] = useState<Comparison>(CompareAgainst.Average);

    const comparison = controlledComparison ?? internalComparison;

    const setComparison = (value: Comparison | ((current: Comparison) => Comparison)) => {
        if (onComparisonChange) {
            onComparisonChange(value);
        } else {
            setInternalComparison(value);
        }
    };

    const comparisonLabel: Record<Comparison, string> = {
        [CompareAgainst.Average]: "Comparing Against: Average",
        [CompareAgainst.Best]: "Comparing Against: Best Run",
        [CompareAgainst.SumOfBest]: "Comparing Against: Sum of Best Segments",
    };

    useEffect(() => {
        const rotate = (dir: number) => {
            setComparison((current) => {
                const index = comparisons.indexOf(current);
                const next = (index + dir + comparisons.length) % comparisons.length;
                return comparisons[next];
            });
            log.debug("Comparison mode changed", comparison);
        };

        const unsubLeft = EventsOn("comparison:left", () => rotate(-1));
        const unsubRight = EventsOn("comparison:right", () => rotate(1));

        return () => {
            unsubLeft();
            unsubRight();
        };
    }, []);

    useEffect(() => {
        if (disableContextMenu) {
            return;
        }

        (async () => {
            setContextMenuItems(await buildContextMenu());
        })();
    }, [disableContextMenu, globalHotkeys, comparison, sessionPayload.loaded_split_file?.wr?.show]);

    // regenerated whenever settings change
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
            label: ((sessionPayload.loaded_split_file?.wr?.show ?? false) ? "✓ " : "") + "Display World Record",
            onClick: async () => {
                await Dispatch(Command.TOGGLEWR, null);
            },
        });

        contextMenuItems.push({ type: "separator" });

        contextMenuItems.push({
            label: (comparison == CompareAgainst.Average ? "✓ " : "") + "Compare Against Average",
            onClick: () => {
                log.debug("Comparison mode changed", comparison);
                setComparison(CompareAgainst.Average);
            },
        });

        contextMenuItems.push({
            label: (comparison == CompareAgainst.Best ? "✓ " : "") + "Compare Against Best Run",
            onClick: () => {
                log.debug("Comparison mode changed", comparison);
                setComparison(CompareAgainst.Best);
            },
        });

        contextMenuItems.push({
            label: (comparison == CompareAgainst.SumOfBest ? "✓ " : "") + "Compare Against Sum of Best Segments",
            onClick: () => {
                log.debug("Comparison mode changed", comparison);
                setComparison(CompareAgainst.SumOfBest);
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
        <div {...(!disableContextMenu ? contextMenu.bind : {})} id="splitter">
            {!disableContextMenu && (
                <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
            )}
            <SegmentList sessionPayload={sessionPayload} comparison={comparison} forceExpandAll={forceExpandAll} />
            <div className="comparison-mode">{comparisonLabel[comparison]}</div>
            <Timer offset={sessionPayload.loaded_split_file?.offset ?? 0} wr={sessionPayload.loaded_split_file?.wr} />
        </div>
    );
}
