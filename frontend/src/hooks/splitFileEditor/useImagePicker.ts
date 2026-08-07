import { ChangeEvent, useRef } from "react";

export default function useImagePicker(onChange: (image: string) => void) {
    const inputRef = useRef<HTMLInputElement>(null);

    const handleUpload = (e: ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];

        if (!file) {
            return;
        }

        const reader = new FileReader();

        reader.onload = () => {
            onChange(reader.result as string);
        };

        reader.readAsDataURL(file);

        e.target.value = "";
    };

    return {
        inputRef,
        handleUpload,
    };
}
