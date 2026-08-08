import { useEffect, useRef } from "react";

import {
    createHighlightOverlays,
    groupHighlightsByDocument,
    removeHighlights,
} from "./elementHighlight/highlightOverlays";
import type { Highlight } from "./elementHighlight/highlightTypes";

export type { Highlight, HighlightType } from "./elementHighlight/highlightTypes";

export default function useElementHighlight(highlights: Highlight[]): void {
    const ownerRef = useRef<string>(crypto.randomUUID());

    useEffect(() => {
        const owner = ownerRef.current;

        if (highlights.length === 0) {
            removeHighlights(document, owner);
            return;
        }

        const grouped = groupHighlightsByDocument(highlights);
        const cleanups: Array<() => void> = [];

        for (const [previewDocument, items] of grouped) {
            removeHighlights(previewDocument, owner);

            const cleanup = createHighlightOverlays(previewDocument, items, owner);

            cleanups.push(cleanup);
        }

        return () => {
            for (const cleanup of cleanups) {
                cleanup();
            }

            for (const previewDocument of grouped.keys()) {
                removeHighlights(previewDocument, owner);
            }
        };
    }, [highlights]);
}
