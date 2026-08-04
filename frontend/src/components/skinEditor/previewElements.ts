import type { SkinElement } from "../../models/skinModel";
import splitGameInfo from "./../splitter/previews/splitGameInfo.preview";
import splitter from "./../splitter/previews/splitter.preview";
import parentSegmentRow from "./../splitter/segments/previews/parentSegmentRow.preview";
import segmentList from "./../splitter/segments/previews/segmentList.preview";
import segmentRow from "./../splitter/segments/previews/segmentRow.preview";
import segmentTime from "./../splitter/segments/previews/segmentTime.preview";
import timer from "./../splitter/timer/timer.preview";

export const previewElements: SkinElement[] = [
    ...splitter,
    ...splitGameInfo,
    ...segmentList,
    ...segmentRow,
    ...parentSegmentRow,
    ...segmentTime,
    ...timer,
];
