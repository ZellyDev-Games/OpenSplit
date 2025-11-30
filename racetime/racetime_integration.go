package racetime

// TODO:
// Convert client_id and client_secret to live site (AFTER getting approval from racetime.gg staff)

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/oauth2"
)

type WebRace struct {
	Token *oauth2.Token
	// restProtocol      string
	// restAddr          string
	// websocketProtocol string
	// websocketAddr     string
	// client_id         string
	// client_secret string
	verifier string
	conf     *oauth2.Config
}

func NewService(RestProtocol string, WebRaceServer string, RedirectURL string) *WebRace {
	client_id := "x4oiff8OAiWwtfQUboFhFlYfgmDMHmxduOFOQgve"
	client_secret := "1BYxBFqyO495W8VCYiZxAEXgortlLa5trpzY0xxDHNAuAWaqfxhgy4435Gq5yp6P76Hw1EIFdp8JjnKvDtDfzLZ2lo6D1TrrWlp0yNbmBTPpNxYVePSqE7eX72ZDAmaU"

	return &WebRace{
		// restProtocol:      RestProtocol,
		// restAddr:          "://" + WebRaceServer,
		// websocketProtocol: "ws",
		// websocketAddr:     "://" + WebRaceServer,
		verifier: oauth2.GenerateVerifier(),
		// client_id:         client_id,
		// client_secret: client_secret,
		conf: &oauth2.Config{
			ClientID:     client_id,
			ClientSecret: client_secret,
			Scopes:       []string{"read", "chat_message", "race_action"},
			RedirectURL:  RestProtocol + "://" + RedirectURL + "/oauth/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  RestProtocol + "://" + WebRaceServer + "/o/authorize",
				TokenURL: RestProtocol + "://" + WebRaceServer + "/o/token",
			},
		},
	}
}

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

// // domain (Domain or IP of the Race-Server)
// racetime.gg

// // Gets all current races
// func (w *WebRace) GetRaces() ([]byte, error) {
// 	// u := url.URL{Scheme: "http", Host: *addr, Path: "/races/data"}
// 	req, err := http.NewRequest("GET", w.restProtocol+w.restAddr+"/races/data", nil)

// 	if err != nil {
// 		log.Fatalf("Error creating request: %v", err)
// 	}
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error reading response body: %v", err)
// 	}

// 	fmt.Printf("Response Status: %s\n", resp.Status)
// 	fmt.Printf("Response Body: %s\n", body)
// 	// Example Response Body: {"races": [{"name": "alttp/perfect-ivysaur-9765", "status": {"value": "open", "verbose_value": "Open", "help_text": "Anyone may join this race"}, "url": "/alttp/perfect-ivysaur-9765", "data_url": "/alttp/perfect-ivysaur-9765/data", "goal": {"name": "100%", "custom": false}, "info": "", "entrants_count": 0, "entrants_count_finished": 0, "entrants_count_inactive": 0, "opened_at": "2025-11-25T14:22:38.834Z", "started_at": null, "time_limit": "P1DT00H00M00S", "opened_by_bot": null, "category": {"name": "The Legend of Zelda: A Link to the Past", "short_name": "ALttP", "slug": "alttp", "url": "/alttp", "data_url": "/alttp/data", "image": null}}]}

// 	return body, err
// }

// Get category details
// func (w *WebRace) GetCategoryDetails(category string) ([]byte, error) {
// 	// category = "alttp"
// 	// u := url.URL{Scheme: "http", Host: *addr, Path: "/" + category + "/data"}
// 	req, err := http.NewRequest("GET", w.restProtocol+w.restAddr+category+"/data", nil)
// 	if err != nil {
// 		log.Fatalf("Error creating request: %v", err)
// 	}
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error reading response body: %v", err)
// 	}

// 	fmt.Printf("Response Status: %s\n", resp.Status)
// 	fmt.Printf("Response Body: %s\n", body)
// 	// Example Response Body: {"name": "The Legend of Zelda: A Link to the Past", "short_name": "ALttP", "slug": "alttp", "url": "/alttp", "data_url": "/alttp/data", "image": null, "info": null, "streaming_required": true, "owners": [{"id": "GvzqPgEyPdZ0RKnr", "full_name": "Luigi#5557", "name": "Luigi", "discriminator": "5557", "url": "/user/GvzqPgEyPdZ0RKnr/luigi", "avatar": null, "pronouns": null, "flair": "moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}], "moderators": [{"id": "xr85vpEMBoX32zJ4", "full_name": "Bowser#7723", "name": "Bowser", "discriminator": "7723", "url": "/user/xr85vpEMBoX32zJ4/bowser", "avatar": null, "pronouns": null, "flair": "moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}], "goals": ["100%", "16 stars", "Beat the game"], "current_races": [], "emotes": {}}

