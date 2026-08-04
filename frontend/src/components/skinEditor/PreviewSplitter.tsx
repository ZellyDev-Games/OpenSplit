import { Dispatch, SetStateAction, useCallback, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

import { EventsEmit } from "../../../wailsjs/runtime/runtime";
import { Comparison } from "../../hooks/splitter/useComparison";
import useElementHighlight, { Highlight } from "../../hooks/useElementHighlight";
import { usePreviewDocument } from "../../hooks/usePreviewDocument";
import { usePreviewFrame } from "../../hooks/usePreviewFrame";
import { usePreviewSelection } from "../../hooks/usePreviewSelection";
import { ConfigPayload } from "../../models/configPayload";
import SessionPayload from "../../models/sessionPayload";
import { PreviewUpdate, RuntimeElement, SkinElement } from "../../models/skinModel";
import Splitter from "../splitter/Splitter";
import CSSPreviewOverride from "./CSSPreviewOverride";

interface Props {
    initialWidth: number;

    initialHeight: number;

    skinCSS?: string;

    overrideCSS?: string;

    comparison: Comparison;

    onComparisonChange: Dispatch<SetStateAction<Comparison>>;

    sessionPayload: SessionPayload;

    configPayload: ConfigPayload;

    elements: SkinElement[];

    selectedElement: RuntimeElement | null;

    onSelect(id: string): void;

    onPreviewUpdate?(update: PreviewUpdate): void;

    disableContextMenu?: boolean;

    forceExpandAll?: boolean;
}

export default function PreviewSplitter({
    initialWidth,
    initialHeight,
    skinCSS,
    overrideCSS = "",
    sessionPayload,
    configPayload,
    elements,
    selectedElement,
    onSelect,
    onPreviewUpdate,
    comparison,
    onComparisonChange,
    disableContextMenu = false,
    forceExpandAll = false,
}: Props) {
    const { iframeRef, container } = usePreviewFrame(skinCSS);

    const workspaceRef = useRef<HTMLDivElement>(null);
    const initialScrollApplied = useRef(false);

    const [viewport, setViewport] = useState({
        width: 0,
        height: 0,
    });

    const defaultPaddingX = Math.max(0, (viewport.width - initialWidth) / 2);

    const defaultPaddingY = Math.max(0, (viewport.height - initialHeight) / 2);

    const previewDocument = container?.ownerDocument ?? null;

    const [preview, setPreview] = useState<PreviewUpdate>({
        elements: [],
        metrics: {
            splitter: new DOMRect(),

            content: new DOMRect(),

            canvasWidth: 0,

            canvasHeight: 0,

            paddingLeft: 0,

            paddingRight: 0,

            paddingTop: 0,

            paddingBottom: 0,

            overflowX: 0,

            overflowY: 0,

            splitterOffsetX: 0,

            splitterOffsetY: 0,

            hasCanvasOverflow: false,

            hasElementOverflow: false,

            overflowingElements: [],
        },
    });

    const runtimeElements = preview.elements;

    const updatePreview = useCallback(
        (update: PreviewUpdate) => {
            setPreview(update);
            onPreviewUpdate?.(update);
        },
        [onPreviewUpdate],
    );

    /*
     * Track available editor space
     */
    useLayoutEffect(() => {
        const workspace = workspaceRef.current;

        if (!workspace) {
            return;
        }

        const update = () => {
            const rect = workspace.getBoundingClientRect();

            setViewport({
                width: rect.width,
                height: rect.height,
            });
        };

        update();

        const observer = new ResizeObserver(update);

        observer.observe(workspace);

        return () => observer.disconnect();
    }, []);

    usePreviewDocument({
        document: previewDocument,
        elements,

        defaultPaddingX,
        defaultPaddingY,

        callback: updatePreview,
    });

    useLayoutEffect(() => {
        const workspace = workspaceRef.current;

        if (!workspace) {
            return;
        }

        if (!preview.metrics.hasCanvasOverflow && !preview.metrics.hasElementOverflow) {
            workspace.scrollLeft = 0;
            workspace.scrollTop = 0;

            initialScrollApplied.current = false;

            return;
        }

        workspace.scrollLeft = preview.metrics.paddingLeft - defaultPaddingX;

        workspace.scrollTop = preview.metrics.paddingTop - defaultPaddingY;

        initialScrollApplied.current = true;
    }, [
        preview.metrics.paddingLeft,
        preview.metrics.paddingTop,
        preview.metrics.hasCanvasOverflow,
        defaultPaddingX,
        defaultPaddingY,
    ]);

    usePreviewSelection({
        document: previewDocument,
        runtime: runtimeElements,
        onSelect,
        disableContextMenu,
    });

    const highlights: Highlight[] = runtimeElements.flatMap<Highlight>((element) => {
        const selected = selectedElement?.id === element.id;

        const overflow = preview.metrics.overflowingElements.includes(element.id);

        if (selected && overflow) {
            return [
                {
                    element: element.element,
                    type: "selected-overflow",
                },
            ];
        }

        if (selected) {
            return [
                {
                    element: element.element,
                    type: "selected",
                },
            ];
        }

        if (overflow) {
            return [
                {
                    element: element.element,
                    type: "overflow",
                },
            ];
        }

        return [];
    });

    useElementHighlight(highlights);

    return (
        <div ref={workspaceRef} className={"skin-preview-workspace"}>
            <div
                className="skin-preview-container"
                style={{
                    width: preview.metrics.canvasWidth || initialWidth,

                    height: preview.metrics.canvasHeight || initialHeight,
                }}
            >
                <iframe
                    ref={iframeRef}
                    className="skin-preview-frame"
                    sandbox="allow-same-origin allow-scripts"
                    style={{
                        width: preview.metrics.canvasWidth || initialWidth,

                        height: preview.metrics.canvasHeight || initialHeight,
                    }}
                />
            </div>

            {previewDocument && <CSSPreviewOverride css={overrideCSS} document={previewDocument} />}

            {container &&
                createPortal(
                    <>
                        <div
                            id="preview-canvas-wrapper"
                            style={{
                                position: "relative",

                                width: preview.metrics.canvasWidth || initialWidth,

                                height: preview.metrics.canvasHeight || initialHeight,
                            }}
                        >
                            <div
                                id="preview-canvas"
                                style={{
                                    position: "relative",

                                    width: preview.metrics.canvasWidth || initialWidth,

                                    height: preview.metrics.canvasHeight || initialHeight,

                                    flex: "none",

                                    overflow: "visible",
                                }}
                            >
                                <PreviewTimer value={sessionPayload.current_run?.total_time ?? 0} />

                                <div
                                    id="splitter"
                                    className="previewSplitter"
                                    style={{
                                        position: "absolute",

                                        left: preview.metrics.paddingLeft,
                                        top: preview.metrics.paddingTop,

                                        width: initialWidth,
                                        height: initialHeight,

                                        overflow: "visible",
                                    }}
                                >
                                    <Splitter
                                        sessionPayload={sessionPayload}
                                        configPayload={configPayload}
                                        disableContextMenu={disableContextMenu}
                                        forceExpandAll={forceExpandAll}
                                        comparison={comparison}
                                        onComparisonChange={onComparisonChange}
                                    />
                                </div>
                            </div>
                        </div>
                    </>,
                    container,
                )}
        </div>
    );
}

function PreviewTimer({ value }: { value: number }) {
    useLayoutEffect(() => {
        EventsEmit("timer:update", value);
    }, [value]);

    return null;
}
