import { Authorize, GetAccessToken, CheckTokens, GenTokens } from "../../wailsjs/go/racetime/WebRace";
import ButtonList, { ButtonData } from "./ButtonList"

const restUrl = "http://localhost:8000"
const socketUrl = "ws://localhost:9999"

// Get list of races to be displayed
export async function RaceList() {
        try {
            const response = await fetch(restUrl+"/races/data");
            const json = await response.json();   // parse JSON
            return json
        } catch (err) {
            console.error(err);
        }
}

//Race List Window
export async function RaceListWindow(w: Window) {
    // Get race list also need to get the X-Date-Exact header value
    const json = await RaceList()

    // Populate buttons with races
    const DATA: ButtonData[] = [
    ];

    for (let index = 0; index < json.races.length; index++) {
        const categoryName = json.races[index].category.name;
        const URL = json.races[index].url;
        const entrantCount = json.races[index].entrants_count;
        const entrantFinishedCount = json.races[index].entrants_count_finished;
        const goal = json.races[index].goal.name;
        const status = json.races[index].status.value;
        // time stamp format 2025-12-06T08:18:13.788Z
        const startedAt = json.races[index].started_at;
        console.log(categoryName);
        console.log(URL);
        console.log(entrantCount);
        console.log(entrantFinishedCount);
        console.log(goal);
        console.log(status);
        console.log(startedAt);

        // TODO: this should be saved from the racelist call
        const x_date_exact_header: Date = new Date("2025-12-06T23:01:07Z");
        var elapsedTime: Date = new Date(x_date_exact_header.getTime() - startedAt.getTime())
        var runTime = status == 'in_progress' ? elapsedTime : "0"
        DATA.push({
            id: index.toString(),
            URL: URL,
            label: "[" + runTime + "] " + categoryName + " - " + goal + " (" + entrantFinishedCount + "/" + entrantCount + " Finished)"
        });
    }

    <ButtonList
      data={DATA}
      onClick={(item) => {
        console.log("Clicked", item);
        RaceWindow(w, item.URL)
      }}
    />
}

export type messageData = {
    message: string;
    pinned?: boolean;
    actions?: any;
    direct_to?: string;
    guid?: string;
};

