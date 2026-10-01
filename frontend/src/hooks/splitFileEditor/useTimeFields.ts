import { ForwardedRef, useEffect, useImperativeHandle, useState } from "react";

import { msToParts } from "../../components/splitter/timer/timerUtils";

type UseTimeFieldsProps = {
    time: number | null;
    onChange?: (millis: number) => void;
    showMilliseconds?: boolean;
};

export type TimeFieldsHandle = {
    getMillis(): number;
};

export function useTimeFields(props: UseTimeFieldsProps, ref: ForwardedRef<TimeFieldsHandle>) {
    const [hours, setHours] = useState("");
    const [minutes, setMinutes] = useState("");
    const [seconds, setSeconds] = useState("");
    const [fraction, setFraction] = useState("");

    useEffect(() => {
        if (props.time == null || props.time < 0) {
            setHours("");
            setMinutes("");
            setSeconds("");
            setFraction("");
            return;
        }

        const p = msToParts(props.time);

        setHours(String(p.hours));
        setMinutes(String(p.minutes));
        setSeconds(String(p.seconds));
        setFraction(props.showMilliseconds ? String(props.time % 1000).padStart(3, "0") : String(p.centis));
    }, [props.time, props.showMilliseconds]);

    const getMillis = (h = hours, m = minutes, s = seconds, f = fraction) => {
        const base = Number(h || 0) * 3600000 + Number(m || 0) * 60000 + Number(s || 0) * 1000;
        return base + (props.showMilliseconds ? Number((f || "0").padEnd(3, "0").slice(0, 3)) : Number(f || 0) * 10);
    };

    const emitChange = (h = hours, m = minutes, s = seconds, f = fraction) => {
        props.onChange?.(getMillis(h, m, s, f));
    };

    useImperativeHandle(ref, () => ({
        getMillis,
    }));

    return {
        hours,
        minutes,
        seconds,
        fraction,

        setHours,
        setMinutes,
        setSeconds,
        setFraction,

        emitChange,
    };
}
