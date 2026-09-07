import { SkinElement } from "../../../models/skin/element";

const elements: SkinElement[] = [
    {
        id: "segment_row",
        label: "Segment Row",
        selector: ".segmentRow",
    },
    {
        id: "active_segment_row",
        label: "Selected Segment",
        selector: ".segmentRow.selected",
    },
    {
        id: "segment_name",
        label: "Segment Name",
        selector: ".splitName",
    },
    {
        id: "segment_icon",
        label: "Segment Icon",
        selector: ".segment-icon",
    },
    {
        id: "segment_delta",
        label: "Segment Delta",
        selector: ".splitDelta",
    },
    {
        id: "segment_comparison",
        label: "Segment Comparison",
        selector: ".splitComparison",
    },
    {
        id: "segment_time",
        label: "Segment Time",
        selector: ".splitTime",
    },
];

export default elements;
