import { useMemo } from "react";

import { Targets } from "../../components/splitter/segments/segmentUtils";
import SegmentPayload from "../../models/segmentPayload";
import { CompareAgainst, Comparison } from "./useComparison";

export function useComparisonTargets(comparison: Comparison, leaves?: SegmentPayload[] | null): Targets {
    /*
     * Comparison targets
     */
    return useMemo<Targets>(() => {
        let cumulative = 0;
        let cumulativeKnown = true;

        const result: Targets = {
            cumulative: {},
            individual: {},
        };

        const selector = {
            [CompareAgainst.Average]: (s: SegmentPayload) => s.average,
            [CompareAgainst.Best]: (s: SegmentPayload) => s.pb,
            [CompareAgainst.SumOfBest]: (s: SegmentPayload) => s.gold,
        };

        leaves?.forEach((segment) => {
            const value = selector[comparison](segment);

            // -1 is the split-file sentinel for an unknown comparison.
            // Keep it out of cumulative targets so an empty history does not
            // create a zero-second delta or a misleading cumulative time.
            if (value < 0) {
                cumulativeKnown = false;
                return;
            }

            result.individual[segment.id] = value;
            if (cumulativeKnown) {
                result.cumulative[segment.id] = cumulative + value;
            }

            cumulative += value;
        });

        return result;
    }, [comparison, leaves]);
}
