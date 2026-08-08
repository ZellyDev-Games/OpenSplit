import type { CSSFile } from "../../../models/skin/css";

interface Props {
    file: CSSFile;
}

export default function ImagePreview({ file }: Props) {
    return (
        <>
            <h3>Preview</h3>

            <div className="skin-image-preview">
                <img src={file.url} alt={file.name} />
            </div>
        </>
    );
}
