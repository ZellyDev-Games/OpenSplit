import { useNavigate } from "react-router";

import { LoadSplitFile } from "../../../wailsjs/go/session/Service";
import { Quit } from "../../../wailsjs/runtime";
import zdgLogo from "../../assets/images/ZG512.png";
import useWindowResize from "../../hooks/useWindowResize";

export default function Welcome() {
    const navigate = useNavigate();
    useWindowResize("welcome");
    return (
        <div className="welcome">
            <img src={zdgLogo} alt="" />
            <hr />
            <h3>OpenSplit</h3>
            <button onClick={() => navigate("/edit")}>Create New Split File</button>
            <button
                onClick={async () => {
                    await LoadSplitFile();
                }}
            >
                Load Split File
            </button>
            <button
                style={{ marginTop: 30 }}
                onClick={async () => {
                    Quit();
                }}
            >
                Exit OpenSplit
            </button>

            <button
                style={{ marginTop: 30 }}
                onClick={async () => {
                    localStorage.clear();
                }}
            >
                Reset All Preferences
            </button>

            <div id="cw">
                <p>Copyright ZellyDev LLC - ZellyDev Games {new Date().getFullYear()}</p>
            </div>
        </div>
    );
}
