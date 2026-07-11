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

type CategorySearchResult struct {
	Data []CategorySearchItem `json:"data"`
}

type CategorySearchItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Service) SearchCategories(gameID string) (CategorySearchResult, error) {
	endpoint := fmt.Sprintf(
		"%s/games/%s/categories",
		s.baseURL,
		url.QueryEscape(gameID),
	)
	log.Println(endpoint)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return CategorySearchResult{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go-HTTPS-Client-Example")

	resp, err := s.client.Do(req)
	if err != nil {
		return CategorySearchResult{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	fmt.Printf("Response Status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return CategorySearchResult{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var result CategorySearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return CategorySearchResult{}, err
	}

	log.Println(result)
	return result, nil
}
