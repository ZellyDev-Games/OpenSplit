export type MinimumSize = {
    width: number;
    height: number;
};

export type SizeContribution = MinimumSize & {
    source?: string;
};

export type LayoutDirection = "row" | "column";
