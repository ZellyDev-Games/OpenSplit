import { useEffect, useRef, useState } from "react";

import SplitFilePayload from "../../models/splitFilePayload";
import { useAvailableSkins } from "./useAvailableSkins";
import { useCategorySearch } from "./useCategorySearch";
import useGameSearch from "./useGameSearch";
import { useVariables } from "./useVariables";

export function useSplitMetadata(splitFilePayload: SplitFilePayload | null) {
    const [gameID, setGameID] = useState(splitFilePayload?.speedrun_game_id ?? "");

    const [gameName, setGameName] = useState(splitFilePayload?.game_name ?? "");

    const [categoryID, setCategoryID] = useState(splitFilePayload?.speedrun_game_category_id ?? "");

    const [categoryName, setCategoryName] = useState(splitFilePayload?.game_category ?? "");

    const [selectedSkin, setSelectedSkin] = useState(splitFilePayload?.selected_skin ?? "");

    const [platform, setPlatform] = useState(splitFilePayload?.platform ?? "");

    const [attempts, setAttempts] = useState(splitFilePayload?.attempts ?? 0);

    const [offsetText, setOffsetText] = useState(String(splitFilePayload?.offset ?? 0));

    const selectingGame = useRef(false);

    const { games } = useGameSearch(gameName, selectingGame);

    const categories = useCategorySearch(gameID);

    const { availableSkins, platforms, setPlatforms } = useAvailableSkins();

    const { variables, selectedVariables, setSelectedVariables } = useVariables({
        categoryID,
        splitFilePayload,
    });

    useEffect(() => {
        if (!gameID) {
            setCategoryID("");
            setCategoryName("");
        }
    }, [gameID]);

    const handleOffsetChange = (value: string) => {
        if (/^-?\d*$/.test(value)) {
            setOffsetText(value);
        }
    };

    return {
        games,
        selectingGame,

        gameID,
        setGameID,

        gameName,
        setGameName,

        categoryID,
        setCategoryID,

        categoryName,
        setCategoryName,

        categories,

        variables,
        selectedVariables,
        setSelectedVariables,

        selectedSkin,
        setSelectedSkin,

        availableSkins,

        platform,
        setPlatform,

        platforms,
        setPlatforms,

        attempts,
        setAttempts,

        offsetText,
        handleOffsetChange,
    };
}
