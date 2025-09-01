import {useState} from 'react';
import Timer from "./components/Timer";

function App() {
    const [session, setSession] = useState(null);

    return (
        <div id="App" className="app">
            <Timer />
        </div>
    )
}

export default App
