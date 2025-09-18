import React, { useEffect } from "react";

export function useClickOutside(ref: React.RefObject<HTMLDivElement | null>, handler: (ev: PointerEvent) => void) {
    useEffect(() => {
        const onPointerDown = (ev: PointerEvent) => {
            const el = ref.current;
            if (!el) return;
            if (ev.target instanceof Node && el.contains(ev.target)) return;
            handler(ev);
        };

        document.addEventListener("pointerdown", onPointerDown);
        return () => document.removeEventListener("pointerdown", onPointerDown);
    }, [ref, handler]);
}
