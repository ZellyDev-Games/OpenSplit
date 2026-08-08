import { type RefObject, useLayoutEffect, useRef, useState } from "react";

const PREVIEW_ROOT_ID = "App";

const PREVIEW_HTML = `
    <style>
        html,
        body {
            margin: 0;
            padding: 0;

            width: 100%;
            height: 100%;

            overflow: visible;
        }

        #root {
            width: 100%;
            height: 100%;
        }

        #App {
            position: relative;

            width: 100%;
            height: 100%;

            overflow: visible;
        }
    </style>
`;

export interface PreviewFrame {
    iframeRef: RefObject<HTMLIFrameElement | null>;
    container: HTMLElement | null;
}

function getSkinBaseUrl(skinCSS?: string): string {
    if (!skinCSS) {
        return "";
    }

    const index = skinCSS.lastIndexOf("/");

    return index >= 0 ? skinCSS.substring(0, index + 1) : "";
}

function createPreviewDocument(document: Document, skinCSS?: string): void {
    const skinBase = getSkinBaseUrl(skinCSS);

    document.open();

    document.write(`
        <!doctype html>
        <html>
            <head>
                <meta charset="utf-8">

                ${skinBase ? `<base href="${skinBase}">` : ""}

                ${skinCSS ? `<link rel="stylesheet" href="${skinCSS}">` : ""}

                ${PREVIEW_HTML}
            </head>

            <body>
                <div id="root">
                    <div id="${PREVIEW_ROOT_ID}"></div>
                </div>
            </body>
        </html>
    `);

    document.close();
}

function copyApplicationStyles(sourceDocument: Document, targetDocument: Document): void {
    for (const node of Array.from(sourceDocument.head.children)) {
        if (node instanceof HTMLStyleElement) {
            copyStyleElement(node, targetDocument);
            continue;
        }

        if (node instanceof HTMLLinkElement && node.rel === "stylesheet") {
            copyStylesheetLink(node, targetDocument);
        }
    }
}

function copyStyleElement(source: HTMLStyleElement, targetDocument: Document): void {
    const style = targetDocument.createElement("style");

    style.textContent = source.textContent;

    targetDocument.head.appendChild(style);
}

function copyStylesheetLink(source: HTMLLinkElement, targetDocument: Document): void {
    const link = targetDocument.createElement("link");

    link.rel = "stylesheet";
    link.href = source.href;

    targetDocument.head.appendChild(link);
}

function initializePreviewFrame(iframe: HTMLIFrameElement, skinCSS?: string): HTMLElement | null {
    const document = iframe.contentDocument;

    if (!document) {
        return null;
    }

    createPreviewDocument(document, skinCSS);
    copyApplicationStyles(window.document, document);

    return document.getElementById(PREVIEW_ROOT_ID);
}

export function usePreviewFrame(skinCSS?: string): PreviewFrame {
    const iframeRef = useRef<HTMLIFrameElement>(null);
    const [container, setContainer] = useState<HTMLElement | null>(null);

    useLayoutEffect(() => {
        const iframe = iframeRef.current;

        if (!iframe) {
            return;
        }

        const initialize = () => {
            const root = initializePreviewFrame(iframe, skinCSS);

            setContainer(root);
        };

        iframe.addEventListener("load", initialize);

        iframe.src = "about:blank";

        return () => {
            iframe.removeEventListener("load", initialize);
            setContainer(null);
        };
    }, [skinCSS]);

    return {
        iframeRef,
        container,
    };
}
