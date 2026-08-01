import { useEffect } from "react";

interface Props {
    css: string;

    document: Document;
}

export default function CSSPreviewOverride({ css, document }: Props) {
    useEffect(() => {
        let style = document.getElementById("opensplit-preview-overrides") as HTMLStyleElement | null;

        if (!style) {
            style = document.createElement("style");
            style.id = "opensplit-preview-overrides";
            document.head.appendChild(style);
        }

        style.textContent = css;
    }, [css, document]);

    return null;
}
