import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { Command } from "../../models/command";
import type { CSSRuleEditor } from "../../models/skin/css";
import type { SkinEditorTarget } from "../../models/skin/editor";

export async function selectElement(id: string): Promise<void> {
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
}

export async function selectFile(file: string): Promise<void> {
    await Dispatch(
        Command.SKIN_FILE,
        JSON.stringify({
            file,
        }),
    );
}

export async function selectRule(rule: CSSRuleEditor): Promise<void> {
    await Dispatch(
        Command.SKIN_RULE,
        JSON.stringify({
            file: rule.file,
            ruleId: rule.id,
        }),
    );
}

export async function createRule(target: SkinEditorTarget): Promise<void> {
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
                parentId: target.parentId ?? "",
                parentAtRule: "",
                body: "",
                create: true,
                delete: false,
                originalFile: "",
                originalId: "",
            },
        }),
    );
}

export async function createFile(name: string): Promise<void> {
    await Dispatch(
        Command.SKIN_CREATE_FILE,
        JSON.stringify({
            name,
        }),
    );
}

export async function updateRule(rule: CSSRuleEditor): Promise<void> {
    await Dispatch(
        Command.SKIN_RULE_UPDATE,
        JSON.stringify({
            rule,
        }),
    );
}

export async function updateFile(file: string, contents: string): Promise<void> {
    await Dispatch(
        Command.SKIN_FILE_UPDATE,
        JSON.stringify({
            file,
            contents,
        }),
    );
}

export async function save(): Promise<void> {
    await Dispatch(Command.SKIN_SAVE, null);
}

export async function reset(): Promise<void> {
    await Dispatch(Command.SKIN_RELOAD, null);
}
