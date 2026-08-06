import { RefObject } from "react";

import { useTimer } from "../../../hooks/splitter/useTimer";
import SegmentRow from "./SegmentRow";
import { FlatSegment } from "./segmentUtils";

type ActiveSegmentRowProps = {
    segmentData: FlatSegment;
    cTarget: number;
    iTarget: number;
    previousCumulative: number;
    activeRowRef: RefObject<HTMLTableRowElement | null>;
};

/**
 * ActiveSegmentRow subscribes to timer updates so only the active
 * segment rerenders during a run instead of the entire segment table.
 */
export default function ActiveSegmentRow({
    segmentData,
    cTarget,
    iTarget,
    previousCumulative,
    activeRowRef,
}: ActiveSegmentRowProps) {
    const time = useTimer();

    return (
        <SegmentRow
            segmentData={segmentData}
            split={null}
            cumulativeTarget={cTarget}
            individualTarget={iTarget}
            activeRow
            time={time}
            previousCumulative={previousCumulative}
            activeRowRef={activeRowRef}
        />
    );
}
