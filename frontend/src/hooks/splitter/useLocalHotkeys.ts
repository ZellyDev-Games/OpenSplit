import { useEffect, useRef } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { EventsOn } from "../../../wailsjs/runtime";
import { Command } from "../../models/command";
import { ConfigPayload, KeyInfo } from "../../models/configPayload";
import { log } from "../../utils/logger";

function normalizeModifier(modifier: string): string {
    switch (modifier.toUpperCase()) {
        case "CONTROL":
            return "CTRL";
        case "COMMAND":
            return "META";
        default:
            return modifier.toUpperCase();
    }
}

function getModifiers(event: KeyboardEvent): string[] {
    return [
        event.ctrlKey ? "CTRL" : "",
        event.altKey ? "ALT" : "",
        event.shiftKey ? "SHIFT" : "",
        event.metaKey ? "META" : "",
    ].filter(Boolean);
}

function matchesKey(event: KeyboardEvent, binding: KeyInfo): boolean {
    if (!binding) {
        return false;
    }

    if (event.repeat) {
        return false;
    }

    if (event.keyCode !== binding.key_code) {
        return false;
    }

    const requiredModifiers = (binding.modifiers ?? []).map(normalizeModifier);
    const eventModifiers = getModifiers(event);

    if (requiredModifiers.length !== eventModifiers.length) {
        return false;
    }

    return requiredModifiers.every((modifier) => eventModifiers.includes(modifier));
}

function shouldIgnoreTarget(target: EventTarget | null): boolean {
    const element = target as HTMLElement | null;

    if (!element) {
        return false;
    }

    if (element.isContentEditable) {
        return true;
    }

    switch (element.tagName) {
        case "INPUT":
        case "TEXTAREA":
        case "SELECT":
            return true;
        default:
            return false;
    }
}

export function useLocalHotkeys(config: ConfigPayload | null) {
    const globalHotkeysActive = useRef(config?.global_hotkeys_active ?? false);

    useEffect(() => {
        globalHotkeysActive.current = config?.global_hotkeys_active ?? false;
    }, [config?.global_hotkeys_active]);

    useEffect(() => {
        const unsubscribe = EventsOn("hotkeys:global-state", (active: boolean) => {
            log.debug("[LocalHotkeys] Global hotkeys state:", active);

            globalHotkeysActive.current = active;
        });

        return () => {
            unsubscribe();
        };
    }, []);

    useEffect(() => {
        if (!config) {
            log.debug("[LocalHotkeys] No config; local hotkeys disabled");
            return;
        }

        const handleKeyDown = (event: KeyboardEvent) => {
            if (globalHotkeysActive.current) {
                return;
            }

            if (shouldIgnoreTarget(event.target)) {
                return;
            }

            for (const [commandString, binding] of Object.entries(config.key_config ?? {})) {
                const command = Number(commandString) as Command;

                if (!matchesKey(event, binding)) {
                    continue;
                }

                log.debug("[LocalHotkeys] Matched command", command, binding);

                event.preventDefault();
                event.stopPropagation();

                void Dispatch(command, null).catch((error) => {
                    log.error("[LocalHotkeys] Failed to dispatch command", command, error);
                });

                return;
            }
        };

        window.addEventListener("keydown", handleKeyDown);

        return () => {
            window.removeEventListener("keydown", handleKeyDown);
        };
    }, [config]);
}