// 	return body, err
// }

// Get category past races
// show_entrants can be either true, false, or empty
// page selects a page of the returned data, -1 ignores option
// per_page controls how many results to return per page
// func (w *WebRace) GetCategoryPastRaces(category string, show_entrants string, page int, per_page int) ([]byte, error) {
// 	// Generate json request body
// 	data := make(map[string]string)
// 	switch show_entrants {
// 	case "":
// 		if page != -1 {
// 			data["page"] = strconv.Itoa(page)
// 			data["per_page"] = strconv.Itoa(per_page)
// 		}
// 	case "true":
// 		if page == -1 {
// 			data["show_entrants"] = show_entrants
// 		} else {
// 			data["show_entrants"] = show_entrants
// 			data["page"] = strconv.Itoa(page)
// 			data["per_page"] = strconv.Itoa(per_page)
// 		}
// 	case "false":
// 		if page == -1 {
// 			data["show_entrants"] = show_entrants
// 		} else {
// 			data["show_entrants"] = show_entrants
// 			data["page"] = strconv.Itoa(page)
// 			data["per_page"] = strconv.Itoa(per_page)
// 		}
// 	}

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		log.Fatalf("Error marshalling JSON: %v", err)
// 	}

// 	// category = "alttp"
// 	// u := url.URL{Scheme: "http", Host: *addr, Path: "/" + category + "/races/data"}
// 	req, err := http.NewRequest("GET", w.restProtocol+w.restAddr+category+"/races/data", bytes.NewBuffer(jsonData))
// 	// Example Response Body: {"count": 0, "num_pages": 1, "races": []}

// 	if err != nil {
// 		log.Fatalf("Error creating request: %v", err)
// 	}
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error reading response body: %v", err)
// 	}

// 	fmt.Printf("Response Status: %s\n", resp.Status)
// 	fmt.Printf("Response Body: %s\n", body)
// 	// Example Response Body: {"name": "The Legend of Zelda: A Link to the Past", "short_name": "ALttP", "slug": "alttp", "url": "/alttp", "data_url": "/alttp/data", "image": null, "info": null, "streaming_required": true, "owners": [{"id": "GvzqPgEyPdZ0RKnr", "full_name": "Luigi#5557", "name": "Luigi", "discriminator": "5557", "url": "/user/GvzqPgEyPdZ0RKnr/luigi", "avatar": null, "pronouns": null, "flair": "moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}], "moderators": [{"id": "xr85vpEMBoX32zJ4", "full_name": "Bowser#7723", "name": "Bowser", "discriminator": "7723", "url": "/user/xr85vpEMBoX32zJ4/bowser", "avatar": null, "pronouns": null, "flair": "moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}], "goals": ["100%", "16 stars", "Beat the game"], "current_races": [], "emotes": {}}

// 	return body, err
// }

// Gets category leaderboard data
// func (w *WebRace) GetCategoryLeaderboards(category string) ([]byte, error) {
// 	// category = "alttp"
// 	// u := url.URL{Scheme: "http", Host: *addr, Path: "/" + category + "/leaderboards/data"}
// 	req, err := http.NewRequest("GET", w.restProtocol+w.restAddr+"/leaderboards/data", nil)
// 	if err != nil {
// 		log.Fatalf("Error creating request: %v", err)
// 	}
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error reading response body: %v", err)
// 	}

// 	fmt.Printf("Response Status: %s\n", resp.Status)
// 	fmt.Printf("Response Body: %s\n", body)
// 	// Example Response Body: {"leaderboards": [{"goal": "100%", "num_ranked": 0, "rankings": []}, {"goal": "16 stars", "num_ranked": 0, "rankings": []}, {"goal": "Beat the game", "num_ranked": 0, "rankings": []}]}

// 	return body, err
// }

// Gets category race info
// func (w *WebRace) GetCategoryRaceInfo(category string, race string) ([]byte, error) {
// 	// category = "alttp"
// 	// race := "funky-link-3070"
// 	// u := url.URL{Scheme: "http", Host: *addr, Path: "/" + category + "/" + race + "/data"}
// 	req, err := http.NewRequest("GET", w.restProtocol+w.restAddr+category+race+"/data", nil)
// 	if err != nil {
// 		log.Fatalf("Error creating request: %v", err)
// 	}
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error reading response body: %v", err)
// 	}

