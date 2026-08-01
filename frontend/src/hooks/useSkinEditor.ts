import { useCallback } from "react";

import { Dispatch } from "../../wailsjs/go/dispatcher/Service";
import { Command } from "../models/command";
import { CSSRuleEditor } from "../models/cssProperty";
import SkinModel, { CSSFile, CSSRule, SkinEditorTarget } from "../models/skinModel";
import { flattenRules } from "./flattenRules";

export interface SkinEditorState {
    model: SkinModel | null;

    target: SkinEditorTarget;

    files: CSSFile[];

    rules: CSSRuleEditor[];

    flatRules: CSSRule[];

    activeRule: CSSRuleEditor | null;

    dirty: boolean;

    selectElement(id: string): Promise<void>;

    selectFile(path: string): Promise<void>;

    selectRule(rule: CSSRuleEditor): Promise<void>;

    createRule(): Promise<void>;

    createFile(name: string): Promise<void>;

    updateRule(rule: CSSRuleEditor): Promise<void>;

    updateFile(file: string, contents: string): Promise<void>;

    save(): Promise<void>;

    reset(): Promise<void>;
}

function cssRuleToEditor(rule: SkinModel["rules"][number]): CSSRuleEditor {
    return {
        id: rule.id,

        originalFile: rule.file,

        originalId: rule.id,

        file: rule.file,

        selector: rule.selector,

        layer: rule.layer,

        body: rule.body,

        parents: rule.parents ?? [],
    };
}

export function useSkinEditor(model: SkinModel | null): SkinEditorState {
    const files = model?.files ?? [];

    const flatRules = flattenRules(model?.rules ?? []);
    const rules = flatRules.map(cssRuleToEditor);

    const target = model?.target ?? {
        elementId: null,
        file: null,
        selector: null,
        ruleId: null,
        mode: "file",
    };

    const activeRule = model?.activeRule ?? null;

    const selectElement = useCallback(async (id: string) => {
        if (!id) {
            await Dispatch(Command.CLEAR_ELEMENT, null);

            return;
        }

        await Dispatch(
            Command.SKIN_ELEMENT,
            JSON.stringify({
                element: id,
            }),
        );
    }, []);

    const selectFile = useCallback(async (file: string) => {
        if (!file) {
            await Dispatch(
                Command.SKIN_FILE,
                JSON.stringify({
                    file: "",
                }),
            );

            return;
        }

        await Dispatch(
            Command.SKIN_FILE,
            JSON.stringify({
                file,
            }),
        );
    }, []);

    const selectRule = useCallback(async (rule: CSSRuleEditor) => {
        await Dispatch(
            Command.SKIN_RULE,
            JSON.stringify({
                file: rule.file,

                ruleId: rule.id,
            }),
        );
    }, []);

    const createRule = useCallback(async () => {
        if (!target.file) {
            throw new Error("No CSS file selected");
        }

        if (!target.selector) {
            throw new Error("No element selector selected");
        }

        await Dispatch(
            Command.SKIN_CREATE_RULE,
            JSON.stringify({
                rule: {
                    id: "",
                    file: target.file,
                    selector: target.selector,
                    layer: "",
                    body: "",
                    parents: [],
                },
            }),
        );
    }, [target]);

    const updateFile = useCallback(async (file: string, contents: string) => {
        await Dispatch(
            Command.SKIN_FILE_UPDATE,
            JSON.stringify({
                file,
                contents,
            }),
        );
    }, []);

    const createFile = useCallback(async (name: string) => {
        await Dispatch(
            Command.SKIN_CREATE_FILE,
            JSON.stringify({
                name,
            }),
        );
    }, []);

    const updateRule = useCallback(async (rule: CSSRuleEditor) => {
        await Dispatch(
            Command.SKIN_RULE_UPDATE,
            JSON.stringify({
                rule,
            }),
        );
    }, []);

    const save = useCallback(async () => {
        await Dispatch(Command.SKIN_SAVE, null);
    }, []);

    const reset = useCallback(async () => {
        await Dispatch(Command.SKIN_RELOAD, null);
    }, []);

    return {
        model,

        target,

        files,

        rules,

        flatRules,

        activeRule,

        dirty: model?.dirty ?? false,

        selectElement,

        selectFile,

        selectRule,

        createRule,

        createFile,

        updateRule,

        updateFile,

        save,

        reset,
    };
}
