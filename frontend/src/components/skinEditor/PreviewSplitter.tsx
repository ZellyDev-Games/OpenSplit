import { Dispatch, SetStateAction, useCallback, useState } from "react";
import { createPortal } from "react-dom";

import { createPreviewHighlights } from "../../hooks/skinEditor/previewDocument/createPreviewHighlights";
import { createEmptyPreviewUpdate } from "../../hooks/skinEditor/previewDocument/previewDefaults";
import { usePreviewScroll } from "../../hooks/skinEditor/previewSplitter/usePreviewScroll";
import { usePreviewViewport } from "../../hooks/skinEditor/previewSplitter/usePreviewViewport";
import useElementHighlight from "../../hooks/skinEditor/useElementHighlight";
import { usePreviewDocument } from "../../hooks/skinEditor/usePreviewDocument";
import { usePreviewFrame } from "../../hooks/skinEditor/usePreviewFrame";
import { usePreviewSelection } from "../../hooks/skinEditor/usePreviewSelection";
import { Comparison } from "../../hooks/splitter/useComparison";
import { SplitterLayout } from "../../hooks/splitter/useSplitterMenu";
import { ConfigPayload } from "../../models/configPayload";
import SessionPayload from "../../models/sessionPayload";
import type { RuntimeElement, SkinElement } from "../../models/skin/element";
import type { PreviewUpdate } from "../../models/skin/preview";
import CSSPreviewOverride from "./CSSPreviewOverride";
import PreviewCanvas from "./previewSplitter/PreviewCanvas";

interface Props {
    initialWidth: number;
    initialHeight: number;

    skinCSS?: string;
    overrideCSS?: string;

    comparison: Comparison;
    onComparisonChange: Dispatch<SetStateAction<Comparison>>;

    layout: SplitterLayout;
    onLayoutChange: Dispatch<SetStateAction<SplitterLayout>>;

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
    layout,
    onLayoutChange,
    disableContextMenu = false,
    forceExpandAll = false,
}: Props) {
    const { iframeRef, container } = usePreviewFrame(skinCSS);

    const viewport = usePreviewViewport();

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

    usePreviewDocument({
        document: previewDocument,
        elements,
        defaultPaddingX,
        defaultPaddingY,
        callback: updatePreview,
    });

    usePreviewScroll({
        workspace: viewport.workspace,
        metrics: preview.metrics,
        defaultPaddingX,
        defaultPaddingY,
    });

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
        <div ref={viewport.workspaceRef} className="skin-preview-workspace">
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
                        layout={layout}
                        onLayoutChange={onLayoutChange}
                    />,
                    container,
                )}
        </div>
    );
}
