import { createBaseSession } from "./previewBase";

export const previewSession = createBaseSession();

previewSession.dirty = false;

previewSession.current_segment_index = -1;
previewSession.current_run = null;
