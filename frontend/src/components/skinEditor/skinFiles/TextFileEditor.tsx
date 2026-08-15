import { useEffect, useRef } from "react";

import type { CSSFile } from "../../../models/skin/css";

interface Props {
    file: CSSFile;

    onChangeFile(file: string, contents: string): Promise<void>;
}

export default function TextFileEditor({ file, onChangeFile }: Props) {
    const textareaRef = useRef<HTMLTextAreaElement | null>(null);

    /*
     * The path identifies the actual document being edited.
     *
     * When the user switches files, replace the textarea contents.
     * Normal backend updates do not recreate the textarea contents.
     */
    const previousPath = useRef(file.path);

    useEffect(() => {
        const textarea = textareaRef.current;

        if (!textarea) {
            return;
        }

        if (previousPath.current === file.path) {
            return;
        }

        previousPath.current = file.path;

        textarea.value = file.contents;
        textarea.scrollTop = 0;
        textarea.scrollLeft = 0;
    }, [file.path, file.contents]);

    /*
     * Synchronize an external file change.
     *
     * If the backend is merely echoing the user's current edit,
     * textarea.value already equals file.contents, so nothing happens.
     *
     * This is important: we never assign textarea.value during the
     * normal typing/backend-echo cycle.
     */
    useEffect(() => {
        const textarea = textareaRef.current;

        if (!textarea) {
            return;
        }

        if (textarea.value === file.contents) {
            return;
        }

        /*
         * The model contains something different from the editor.
         * Preserve the current position when possible.
         */
        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const direction = textarea.selectionDirection;
        const scrollTop = textarea.scrollTop;
        const scrollLeft = textarea.scrollLeft;

        textarea.value = file.contents;

        textarea.setSelectionRange(start, end, direction);
        textarea.scrollTop = scrollTop;
        textarea.scrollLeft = scrollLeft;
    }, [file.contents]);

    function handleChange(event: React.ChangeEvent<HTMLTextAreaElement>): void {
        void onChangeFile(file.path, event.currentTarget.value);
    }

    return (
        <div className="panel text-file-editor">
            <h3>Text</h3>

            <textarea ref={textareaRef} defaultValue={file.contents} onChange={handleChange} />
        </div>
    );
}
