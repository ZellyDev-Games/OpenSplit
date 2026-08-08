import { CSSFile, CSSRuleEditor, SkinCSSRule } from "../../models/skin/css";
import { SkinEditorTarget, SkinModel } from "../../models/skin/editor";

export interface SkinEditorState {
    model: SkinModel | null;
    target: SkinEditorTarget;
    files: CSSFile[];
    rules: CSSRuleEditor[];
    flatRules: SkinCSSRule[];
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
