/**
 * Primary race timer window.
 *
 * Hosts:
 *  - Timer
 *  - Segment list
 *  - Context menu
 *  - Comparison mode
 */

import { SetStateAction } from "react";

import { CompareAgainst, Comparison, useComparison } from "../../hooks/splitter/useComparison";
import { useContextMenu } from "../../hooks/splitter/useContextMenu";
import { useSplitterMenu } from "../../hooks/splitter/useSplitterMenu";
import { ConfigPayload } from "../../models/configPayload";
import SessionPayload from "../../models/sessionPayload";
import { ContextMenu } from "./ContextMenu";
import SegmentList from "./segments/SegmentList";
import Timer from "./timer/Timer";

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

    const { comparison, setComparison } = useComparison(controlledComparison, onComparisonChange);

    const { items: contextMenuItems } = useSplitterMenu({
        disableContextMenu,
        globalHotkeysInitial: configPayload.global_hotkeys_active,
        comparison,
        setComparison,
        sessionPayload,
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
            <SegmentList sessionPayload={sessionPayload} comparison={comparison} forceExpandAll={forceExpandAll} />
            <div className="comparison-mode">{comparisonLabel[comparison]}</div>
            <Timer offset={sessionPayload.loaded_split_file?.offset ?? 0} wr={sessionPayload.loaded_split_file?.wr} />
        </div>
    );
}
