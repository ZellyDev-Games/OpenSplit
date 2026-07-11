package speedrun

import (
	"net/http"
	"time"

	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/session"
)

const logModule = "main"

const speedRunBase = "https://www.speedrun.com/api/v1"

type Service struct {
	baseURL string
	client  *http.Client

	platforms map[string]string
}

func NewService() *Service {
	return &Service{
		baseURL: speedRunBase,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *Service) Startup() {
	platforms, err := s.fetchPlatforms()
	if err != nil {
		logger.Errorf(logModule, "failed to load speedrun platforms: %v", err)
	}

	s.platforms = make(map[string]string, len(platforms.Data))
	for _, p := range platforms.Data {
		s.platforms[p.ID] = p.Name
	}
}

func (s *Service) ToWorldRecord(result WRSearchResult) session.WorldRecord {
	if len(result.Data) == 0 {
		return session.WorldRecord{}
	}

	if len(result.Data[0].Runs) == 0 {
		return session.WorldRecord{}
	}

	players := make([]string, 0, len(result.Data[0].Players.Data))
	for _, p := range result.Data[0].Players.Data {
		players = append(players, p.Names.International)
	}

	run := result.Data[0].Runs[0].Run

	return session.WorldRecord{
		Show:       true,
		RunID:      run.ID,
		Players:    players,
		RealTime:   run.Time.RealTime,
		InGameTime: run.Time.InGame,
	}
}
