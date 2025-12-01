import {useEffect, useState } from "react";
import * as webRace from "../../wailsjs/go/racetime/WebRace";

// RaceList can be run without authenticating to get a list of current races
// LoginWithOAuth should be run before before running RaceData or initializing the WebSocketManager
// WebSocketManager is unique to each race, so it should be 

// type UserRole int

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
// )

// type UserStatus int

// const (
// 	Unknown UserStatus = iota
// 	NotInRace
// 	NotReady
// 	Ready
// 	Finished
// 	Disqualified
// 	Forfeit
// 	Racing
// )

// type RaceState int

// const (
// 	Unknown UserStatus = iota
// 	Open
// 	OpenInviteOnly
// 	Ready
// 	Starting
// 	Started
// 	Ended
// 	Cancelled
// )

const restUrl = "https://localhost:8000"
const socketUrl = "ws://localhost:9999"

export class WebSocketManager {
    private ws: WebSocket;

    // websocket_oauth_url used for authenticated chat messages and real-time updates
    // Retrieve from RaceData function
    // Example socket url "websocket_oauth_url": "/ws/o/race/funky-link-3070"
    private constructor(raceURL: string, accessToken: string) {
        this.ws = new WebSocket(socketUrl+raceURL+"?token="+accessToken);
        
        this.ws.onopen = () => {
            console.log("Connected to WebSocket server");
            
            // Start ping interval
            setInterval(() => {
                if (this.ws.readyState === WebSocket.OPEN) {
                    this.ws.send(JSON.stringify({ type: "ping" }));
                    console.log("Ping sent");
                }
            }, 10_000);
        };

        // Listen for messages
        this.ws.addEventListener("message", (event) => {
            console.log("Message from server:", event.data);
            // TODO: add function here to pass message data out of this event handler.
            // where to put it? go file? whatever frontend file will do the display work?
            // frontend would make more sense since the object is there and sendMessage is called from there.
        });

        // Handle connection close
        this.ws.addEventListener("close", () => {
            console.log("WebSocket connection closed");
        });

        // Handle errors
        this.ws.addEventListener("error", (err) => {
            console.error("WebSocket error:", err);
        });
    }

    public getState() {
        return this.ws.readyState
    }

    static async create(raceURL: string): Promise<WebSocketManager> {
        const accessToken = await webRace.GetAccessToken();
        return new WebSocketManager(raceURL, accessToken);
    }

    // frontend display should call this when the user sends a message.
    // message should be sent as stringified JSON
    public sendMessage(message: string) {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(message);
        } else {
            console.warn("WebSocket is not open, message not sent.");
        }
    }
}

// Get list of races to be displayed
export async function RaceList() {
    // for each race data to display:
    // goal: name, Category: name, entrants_count,entrants_count_finished+entrants_count_inactive, elapsed_time: current_time-started_at, status: value
    const [data, setData] = useState(null);

    useEffect(() => {
        async function loadData() {
          try {
            const response = await fetch(restUrl+"/races/data");
            const json = await response.json();   // parse JSON
            setData(json);
          } catch (err) {
            console.error(err);
          }
        }

        loadData();
      }, []);

  return <pre>{JSON.stringify(data, null, 2)}</pre>;
}

// Get data about specific race
// data url must be used to construct the correct request
export async function RaceData(dataURL: string){
    // data to display for race:
    // slug, status: value, category: name, goal: name, info, entrants_count, entrants_count_finished, entrants_count_inactive, entrants: [], "start_delay": "P0DT00H00M15S", "started_at": null, "ended_at": null, "cancelled_at": null, "allow_comments": true, "hide_comments": false, "hide_entrants": false, "chat_restricted": false, "allow_prerace_chat": true, "allow_midrace_chat": true, "allow_non_entrant_chat": true
    const [data, setData] = useState(null);

    useEffect(() => {
        async function loadData() {
          try {
            const response = await fetch(restUrl+dataURL);
            const json = await response.json();   // parse JSON
            setData(json);
          } catch (err) {
            console.error(err);
          }
        }

        loadData();
      }, []);

  return <pre>{JSON.stringify(data, null, 2)}</pre>;
}

// Authenticate and get user tokens
export async function LoginWithOAuth() {
    try {
        if (await webRace.CheckTokens()) {
            return
        }

        const AuthURL = await webRace.Authorize();
        console.log(AuthURL);

        // Open OAuth popup window
        window.open(AuthURL, "RaceTime.gg OAuth", "width=800,height=700,resizable=yes");

        // Listen for messages coming from the popup OAuth window
        window.addEventListener("message", (event) => {
            if (event.origin === window.location.origin) { // Check origin for security
                const accessToken = event.data;
                // Process the auth code
                webRace.GenTokens(accessToken);
            }
        }, false);

    } catch (error) {
           console.error("Error initiating OAuth:", error);
    }
}