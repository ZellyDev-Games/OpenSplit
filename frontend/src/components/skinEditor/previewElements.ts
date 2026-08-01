import type { SkinElement } from "../../models/skinModel";
import parentSegmentRow from "./../splitter/parentSegmentRow.preview";
import segmentList from "./../splitter/segmentList.preview";
import segmentRow from "./../splitter/segmentRow.preview";
import segmentTime from "./../splitter/segmentTime.preview";
import splitGameInfo from "./../splitter/splitGameInfo.preview";
import splitter from "./../splitter/splitter.preview";
import timer from "./../splitter/timer.preview";

export const previewElements: SkinElement[] = [
    ...splitter,
    ...splitGameInfo,
    ...segmentList,
    ...segmentRow,
    ...parentSegmentRow,
    ...segmentTime,
    ...timer,
];
