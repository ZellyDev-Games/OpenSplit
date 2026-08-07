import { useEffect, useState } from "react";

import { SearchCategories } from "../../../wailsjs/go/speedrun/Service";
import { Category } from "../../components/splitFileEditor/types/game";
import { log } from "../../utils/logger";

export function useCategorySearch(gameID: string) {
    const [categories, setCategories] = useState<Category[]>([]);

    useEffect(() => {
        log.info("gameID changed:", gameID);
        if (!gameID) {
            setCategories([]);
            return;
        }

        const load = async () => {
            const result = await SearchCategories(gameID);
            log.debug("categories:", result);

            setCategories(result.data);
        };

        void load();
    }, [gameID]);

    return categories;
}
