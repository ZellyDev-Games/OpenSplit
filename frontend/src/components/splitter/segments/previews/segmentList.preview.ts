import { SkinElement } from "../../../../models/skin/element";

const elements: SkinElement[] = [
    {
        id: "segment_list",
        label: "Segment List",
        selector: "#splitList",
    },
    {
        id: "segment_body",
        label: "Segment Body",
        selector: "#splitBody",
    },
    {
        id: "segment_container",
        label: "Segment Container",
        selector: "#splitContainer",
    },
    {
        id: "final_segment",
        label: "Final Segment",
        selector: "#finalSegment",
    },
    {
        id: "completed_run",
        label: "Completed Run",
        selector: ".complete",
    },
    {
        id: "personal_best",
        label: "Personal Best",
        selector: ".pb",
    },
];

export default elements;