// 	fmt.Printf("Response Status: %s\n", resp.Status)
// 	fmt.Printf("Response Body: %s\n", body)
// 	// Example Response Body: {"version": 1, "name": "alttp/funky-link-3070", "slug": "funky-link-3070", "status": {"value": "open", "verbose_value": "Open", "help_text": "Anyone may join this race"}, "url": "/alttp/funky-link-3070", "data_url": "/alttp/funky-link-3070/data", "websocket_url": "/ws/race/funky-link-3070", "websocket_bot_url": "/ws/o/bot/funky-link-3070", "websocket_oauth_url": "/ws/o/race/funky-link-3070", "category": {"name": "The Legend of Zelda: A Link to the Past", "short_name": "ALttP", "slug": "alttp", "url": "/alttp", "data_url": "/alttp/data", "image": null}, "goal": {"name": "100%", "custom": false}, "info": "", "info_bot": null, "info_user": "", "team_race": false, "entrants_count": 0, "entrants_count_finished": 0, "entrants_count_inactive": 0, "entrants": [], "opened_at": "2025-11-25T15:05:51.047Z", "start_delay": "P0DT00H00M15S", "started_at": null, "ended_at": null, "cancelled_at": null, "ranked": true, "unlisted": false, "time_limit": "P1DT00H00M00S", "time_limit_auto_complete": false, "require_even_teams": false, "streaming_required": true, "auto_start": true, "opened_by": {"id": "5BRGVMd30E368Lzv", "full_name": "Douglas Kirby", "name": "Douglas Kirby", "discriminator": null, "url": "/user/5BRGVMd30E368Lzv/douglas-kirby", "avatar": null, "pronouns": null, "flair": "staff moderator", "twitch_name": null, "twitch_display_name": null, "twitch_channel": null, "can_moderate": true}, "opened_by_bot": null, "monitors": [], "recordable": true, "recorded": false, "recorded_by": null, "disqualify_unready": false, "allow_comments": true, "hide_comments": false, "hide_entrants": false, "chat_restricted": false, "allow_prerace_chat": true, "allow_midrace_chat": true, "allow_non_entrant_chat": true, "chat_message_delay": "P0DT00H00M00S", "bot_meta": {}}

// 	return body, err
// }

// Gets user past race info
// show_entrants can be either true, false, or empty
// page selects a page of the returned data, -1 ignores option
// per_page controls how many results to return per page
// func (w *WebRace) GetUserPastRaces(user string, show_entrants string, page int, per_page int) ([]byte, error) {
// 	// Generate json request body
// 	data := make(map[string]string)
// 	switch show_entrants {
// 	case "":
// 		if page != -1 {
// 			data["page"] = strconv.Itoa(page)
// 			data["per_page"] = strconv.Itoa(per_page)
// 		}
// 	case "true":
// 		if page == -1 {
// 			data["show_entrants"] = show_entrants
// 		} else {
// 			data["show_entrants"] = show_entrants
// 			data["page"] = strconv.Itoa(page)
// 			data["per_page"] = strconv.Itoa(per_page)
// 		}
// 	case "false":
// 		if page == -1 {
// 			data["show_entrants"] = show_entrants
// 		} else {
// 			data["show_entrants"] = show_entrants
// 			data["page"] = strconv.Itoa(page)
// 			data["per_page"] = strconv.Itoa(per_page)
// 		}
// 	}

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		log.Fatalf("Error marshalling JSON: %v", err)
// 	}

// 	// u := url.URL{Scheme: "http", Host: *addr, Path: "/user/" + user + "/races/data"}
// 	req, err := http.NewRequest("GET", w.restProtocol+w.restAddr+user+"/races/data", bytes.NewBuffer(jsonData))

// 	if err != nil {
// 		log.Fatalf("Error creating request: %v", err)
// 	}
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error reading response body: %v", err)
// 	}

// 	fmt.Printf("Response Status: %s\n", resp.Status)
// 	fmt.Printf("Response Body: %s\n", body)

// 	return body, err
// }

// User search
// func (w *WebRace) UserSearch(user string, discriminator string) ([]byte, error) {
// 	// Generate json request body
// 	data := make(map[string]string)
// 	data["name"] = user
// 	if discriminator != "" {
// 		data["discriminator"] = discriminator
// 	}

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		log.Fatalf("Error marshalling JSON: %v", err)
// 	}

// 	// u := url.URL{Scheme: "http", Host: *addr, Path: "/user/search"}
// 	req, err := http.NewRequest("GET", w.restProtocol+w.restAddr+"/user/search", bytes.NewBuffer(jsonData))

// 	if err != nil {
// 		log.Fatalf("Error creating request: %v", err)
// 	}
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error reading response body: %v", err)
// 	}

// 	fmt.Printf("Response Status: %s\n", resp.Status)
// 	fmt.Printf("Response Body: %s\n", body)

// 	return body, err
// }

// TODO: make token available
// func (*WebRace) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
// ctx := context.Background()
// code := r.URL.Query().Get("code")
// tok, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
// if err != nil {
// log.Fatal(err)
// }
// }

func (w *WebRace) CheckTokens() bool {
	if w.Token == nil || (w.Token.AccessToken == "" && w.Token.RefreshToken == "") {
		return false
	}
	if !w.Token.Valid() {
		if w.Token.RefreshToken != "" {
			w.RefreshTokens()
			return true
		} else {
			return false
		}
	}
	return true
}

