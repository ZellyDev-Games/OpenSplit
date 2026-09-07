import { SkinElement } from "../../../models/skin/element";

const elements: SkinElement[] = [
    {
        id: "parent_row",
        label: "Parent Row",
        selector: ".parentRow",
    },
    {
        id: "collapse_toggle",
        label: "Collapse Toggle",
        selector: ".collapseToggle",
    },

    {
        id: "parent_name",
        label: "Parent Name",
        selector: ".parentName",
    },
    {
        id: "parent_name_complete",
        label: "Parent Name — Complete",
        selector: ".parentName.complete",
    },
    {
        id: "parent_name_pb",
        label: "Parent Name — PB",
        selector: ".parentName.complete.pb",
    },

    {
        id: "parent_icon",
        label: "Parent Icon",
        selector: ".segment-icon",
    },

    {
        id: "parent_delta",
        label: "Parent Delta",
        selector: ".parentDelta",
    },
    {
        id: "parent_delta_complete",
        label: "Parent Delta — Complete",
        selector: ".parentDelta.complete",
    },
    {
        id: "parent_delta_pb",
        label: "Parent Delta — PB",
        selector: ".parentDelta.complete.pb",
    },

    {
        id: "parent_comparison",
        label: "Parent Comparison",
        selector: ".parentComparison",
    },
    {
        id: "parent_comparison_complete",
        label: "Parent Comparison — Complete",
        selector: ".parentComparison.complete",
    },
    {
        id: "parent_comparison_pb",
        label: "Parent Comparison — PB",
        selector: ".parentComparison.complete.pb",
    },

    {
        id: "parent_time",
        label: "Parent Time",
        selector: ".parentTime",
    },
    {
        id: "parent_time_complete",
        label: "Parent Time — Complete",
        selector: ".parentTime.complete",
    },
    {
        id: "parent_time_pb",
        label: "Parent Time — PB",
        selector: ".parentTime.complete.pb",
    },
];

export default elements;
