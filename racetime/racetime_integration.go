package racetime

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

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

// type MessageType int

// const (
// 	Unknown MessageType = iota
// 	User
// 	Error
// 	Race
// 	System
// 	LiveSplit
// 	SplitUpdate
// 	Bot
// )

// // domain (Domain or IP of the Race-Server)
// racetime.gg
var addr = flag.String("addr", "localhost:8000", "http service address")

var client_id = "x4oiff8OAiWwtfQUboFhFlYfgmDMHmxduOFOQgve"
var client_secret = "1BYxBFqyO495W8VCYiZxAEXgortlLa5trpzY0xxDHNAuAWaqfxhgy4435Gq5yp6P76Hw1EIFdp8JjnKvDtDfzLZ2lo6D1TrrWlp0yNbmBTPpNxYVePSqE7eX72ZDAmaU"

// Gets all current races
func GetRaces() ([]byte, error) {
	u := url.URL{Scheme: "http", Host: *addr, Path: "/races/data"}
	req, err := http.NewRequest("GET", u.String(), nil)

	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)
	// Example Response Body: {"races": [{"name": "alttp/perfect-ivysaur-9765", "status": {"value": "open", "verbose_value": "Open", "help_text": "Anyone may join this race"}, "url": "/alttp/perfect-ivysaur-9765", "data_url": "/alttp/perfect-ivysaur-9765/data", "goal": {"name": "100%", "custom": false}, "info": "", "entrants_count": 0, "entrants_count_finished": 0, "entrants_count_inactive": 0, "opened_at": "2025-11-25T14:22:38.834Z", "started_at": null, "time_limit": "P1DT00H00M00S", "opened_by_bot": null, "category": {"name": "The Legend of Zelda: A Link to the Past", "short_name": "ALttP", "slug": "alttp", "url": "/alttp", "data_url": "/alttp/data", "image": null}}]}

	return body, err
}

// Get category details
func GetCategoryDetails(category string) ([]byte, error) {
	// category = "alttp"
	u := url.URL{Scheme: "http", Host: *addr, Path: "/" + category + "/data"}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)
	// Example Response Body: {"name": "The Legend of Zelda: A Link to the Past", "short_name": "ALttP", "slug": "alttp", "url": "/alttp", "data_url": "/alttp/data", "image": null, "info": null, "streaming_required": true, "owners": [{"id": "GvzqPgEyPdZ0RKnr", "full_name": "Luigi#5557", "name": "Luigi", "discriminator": "5557", "url": "/user/GvzqPgEyPdZ0RKnr/luigi", "avatar": null, "pronouns": null, "flair": "moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}], "moderators": [{"id": "xr85vpEMBoX32zJ4", "full_name": "Bowser#7723", "name": "Bowser", "discriminator": "7723", "url": "/user/xr85vpEMBoX32zJ4/bowser", "avatar": null, "pronouns": null, "flair": "moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}], "goals": ["100%", "16 stars", "Beat the game"], "current_races": [], "emotes": {}}

	return body, err
}

// Get category past races
// show_entrants can be either true, false, or empty
// page selects a page of the returned data, -1 ignores option
// per_page controls how many results to return per page
func GetCategoryPastRaces(category string, show_entrants string, page int, per_page int) ([]byte, error) {
	// Generate json request body
	data := make(map[string]string)
	switch show_entrants {
	case "":
		if page != -1 {
			data["page"] = strconv.Itoa(page)
			data["per_page"] = strconv.Itoa(per_page)
		}
	case "true":
		if page == -1 {
			data["show_entrants"] = show_entrants
		} else {
			data["show_entrants"] = show_entrants
			data["page"] = strconv.Itoa(page)
			data["per_page"] = strconv.Itoa(per_page)
		}
	case "false":
		if page == -1 {
			data["show_entrants"] = show_entrants
		} else {
			data["show_entrants"] = show_entrants
			data["page"] = strconv.Itoa(page)
			data["per_page"] = strconv.Itoa(per_page)
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	// category = "alttp"
	u := url.URL{Scheme: "http", Host: *addr, Path: "/" + category + "/races/data"}
	req, err := http.NewRequest("GET", u.String(), bytes.NewBuffer(jsonData))
	// Example Response Body: {"count": 0, "num_pages": 1, "races": []}

	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)
	// Example Response Body: {"name": "The Legend of Zelda: A Link to the Past", "short_name": "ALttP", "slug": "alttp", "url": "/alttp", "data_url": "/alttp/data", "image": null, "info": null, "streaming_required": true, "owners": [{"id": "GvzqPgEyPdZ0RKnr", "full_name": "Luigi#5557", "name": "Luigi", "discriminator": "5557", "url": "/user/GvzqPgEyPdZ0RKnr/luigi", "avatar": null, "pronouns": null, "flair": "moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}], "moderators": [{"id": "xr85vpEMBoX32zJ4", "full_name": "Bowser#7723", "name": "Bowser", "discriminator": "7723", "url": "/user/xr85vpEMBoX32zJ4/bowser", "avatar": null, "pronouns": null, "flair": "moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}], "goals": ["100%", "16 stars", "Beat the game"], "current_races": [], "emotes": {}}

	return body, err
}

// Gets category leaderboard data
func GetCategoryLeaderboards(category string) ([]byte, error) {
	// category = "alttp"
	u := url.URL{Scheme: "http", Host: *addr, Path: "/" + category + "/leaderboards/data"}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)
	// Example Response Body: {"leaderboards": [{"goal": "100%", "num_ranked": 0, "rankings": []}, {"goal": "16 stars", "num_ranked": 0, "rankings": []}, {"goal": "Beat the game", "num_ranked": 0, "rankings": []}]}

	return body, err
}

