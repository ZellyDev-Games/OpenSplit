import { useEffect } from "react";

import { WindowCenter } from "../../../wailsjs/runtime/runtime";
import SplitFilePayload from "../../models/splitFilePayload";
import { log } from "../../utils/logger";
import { useSaveSplitFile } from "./useSaveSplitFile";
import { useSegmentEditor } from "./useSegmentEditor";
import { useSplitMetadata } from "./useSplitMetadata";

export default function useSplitEditor(splitFilePayload: SplitFilePayload | null) {
    useEffect(() => {
        WindowCenter();
    }, []);

    const metadata = useSplitMetadata(splitFilePayload);

    const segments = useSegmentEditor(splitFilePayload?.segments ?? []);

    log.debug(metadata.categoryID);
    log.debug(metadata.gameID);

    const saveSplitFile = useSaveSplitFile({
        splitFilePayload,

        gameName: metadata.gameName,

        gameID: metadata.gameID,

        categoryName: metadata.categoryName,

        categoryID: metadata.categoryID,

        selectedSkin: metadata.selectedSkin,

        variables: metadata.selectedVariables,

        segments: segments.segments,

        attempts: metadata.attempts,

        offsetText: metadata.offsetText,

        platform: metadata.platform,
    });

    return {
        ...metadata,
        ...segments,

        editing: splitFilePayload !== null,

        saveSplitFile,
    };
}