//Race Window
export async function RaceWindow(w: Window, dataURL: string) {
    // variables
    var goal
    var info
    let entrants: any[] = []
    var category
    var raceID
    var joined = false
    var forfeit = false
    var done = false
    // const tempURL = dataURL.split("/")
    const authenticatedRaceURL = "/ws/o/race/"+dataURL.split("/")[1]
    const accessToken = await GetAccessToken()


    // open websocket for selected race
    // websocket_oauth_url used for authenticated chat messages and real-time updates
    // Example socket url "websocket_oauth_url": "/ws/o/race/funky-link-3070"
    const ws = new WebSocket(socketUrl+authenticatedRaceURL+"?token="+accessToken);
        
    ws.onopen = () => {
        console.log("Connected to WebSocket server");
        
        // Start ping interval
        setInterval(() => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: "ping" }));
                console.log("Ping sent");
            }
        }, 10_000);
    };

    // Listen for messages
    ws.addEventListener("message", (event) => {
        console.log("Message from server:", event.data);

        const paragraph = w.document.getElementById('text') as HTMLParagraphElement
        switch (event.data.type) {
            case "chat.history":
                // {
                //   "type": "chat.history",
                //   "messages": [
                //      {"id":"xa2wrRW32bl48fJq", ...},
                //      {"id":"g6Kem5bewJfG3ds2", ...},
                //   ]
                // }
                if (paragraph) {
                    for (let index = 0; index < event.data.messages.length; index++) {
                        paragraph.textContent += event.data.messages[index]                        
                    }
                }
                break;
            
            case "chat.message":
                // {
                //   "type": "chat.message",
                //   "message": {
                //     "id": "<string>",
                //     "user": { <user info object> },
                //     "bot": "<string>",
                //     "direct_to": { <user info object> },
                //     "posted_at": "<iso date string>"
                //     "message": "<string>",
                //     "message_plain": "<string>",
                //     "highlight": <bool>,
                //     "is_dm": <bool>,
                //     "is_bot": <bool>,
                //     "is_system": <bool>,
                //     "is_pinned": <bool>,
                //     "delay": "<iso duration string>",
                //     "actions" { <action objects> }
                //   }
                // }
                // Get the hours (0-23)
                const hours: number = event.data.message.posted_at.getHours();
                // Get the minutes (0-59)
                const minutes: number = event.data.message.posted_at.getMinutes();
                // You can then format them as needed, for example, with leading zeros
                const formattedHours: string = String(hours).padStart(2, '0');
                const formattedMinutes: string = String(minutes).padStart(2, '0');

                console.log(`Hours: ${hours}`);
                console.log(`Minutes: ${minutes}`);
                console.log(`Formatted Time: ${formattedHours}:${formattedMinutes}`);

                if (paragraph) {
                    paragraph.textContent += formattedHours + ":" + formattedMinutes + " " + event.data.message.user.name + event.data.message.message
                }
                break

            case "chat.dm":
                // {
                //   "type": "chat.dm",
                //   "message": "<string>",
                //   "from_user": { <user info object> },
                //   "from_bot": "<string>",
                //   "to": { <user info object> },
                // }
                break

            case "chat.pin":
                // {
                //   "type": "chat.pin",
                //   "message": { ... }
                // }
                break

            case "chat.unpin":
                // {
                //   "type": "chat.pin",
                //   "message": { ... }
                // }
                break

            case "chat.delete":
                // chat.delete
                // {
                //     "type": "chat.delete",
                //     "delete": {
                //         "id": "<string>",
                //         "user": { <user info object> },
                //         "bot": "<string>",
                //         "is_bot": <bool>,
                //         "deleted_by": { <user info object> }
                //     }
                // }
                break

            case "chat.purge":
                // {
                    // "type": "chat.purge",
                    // "purge": {
                        // "user": { <user info object> },
                        // "purged_by": { <user info object> }
                    // }
                // }
                break

            case "error":
                // {
                //   "type": "error",
                //   "errors": [
                    // "Permission denied, you may need to re-authorise this application.",
                    // "..."
                //   ]
                // }
                console.log(event.data.errors)
                break

            case "pong":
                // {
                //   "type": "pong"
                // }
                console.log("Pong received");
                break

            case "race.data":
                // {
                //   "type": "race.data",
                //   "race": {
                    // ...
                //   }
                // }
                goal = event.data.race.goal.name
                info = event.data.race.info
                entrants = event.data.race.entrants
                category = event.data.race.category.name
                raceID = event.data.race.slug
                
                const enter = w.document.getElementById('enterRaceButton') as HTMLButtonElement
                const finish = w.document.getElementById('finishButton') as HTMLButtonElement
                const forfeit = w.document.getElementById('forfeitButton') as HTMLButtonElement
                
                // type RaceState
                // invitational
                // pending
                // partitioned //(only for ladder 1v1 races)
                // open
                // in_progress
                // finished
                // cancelled
                switch (event.data.race.status) {
                    case "open":
                        enterRaceButton.hidden = false
                        enterRaceButton.disabled = false
                        finishButton.hidden = true
                        finishButton.disabled = true
                        forfeitButton.hidden = true
                        forfeitButton.disabled = true
                    break

                    case "in_progress":
                        enterRaceButton.hidden = true
                        enterRaceButton.disabled = true
                        finishButton.hidden = false
                        finishButton.disabled = false
                        forfeitButton.hidden = false
                        forfeitButton.disabled = false
                    break

                    case  "finished":
                    case "cancelled":
                        enterRaceButton.hidden = true
                        enterRaceButton.disabled = true
                        finishButton.hidden = true
                        finishButton.disabled = true
                        forfeitButton.hidden = true
                        forfeitButton.disabled = true
                    break
                }

                // update entrants list
                break

            default:
                break;
        }
    });

    // Handle connection close
    ws.addEventListener("close", () => {
        console.log("WebSocket connection closed");
    });

    // Handle errors
    ws.addEventListener("error", (err) => {
        console.error("WebSocket error:", err);
    });

    // clear window contents
    w.document.body.innerHTML = "";
    
    // title format
    // {goal} [{category}] - {URL}
    w.document.title = goal + " [" + category + "] - " + raceID

    // top of window
    // Goal: {goal} Info: {info}
    const raceInfoBar: HTMLDivElement = w.document.createElement('div')
    raceInfoBar.textContent = "Goal: " + goal + " Info: " + info
    raceInfoBar.classList.add('race-info')
    w.document.body.appendChild(raceInfoBar)

    // right side of window
    // List of entrants with stream status (color coded icon??), ready status (color code name??)

    // type UserStatus
    // ready
    // not_ready
    // in_progress
    // done
    // dnf //(did not finish, i.e. forfeited)
    // dq //(disqualified)

    // type UserRole
    // const (
    // 	Unknown UserRole = iota
    // 	Anonymous
    // 	Regular
    // 	ChannelCreator UserRole = 4
    // 	Monitor        UserRole = 8
    // 	Moderator      UserRole = 16
    // 	Staff          UserRole = 32
    // 	Bot            UserRole = 64
    // 	System         UserRole = 128
    const entrantList: HTMLUListElement = document.createElement('ul');
    for (let index = 0; index < entrants.length; index++) {
        const element = entrants[index].name;
        
        const listItem: HTMLLIElement = document.createElement('li');
        listItem.textContent = element;
        entrantList.appendChild(listItem);        
    }

    w.document.body.appendChild(entrantList);

    // chat display window
    const text = w.document.createElement('p')
    text.id = 'text'
    text.textContent = "";
    w.document.body.appendChild(text)

    // bottom of window
    // [hide results checkbox] [Save Log button] [Ready checkbox] [Enter Race button]
    // Create the hide results input element
    const hideResultsCheckBox = w.document.createElement('input');
    hideResultsCheckBox.type = 'checkbox';
    hideResultsCheckBox.id = 'hideResults';
    hideResultsCheckBox.name = 'hideResults';
    hideResultsCheckBox.checked = false;

    // Create the hide label element
    const hideResultsLabel = w.document.createElement('label');
    hideResultsLabel.htmlFor = 'hideResults'; // Associate label with the checkbox
    hideResultsLabel.textContent = 'Hide Results';
    hideResultsCheckBox.addEventListener('change', (event: Event) => {
        const target = event.target as HTMLInputElement
        if (ws.readyState === WebSocket.OPEN) {
            if (target.checked) {
                console.log('Checkbox is checked')
                
                // TODO: hide entrants
            } else {
                console.log('Checkbox is unchecked')
                // TODO: show entrants
            }
        } else {
            console.warn("WebSocket is not open, message not sent.");
        }
    })

    // Create a new button element
    const saveChatLogButton: HTMLButtonElement = document.createElement('button');
    saveChatLogButton.textContent = 'Save Chat Log!';
    saveChatLogButton.id = 'saveChatLogButton';
    saveChatLogButton.classList.add('save-chat-log');

    // Set the type attribute (important for form submission behavior)
    saveChatLogButton.type = 'button'; // or 'submit', 'reset'
    saveChatLogButton.addEventListener('click', () => {
        console.log('Save chat log clicked!');
        // TODO: output chat log to file here
    });

    // Create the ready checkbox input element
    const readyCheckBox = w.document.createElement('input');
    readyCheckBox.type = 'checkbox';
    readyCheckBox.id = 'ready';
    readyCheckBox.name = 'ready';
    readyCheckBox.checked = false;
    readyCheckBox.addEventListener('change', (event: Event) => {
        const target = event.target as HTMLInputElement
        if (ws.readyState === WebSocket.OPEN) {
            if (target.checked) {
                console.log('Checkbox is checked')
                // message format
                // {
                //     "action": "message",
                //     "data": {
                //         "message": "Your message goes here",
                //         // "pinned": <bool>,
                //         "actions": <object or null>,
                //         "direct_to": <string or null>,
                //         "guid": "<random string>"
                //     }
                // }
                
                const mData: messageData = {
                    message: ".ready"
                }
                const ready_message: {action: string; data: messageData} = {
                    action: "message",
                    data: mData
                }

                ws.send(JSON.stringify(ready_message));
            } else {
                console.log('Checkbox is unchecked')
                // message format
                // {
                //     "action": "message",
                //     "data": {
                //         "message": "Your message goes here",
                //         // "pinned": <bool>,
                //         "actions": <object or null>,
                //         "direct_to": <string or null>,
                //         "guid": "<random string>"
                //     }
                // }
                const mData: messageData = {
                    message: ".unready"
                }
                const ready_message: {action: string; data: messageData} = {
                    action: "message",
                    data: mData
                }

                ws.send(JSON.stringify(ready_message));
            }
        } else {
            console.warn("WebSocket is not open, message not sent.");
        }
    })

    // Create the ready label element
    const readyLabel = w.document.createElement('label');
    readyLabel.htmlFor = 'ready'; // Associate label with the checkbox
    readyLabel.textContent = 'Ready';

    // Create a new button element
    const enterRaceButton: HTMLButtonElement = w.document.createElement('button');
    enterRaceButton.textContent = 'Enter Race';
    enterRaceButton.id = 'enterRaceButton';
    enterRaceButton.classList.add('enter');
    enterRaceButton.type = 'button'; // or 'submit', 'reset'
    enterRaceButton.hidden = false
    enterRaceButton.addEventListener('click', () => {
        // message format
        // {
        //     "action": "message",
        //     "data": {
        //         "message": "Your message goes here",
        //         // "pinned": <bool>,
        //         "actions": <object or null>,
        //         "direct_to": <string or null>,
        //         "guid": "<random string>"
        //     }
        // }
        if (ws.readyState === WebSocket.OPEN) {
            console.log('Race join status changed!');
            // if in race leave otherwise enter
            if (joined) {
                const mData: messageData = {
                    message: ".leave"
                }
                const ready_message: {action: string; data: messageData} = {
                    action: "message",
                    data: mData
                }

                joined = !joined
                ws.send(JSON.stringify(ready_message));
            } else {
                const mData: messageData = {
                    message: ".join"
                }
                const ready_message: {action: string; data: messageData} = {
                    action: "message",
                    data: mData
                }

                joined = !joined
                ws.send(JSON.stringify(ready_message));
            }
        } else {
            console.warn("WebSocket is not open, message not sent.");
        }
    });

    // Create a new button element
    const finishButton: HTMLButtonElement = w.document.createElement('button');
    finishButton.textContent = 'Done';
    finishButton.id = 'finishButton';
    finishButton.classList.add('done');
    finishButton.type = 'button'; // or 'submit', 'reset'
    finishButton.hidden = true
    enterRaceButton.addEventListener('click', () => {
        // message format
        // {
        //     "action": "message",
        //     "data": {
        //         "message": "Your message goes here",
        //         // "pinned": <bool>,
        //         "actions": <object or null>,
        //         "direct_to": <string or null>,
        //         "guid": "<random string>"
        //     }
        // }
        if (ws.readyState === WebSocket.OPEN) {
            console.log('Race join status changed!');
            // if done undone otherwise done
            if (done) {
                const mData: messageData = {
                    message: ".undone"
                }
                const ready_message: {action: string; data: messageData} = {
                    action: "message",
                    data: mData
                }

                done = !done
                ws.send(JSON.stringify(ready_message));
            } else {
                const mData: messageData = {
                    message: ".done"
                }
                const ready_message: {action: string; data: messageData} = {
                    action: "message",
                    data: mData
                }

                done = !done
                ws.send(JSON.stringify(ready_message));
            }
        } else {
            console.warn("WebSocket is not open, message not sent.");
        }
    });

    // Create a new button element
    const forfeitButton: HTMLButtonElement = w.document.createElement('button');
    forfeitButton.textContent = 'Forfeit';
    forfeitButton.id = 'forfeitButton';
    forfeitButton.classList.add('forfeit');
    forfeitButton.type = 'button'; // or 'submit', 'reset'
    forfeitButton.hidden = true
    forfeitButton.addEventListener('click', () => {
        // message format
        // {
        //     "action": "message",
        //     "data": {
        //         "message": "Your message goes here",
        //         // "pinned": <bool>,
        //         "actions": <object or null>,
        //         "direct_to": <string or null>,
        //         "guid": "<random string>"
        //     }
        // }
        if (ws.readyState === WebSocket.OPEN) {
            console.log('Forfeit status changed!');
            // if forfeited unforfeit otherwise forfeit
            if (forfeit) {
                const mData: messageData = {
                    message: ".unforfeit"
                }
                const ready_message: {action: string; data: messageData} = {
                    action: "message",
                    data: mData
                }

                forfeit = !forfeit
                ws.send(JSON.stringify(ready_message));
            } else {
                const mData: messageData = {
                    message: ".forfeit"
                }
                const ready_message: {action: string; data: messageData} = {
                    action: "message",
                    data: mData
                }

                forfeit = !forfeit
                ws.send(JSON.stringify(ready_message));
            }
        } else {
            console.warn("WebSocket is not open, message not sent.");
        }
    });

    // [chat bar]
    const textInput: HTMLInputElement = w.document.createElement('input');
    textInput.type = 'text';
    textInput.id = 'myTextInput'; // Assign an ID for easy access
    textInput.addEventListener('input', (event: Event) => {
        const target = event.target as HTMLInputElement; // Type assertion for type safety
        const enteredText: string = target.value;
        console.log('User entered:', enteredText);
        if (ws.readyState === WebSocket.OPEN) {
            const mData: messageData = {
                message: enteredText
            }
            const message: {action: string; data: messageData} = {
                action: "message",
                data: mData
            }
            ws.send(JSON.stringify(message));
        } else {
            console.warn("WebSocket is not open, message not sent.");
        }
    });
    
    // Create a container for better styling/layout
    const container = w.document.createElement('div');
    container.appendChild(hideResultsCheckBox);
    container.appendChild(hideResultsLabel);
    container.appendChild(saveChatLogButton);
    container.appendChild(readyCheckBox);
    container.appendChild(readyLabel);
    container.appendChild(enterRaceButton);
    
    w.document.body.appendChild(textInput)
}

