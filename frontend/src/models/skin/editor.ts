import type { CSSFile, CSSRuleEditor, SkinCSSRule } from "./css";
import type { SkinElement } from "./element";

export type SkinEditorMode = "file" | "existing" | "create";

export interface SkinEditorTarget {
    elementId: string | null;
    file: string | null;
    selector: string | null;
    ruleId: string | null;
    parentId: string | null;
    mode: SkinEditorMode;
}

export interface SkinModel {
    name: string;
    skins: string[];
    directory: string;
    styleSheet: string;

    files: CSSFile[];
    rules: SkinCSSRule[];
    elements: SkinElement[];

    target: SkinEditorTarget;
    activeRule: CSSRuleEditor | null;

    dirty: boolean;
    revision: number;
}
