import type { SkinElement } from "../../../models/skinModel";

const elements: SkinElement[] = [
    {
        id: "game_info",
        label: "Game Info",
        selector: "#gameInfo",
    },
    {
        id: "game_title",
        label: "Game Title",
        selector: "#gameTitle",
    },
    {
        id: "game_category",
        label: "Game Category",
        selector: "#gameCategory",
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
];

export default elements;
