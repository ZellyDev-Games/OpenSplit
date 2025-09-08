import {Route, Routes, useNavigate} from "react-router";
import Splitter from "./components/splitter/Splitter";
import SplitEditor from "./components/editor/SplitEditor";

function App() {
    const navigate = useNavigate()

    return (
        <div id="App" className="app">
            <Routes>
                <Route path="/" element={<Splitter />}/>
                <Route path="/edit" element={<SplitEditor />}/>
            </Routes>
        </div>
    )
}

export default App