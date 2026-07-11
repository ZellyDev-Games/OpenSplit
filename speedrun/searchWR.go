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

type WRSearchResult struct {
	Data []WRSearchItem `json:"data"`
}

type WRSearchItem struct {
	Runs    []WRSearchRuns `json:"runs"`
	Players WRPlayerEmbed  `json:"players"`
}

type WRPlayerEmbed struct {
	Data []WRPlayer `json:"data"`
}

type WRPlayer struct {
	Names WRPlayerNames `json:"names"`
}

type WRPlayerNames struct {
	International string `json:"international"`
}

type WRSearchRuns struct {
	Place int   `json:"place"`
	Run   WRRun `json:"run"`
}

type WRRun struct {
	ID   string `json:"id"`
	Time WRTime `json:"times"`
}

type WRTime struct {
	RealTime float64 `json:"realtime_t"`
	InGame   float64 `json:"ingame_t"`
}

func (s *Service) SearchWR(categoryID string) (WRSearchResult, error) {
	endpoint := fmt.Sprintf(
		"%s/categories/%s/records?top=1&embed=players",
		s.baseURL,
		url.QueryEscape(categoryID),
	)
	log.Println(endpoint)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return WRSearchResult{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go-HTTPS-Client-Example")

	resp, err := s.client.Do(req)
	if err != nil {
		return WRSearchResult{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	fmt.Printf("Response Status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return WRSearchResult{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var result WRSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return WRSearchResult{}, err
	}

	log.Println(result)
	return result, nil
}
