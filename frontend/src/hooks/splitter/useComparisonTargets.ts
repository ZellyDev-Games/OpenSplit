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

            result.individual[segment.id] = value;
            result.cumulative[segment.id] = cumulative + value;

            cumulative += value;
        });

        return result;
    }, [comparison, leaves]);
}
