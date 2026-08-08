import { Dispatch, SetStateAction, useCallback, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

import { EventsEmit } from "../../../wailsjs/runtime/runtime";
import { createPreviewHighlights } from "../../hooks/skinEditor/previewDocument/createPreviewHighlights";
import { createEmptyPreviewUpdate } from "../../hooks/skinEditor/previewDocument/previewDefaults";
import { usePreviewDocument } from "../../hooks/skinEditor/usePreviewDocument";
import { Comparison } from "../../hooks/splitter/useComparison";
import useElementHighlight from "../../hooks/useElementHighlight";
import { usePreviewFrame } from "../../hooks/usePreviewFrame";
import { usePreviewSelection } from "../../hooks/usePreviewSelection";
import { ConfigPayload } from "../../models/configPayload";
import SessionPayload from "../../models/sessionPayload";
import type { RuntimeElement, SkinElement } from "../../models/skin/element";
import type { PreviewUpdate } from "../../models/skin/preview";
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

    const [viewport, setViewport] = useState({
        width: 0,
        height: 0,
    });

    const [preview, setPreview] = useState<PreviewUpdate>(createEmptyPreviewUpdate);

    const runtimeElements = preview.elements;

    const defaultPaddingX = Math.max(0, (viewport.width - initialWidth) / 2);

    const defaultPaddingY = Math.max(0, (viewport.height - initialHeight) / 2);

    const previewDocument = container?.ownerDocument ?? null;

    const updatePreview = useCallback(
        (update: PreviewUpdate) => {
            setPreview(update);
            onPreviewUpdate?.(update);
        },
        [onPreviewUpdate],
    );

    useLayoutEffect(() => {
        const workspace = workspaceRef.current;

        if (!workspace) {
            return;
        }

        const updateViewport = () => {
            const rect = workspace.getBoundingClientRect();

            setViewport({
                width: rect.width,
                height: rect.height,
            });
        };

        updateViewport();

        const observer = new ResizeObserver(updateViewport);

        observer.observe(workspace);

        return () => {
            observer.disconnect();
        };
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

            return;
        }

        workspace.scrollLeft = preview.metrics.paddingLeft - defaultPaddingX;

        workspace.scrollTop = preview.metrics.paddingTop - defaultPaddingY;
    }, [
        preview.metrics.paddingLeft,
        preview.metrics.paddingTop,
        preview.metrics.hasCanvasOverflow,
        preview.metrics.hasElementOverflow,
        defaultPaddingX,
        defaultPaddingY,
    ]);

    usePreviewSelection({
        document: previewDocument,
        runtime: runtimeElements,
        onSelect,
        disableContextMenu,
    });

    const highlights = createPreviewHighlights(runtimeElements, preview.metrics, selectedElement);

    useElementHighlight(highlights);

    const canvasWidth = preview.metrics.canvasWidth || initialWidth;

    const canvasHeight = preview.metrics.canvasHeight || initialHeight;

    return (
        <div ref={workspaceRef} className="skin-preview-workspace">
            <div
                className="skin-preview-container"
                style={{
                    width: canvasWidth,
                    height: canvasHeight,
                }}
            >
                <iframe
                    ref={iframeRef}
                    className="skin-preview-frame"
                    sandbox="allow-same-origin allow-scripts"
                    style={{
                        width: canvasWidth,
                        height: canvasHeight,
                    }}
                />
            </div>

            {previewDocument && <CSSPreviewOverride css={overrideCSS} document={previewDocument} />}

            {container &&
                createPortal(
                    <PreviewCanvas
                        width={canvasWidth}
                        height={canvasHeight}
                        splitterWidth={initialWidth}
                        splitterHeight={initialHeight}
                        paddingLeft={preview.metrics.paddingLeft}
                        paddingTop={preview.metrics.paddingTop}
                        sessionPayload={sessionPayload}
                        configPayload={configPayload}
                        disableContextMenu={disableContextMenu}
                        forceExpandAll={forceExpandAll}
                        comparison={comparison}
                        onComparisonChange={onComparisonChange}
                    />,
                    container,
                )}
        </div>
    );
}

interface PreviewCanvasProps {
    width: number;
    height: number;

    splitterWidth: number;
    splitterHeight: number;

    paddingLeft: number;
    paddingTop: number;

    sessionPayload: SessionPayload;
    configPayload: ConfigPayload;

    disableContextMenu: boolean;
    forceExpandAll: boolean;

    comparison: Comparison;
    onComparisonChange: Dispatch<SetStateAction<Comparison>>;
}

function PreviewCanvas({
    width,
    height,
    splitterWidth,
    splitterHeight,
    paddingLeft,
    paddingTop,
    sessionPayload,
    configPayload,
    disableContextMenu,
    forceExpandAll,
    comparison,
    onComparisonChange,
}: PreviewCanvasProps) {
    return (
        <div
            id="preview-canvas-wrapper"
            style={{
                position: "relative",
                width,
                height,
            }}
        >
            <div
                id="preview-canvas"
                style={{
                    position: "relative",
                    width,
                    height,
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
                        left: paddingLeft,
                        top: paddingTop,
                        width: splitterWidth,
                        height: splitterHeight,
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
    );
}

function PreviewTimer({ value }: { value: number }) {
    useLayoutEffect(() => {
        EventsEmit("timer:update", value);
    }, [value]);

    return null;
}
