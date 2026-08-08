import { type RefObject, useLayoutEffect, useRef, useState } from "react";

interface PreviewViewport {
    width: number;
    height: number;
}

interface UsePreviewViewportResult extends PreviewViewport {
    workspaceRef: RefObject<HTMLDivElement | null>;
    workspace: HTMLDivElement | null;
}

export function usePreviewViewport(): UsePreviewViewportResult {
    const workspaceRef = useRef<HTMLDivElement>(null);

    const [viewport, setViewport] = useState<PreviewViewport>({
        width: 0,
        height: 0,
    });

    const [workspace, setWorkspace] = useState<HTMLDivElement | null>(null);

    useLayoutEffect(() => {
        const element = workspaceRef.current;

        if (!element) {
            return;
        }

        setWorkspace(element);

        const updateViewport = () => {
            const rect = element.getBoundingClientRect();

            setViewport({
                width: rect.width,
                height: rect.height,
            });
        };

        updateViewport();

        const observer = new ResizeObserver(updateViewport);

        observer.observe(element);

        return () => {
            observer.disconnect();
            setWorkspace(null);
        };
    }, []);

    return {
        ...viewport,
        workspaceRef,
        workspace,
    };
}
