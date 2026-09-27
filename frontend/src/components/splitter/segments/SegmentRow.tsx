import { CSSProperties, RefObject } from "react";

import SegmentPayload from "../../../models/segmentPayload";
import SplitPayload from "../../../models/splitPayload";
import { ComparisonDisplay, CumulativeTimeDisplay, DeltaDisplay } from "./SegmentTime";
import { FlatSegment } from "./segmentUtils";

type SegmentRowProps = {
    segmentData: FlatSegment;
    split: SplitPayload | null;
    cumulativeTarget: number | null;
    individualTarget: number | null;
    hasSegmentIcons: boolean;
    activeRow?: boolean;
    time?: number | null;
    previousCumulative?: number;
    activeRowRef?: RefObject<HTMLTableRowElement | null>;
};

export default function SegmentRow({
    segmentData,
    split,
    cumulativeTarget,
    individualTarget,
    hasSegmentIcons,
    activeRow = false,
    time = null,
    previousCumulative = 0,
    activeRowRef,
}: SegmentRowProps) {
    let delta: number | null = null;

    if (split != null && cumulativeTarget != null) {
        delta = split.current_cumulative - cumulativeTarget;
    } else if (activeRow && time !== null && cumulativeTarget != null && time > cumulativeTarget - 60000) {
        delta = time - cumulativeTarget;
    }

    const runningSegmentTime = activeRow && time != null ? time - previousCumulative : null;

    const segment: SegmentPayload = segmentData.segment;

    return (
        <tr
            ref={activeRow ? (activeRowRef ?? null) : null}
            className={"segmentRow" + (activeRow ? " selected" : "")}
            style={
                {
                    "--segment-depth": segmentData.depth,
                } as CSSProperties
            }
        >
            {hasSegmentIcons && (
                <td className="segmentIcon">
                    {segment.icon && <img src={segment.icon} alt="" draggable={false} className="segment-icon" />}
                </td>
            )}

            <td className="splitName">{segment.name}</td>

            <td className="splitDelta">{delta !== null && <DeltaDisplay delta={delta} />}</td>

            <td className="splitComparison">
                <ComparisonDisplay
                    segment={segment}
                    split={split}
                    targetIndividual={individualTarget}
                    runningTime={runningSegmentTime}
                />
            </td>

            <td className="splitTime">
                <CumulativeTimeDisplay split={split} targetCumulative={cumulativeTarget} />
            </td>
        </tr>
    );
}
