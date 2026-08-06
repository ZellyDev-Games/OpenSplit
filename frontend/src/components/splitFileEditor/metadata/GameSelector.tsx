import { RefObject } from "react";

import { log } from "../../../utils/logger";
import { GameMatch } from "../types/game";
import MetadataAutocomplete from "./MetadataAutocomplete";

type GameSelectorProps = {
    value: string;

    games: GameMatch[];

    selectingGame: RefObject<boolean>;

    onChange: (value: string) => void;

    onSelect: (game: GameMatch) => void;

    onClearSelection: () => void;
};

export default function GameSelector({
    value,
    games,
    selectingGame,
    onChange,
    onSelect,
    onClearSelection,
}: GameSelectorProps) {
    return (
        <MetadataAutocomplete
            label="Game Name"
            htmlFor="game-name"
            value={value}
            items={games}
            getKey={(game) => game.id}
            getLabel={(game) => game.name}
            onChange={(value) => {
                onChange(value);
                onClearSelection();
            }}
            onSelect={(game) => {
                selectingGame.current = true;

                log.info("Selected gameID:", game.id);
                log.info("Selected game:", game.name);

                onSelect(game);
            }}
        />
    );
}
