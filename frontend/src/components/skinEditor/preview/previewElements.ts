import type { SkinElement } from "../../../models/skin/element";
import parentSegmentRow from "../../splitter/previews/parentSegmentRow.preview";
import segmentList from "../../splitter/previews/segmentList.preview";
import segmentRow from "../../splitter/previews/segmentRow.preview";
import segmentTime from "../../splitter/previews/segmentTime.preview";
import splitGameInfo from "../../splitter/previews/splitGameInfo.preview";
import splitter from "../../splitter/previews/splitter.preview";
import timer from "../../splitter/previews/timer.preview";
import worldRecord from "../../splitter/previews/worldRecord.preview";

export const previewElements: SkinElement[] = [
    ...splitter,
    ...splitGameInfo,
    ...segmentList,
    ...segmentRow,
    ...parentSegmentRow,
    ...segmentTime,
    ...timer,
    ...worldRecord,
];
