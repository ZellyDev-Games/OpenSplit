import { SkinElement } from "../../../models/skin/element";

const elements: SkinElement[] = [
    {
        id: "timer",
        label: "Timer",
        selector: "#time-container",
    },

    {
        id: "time_sign",
        label: "Time Sign",
        selector: "#time-sign",
    },
    {
        id: "time_hours",
        label: "Hours",
        selector: "#time-hours",
    },
    {
        id: "time_separator_hours_minutes",
        label: "Hours/Minutes Separator",
        selector: "#time-sep-hm",
    },
    {
        id: "time_minutes",
        label: "Minutes",
        selector: "#time-minutes",
    },
    {
        id: "time_separator_minutes_seconds",
        label: "Minutes/Seconds Separator",
        selector: "#time-sep-ms",
    },
    {
        id: "time_seconds",
        label: "Seconds",
        selector: "#time-seconds",
    },
    {
        id: "time_separator_seconds_centis",
        label: "Seconds/Centiseconds Separator",
        selector: "#time-sep-sc",
    },
    {
        id: "time_centis",
        label: "Centiseconds",
        selector: "#time-centis",
    },
];

export default elements;
