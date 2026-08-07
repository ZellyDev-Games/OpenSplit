import { RunningTotals } from "../../components/splitFileEditor/types/render";

export function addTime(total: number, value: number): number {
    return value < 0 ? total : total + value;
}

export function cloneTotals(totals: RunningTotals): RunningTotals {
    return {
        avg: totals.avg,
        pb: totals.pb,
        gold: totals.gold,
    };
}
