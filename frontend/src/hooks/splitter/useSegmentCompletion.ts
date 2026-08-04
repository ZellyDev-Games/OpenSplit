import { useEffect, useState } from "react";

import SessionPayload from "../../models/sessionPayload";
import { log } from "../../utils/logger";

export function useSegmentCompletion(sessionPayload: SessionPayload) {
    const [completeClassName, setCompleteClassName] = useState("");

    /*
     * Completion state
     */
    useEffect(() => {
        let className = "";

        const splitFile = sessionPayload.loaded_split_file;
        const run = sessionPayload.current_run;
        const leaves = sessionPayload.leaf_segments;

        if (splitFile && run && leaves && Object.keys(run.splits).length === leaves.length) {
            className = "complete";

            const pb = splitFile.pb;

            log.info("[SegmentList] Run completed", {
                pb: className.includes("pb"),
                totalSplits: leaves.length,
            });
            if (pb) {
                const finalSplit = leaves[leaves.length - 1];
                const finalTime = run.splits[finalSplit.id].current_cumulative;

                if (finalTime < pb.total_time) {
                    className += " pb";
                }
            }

            log.info("[SegmentList] Run completed", {
                totalSplits: leaves.length,
                pb: className.includes("pb"),
            });
        }

        log.debug(sessionPayload.loaded_split_file);
        log.debug(sessionPayload.loaded_split_file?.variables);
        setCompleteClassName(className);
    }, [sessionPayload]);

    return completeClassName;
}
