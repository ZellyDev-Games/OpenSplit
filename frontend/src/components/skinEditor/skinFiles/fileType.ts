import type { CSSFile } from "../../../models/skin/css";

export function isImageFile(file: CSSFile): boolean {
    if (file.text) {
        return false;
    }

    return file.type.startsWith("image") || /\.(png|jpg|jpeg|gif|svg|webp)$/i.test(file.path);
}

export function isCSSFile(file: CSSFile): boolean {
    return file.type === "css" || file.path.endsWith(".css");
}
