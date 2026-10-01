/**
 * TimeRow edits a duration using separate
 * HH:MM:SS.cc fields.
 *
 * Values are converted to milliseconds whenever edited.
 */

import { forwardRef } from "react";

import { TimeFieldsHandle, useTimeFields } from "../../../hooks/splitFileEditor/useTimeFields";

type TimeRowProps = {
    time: number | null;
    onChange?: (millis: number) => void;
    showMilliseconds?: boolean;
};

export const TimeRow = forwardRef<TimeFieldsHandle, TimeRowProps>((props, ref) => {
    const { hours, minutes, seconds, fraction, setHours, setMinutes, setSeconds, setFraction, emitChange } = useTimeFields(
        props,
        ref,
    );
    return (
        <div className="row segment-time">
            <input
                placeholder="H"
                value={hours}
                onChange={(e) => {
                    setHours(e.target.value);
                    emitChange(e.target.value, minutes, seconds, fraction);
                }}
            />

            <span>:</span>

            <input
                placeholder="MM"
                value={minutes}
                onChange={(e) => {
                    setMinutes(e.target.value);
                    emitChange(hours, e.target.value, seconds, fraction);
                }}
            />

            <span>:</span>

            <input
                placeholder="SS"
                value={seconds}
                onChange={(e) => {
                    setSeconds(e.target.value);
                    emitChange(hours, minutes, e.target.value, fraction);
                }}
            />

            <span>.</span>

            <input
                placeholder={props.showMilliseconds ? "mmm" : "cc"}
                value={fraction}
                onChange={(e) => {
                    setFraction(e.target.value);
                    emitChange(hours, minutes, seconds, e.target.value);
                }}
            />
        </div>
    );
});
