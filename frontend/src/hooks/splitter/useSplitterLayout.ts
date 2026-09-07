import { useEffect, useState } from "react";

import type SessionPayload from "../../models/sessionPayload";
import type { SplitterLayout } from "./useSplitterMenu";

const DEFAULT_LAYOUT: SplitterLayout = "vertical";

function getSkinDefaultLayout(element: HTMLElement): SplitterLayout {
    const layout = getComputedStyle(element).getPropertyValue("--splitter-layout").trim();

    if (layout === "horizontal" || layout === "vertical") {
        return layout;
    }

    return DEFAULT_LAYOUT;
}

function getSavedLayout(sessionPayload: SessionPayload): SplitterLayout | null {
    const layout = sessionPayload.loaded_split_file?.layout;

    if (layout === "horizontal" || layout === "vertical") {
        return layout;
    }

    return null;
}

type UseSplitterLayoutParams = {
    splitterRef: React.RefObject<HTMLDivElement | null>;
    sessionPayload: SessionPayload;
    controlledLayout?: SplitterLayout;
};

export function useSplitterLayout({ splitterRef, sessionPayload, controlledLayout }: UseSplitterLayoutParams) {
    const [initialLayout, setInitialLayout] = useState<SplitterLayout>(DEFAULT_LAYOUT);

    useEffect(() => {
        if (controlledLayout !== undefined) {
            return;
        }

        const savedLayout = getSavedLayout(sessionPayload);

        if (savedLayout !== null) {
            setInitialLayout(savedLayout);
            return;
        }

        const element = splitterRef.current;

        if (!element) {
            return;
        }

        setInitialLayout(getSkinDefaultLayout(element));
    }, [controlledLayout, sessionPayload.loaded_split_file?.layout, splitterRef]);

    return {
        initialLayout,
    };
}
