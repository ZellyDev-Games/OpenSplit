import { useEffect, useState } from "react";

import { GetAvailableSkins } from "../../../wailsjs/go/skin/Service";
import { Platforms } from "../../../wailsjs/go/speedrun/Service";
import { Platform } from "../../components/splitFileEditor/types/game";
import { log } from "../../utils/logger";

export function useAvailableSkins() {
    const [availableSkins, setAvailableSkins] = useState<string[]>([]);
    const [platforms, setPlatforms] = useState<Platform[]>([]);

    useEffect(() => {
        const load = async () => {
            const [skinsResult, platformsResult] = await Promise.allSettled([GetAvailableSkins(), Platforms()]);

            if (skinsResult.status === "fulfilled") {
                setAvailableSkins(skinsResult.value);
            } else {
                log.error("Unable to load skins", skinsResult.reason);
            }

            if (platformsResult.status === "fulfilled") {
                setPlatforms(platformsResult.value);
            } else {
                log.error("Unable to load platforms", platformsResult.reason);
            }
        };

        void load();
    }, []);

    return {
        availableSkins,
        platforms,
        setPlatforms,
    };
}
