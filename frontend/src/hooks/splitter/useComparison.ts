import { Dispatch, SetStateAction, useEffect, useState } from "react";

import { EventsOn } from "../../../wailsjs/runtime";
import { log } from "../../utils/logger";

export enum CompareAgainst {
    Best = "best",
    Average = "average",
    SumOfBest = "sumOfBest",
}

export type Comparison = CompareAgainst.Best | CompareAgainst.Average | CompareAgainst.SumOfBest;

const comparisons = [CompareAgainst.Average, CompareAgainst.Best, CompareAgainst.SumOfBest];

export function useComparison(controlled?: Comparison, onChange?: Dispatch<SetStateAction<Comparison>>) {
    const [internal, setInternal] = useState<Comparison>(CompareAgainst.Average);

    const comparison = controlled ?? internal;

    const setComparison = (value: Comparison | ((x: Comparison) => Comparison)) => {
        if (onChange) {
            onChange(value);
        } else {
            setInternal(value);
        }
    };

    useEffect(() => {
        const rotate = (dir: number) => {
            setComparison((current) => {
                const index = comparisons.indexOf(current);

                return comparisons[(index + dir + comparisons.length) % comparisons.length];
            });
            log.debug("Comparison mode changed", comparison);
        };

        const left = EventsOn("comparison:left", () => rotate(-1));

        const right = EventsOn("comparison:right", () => rotate(1));

        return () => {
            left();
            right();
        };
    }, []);

    return {
        comparison,
        setComparison,
    };
}
