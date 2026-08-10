import { SkinElement } from "../../../models/skin/element";

const elements: SkinElement[] = [
    {
        id: "game_info",
        label: "Game Info",
        selector: "#gameInfo",
        stateful: true,
    },
    {
        id: "game_title",
        label: "Game Title",
        selector: "#gameTitle",
        stateful: true,
    },
    {
        id: "game_category",
        label: "Game Category",
        selector: "#gameCategory",
        stateful: true,
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
        stateful: true,
    },
];

export default elements;
