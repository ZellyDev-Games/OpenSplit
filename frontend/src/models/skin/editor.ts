import type { CSSFile, CSSRule, CSSRuleEditor } from "./css";
import type { SkinElement } from "./element";

export type SkinEditorMode = "file" | "existing" | "create";

export interface SkinEditorTarget {
    elementId: string | null;
    file: string | null;
    selector: string | null;
    ruleId: string | null;
    mode: SkinEditorMode;
}

export interface SkinModel {
    name: string;
    skins: string[];
    directory: string;
    styleSheet: string;

    files: CSSFile[];
    rules: CSSRule[];
    elements: SkinElement[];

    target: SkinEditorTarget;
    activeRule: CSSRuleEditor | null;

    dirty: boolean;
    revision: number;
}
