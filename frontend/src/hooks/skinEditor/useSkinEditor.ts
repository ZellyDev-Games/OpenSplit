import { useCallback } from "react";

import type { CSSFile, CSSRuleEditor, SkinCSSRule } from "../../models/skin/css";
import type { SkinEditorTarget, SkinModel } from "../../models/skin/editor";
import { cssRuleToEditor } from "./cssRuleToEditor";
import { flattenRules } from "./flattenRules";
import {
    createFile,
    createRule,
    reset,
    save,
    selectElement,
    selectFile,
    selectRule,
    updateFile,
    updateRule,
} from "./skinEditorActions";
import type { SkinEditorState } from "./skinEditorTypes";

const EMPTY_TARGET: SkinEditorTarget = {
    elementId: null,
    file: null,
    selector: null,
    ruleId: null,
    parentId: null,
    mode: "file",
};

export function useSkinEditor(model: SkinModel | null): SkinEditorState {
    const files: CSSFile[] = model?.files ?? [];

    const flatRules: SkinCSSRule[] = flattenRules(model?.rules ?? []);

    const rules: CSSRuleEditor[] = flatRules.map(cssRuleToEditor);

    const target = model?.target ?? EMPTY_TARGET;
    const activeRule = model?.activeRule ?? null;
    const dirty = model?.dirty ?? false;

    const handleSelectElement = useCallback((id: string) => selectElement(id), []);

    const handleSelectFile = useCallback((file: string) => selectFile(file), []);

    const handleSelectRule = useCallback((rule: CSSRuleEditor) => selectRule(rule), []);

    const handleCreateRule = useCallback(() => createRule(target), [target]);

    const handleCreateFile = useCallback((name: string) => createFile(name), []);

    const handleUpdateRule = useCallback((rule: CSSRuleEditor) => updateRule(rule), []);

    const handleUpdateFile = useCallback((file: string, contents: string) => updateFile(file, contents), []);

    const handleSave = useCallback(() => save(), []);

    const handleReset = useCallback(() => reset(), []);

    return {
        model,
        target,
        files,
        rules,
        flatRules,
        activeRule,
        dirty,

        selectElement: handleSelectElement,
        selectFile: handleSelectFile,
        selectRule: handleSelectRule,

        createRule: handleCreateRule,
        createFile: handleCreateFile,

        updateRule: handleUpdateRule,
        updateFile: handleUpdateFile,

        save: handleSave,
        reset: handleReset,
    };
}
