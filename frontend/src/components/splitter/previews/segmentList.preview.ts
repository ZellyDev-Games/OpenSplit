import { SkinElement } from "../../../../models/skin/element";

const elements: SkinElement[] = [
    {
        id: "segment_list",
        label: "Segment List",
        selector: "#splitList",
    },
    {
        id: "segment_list_complete",
        label: "Segment List — Complete",
        selector: "#splitList.complete",
    },
    {
        id: "segment_list_pb",
        label: "Segment List — PB",
        selector: "#splitList.complete.pb",
    },

    {
        id: "segment_container",
        label: "Segment Container",
        selector: "#splitContainer",
    },
    {
        id: "segment_container_complete",
        label: "Segment Container — Complete",
        selector: "#splitContainer.complete",
    },
    {
        id: "segment_container_pb",
        label: "Segment Container — PB",
        selector: "#splitContainer.complete.pb",
    },

    {
        id: "final_segment",
        label: "Final Segment",
        selector: "#finalSegment",
    },
    {
        id: "final_segment_complete",
        label: "Final Segment — Complete",
        selector: "#finalSegment.complete",
    },
    {
        id: "final_segment_pb",
        label: "Final Segment — PB",
        selector: "#finalSegment.complete.pb",
    },
];

export default elements;
