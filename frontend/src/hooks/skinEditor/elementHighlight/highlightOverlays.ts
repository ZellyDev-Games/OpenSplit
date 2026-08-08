import { getHighlightStyle } from "./highlightStyles";
import type { Highlight } from "./highlightTypes";

const HIGHLIGHT_ATTRIBUTE = "data-highlight-owner";
const HIGHLIGHT_CLASS = "skin-element-highlight";

export function groupHighlightsByDocument(highlights: Highlight[]): Map<Document, Highlight[]> {
    const grouped = new Map<Document, Highlight[]>();

    for (const highlight of highlights) {
        const previewDocument = highlight.element.ownerDocument;

        const items = grouped.get(previewDocument) ?? [];

        items.push(highlight);

        grouped.set(previewDocument, items);
    }

    return grouped;
}

export function createHighlightOverlays(previewDocument: Document, highlights: Highlight[], owner: string): () => void {
    const window = previewDocument.defaultView;

    if (!window) {
        return () => {};
    }

    const cleanups: Array<() => void> = [];

    for (const highlight of highlights) {
        const cleanup = createHighlightOverlay(previewDocument, window, highlight, owner);

        cleanups.push(cleanup);
    }

    return () => {
        for (const cleanup of cleanups) {
            cleanup();
        }
    };
}

function createHighlightOverlay(
    previewDocument: Document,
    window: Window,
    highlight: Highlight,
    owner: string,
): () => void {
    const overlay = previewDocument.createElement("div");

    overlay.className = HIGHLIGHT_CLASS;
    overlay.dataset.highlightOwner = owner;

    Object.assign(overlay.style, {
        position: "fixed",
        pointerEvents: "none",
        zIndex: "2147483647",
        boxSizing: "border-box",
    });

    Object.assign(overlay.style, getHighlightStyle(highlight.type));

    previewDocument.body.appendChild(overlay);

    const update = (): void => {
        if (!highlight.element.isConnected) {
            overlay.remove();
            return;
        }

        updateOverlayBounds(overlay, highlight.element);
    };

    update();

    const resizeObserver = new ResizeObserver(update);

    resizeObserver.observe(highlight.element);

    window.addEventListener("resize", update);
    window.addEventListener("scroll", update, true);

    return () => {
        resizeObserver.disconnect();

        window.removeEventListener("resize", update);
        window.removeEventListener("scroll", update, true);

        overlay.remove();
    };
}

function updateOverlayBounds(overlay: HTMLElement, element: Element): void {
    const rect = element.getBoundingClientRect();

    overlay.style.left = `${rect.left}px`;
    overlay.style.top = `${rect.top}px`;
    overlay.style.width = `${rect.width}px`;
    overlay.style.height = `${rect.height}px`;
}

export function removeHighlights(previewDocument: Document, owner: string): void {
    previewDocument.querySelectorAll(`[${HIGHLIGHT_ATTRIBUTE}="${owner}"]`).forEach((node) => node.remove());
}
