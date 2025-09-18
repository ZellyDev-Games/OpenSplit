import { Route, Routes } from "react-router";

import SplitEditor from "./components/editor/SplitEditor";
import Splitter from "./components/splitter/Splitter";

function App() {
    return (
        <div id="App" className="app">
            <Routes>
                <Route path="/" element={<Splitter />} />
                <Route path="/edit" element={<SplitEditor />} />
            </Routes>
        </div>
    );
}

export default App;
