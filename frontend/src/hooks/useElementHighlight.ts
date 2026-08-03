import { useEffect, useRef } from "react";

export type HighlightType = "selected" | "overflow" | "selected-overflow";

export interface Highlight {
    element: Element;
    type: HighlightType;
}

export default function useElementHighlight(highlights: Highlight[]) {
    const ownerRef = useRef<string>(crypto.randomUUID());

    useEffect(() => {
        const owner = ownerRef.current;

        if (highlights.length === 0) {
            removeHighlights(document, ownerRef.current);
            return;
        }

        const grouped = new Map<Document, Highlight[]>();

        for (const highlight of highlights) {
            const document = highlight.element.ownerDocument;

            const list = grouped.get(document) ?? [];

            list.push(highlight);

            grouped.set(document, list);
        }

        const cleanups: (() => void)[] = [];

        for (const [document, items] of grouped) {
            removeHighlights(document, owner);

            const win = document.defaultView;

            if (!win) {
                continue;
            }

            for (const item of items) {
                const overlay = document.createElement("div");

                overlay.className = "skin-element-highlight";
                overlay.dataset.highlightOwner = owner;

                Object.assign(overlay.style, {
                    position: "fixed",
                    pointerEvents: "none",
                    zIndex: "2147483647",
                    boxSizing: "border-box",
                });

                switch (item.type) {
                    case "selected":
                        Object.assign(overlay.style, {
                            border: "3px solid #00b7ff",
                            background: "rgba(0,183,255,0.20)",
                        });
                        break;

                    case "overflow":
                        Object.assign(overlay.style, {
                            border: "3px solid #ff3333",
                            background: "rgba(255,0,0,0.15)",
                        });
                        break;

                    case "selected-overflow":
                        Object.assign(overlay.style, {
                            border: "3px solid #ff3333",
                            outline: "3px solid #00b7ff",
                            outlineOffset: "-6px",
                            background: "rgba(255,0,0,0.20)",
                        });
                        break;
                }

                document.body.appendChild(overlay);

                const update = () => {
                    if (!item.element.isConnected) {
                        overlay.remove();
                        return;
                    }

                    const rect = item.element.getBoundingClientRect();

                    overlay.style.left = `${rect.left}px`;
                    overlay.style.top = `${rect.top}px`;
                    overlay.style.width = `${rect.width}px`;
                    overlay.style.height = `${rect.height}px`;
                };

                update();

                const observer = new ResizeObserver(update);

                observer.observe(item.element);

                win.addEventListener("resize", update);
                win.addEventListener("scroll", update, true);

                cleanups.push(() => {
                    observer.disconnect();

                    win.removeEventListener("resize", update);
                    win.removeEventListener("scroll", update, true);

                    overlay.remove();
                });
            }
        }

        return () => {
            cleanups.forEach((cleanup) => cleanup());

            for (const document of grouped.keys()) {
                removeHighlights(document, owner);
            }
        };
    }, [highlights]);
}

function removeHighlights(document: Document, owner: string) {
    document.querySelectorAll(`[data-highlight-owner="${owner}"]`).forEach((node) => node.remove());
}