func (w *WebRace) Authorize() (url string) {
	// generates "correctly". Throws a csrf error constantly
	url = w.conf.AuthCodeURL("state", oauth2.AccessTypeOnline, oauth2.S256ChallengeOption(w.verifier))
	fmt.Printf("URL for the auth dialog: %v", url)
	return url
}

// func Authorize() {
// 	req, err := http.NewRequest("GET", "http://localhost:8000/o/authorize", nil)

// 	if err != nil {
// 		log.Fatalf("Error creating request: %v", err)
// 	}

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error reading response body: %v", err)
// 	}

// 	fmt.Printf("Response Status: %s\n", resp.Status)
// 	//returns full html page
// 	fmt.Printf("Response Body: %s\n", body)
// }

// Requests tokens from authorization code
// Can only be done if the user is authorized. Creates access and refresh tokens that needs to be stored. Expires eventually and needs to be refreshed with the refresh token.
// Example response should include: access_token, refresh_token, token_type, expires_in, scope
func (w *WebRace) GenTokens(code string) {
	// // Use the authorization code that is pushed to the redirect
	// // URL. Exchange will do the handshake to retrieve the
	// // initial access token.
	ctx := context.Background()
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := w.conf.Exchange(ctx, code, oauth2.VerifierOption(w.verifier))
	if err != nil {
		log.Fatal(err)
	}

	w.Token = tok

	// TODO: STORE THESE BETTER
	fmt.Printf("Access token: %s\n", w.Token.AccessToken)
	fmt.Printf("Refresh token: %s\n", w.Token.RefreshToken)
	fmt.Printf("Access token expires: %s\n", w.Token.Expiry)
	fmt.Printf("Access token expires: %v\n", w.Token.ExpiresIn)
}

// Can only be done if the user is logged in. Refreshes tokens that needs to be stored.
// Example response should include: access_token, refresh_token, token_type, expires_in, scope
func (w *WebRace) RefreshTokens() {
	ctx := context.Background()

	// TODO: catch errors
	// no token, auth revoked
	w.conf.TokenSource(ctx, w.Token)

	fmt.Printf("Access token: %s\n", w.Token.AccessToken)
	fmt.Printf("Refresh token: %s\n", w.Token.RefreshToken)
	fmt.Printf("Access token expires: %s\n", w.Token.Expiry)
	fmt.Printf("Access token expires: %v\n", w.Token.ExpiresIn)
}

func (w *WebRace) GetAccessToken() (accessToken string) {
	return w.Token.AccessToken
}

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

// func (w *WebRace) RaceData(dataURL string) {
// "data_url": "/alttp/funky-link-3070/data"
// }

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

// Connects to websocket to get and send chat commands
// Example socket url "websocket_oauth_url": "/ws/o/race/funky-link-3070"
// func (w *WebRace) RaceChat(chatURL string) {
// 	// Connect to race
// 	interrupt := make(chan os.Signal, 1)
// 	signal.Notify(interrupt, os.Interrupt)

// 	// chatURL = "/ws/o/race/funky-link-3070"
// 	// requires authentication
// 	addr := w.websocketProtocol + w.websocketAddr + chatURL
// 	log.Printf("connecting to %s", addr)

// 	dialer := websocket.DefaultDialer
// 	dialer.HandshakeTimeout = 45 * time.Second

// 	// // Add custom headers to the WebSocket handshake request
// 	requestHeader := http.Header{}
// 	requestHeader.Set("Authorization", w.Token.AccessToken)

// 	c, _, err := dialer.Dial(addr, requestHeader)
// 	if err != nil {
// 		log.Fatal("Dial error:", err)
// 	}
// 	defer c.Close()

// 	done := make(chan struct{})

// 	go func() {
// 		defer close(done)
// 		for {
// 			_, message, err := c.ReadMessage()
// 			if err != nil {
// 				log.Println("read:", err)
// 				return
// 			}
// 			log.Printf("recv: %s", message)
// 		}
// 	}()

// 	ticker := time.NewTicker(time.Second)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-done:
// 			return
// 		case t := <-ticker.C:
// 			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
// 			if err != nil {
// 				log.Println("write:", err)
// 				return
// 			}
// 		case <-interrupt:
// 			log.Println("interrupt")

// 			// Cleanly close the connection by sending a close message and then
// 			// waiting (with timeout) for the server to close the connection.
// 			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
// 			if err != nil {
// 				log.Println("write close:", err)
// 				return
// 			}
// 			select {
// 			case <-done:
// 			case <-time.After(time.Second):
// 			}
// 			return
// 		}
// 	}
// }

// func (w *WebRace) Run() {
// }

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