// Gets category race info
func GetCategoryRaceInfo(category string, race string) ([]byte, error) {
	// category = "alttp"
	// race := "funky-link-3070"
	u := url.URL{Scheme: "http", Host: *addr, Path: "/" + category + "/" + race + "/data"}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)
	// Example Response Body: {"version": 1, "name": "alttp/funky-link-3070", "slug": "funky-link-3070", "status": {"value": "open", "verbose_value": "Open", "help_text": "Anyone may join this race"}, "url": "/alttp/funky-link-3070", "data_url": "/alttp/funky-link-3070/data", "websocket_url": "/ws/race/funky-link-3070", "websocket_bot_url": "/ws/o/bot/funky-link-3070", "websocket_oauth_url": "/ws/o/race/funky-link-3070", "category": {"name": "The Legend of Zelda: A Link to the Past", "short_name": "ALttP", "slug": "alttp", "url": "/alttp", "data_url": "/alttp/data", "image": null}, "goal": {"name": "100%", "custom": false}, "info": "", "info_bot": null, "info_user": "", "team_race": false, "entrants_count": 0, "entrants_count_finished": 0, "entrants_count_inactive": 0, "entrants": [], "opened_at": "2025-11-25T15:05:51.047Z", "start_delay": "P0DT00H00M15S", "started_at": null, "ended_at": null, "cancelled_at": null, "ranked": true, "unlisted": false, "time_limit": "P1DT00H00M00S", "time_limit_auto_complete": false, "require_even_teams": false, "streaming_required": true, "auto_start": true, "opened_by": {"id": "5BRGVMd30E368Lzv", "full_name": "Douglas Kirby", "name": "Douglas Kirby", "discriminator": null, "url": "/user/5BRGVMd30E368Lzv/douglas-kirby", "avatar": null, "pronouns": null, "flair": "staff moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}, "opened_by_bot": null, "monitors": [], "recordable": true, "recorded": false, "recorded_by": null, "disqualify_unready": false, "allow_comments": true, "hide_comments": false, "hide_entrants": false, "chat_restricted": false, "allow_prerace_chat": true, "allow_midrace_chat": true, "allow_non_entrant_chat": true, "chat_message_delay": "P0DT00H00M00S", "bot_meta": {}}

	return body, err
}

