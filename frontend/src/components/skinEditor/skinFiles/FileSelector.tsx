import type { CSSFile } from "../../../models/skin/css";

interface Props {
    files: CSSFile[];

    selectedFile: string | null;

    onSelectFile(path: string): Promise<void>;

    onCreateFile(name: string): Promise<void>;
}

export default function FileSelector({ files, selectedFile, onSelectFile, onCreateFile }: Props) {
    function handleChange(value: string): void {
        if (value === "__new__") {
            const name = prompt("CSS filename");

            if (name) {
                void onCreateFile(name.endsWith(".css") ? name : `${name}.css`);
            }

            return;
        }

        void onSelectFile(value);
    }

    return (
        <>
            <h3>CSS File</h3>

            <select value={selectedFile ?? ""} onChange={(event) => handleChange(event.target.value)}>
                <option value="">Select file</option>

                {files.map((file) => (
                    <option key={file.path} value={file.path}>
                        {file.name}
                    </option>
                ))}

                <option value="__new__">Create new CSS file…</option>
            </select>
        </>
    );
}
