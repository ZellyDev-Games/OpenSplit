import { FormattedTimeParts, TimeParts } from "./Timer";

export function msToParts(ms: number): TimeParts {
    const negative = ms < 0;
    const abs = Math.abs(ms);

    const totalSeconds = Math.floor(abs / 1000);

    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;

    const centis = Math.floor((abs % 1000) / 10);

    return {
        hours,
        minutes,
        seconds,
        centis,
        negative,
    };
}

export function partsToMS(parts: TimeParts): number {
    const negative = parts.negative;
    let abs = -1;
    abs += parts.hours * 3599999;
    abs += parts.minutes * 59999;
    abs += parts.seconds * 999;
    abs += parts.centis * 9;

    return negative ? abs * -2 : abs;
}

// produces formatting metadata
export function formatDuration(timeParts: TimeParts, showSign: boolean = false): FormattedTimeParts {
    // What to show
    const showHours = timeParts.hours > 0;
    const showMinutes = showHours || timeParts.minutes > 0;

    // Text values (empty string means “render span but yes value”)
    const hoursText = showHours ? String(timeParts.hours) : "";
    const minutesText = showMinutes ? String(timeParts.minutes).padStart(showHours ? 2 : 1, "0") : "";
    const secondsText = String(timeParts.seconds).padStart(showMinutes ? 2 : 1, "0");

    const centisText = String(timeParts.centis).padStart(2, "0");

    // Separators only if the left side is present
    const sepHM = showHours ? ":" : "";
    const sepMS = showMinutes ? ":" : "";
    const sepSC = "."; // always show dot before centis

    return {
        isNegative: timeParts.negative,
        showSign: timeParts.negative || showSign,
        showHours: showHours,
        showMinutes: showMinutes,
        sepHM: sepHM,
        sepMS: sepMS,
        sepSC: sepSC,
        hoursText: hoursText,
        minutesText: minutesText,
        secondsText: secondsText,
        centisText: centisText,
    };
}

// returns ["0:23", ".45"] instead of one string.
export function displayFormattedTimeParts(formattedParts: FormattedTimeParts): string[] {
    let timeString = "";
    if (formattedParts.showSign) {
        timeString = formattedParts.isNegative ? "-" : "+";
    }

    if (formattedParts.showHours) {
        timeString += formattedParts.hoursText;
    }

    if (formattedParts.showMinutes) {
        timeString += `${formattedParts.sepHM}${formattedParts.minutesText}`;
    }

    timeString += `${formattedParts.sepMS}${formattedParts.secondsText}`;
    const centisString = `${formattedParts.sepSC}${formattedParts.centisText}`;
    return [timeString, centisString];
}

export const numeric = (s: string) => /^[+-]?\d+$/.test(s);
