/**
 * Displays world record information.
 *
 * Shows:
 *  - World record players
 *  - Real-time world record
 *  - In-game-time world record when available
 */

import WorldRecord from "../../../models/worldRecord";
import { displayFormattedTimeParts, formatDuration, msToParts } from "./timerUtils";

type WorldRecordParams = {
    worldRecord?: WorldRecord;
};

export default function WorldRecordDisplay({ worldRecord }: WorldRecordParams) {
    if (!worldRecord?.show) {
        return null;
    }

    const realTime = displayFormattedTimeParts(formatDuration(msToParts(worldRecord.real_time * 1000)));

    const inGameTime = displayFormattedTimeParts(formatDuration(msToParts(worldRecord.in_game_time * 1000)));

    const players = worldRecord.players?.length ? worldRecord.players.join(", ") : "Unknown";

    return (
        <div id="world-record">
            <div id="world-record-player">
                <strong id="world-record-label">WR</strong>
                <span id="world-record-players">{players}</span>
            </div>

            <div id="world-record-real-time">
                <strong id="world-record-rt-label">RT</strong>
                <span id="world-record-rt-time">{realTime[0]}</span>
                <small id="world-record-rt-centiseconds">{realTime[1]}</small>
            </div>

            {worldRecord.in_game_time > 0 && (
                <div id="world-record-in-game-time">
                    <strong id="world-record-igt-label">IGT</strong>
                    <span id="world-record-igt-time">{inGameTime[0]}</span>
                    <small id="world-record-igt-centiseconds">{inGameTime[1]}</small>
                </div>
            )}
        </div>
    );
}
