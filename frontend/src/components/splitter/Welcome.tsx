import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { WindowSetSize } from "../../../wailsjs/runtime";
import { Command } from "../../App";
import zdgLogo from "../../assets/images/ZG512.png";
import { LoginWithOAuth, RaceListWindow } from "../racetime_gg";

export default function Welcome() {
    WindowSetSize(320, 580);
    return (
        <div className="welcome">
            <img src={zdgLogo} alt="" />
            <hr />
            <h3>OpenSplit</h3>
            <button
                onClick={async () => {
                    await Dispatch(Command.NEW, null);
                }}
            >
                Create New Split File
            </button>

            <button
                onClick={async () => {
                    await Dispatch(Command.LOAD, null);
                }}
            >
                Load Split File
            </button>

            <button
                onClick={async () => {
                    await Dispatch(Command.EDIT, null);
                }}
            >
                OpenSplit Settings
            </button>

            <button
                onClick={async () => {
                    // TODO: Get this shit to work
                    // This button should open a popup OAuth window
                    // The window needs to opened here otherwise it takes too long and gets blocked
                    const w = window.open("", "_blank")
                    // const w = window.open("", "RaceTime.gg OAuth", "width=800,height=700,resizable=yes")
                    await LoginWithOAuth(w!);
                }}
            >
                Racetime.gg Auth
            </button>

            <button
                hidden
                onClick={async () => {
                    // TODO: Get this shit to work
                    // This button should open a popup window to interact with racetime.gg races
                    // The window needs to opened here otherwise it takes too long and gets blocked
                    // Hidden until the user is authorized
                    const w = window.open("", "_blank")
                    // const w = window.open("", "RaceTime.gg Races", "width=800,height=700,resizable=yes")
                    await RaceListWindow(w!);
                }}
            >
                Racetime.gg Races
            </button>

            <button
                style={{ marginTop: 30 }}
                onClick={async () => {
                    await Dispatch(Command.QUIT, null);
                }}
            >
                Exit OpenSplit
            </button>

            <button
                style={{ marginTop: 30 }}
                onClick={async () => {
                    localStorage.clear();
                    await Dispatch(Command.RESET, null);
                }}
            >
                <small>Reset All Preferences</small>
            </button>

            <div id="cw">
                <p>Copyright ZellyDev LLC - ZellyDev Games {new Date().getFullYear()}</p>
            </div>
        </div>
    );
}
