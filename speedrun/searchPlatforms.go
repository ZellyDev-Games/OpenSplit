package speedrun

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

type PlatformResult struct {
	Data []Platforms `json:"data"`
}

type Platforms struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Service) fetchPlatforms() (PlatformResult, error) {
	const pageSize = 20

	var all PlatformResult

	for offset := 0; ; offset += pageSize {
		var page PlatformResult

		err := func() error {
			endpoint := fmt.Sprintf(
				"%s/platforms?max=%d&offset=%d",
				s.baseURL,
				pageSize,
				offset,
			)
			log.Println(endpoint)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if err != nil {
				return err
			}

			req.Header.Set("Accept", "application/json")
			req.Header.Set("User-Agent", "Go-HTTPS-Client-Example")

			resp, err := s.client.Do(req)
			if err != nil {
				return err
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Printf("error closing response body: %v", err)
				}
			}()

			fmt.Printf("Response Status: %s\n", resp.Status)
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status: %s", resp.Status)
			}

			return json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return PlatformResult{}, err
		}

		all.Data = append(all.Data, page.Data...)

		if len(page.Data) < pageSize {
			break
		}
	}

	log.Println(all)
	return all, nil
}

func (s *Service) getPlatformMatches(ids []string) []Platforms {
	platforms := make([]Platforms, 0, len(ids))

	for _, id := range ids {
		if name, ok := s.platforms[id]; ok {
			platforms = append(platforms, Platforms{
				ID:   id,
				Name: name,
			})
		}
	}

	return platforms
}

func (s *Service) Platforms() []Platforms {
	result := make([]Platforms, 0, len(s.platforms))

	for id, name := range s.platforms {
		result = append(result, Platforms{
			ID:   id,
			Name: name,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}
