import { JSX } from "react";

import SegmentPayload from "../../../models/segmentPayload";
import SplitPayload from "../../../models/splitPayload";
import { displayFormattedTimeParts, formatDuration, msToParts } from "../timer/timerUtils";

/**
 * Delta time display for splits and active rows.
 */
export function DeltaDisplay({ delta, gold = false }: { delta: number; gold?: boolean }): JSX.Element {
    const t = displayFormattedTimeParts(formatDuration(msToParts(delta), true));

    let className = "";

    if (gold) {
        className = "timer-gold";
    } else if (delta > 0) {
        className = "timer-behind";
    } else if (delta < 0) {
        className = "timer-ahead";
    }

    return (
        <strong className={className}>
            {t[0]}
            <small>{t[1]}</small>
        </strong>
    );
}

export function CumulativeTimeDisplay({
    split,
    targetCumulative,
}: {
    split: SplitPayload | null;
    targetCumulative: number | null;
}): JSX.Element {
    const value = split?.current_cumulative ?? targetCumulative;

    if (value == null) {
        return <strong className="target">-</strong>;
    }

    const t = displayFormattedTimeParts(formatDuration(msToParts(value)));

    return (
        <strong className="target">
            {t[0]}
            <small>{t[1]}</small>
        </strong>
    );
}

export function ComparisonDisplay({
    segment,
    split,
    targetIndividual,
    runningTime,
}: {
    segment: SegmentPayload;
    split: SplitPayload | null;
    targetIndividual: number | null;
    runningTime: number | null;
}): JSX.Element {
    const gold = segment.gold;

    // Completed segment -> show segment delta
    if (split) {
        if (targetIndividual == null) {
            return <strong>-</strong>;
        }

        const delta = split.current_duration - targetIndividual;

        return <DeltaDisplay delta={delta} gold={gold !== 0 && split.current_duration < gold} />;
    }

    // Active segment
    if (runningTime != null && targetIndividual != null) {
        if (runningTime >= targetIndividual - 60000) {
            return <DeltaDisplay delta={runningTime - targetIndividual} />;
        }
    }

    // Otherwise show comparison segment time
    if (targetIndividual == null) {
        return <strong className="target">-</strong>;
    }

    const t = displayFormattedTimeParts(formatDuration(msToParts(targetIndividual)));

    return (
        <strong className="target">
            {t[0]}
            <small>{t[1]}</small>
        </strong>
    );
}
