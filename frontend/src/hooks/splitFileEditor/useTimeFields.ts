import { ForwardedRef, useEffect, useImperativeHandle, useState } from "react";

import { msToParts, partsToMS } from "../../components/splitter/timer/timerUtils";

type UseTimeFieldsProps = {
    time: number | null;
    onChange?: (millis: number) => void;
};

export type TimeFieldsHandle = {
    getMillis(): number;
};

export function useTimeFields(props: UseTimeFieldsProps, ref: ForwardedRef<TimeFieldsHandle>) {
    const [hours, setHours] = useState("");
    const [minutes, setMinutes] = useState("");
    const [seconds, setSeconds] = useState("");
    const [centis, setCentis] = useState("");

    useEffect(() => {
        if (props.time == null || props.time < 0) {
            setHours("");
            setMinutes("");
            setSeconds("");
            setCentis("");
            return;
        }

        const p = msToParts(props.time);

        setHours(String(p.hours));
        setMinutes(String(p.minutes));
        setSeconds(String(p.seconds));
        setCentis(String(p.centis));
    }, [props.time]);

    const getMillis = (h = hours, m = minutes, s = seconds, c = centis) =>
        partsToMS({
            negative: false,
            hours: Number(h || 0),
            minutes: Number(m || 0),
            seconds: Number(s || 0),
            centis: Number(c || 0),
        });

    const emitChange = (h = hours, m = minutes, s = seconds, c = centis) => {
        props.onChange?.(getMillis(h, m, s, c));
    };

    useImperativeHandle(ref, () => ({
        getMillis,
    }));

    return {
        hours,
        minutes,
        seconds,
        centis,

        setHours,
        setMinutes,
        setSeconds,
        setCentis,

        emitChange,
    };
}
