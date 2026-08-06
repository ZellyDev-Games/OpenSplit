/**
 * SegmentIconPicker allows a segment icon to be selected, replaced, or removed.
 *
 * Images are stored as base64 data URLs so they can be embedded directly into
 * the split file without requiring external asset management.
 */

import { CSSProperties } from "react";

import addIcon from "../../../assets/images/add.png";
import removeIcon from "../../../assets/images/remove.png";
import useImagePicker from "../../../hooks/splitFileEditor/useImagePicker";

type SegmentIconPickerProps = {
    icon: string;
    onChange: (icon: string) => void;
};

const iconStyle: CSSProperties = {
    width: 24,
    height: 24,
    cursor: "pointer",
    border: "1px solid #666",
    borderRadius: 2,
    padding: 2,
    boxSizing: "border-box",
};

const previewStyle: CSSProperties = {
    width: 24,
    height: 24,
    objectFit: "contain",
    border: "1px solid #666",
    borderRadius: 2,
    cursor: "pointer",
};

export default function SegmentIconPicker({ icon, onChange }: SegmentIconPickerProps) {
    const { inputRef, handleUpload } = useImagePicker(onChange);

    return (
        <>
            <input ref={inputRef} hidden type="file" accept="image/*" onChange={handleUpload} />

            {!icon ? (
                <img src={addIcon} alt="" style={iconStyle} onClick={() => inputRef.current?.click()} />
            ) : (
                <>
                    <img src={icon} alt="" style={previewStyle} onClick={() => inputRef.current?.click()} />

                    <img src={removeIcon} alt="" style={iconStyle} onClick={() => onChange("")} />
                </>
            )}
        </>
    );
}
