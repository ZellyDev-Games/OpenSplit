import { hashStringToInt } from "./hash";

export type GroupCtx = { bg: string };

export function colorFromId(id: string): string {
    const n = hashStringToInt(id);

    const hue = n % 360;

    // Keep saturation strong but not neon
    const sat = 45 + (n % 15); // 45–59%

    // Dark background range
    const light = 18 + (n % 10); // 18–27%

    return `hsl(${hue} ${sat}% ${light}%)`;
}
