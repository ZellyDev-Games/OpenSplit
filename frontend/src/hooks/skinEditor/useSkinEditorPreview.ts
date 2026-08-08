import { Dispatch, SetStateAction, useEffect, useMemo, useState } from "react";

import { mergeSkinElements } from "../../components/skinEditor/preview/mergeSkinElements";
import { previewElements } from "../../components/skinEditor/preview/previewElements";
import { registerPreviewElements } from "../../components/skinEditor/preview/registerPreviewElements";
import { collectRuntimeCSS } from "../../components/skinEditor/preview/runtimeCSSCollector";
import { previewConfig } from "../../components/skinEditor/preview/sessions/previewBase";
import { previewSession as completedPreview } from "../../components/skinEditor/preview/sessions/previewCompletedSession";
import { previewSession as emptyPreview } from "../../components/skinEditor/preview/sessions/previewEmptySession";
import { previewSession as inProgressPreview } from "../../components/skinEditor/preview/sessions/previewInProgressSession";
import type { SkinCSSRule } from "../../models/skin/css";
import type { SkinModel } from "../../models/skin/editor";
import type { RuntimeElement, SkinElement } from "../../models/skin/element";
import type { PreviewUpdate } from "../../models/skin/preview";
import { CompareAgainst, type Comparison } from "../splitter/useComparison";
import type { SkinEditorState } from "./skinEditorTypes";

export type PreviewMode = "empty" | "running" | "completed";

export interface SkinEditorPreview {
    mode: PreviewMode;
    setMode: Dispatch<SetStateAction<PreviewMode>>;

    comparison: Comparison;
    setComparison: Dispatch<SetStateAction<Comparison>>;

    session: SkinEditorPreviewSession;

    previewUpdate: PreviewUpdate;
    setPreviewUpdate: Dispatch<SetStateAction<PreviewUpdate>>;

    runtimeElements: RuntimeElement[];

    elements: SkinElement[];

    availableIds: Set<string>;

    selectedRuntime: RuntimeElement | null;

    combinedModel: SkinModel;

    overrideCSS: string;

    hasPreviewOverflow: boolean;

    config: typeof previewConfig;
}

type SkinEditorPreviewSession = typeof emptyPreview | typeof inProgressPreview | typeof completedPreview;

function getPreviewSession(mode: PreviewMode): SkinEditorPreviewSession {
    switch (mode) {
        case "empty":
            return emptyPreview;

        case "completed":
            return completedPreview;

        case "running":
        default:
            return inProgressPreview;
    }
}

function createEmptyPreviewUpdate(): PreviewUpdate {
    return {
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

            overflowingIds: new Set<string>(),
        },
    };
}

function buildPreviewOverrideCSS(rules: SkinCSSRule[]): string {
    return rules
        .filter((rule) => rule.selector)
        .map(
            (rule) => `
${rule.selector} {
${rule.body}
}
`,
        )
        .join("\n");
}

export function useSkinEditorPreview(model: SkinModel, editor: SkinEditorState): SkinEditorPreview {
    const [mode, setMode] = useState<PreviewMode>("running");

    const [comparison, setComparison] = useState<Comparison>(CompareAgainst.Average);

    const [previewUpdate, setPreviewUpdate] = useState<PreviewUpdate>(createEmptyPreviewUpdate);

    useEffect(() => {
        void registerPreviewElements(previewElements);
    }, []);

    const session = useMemo(() => getPreviewSession(mode), [mode]);

    const runtimeRules = useMemo<SkinCSSRule[]>(() => collectRuntimeCSS(), []);

    const runtimeElements = previewUpdate.elements;

    const elements = useMemo(
        () => mergeSkinElements(previewElements, model.elements, runtimeElements),
        [model.elements, runtimeElements],
    );

    const availableIds = useMemo(() => new Set(runtimeElements.map((element) => element.id)), [runtimeElements]);

    const selectedRuntime = editor.target.elementId
        ? (runtimeElements.find((element) => element.id === editor.target.elementId) ?? null)
        : null;

    const combinedModel = useMemo<SkinModel>(
        () => ({
            ...model,
            rules: [...editor.flatRules, ...runtimeRules],
        }),
        [model, editor.flatRules, runtimeRules],
    );

    const overrideCSS = useMemo(() => buildPreviewOverrideCSS(editor.flatRules), [editor.flatRules]);

    const hasPreviewOverflow = previewUpdate.metrics.hasElementOverflow;

    return {
        mode,
        setMode,

        comparison,
        setComparison,

        session,

        previewUpdate,
        setPreviewUpdate,

        runtimeElements,

        elements,

        availableIds,

        selectedRuntime,

        combinedModel,

        overrideCSS,

        hasPreviewOverflow,

        config: previewConfig,
    };
}
