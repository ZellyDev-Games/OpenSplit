import { Dispatch } from "../../../../wailsjs/go/dispatcher/Service";
import { Command } from "../../../models/command";
import type { SkinElement } from "../../../models/skin/element";

export async function registerPreviewElements(elements: SkinElement[]) {
    await Dispatch(Command.SKIN_PREVIEW_SET, JSON.stringify(elements));
}
