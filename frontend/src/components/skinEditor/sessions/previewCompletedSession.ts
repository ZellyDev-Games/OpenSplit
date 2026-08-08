import { createPreviewRun } from "./createPreviewRun";
import { createBaseSession } from "./previewBase";

export const previewSession = createBaseSession();

previewSession.current_segment_index = 8;
previewSession.dirty = false;
previewSession.current_run = createPreviewRun(
    true,
    [
        {
            id: "opening",
            duration: 12000,
        },
        {
            id: "world1-level1",
            duration: 30000,
        },
        {
            id: "world1-level2",
            duration: 46100,
        },
        {
            id: "world1-boss",
            duration: 18700,
        },
        {
            id: "castle-entry",
            duration: 15100,
        },
        {
            id: "castle-tower",
            duration: 61400,
        },
        {
            id: "final-boss",
            duration: 96200,
        },
        {
            id: "credits",
            duration: 25000,
        },
    ],
    previewSession.leaf_segments ?? [],
);
