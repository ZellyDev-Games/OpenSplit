package speedrun

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type gameSearchResponse struct {
	Data []gameSearchAPIItem `json:"data"`
}

type gameSearchAPIItem struct {
	ID          string          `json:"id"`
	Names       GameSearchNames `json:"names"`
	PlatformIDs []string        `json:"platforms"`
}

type GameSearchNames struct {
	International string `json:"international"`
}

type GameSearchResult struct {
	Data []GameSearchItem `json:"data"`
}

type GameSearchItem struct {
	ID        string          `json:"id"`
	Names     GameSearchNames `json:"names"`
	Platforms []Platforms     `json:"platforms"`
}

func (s *Service) SearchGames(query string) (GameSearchResult, error) {
	endpoint := fmt.Sprintf(
		"%s/games?name=%s&max=10",
		s.baseURL,
		url.QueryEscape(query),
	)
	log.Println(endpoint)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return GameSearchResult{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go-HTTPS-Client-Example")

	resp, err := s.client.Do(req)
	if err != nil {
		return GameSearchResult{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	fmt.Printf("Response Status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return GameSearchResult{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var api gameSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&api); err != nil {
		return GameSearchResult{}, err
	}

	log.Printf("%+v\n", api.Data[0])

	result := GameSearchResult{
		Data: make([]GameSearchItem, len(api.Data)),
	}

	for i, game := range api.Data {
		result.Data[i] = GameSearchItem{
			ID:        game.ID,
			Names:     game.Names,
			Platforms: s.getPlatformMatches(game.PlatformIDs),
		}
	}

	log.Println(result)
	return result, nil
}
