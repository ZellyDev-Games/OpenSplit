/**
 * Primary race timer window.
 *
 * Hosts:
 *  - Timer
 *  - Segment list
 *  - Split game information
 *  - Context menu
 *  - Comparison mode
 */

import { Dispatch, SetStateAction } from "react";

import { CompareAgainst, Comparison, useComparison } from "../../hooks/splitter/useComparison";
import { useSegmentList } from "../../hooks/splitter/useSegmentList";
import { useSplitterMenu } from "../../hooks/splitter/useSplitterMenu";
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

export default function Splitter({
    sessionPayload,
    configPayload,
    disableContextMenu = false,
    forceExpandAll = false,
    comparison: controlledComparison,
    onComparisonChange,
}: SplitterParams) {
    const contextMenu = useContextMenu();

    const { comparison, setComparison } = useComparison(controlledComparison, onComparisonChange);

    const { items: contextMenuItems } = useSplitterMenu({
        disableContextMenu,
        globalHotkeysInitial: configPayload.global_hotkeys_active,
        comparison,
        setComparison,
        sessionPayload,
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
        <div {...(!disableContextMenu ? contextMenu.bind : {})} id="splitter">
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

            <div className="comparison-mode">{comparisonLabel[comparison]}</div>

            <Timer offset={sessionPayload.loaded_split_file?.offset ?? 0} />
            <WorldRecordDisplay worldRecord={sessionPayload.loaded_split_file?.wr} />
        </div>
    );
}
