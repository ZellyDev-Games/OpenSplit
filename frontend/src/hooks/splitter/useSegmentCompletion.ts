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
            const isPB = pb?.id === run.id;

            if (isPB) {
                className += " pb";
            }

            log.info("[SegmentList] Run completed", {
                runId: run.id,
                pbId: pb?.id ?? null,
                pb: isPB,
                totalSplits: leaves.length,
                runTotal: run.total_time,
                pbTotal: pb?.total_time ?? null,
            });
        }

        setCompleteClassName(className);
    }, [sessionPayload]);

    return completeClassName;
}
