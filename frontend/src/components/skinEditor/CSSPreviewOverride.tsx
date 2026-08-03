import { useEffect } from "react";

interface Props {
    css: string;

    document: Document;
}

const previewHelperCSS = `
/*
 * Skin editor preview helpers.
 * These styles intentionally override skin CSS.
 */

.previewSplitter {
    outline: 1px solid rgba(128, 128, 128, 0.65) !important;

    outline-offset: -1px !important;
}
`;

export default function CSSPreviewOverride({ css, document }: Props) {
    useEffect(() => {
        let style = document.getElementById("opensplit-preview-overrides") as HTMLStyleElement | null;

        if (!style) {
            style = document.createElement("style");

            style.id = "opensplit-preview-overrides";

            document.head.appendChild(style);
        }

        style.textContent = `
${css}

/*
 * OpenSplit skin editor overrides.
 */
${previewHelperCSS}
`;
    }, [css, document]);

    return null;
}
