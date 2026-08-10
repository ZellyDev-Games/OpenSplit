import { MouseEvent, useLayoutEffect, useRef, useState } from "react";

import { ContextMenuProps, MenuAction, MenuSeparator } from "../hooks/useContextMenu";

const MENU_MARGIN = 8;
const MIN_MENU_HEIGHT = 220;

export function ContextMenu({ state, close, items = [] }: ContextMenuProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const [position, setPosition] = useState({
        left: state.x,
        top: state.y,
    });

    useLayoutEffect(() => {
        if (!state.open || !containerRef.current) return;

        const vw = window.innerWidth;
        const vh = window.innerHeight;

        const menuWidth = 200;
        const menuHeight = Math.max(MIN_MENU_HEIGHT, containerRef.current.offsetHeight);

        setPosition({
            left: Math.max(MENU_MARGIN, Math.min(state.x, vw - menuWidth - MENU_MARGIN)),
            top: Math.max(MENU_MARGIN, Math.min(state.y, vh - menuHeight - MENU_MARGIN)),
        });
    }, [state.open, state.x, state.y, items]);

    if (!state.open) return null;

    const handleOverlayClick = (e: MouseEvent<HTMLDivElement>) => {
        e.stopPropagation();
        close();
    };

    const handleOverlayContextMenu = (e: MouseEvent<HTMLDivElement>) => {
        e.preventDefault();
        e.stopPropagation();
        close();
    };

    const stopPropagation = (e: MouseEvent<HTMLDivElement>) => {
        e.stopPropagation();
    };

    return (
        <div
            className="cm-overlay"
            onClick={handleOverlayClick}
            onContextMenu={handleOverlayContextMenu}
            role="presentation"
        >
            <div
                ref={containerRef}
                className="cm-container"
                style={{
                    left: position.left,
                    top: position.top,
                    minHeight: MIN_MENU_HEIGHT,
                }}
                onClick={stopPropagation}
            >
                <div className="cm-panel" role="menu" aria-label="Context menu">
                    <ul className="cm-list">
                        {items.map((it, i) => {
                            if ((it as MenuSeparator).type === "separator") {
                                return <li key={`sep-${i}`} className="cm-separator" role="separator" />;
                            }

                            const item = it as MenuAction;
                            const disabled = !!item.disabled;

                            const onItemClick = (e: MouseEvent<HTMLButtonElement>) => {
                                e.stopPropagation();

                                if (disabled) return;

                                item.onClick?.();
                                close();
                            };

                            return (
                                <li key={`item-${i}`} role="none">
                                    <button
                                        type="button"
                                        role="menuitem"
                                        className={["cm-item", disabled ? "cm-item--disabled" : ""].join(" ")}
                                        onClick={onItemClick}
                                        disabled={disabled}
                                    >
                                        {item.label}
                                    </button>
                                </li>
                            );
                        })}
                    </ul>
                </div>
            </div>
        </div>
    );
}
