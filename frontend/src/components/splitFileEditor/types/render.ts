import { ReactElement } from "react";

export type RunningTotals = {
    avg: number;
    pb: number;
    gold: number;
};

export type RenderResult = {
    rows: ReactElement[];
    totals: RunningTotals;
};
