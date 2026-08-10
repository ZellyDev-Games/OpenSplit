import { SkinElement } from "../../../models/skin/element";

const elements: SkinElement[] = [
    {
        id: "game_info",
        label: "Game Info",
        selector: "#gameInfo",
    },
    {
        id: "game_info_complete",
        label: "Game Info — Complete",
        selector: "#gameInfo.complete",
    },
    {
        id: "game_info_pb",
        label: "Game Info — PB",
        selector: "#gameInfo.complete.pb",
    },

    {
        id: "game_title",
        label: "Game Title",
        selector: "#gameTitle",
    },
    {
        id: "game_title_complete",
        label: "Game Title — Complete",
        selector: "#gameTitle.complete",
    },
    {
        id: "game_title_pb",
        label: "Game Title — PB",
        selector: "#gameTitle.complete.pb",
    },

    {
        id: "game_category",
        label: "Game Category",
        selector: "#gameCategory",
    },
    {
        id: "game_category_complete",
        label: "Game Category — Complete",
        selector: "#gameCategory.complete",
    },
    {
        id: "game_category_pb",
        label: "Game Category — PB",
        selector: "#gameCategory.complete.pb",
    },

    {
        id: "game_variable",
        label: "Game Variable",
        selector: ".game-variable",
    },

    {
        id: "attempts",
        label: "Attempts",
        selector: "#attempts",
    },
    {
        id: "attempts_complete",
        label: "Attempts — Complete",
        selector: "#attempts.complete",
    },
    {
        id: "attempts_pb",
        label: "Attempts — PB",
        selector: "#attempts.complete.pb",
    },
];

export default elements;
