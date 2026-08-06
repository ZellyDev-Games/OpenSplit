export function hashStringToInt(s: string): number {
    let h = 2166136261; // FNV-1a-ish
    for (let i = 0; i < s.length; i++) {
        h ^= s.charCodeAt(i);
        h = Math.imul(h, 16777619);
    }
    return h >>> 0;
}
