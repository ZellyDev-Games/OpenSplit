import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { WindowSetSize } from "../../../wailsjs/runtime";
import { Command } from "../../App";
import zdgLogo from "../../assets/images/ZG512.png";

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
