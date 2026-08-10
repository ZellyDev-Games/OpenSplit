import { SkinElement } from "../../../../models/skin/element";

const elements: SkinElement[] = [
    {
        id: "segment_list",
        label: "Segment List",
        selector: "#splitList",
        stateful: true,
    },
    {
        id: "segment_body",
        label: "Segment Body",
        selector: "#splitBody",
        stateful: true,
    },
    {
        id: "segment_container",
        label: "Segment Container",
        selector: "#splitContainer",
        stateful: true,
    },
    {
        id: "final_segment",
        label: "Final Segment",
        selector: "#finalSegment",
        stateful: true,
    },
];

export default elements;
