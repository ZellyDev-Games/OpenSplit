import { useEffect } from "react";

import { RuntimeElement } from "../../models/skin/element";

interface PreviewSelectionOptions {
    document: Document | null;
    runtime: RuntimeElement[];
    onSelect(id: string): void;
    disableContextMenu: boolean;
}

function createElementLookup(runtime: RuntimeElement[]): WeakMap<Element, string> {
    const lookup = new WeakMap<Element, string>();

    for (const { element, id } of runtime) {
        lookup.set(element, id);
    }

    return lookup;
}

function findRuntimeElementId(target: Element, lookup: WeakMap<Element, string>): string | null {
    let current: Element | null = target;

    while (current) {
        const id = lookup.get(current);

        if (id) {
            return id;
        }

        current = current.parentElement;
    }

    return null;
}

function handlePreviewClick(document: Document, runtime: RuntimeElement[], onSelect: (id: string) => void): () => void {
    const win = document.defaultView;

    if (!win) {
        return () => {};
    }

    const lookup = createElementLookup(runtime);

    const handleClick = (event: MouseEvent) => {
        if (!(event.target instanceof win.Element)) {
            return;
        }

        const id = findRuntimeElementId(event.target, lookup);

        if (!id) {
            return;
        }

        event.preventDefault();
        event.stopPropagation();
        event.stopImmediatePropagation();

        onSelect(id);
    };

    document.addEventListener("click", handleClick, true);

    return () => {
        document.removeEventListener("click", handleClick, true);
    };
}

function handleContextMenu(document: Document): () => void {
    const handler = (event: MouseEvent) => {
        event.preventDefault();
    };

    document.addEventListener("contextmenu", handler, true);

    return () => {
        document.removeEventListener("contextmenu", handler, true);
    };
}

export function usePreviewSelection({
    document,
    runtime,
    onSelect,
    disableContextMenu,
}: PreviewSelectionOptions): void {
    useEffect(() => {
        if (!document) {
            return;
        }

        return handlePreviewClick(document, runtime, onSelect);
    }, [document, runtime, onSelect]);

    useEffect(() => {
        if (!document || !disableContextMenu) {
            return;
        }

        return handleContextMenu(document);
    }, [document, disableContextMenu]);
}
