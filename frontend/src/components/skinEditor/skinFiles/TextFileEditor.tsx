import { useEffect, useState } from "react";

import type { CSSFile } from "../../../models/skin/css";

interface Props {
    file: CSSFile;

    onChangeFile(file: string, contents: string): Promise<void>;
}

export default function TextFileEditor({ file, onChangeFile }: Props) {
    const [contents, setContents] = useState(file.contents);

    useEffect(() => {
        setContents(file.contents);
    }, [file.path, file.contents]);

    function handleChange(value: string): void {
        setContents(value);

        void onChangeFile(file.path, value);
    }

    return (
        <div className="panel text-file-editor">
            <h3>Text</h3>

            <textarea value={contents} onChange={(event) => handleChange(event.target.value)} />
        </div>
    );
}
