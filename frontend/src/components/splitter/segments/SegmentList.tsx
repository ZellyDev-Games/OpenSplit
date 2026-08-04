/**
 * Displays the active split list during a run.
 *
 * Features:
 *  - Hierarchical segment display
 *  - Expand/collapse parent segments
 *  - Live timer updates
 *  - Comparison against Average / PB / Sum of Best
 *  - Automatic scrolling to the active split
 *  - Parent aggregate timing and delta calculations
 */

import { Comparison } from "../../../hooks/splitter/useComparison";
import { useSegmentList } from "../../../hooks/splitter/useSegmentList";
import SessionPayload from "../../../models/sessionPayload";
import SplitGameInfo from "../SplitGameInfo";

type SegmentListParameters = {
    sessionPayload: SessionPayload;
    comparison: Comparison;
    forceExpandAll?: boolean;
};

export default function SegmentList({ sessionPayload, comparison, forceExpandAll }: SegmentListParameters) {
    const { completeClassName, rows, finalRow, containerRef } = useSegmentList({
        sessionPayload,
        comparison,
        forceExpandAll,
    });

    return (
        <div id="splitList" className={completeClassName}>
            <SplitGameInfo sessionPayload={sessionPayload} completeClassName={completeClassName} />

            <div id="splitBody" className={completeClassName}>
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
        </div>
    );
}
