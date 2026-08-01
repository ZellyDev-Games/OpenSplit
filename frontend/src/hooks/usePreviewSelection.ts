import { useEffect } from "react";

import { RuntimeElement } from "../models/skinModel";

interface Props {
    document: Document | null;

    runtime: RuntimeElement[];

    onSelect(id: string): void;

    disableContextMenu: boolean;
}

export function usePreviewSelection({ document, runtime, onSelect, disableContextMenu }: Props) {
    useEffect(() => {
        if (!document) {
            return;
        }

        const win = document.defaultView;

        if (!win) {
            return;
        }

        const lookup = new WeakMap<Element, string>();

        for (const element of runtime) {
            lookup.set(element.element, element.id);
        }

        const click = (event: MouseEvent) => {
            if (!(event.target instanceof win.Element)) {
                return;
            }

            let current: Element | null = event.target;

            while (current) {
                const id = lookup.get(current);

                if (id) {
                    event.preventDefault();
                    event.stopPropagation();
                    event.stopImmediatePropagation();

                    onSelect(id);

                    return;
                }

                current = current.parentElement;
            }
        };

        document.addEventListener("click", click, true);

        return () => {
            document.removeEventListener("click", click, true);
        };
    }, [document, runtime, onSelect]);

    useEffect(() => {
        if (!document || !disableContextMenu) {
            return;
        }

        const handler = (event: MouseEvent) => {
            event.preventDefault();
        };

        document.addEventListener("contextmenu", handler, true);

        return () => {
            document.removeEventListener("contextmenu", handler, true);
        };
    }, [document, disableContextMenu]);
}
