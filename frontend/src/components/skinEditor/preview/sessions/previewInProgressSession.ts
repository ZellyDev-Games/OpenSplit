import { createPreviewRun } from "./createPreviewRun";
import { createBaseSession } from "./previewBase";

export const previewSession = createBaseSession();

previewSession.current_segment_index = 5;

previewSession.dirty = false;

previewSession.current_run = createPreviewRun(
    false,
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
    ],
    previewSession.leaf_segments ?? [],
);
