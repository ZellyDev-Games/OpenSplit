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
	Token    *oauth2.Token
	verifier string
	conf     *oauth2.Config
}

func NewService(RestProtocol string, WebRaceServer string, RedirectURL string) *WebRace {
	client_id := "x4oiff8OAiWwtfQUboFhFlYfgmDMHmxduOFOQgve"
	client_secret := "1BYxBFqyO495W8VCYiZxAEXgortlLa5trpzY0xxDHNAuAWaqfxhgy4435Gq5yp6P76Hw1EIFdp8JjnKvDtDfzLZ2lo6D1TrrWlp0yNbmBTPpNxYVePSqE7eX72ZDAmaU"

	return &WebRace{
		verifier: oauth2.GenerateVerifier(),
		conf: &oauth2.Config{
			ClientID:     client_id,
			ClientSecret: client_secret,
			Scopes:       []string{"read", "chat_message", "race_action"},
			// RedirectURL:  RestProtocol + "://" + RedirectURL + "/oauth/callback",
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
	url = w.conf.AuthCodeURL("state", oauth2.AccessTypeOnline, oauth2.S256ChallengeOption(w.verifier))
	fmt.Printf("URL for the auth dialog: %v\n", url)
	return url
}

// Requests tokens from authorization code
// Can only be done if the user is authorized. Creates access and refresh tokens that needs to be stored. Expires eventually and needs to be refreshed with the refresh token.
// Example response should include: access_token, refresh_token, token_type, expires_in, scope
func (w *WebRace) GenTokens(code string) (accessToken string, refreshToken string) {
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

	return w.Token.AccessToken, w.Token.RefreshToken
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
