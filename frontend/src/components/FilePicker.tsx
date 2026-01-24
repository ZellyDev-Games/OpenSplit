import React from "react";

import { PickAutoSplitterFile } from "../../wailsjs/go/dispatcher/Service";

type FilePickerProps = {
    setFilename: React.Dispatch<React.SetStateAction<string>>;
    fileName: string;
};

export function FilePicker({ setFilename, fileName }: FilePickerProps) {
    const openFileDialog = async () => {
        setFilename(await PickAutoSplitterFile());
    };

    const clearFile = () => {
        setFilename("");
    };

    return (
        <>
            <label htmlFor="autosplitter_file">Autosplitter File</label>
            <button style={{ marginLeft: 10 }} type="button" onClick={openFileDialog}>
                Choose File
            </button>

            <button style={{ marginLeft: 10 }} type="button" onClick={clearFile}>
                Clear File
            </button>

            <input id="autosplitter_file" readOnly value={fileName} />
        </>
    );
}
