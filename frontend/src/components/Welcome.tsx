import { Dispatch } from "../../wailsjs/go/dispatcher/Service";
import zdgLogo from "../assets/images/ZG512.png";
import { Command } from "../models/command";

export default function Welcome() {
    const loadSplitFile = async () => {
        try {
            await Dispatch(Command.LOAD, null);
        } catch (error) {
            // The backend rejects the promise when the user closes the open dialog.
            const message = error instanceof Error ? error.message : String(error);
            if (message !== "user cancelled file open operation") {
                console.error("Failed to load split file", error);
            }
        }
    };

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
                onClick={loadSplitFile}
            >
                Load Split File
            </button>
            <button
                onClick={async () => {
                    await Dispatch(Command.NEW_SKIN, null);
                }}
            >
                Create New Skin
            </button>
            <button
                onClick={async () => {
                    await Dispatch(Command.EDIT_SKIN, null);
                }}
            >
                Edit Skin
            </button>
            <button
                onClick={async () => {
                    await Dispatch(Command.EDIT, null);
                }}
            >
                OpenSplit Settings
            </button>

            <button
                className="welcome-exit"
                onClick={async () => {
                    await Dispatch(Command.QUIT, null);
                }}
            >
                Exit OpenSplit
            </button>

            <div className="welcome-footer">
                <p>Copyright ZellyDev LLC - ZellyDev Games {new Date().getFullYear()}</p>
            </div>
        </div>
    );
}
