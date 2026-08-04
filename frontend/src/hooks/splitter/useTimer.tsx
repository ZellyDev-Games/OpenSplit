import { useEffect, useState } from "react";

import { EventsOn } from "../../../wailsjs/runtime";

export function useTimer(initial = 0) {
    const [time, setTime] = useState(initial);

    useEffect(() => {
        return EventsOn("timer:update", (value: number) => {
            setTime(value);
        });
    }, []);

    return time;
}
