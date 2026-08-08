import { useEffect } from "react";

import { EventsEmit } from "../../../../wailsjs/runtime/runtime";

interface Props {
    value: number;
}

export default function PreviewTimer({ value }: Props) {
    useEffect(() => {
        const frame = requestAnimationFrame(() => {
            EventsEmit("timer:update", value);
        });

        return () => {
            cancelAnimationFrame(frame);
        };
    }, [value]);

    return null;
}