// Gets user past race info
// show_entrants can be either true, false, or empty
// page selects a page of the returned data, -1 ignores option
// per_page controls how many results to return per page
func GetUserPastRaces(category string, user string, show_entrants string, page int, per_page int) ([]byte, error) {
	// Generate json request body
	data := make(map[string]string)
	switch show_entrants {
	case "":
		if page != -1 {
			data["page"] = strconv.Itoa(page)
			data["per_page"] = strconv.Itoa(per_page)
		}
	case "true":
		if page == -1 {
			data["show_entrants"] = show_entrants
		} else {
			data["show_entrants"] = show_entrants
			data["page"] = strconv.Itoa(page)
			data["per_page"] = strconv.Itoa(per_page)
		}
	case "false":
		if page == -1 {
			data["show_entrants"] = show_entrants
		} else {
			data["show_entrants"] = show_entrants
			data["page"] = strconv.Itoa(page)
			data["per_page"] = strconv.Itoa(per_page)
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	u := url.URL{Scheme: "http", Host: *addr, Path: "/user/" + user + "/races/data"}
	req, err := http.NewRequest("GET", u.String(), bytes.NewBuffer(jsonData))

	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)

	return body, err
}

// User search
func UserSearch(user string, discriminator string) ([]byte, error) {
	// Generate json request body
	data := make(map[string]string)
	data["name"] = user
	if discriminator != "" {
		data["discriminator"] = discriminator
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	u := url.URL{Scheme: "http", Host: *addr, Path: "/user/search"}
	req, err := http.NewRequest("GET", u.String(), bytes.NewBuffer(jsonData))

	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)

	return body, err
}

func Authorize() {
	// app := application.Get()
	// // New windows inherit the same frontend assets by default
	// err := runtime.NewWindow(ctx, options.Window{
	// 	Title:  "My Popup Window",
	// 	Width:  400,
	// 	Height: 300,
	// 	// You can set other options like Frameless, MinSize, etc. here
	// })
	// if err != nil {
	// 	// Handle error appropriately
	// 	println("Error opening new window:", err.Error())
	// }

	// This requires processing the login page.
	req, err := http.NewRequest("GET", "http://localhost:8000/o/authorize", nil)

	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)
}

func GenTokens() {
	// Generate json request body
	data := map[string]string{"client_id": client_id, "client_secret": client_secret, "grant_type": "authorization_code"}
	// data := map[string]string{"client_id": client_id, "client_secret": client_secret, "grant_type": "client_credentials"}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}
	// Can only be done if the user is authorized. Creates access and refresh tokens that needs to be stored. Expires eventually and needs to be refreshed with the refresh token.
	u := url.URL{Scheme: "http", Host: *addr, Path: "/o/token"}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)
	// Example response should include: access_token, refresh_token, token_type, expires_in, scope
}

func RefreshToken(refreshToken string) {
	// Generate json request body
	data := map[string]string{"client_id": client_id, "client_secret": client_secret, "grant_type": "refresh_token", "refresh_token": refreshToken}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}
	// Can only be done if the user is logged in. Creates access and refresh tokens that needs to be stored. Expires eventually and needs to be refreshed with the refresh token.
	req, err := http.NewRequest("POST", "http://localhost:8000/o/token", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", body)
	// Example response should include: access_token, refresh_token, token_type, expires_in, scope
}

// var assets embed.FS

func Run() {
	// err := wails.Run(&options.App{
	// 	Title:     "Authorize",
	// 	Width:     1024,
	// 	Height:    768,
	// 	Frameless: true,
	// 	AssetServer: &assetserver.Options{
	// 		Assets: assets,
	// 		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 			if len(r.URL.Path) > 7 && r.URL.Path[:7] == "/skins/" {
	// 				//skinsFileServer.ServeHTTP(w, r)
	// 				return
	// 			}
	// 			http.NotFound(w, r)
	// 		}),
	// 	},
	// 	BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},

	// 	OnStartup: func(ctx context.Context) {
	// 		Authorize()
	// 	},
	// })

	// if err != nil {
	// 	println("Error:", err.Error())
	// }

	// Connect to race
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// requires authentication
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws/o/bot/clean-pacman-8175"}
	log.Printf("connecting to %s", u.String())

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 45 * time.Second

	// Add custom headers to the WebSocket handshake request
	requestHeader := http.Header{}
	// requestHeader.Set("Authorization", access_token)

	c, _, err := dialer.Dial(u.String(), requestHeader)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

// // create race address (Webpage to be opened when clicking on New Race. Relative to race server)
// /

// // OAUTH_CHALLENGE_METHOD (Plain or S256)
// S256

// // OAUTH_ENDPOINT_FAILURE
// o/done?error=access_denied

// // OAUTH_ENDPOINT_REVOKE
// o/revoke_token

// // OAUTH_ENDPOINT_SUCCESS
// o/done

// // OAUTH_ENDPOINT_USERINFO
// o/userinfo

// // OAUTH_REDIRECT_ADDRESS
// 127.0.0.1

// // OAUTH_REDIRECT_PORT
// 4888

// // OAUTH_SCOPES
// read chat_message race_action

// // OAUTH_SERVER
// https://racetime.gg/

// // PROTOCOL_REST (http or https)
// https

// // PROTOCOL_WEBSOCKET (ws or wss)
// wss
