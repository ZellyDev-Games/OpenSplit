import { useLayoutEffect } from "react";

import { EventsEmit } from "../../../../wailsjs/runtime/runtime";

interface Props {
    value: number;
}

export default function PreviewTimer({ value }: Props) {
    useLayoutEffect(() => {
        EventsEmit("timer:update", value);
    }, [value]);

    return null;
}
