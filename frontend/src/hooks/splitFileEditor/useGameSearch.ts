import { RefObject, useEffect, useState } from "react";

import { SearchGames } from "../../../wailsjs/go/speedrun/Service";
import { GameMatch } from "../../components/splitFileEditor/types/game";

export default function useGameSearch(value: string, selectingGame: RefObject<boolean>) {
    const [games, setGames] = useState<GameMatch[]>([]);

    useEffect(() => {
        if (selectingGame.current) {
            selectingGame.current = false;
            return;
        }

        const query = value.trim();

        const timeout = setTimeout(async () => {
            if (query.length === 0) {
                setGames([]);
                return;
            }

            const result = await SearchGames(query);

            setGames(
                result.data.map((g) => ({
                    id: g.id,
                    name: g.names.international,
                    platforms: g.platforms,
                })),
            );
        }, 200);

        return () => clearTimeout(timeout);
    }, [value]);

    return { games, selectingGame };
}