// Authenticate and get user tokens
export async function LoginWithOAuth(w: Window) {
    try {
        // Gets the racetime.gg buttons
        const loginBtn = document.getElementsByTagName("BUTTON")[3] as HTMLButtonElement
        const raceBtn = document.getElementsByTagName("BUTTON")[4] as HTMLButtonElement

        loginBtn.style.display = "none"
        raceBtn.style.display = "block"

        if (await CheckTokens()) {
            w.close
            return
        }

        const AuthURL = await Authorize()
        console.log(AuthURL)

        // Open OAuth popup window
        // window.open(AuthURL, "RaceTime.gg OAuth", "width=800,height=700,resizable=yes");
        w.location.href = AuthURL;

        // TODO: Get this shit to work
        // It should detect when the popup window changes due to the user clicking authorize and the code being sent back via configured redirect
        // example redirect http://localhost:34115/?code=GDaChqJIf16CdrL8XsWJtmn3gmm64j&state=state
        w.addEventListener("DOMContentLoaded", () => {
            const params = new URLSearchParams(w.location.search);

            if (params.has("code") && params.has("state")) {
                console.log("Detected OAuth redirect on load");

                console.log("Current URL:", w.location.href);

                const accessCode = params.get("code")
                console.log(accessCode)
                // Process the auth code
                const accessToken = GenTokens(accessCode!);
                console.log(accessToken)
                w.close
            }
        })

    } catch (error) {
           console.error("Error initiating OAuth:", error);
    }
}