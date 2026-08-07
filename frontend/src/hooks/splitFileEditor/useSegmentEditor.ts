import { useState } from "react";

import { addChildRecursive, deleteSegmentRecursive } from "../../components/splitFileEditor/segments/SegmentActions";
import SegmentPayload from "../../models/segmentPayload";
import { cloneSegments } from "./segmentClone";
import { updateSegmentRecursive } from "./segmentUpdate";

export function useSegmentEditor(initialSegments: SegmentPayload[]) {
    const [segments, setSegments] = useState<SegmentPayload[]>(cloneSegments(initialSegments));

    const [showCumulativeTimes, setShowCumulativeTimes] = useState(false);

    const addSegment = (parent: SegmentPayload | null) => {
        if (!parent) {
            setSegments((prev) => [...prev, new SegmentPayload()]);

            return;
        }

        setSegments((prev) => addChildRecursive(prev, parent));
    };

    const updateSegment = (id: string, updater: (segment: SegmentPayload) => SegmentPayload) => {
        setSegments((prev) => updateSegmentRecursive(prev, id, updater));
    };

    const deleteSegment = (id: string) => {
        setSegments((prev) => deleteSegmentRecursive(prev, id));
    };

    return {
        segments,
        setSegments,

        showCumulativeTimes,
        setShowCumulativeTimes,

        addSegment,
        updateSegment,
        deleteSegment,
    };
}
