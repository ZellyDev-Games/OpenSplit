import { MouseEvent } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { SelectedVariable } from "../../components/splitFileEditor/types/variables";
import { Command } from "../../models/command";
import SegmentPayload from "../../models/segmentPayload";
import SplitFilePayload from "../../models/splitFilePayload";

type SaveProps = {
    splitFilePayload: SplitFilePayload | null;

    gameName: string;

    gameID: string;

    categoryName: string;

    categoryID: string;

    selectedSkin: string;

    variables: Record<string, SelectedVariable>;

    segments: SegmentPayload[];

    attempts: number;

    offsetText: string;

    platform: string;
};

export function useSaveSplitFile(props: SaveProps) {
    return async function saveSplitFile(e: MouseEvent<HTMLButtonElement>) {
        e.preventDefault();

        const variables = Object.entries(props.variables).map(([id, variable]) => ({
            id,
            name: variable.name,
            value: variable.value,
            label: variable.label,
        }));

        const payload = SplitFilePayload.createFrom({
            id: props.splitFilePayload?.id ?? "",

            game_name: props.gameName,

            speedrun_game_id: props.gameID,

            game_category: props.categoryName,

            speedrun_game_category_id: props.categoryID,

            variables,

            version: props.splitFilePayload?.version ?? 0,

            selected_skin: props.selectedSkin,

            segments: props.segments,

            runs: props.splitFilePayload?.runs ?? [],

            pb: props.splitFilePayload?.pb ?? null,

            sob: props.splitFilePayload?.sob ?? 0,

            attempts: props.attempts,

            offset: props.offsetText === "" ? 0 : Number(props.offsetText),

            platform: props.platform,

            layout: props.splitFilePayload?.layout ?? "",

            windows: props.splitFilePayload?.windows,
        });

        await Dispatch(Command.SUBMIT, JSON.stringify(payload));
    };
}
