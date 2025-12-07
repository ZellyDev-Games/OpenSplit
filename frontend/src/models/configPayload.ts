import { Command } from "../App";

export type KeyInfo = {
    key_code: number;
    locale_name: string;
    modifiers: string[];
    modifier_locale_names: string[];
};

export type ConfigPayload = {
    speed_run_API_base: string;
    key_config: Record<Command, KeyInfo>;
};
