/**
 * Primary race timer window.
 *
 * Hosts:
 *  - Timer
 *  - Segment list
 *  - Split game information
 *  - Context menu
 *  - Comparison mode
 *  - World record
 *
 * The splitter supports two layouts:
 *
 *  - vertical
 *  - horizontal
 *
 * The active layout is exposed through:
 *
 *  data-layout="vertical"
 *  data-layout="horizontal"
 */

import { Dispatch, SetStateAction, useEffect, useRef, useState } from "react";

import { CompareAgainst, Comparison, useComparison } from "../../hooks/splitter/useComparison";
import { useSegmentList } from "../../hooks/splitter/useSegmentList";
import { SplitterLayout, useSplitterMenu } from "../../hooks/splitter/useSplitterMenu";
import { useContextMenu } from "../../hooks/useContextMenu";
import { ConfigPayload } from "../../models/configPayload";
import SessionPayload from "../../models/sessionPayload";
import { ContextMenu } from "../ContextMenu";
import SplitGameInfo from "./SplitGameInfo";
import Timer from "./timer/Timer";
import WorldRecordDisplay from "./timer/WorldRecord";

type SplitterParams = {
    sessionPayload: SessionPayload;
    configPayload: ConfigPayload;
    disableContextMenu?: boolean;
    forceExpandAll?: boolean;

    comparison?: Comparison;
    onComparisonChange?: Dispatch<SetStateAction<Comparison>>;
};

const DEFAULT_LAYOUT: SplitterLayout = "vertical";

function getSkinDefaultLayout(element: HTMLElement): SplitterLayout {
    const layout = getComputedStyle(element).getPropertyValue("--splitter-layout").trim();

    if (layout === "horizontal" || layout === "vertical") {
        return layout;
    }

    return DEFAULT_LAYOUT;
}

function getSavedLayout(sessionPayload: SessionPayload): SplitterLayout | null {
    const layout = sessionPayload.loaded_split_file?.layout;

    if (layout === "horizontal" || layout === "vertical") {
        return layout;
    }

    return null;
}

export default function Splitter({
    sessionPayload,
    configPayload,
    disableContextMenu = false,
    forceExpandAll = false,
    comparison: controlledComparison,
    onComparisonChange,
}: SplitterParams) {
    const splitterRef = useRef<HTMLDivElement>(null);
    const [initialLayout, setInitialLayout] = useState<SplitterLayout>(DEFAULT_LAYOUT);

    const contextMenu = useContextMenu();

    const { comparison, setComparison } = useComparison(controlledComparison, onComparisonChange);

    /*
     * The split file's saved layout takes precedence over the
     * skin's preferred layout.
     *
     * If no layout has been saved yet, use:
     *
     *     --splitter-layout: vertical;
     *
     * or:
     *
     *     --splitter-layout: horizontal;
     */
    useEffect(() => {
        const savedLayout = getSavedLayout(sessionPayload);

        if (savedLayout !== null) {
            setInitialLayout(savedLayout);
            return;
        }

        if (!splitterRef.current) {
            return;
        }

        setInitialLayout(getSkinDefaultLayout(splitterRef.current));
    }, [sessionPayload.loaded_split_file?.layout]);

    const { layout, items: contextMenuItems } = useSplitterMenu({
        disableContextMenu,
        globalHotkeysInitial: configPayload.global_hotkeys_active,
        comparison,
        setComparison,
        sessionPayload,
        initialLayout,
    });

    const { completeClassName, rows, finalRow, containerRef } = useSegmentList({
        sessionPayload,
        comparison,
        forceExpandAll,
    });

    const comparisonLabel: Record<Comparison, string> = {
        [CompareAgainst.Average]: "Comparing Against: Average",
        [CompareAgainst.Best]: "Comparing Against: Best Run",
        [CompareAgainst.SumOfBest]: "Comparing Against: Sum of Best Segments",
    };

    return (
        <div ref={splitterRef} {...(!disableContextMenu ? contextMenu.bind : {})} id="splitter" data-layout={layout}>
            {!disableContextMenu && (
                <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
            )}

            <SplitGameInfo sessionPayload={sessionPayload} completeClassName={completeClassName} />

            <div id="splitList" className={completeClassName}>
                <div ref={containerRef} id="splitContainer" className={completeClassName}>
                    <table cellSpacing={0} className={completeClassName}>
                        <tbody>{rows}</tbody>
                    </table>
                </div>

                <div id="finalSegment" className={completeClassName}>
                    <table cellSpacing={0} className={completeClassName}>
                        <tbody>{finalRow}</tbody>
                    </table>
                </div>
            </div>

            <div id="splitterInfo">
                <div className="comparison-mode">{comparisonLabel[comparison]}</div>

                <Timer offset={sessionPayload.loaded_split_file?.offset ?? 0} />

                <WorldRecordDisplay worldRecord={sessionPayload.loaded_split_file?.wr} />
            </div>
        </div>
    );
}
