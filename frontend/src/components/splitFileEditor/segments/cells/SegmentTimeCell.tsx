import { displayFormattedTimeParts, formatDuration, msToParts } from "../../../splitter/timer/timerUtils";

type SegmentTimeCellProps = {
    value: number;
};

export default function SegmentTimeCell({ value }: SegmentTimeCellProps) {
    if (value < 0) {
        return <td />;
    }

    return <td>{displayFormattedTimeParts(formatDuration(msToParts(value))).join("")}</td>;
}
