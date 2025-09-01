import {ContextMenuProps, MenuAction, MenuSeparator} from "../hooks/useContextMenu";
import React from "react";

export function ContextMenu({state, close, items = []} : ContextMenuProps) {
    if(!state.open) return null;

    // clamp position to screen based on width/height estimates
    const vw = typeof window !== "undefined" ? window.innerWidth : 1000;
    const vh = typeof window !== "undefined" ? window.innerHeight : 800;
    const MENU_W = 200;
    const MENU_H = Math.min(320, items.length * 36 + 12);
    const margin = 8;
    const left = Math.max(margin, Math.min(state.x, vw - MENU_W - margin));
    const top = Math.max(margin, Math.min(state.y, vh - MENU_H - margin));

    const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
        e.stopPropagation();
        close();
    }

    const handleOverlayContextMenu = (e: React.MouseEvent<HTMLDivElement>) => {
        e.stopPropagation();
        close();
    }

    const stopPropagation = (e: React.MouseEvent<HTMLDivElement>) => {
        e.stopPropagation();
    }

    return (
        <div
            className={"cm-overlay"}
            onClick={handleOverlayClick}
            onContextMenu={handleOverlayContextMenu}
            role={"presentation"}
        >
            <div className={"cm-container"} style={{ left, top }} onClick={stopPropagation}>
                <div className="cm-panel" role="menu" aria-label="Context menu">
                    <ul className="cm-list">
                        {items.map((it, i) => {
                            if ((it as MenuSeparator).type === "separator") {
                                return <li key={`sep-${i}`} className="cm-separator" role="separator" />;
                            }

                            const item = it as MenuAction;
                            const disabled = !!item.disabled;

                            const onItemClick = (e: React.MouseEvent<HTMLButtonElement>) => {
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
    )
}
