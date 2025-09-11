import { useEffect, RefObject } from "react";

export function useClickOutside<T extends HTMLElement>(ref: RefObject<T>, handler: (ev: PointerEvent) => void) {
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
