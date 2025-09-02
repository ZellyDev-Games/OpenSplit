import React, {useCallback, useState} from "react";

export type MenuSeparator = { type: "separator" };
export type MenuAction = {
    label: string;
    onClick: () => void;
    disabled?: boolean;
};

export type MenuItem = MenuSeparator | MenuAction;

export type ContextMenuState = {
    open: boolean;
    x: number;
    y: number;
};

export type UseContextMenu = {
    bind: {
        onContextMenu: (e: React.MouseEvent<HTMLElement>) => void;
    };
    state: ContextMenuState;
    close: () => void;
};

export type ContextMenuProps = {
    state: ContextMenuState;
    close: () => void;
    items?: MenuItem[];
};

export function useContextMenu() {
    const [state, setState] = useState<ContextMenuState>({ open: false, x: 0, y: 0 });

    const onContextMenu = useCallback((e: React.MouseEvent<HTMLElement>) => {
        console.log("Attempting to open context menu");
        e.preventDefault();
        setState({ open: true, x: e.clientX, y: e.clientY });
    }, []);

    const close = useCallback(() => {
        setState((s) => ({ ...s, open: false }));
    }, []);

    return { bind: { onContextMenu }, state, close };
}