package config

import (
	"os"
	"runtime"
	"sync"

	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
)

// Service holds configuration options so that Service.GetEnvironment can work for both backend and frontend.
type Service struct {
	mu                   sync.Mutex
	SpeedRunAPIBase      string                                 `json:"speed_run_API_base"`
	KeyConfig            map[dispatcher.Command]keyinfo.KeyData `json:"key_config"`
	configUpdatedChannel chan<- *Service
}

func NewService() (*Service, chan *Service) {
	updateChannel := make(chan *Service)
	return &Service{
		SpeedRunAPIBase:      "",
		KeyConfig:            map[dispatcher.Command]keyinfo.KeyData{},
		configUpdatedChannel: updateChannel,
	}, updateChannel
}

// GetEnvironment is designed to expose configuration options from the environment or other sources (config files) to the
// frontend.  Go services can just read the environment, but the frontend has no reliable way to do so, so this func
// is bound to the app in main which generates a typescript function for the frontend.
func (s *Service) GetEnvironment() *Service {
	speedRunBase := os.Getenv("SPEEDRUN_API_BASE")
	if speedRunBase == "" {
		speedRunBase = "https://www.speedrun.com/api/v1"
	}
	return &Service{
		SpeedRunAPIBase: speedRunBase,
	}
}

// UpdateKeyBinding changes the ConfigPayload for the given command.
func (s *Service) UpdateKeyBinding(command dispatcher.Command, data keyinfo.KeyData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.KeyConfig[command] = data
	s.sendUIBridgeUpdate()

}

// CreateDefaultConfig sets the service's options to reasonable defaults.
//
// Useful if the config file hasn't been created yet (first run)
func (s *Service) CreateDefaultConfig() {
	s.KeyConfig = map[dispatcher.Command]keyinfo.KeyData{}
	switch runtime.GOOS {
	case "windows":
		s.KeyConfig[dispatcher.SPLIT] = keyinfo.KeyData{
			KeyCode:    32,
			LocaleName: "SPACE",
		}
		s.KeyConfig[dispatcher.UNDO] = keyinfo.KeyData{}
		s.KeyConfig[dispatcher.SKIP] = keyinfo.KeyData{}
		s.KeyConfig[dispatcher.PAUSE] = keyinfo.KeyData{}
		s.KeyConfig[dispatcher.RESET] = keyinfo.KeyData{}

	default:
		logger.Warn("OS not yet supported, setting zero value defaults to prevent crash, but hotkeys almost certainly will not work")
		s.KeyConfig[dispatcher.SPLIT] = keyinfo.KeyData{}
		s.KeyConfig[dispatcher.UNDO] = keyinfo.KeyData{}
		s.KeyConfig[dispatcher.SKIP] = keyinfo.KeyData{}
		s.KeyConfig[dispatcher.PAUSE] = keyinfo.KeyData{}
		s.KeyConfig[dispatcher.RESET] = keyinfo.KeyData{}
	}

	s.sendUIBridgeUpdate()
}

func (s *Service) sendUIBridgeUpdate() {
	select {
	case s.configUpdatedChannel <- s:
	default:
	}
}
