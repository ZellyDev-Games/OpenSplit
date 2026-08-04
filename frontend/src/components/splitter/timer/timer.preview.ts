import type { SkinElement } from "../../../models/skinModel";

const elements: SkinElement[] = [
    {
        id: "timer",
        label: "Timer",
        selector: "#timer-container",
    },
    {
        id: "timer_display",
        label: "Timer Display",
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
        id: "time_minutes",
        label: "Minutes",
        selector: "#time-minutes",
    },
    {
        id: "time_seconds",
        label: "Seconds",
        selector: "#time-seconds",
    },
    {
        id: "time_centis",
        label: "Centiseconds",
        selector: "#time-centis",
    },
    {
        id: "world_record",
        label: "World Record",
        selector: "#world-record",
    },
    {
        id: "time_separator_hours_minutes",
        label: "Hours/Minutes Separator",
        selector: "#time-sep-hm",
    },
    {
        id: "time_separator_minutes_seconds",
        label: "Minutes/Seconds Separator",
        selector: "#time-sep-ms",
    },
    {
        id: "time_separator_seconds_centis",
        label: "Seconds/Centiseconds Separator",
        selector: "#time-sep-sc",
    },
];

export default elements;
