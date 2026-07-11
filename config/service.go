package config

import (
	"sync"

	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
)

const logModule = "config"

// Service holds configuration options so that Service.GetEnvironment can work for both backend and frontend.
type Service struct {
	mu                   sync.Mutex
	SpeedRunAPIBase      string                              `json:"speed_run_API_base"`
	KeyConfig            map[command.Command]keyinfo.KeyData `json:"key_config"`
	GlobalHotkeysActive  bool                                `json:"global_hotkeys_active"`
	SplitFileDir         string                              `json:"splitfile_dir"`
	SkinsDir             string                              `json:"skins_dir"`
	SelectedSkin         string                              `json:"selected_skin"`
	configUpdatedChannel chan<- *Service
}

func NewService(splitFileDir string, skinsDir string) (*Service, chan *Service) {
	updateChannel := make(chan *Service)
	return &Service{
		SplitFileDir:         splitFileDir,
		SkinsDir:             skinsDir,
		SpeedRunAPIBase:      "",
		KeyConfig:            map[command.Command]keyinfo.KeyData{},
		configUpdatedChannel: updateChannel,
	}, updateChannel
}

// GetEnvironment is designed to expose configuration options from the environment or other sources (config files) to the
// frontend.  Go services can just read the environment, but the frontend has no reliable way to do so, so this func
// is bound to the app in main which generates a typescript function for the frontend.
// func (s *Service) GetEnvironment() *Service {
// 	speedRunBase := os.Getenv("SPEEDRUN_API_BASE")
// 	if speedRunBase == "" {
// 		speedRunBase = "https://www.speedrun.com/api/v1"
// 	}
// 	return &Service{
// 		SpeedRunAPIBase: speedRunBase,
// 	}
// }

// UpdateKeyBinding changes the ConfigPayload for the given command.
func (s *Service) UpdateKeyBinding(c command.Command, data keyinfo.KeyData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.KeyConfig[c] = data
	s.sendUIBridgeUpdate()
	logger.Infof(logModule, "updated key binding for c %v to %s", c, data.LocaleName)
}

// CreateDefaultConfig sets the service's options to reasonable defaults.
//
// Useful if the config file hasn't been created yet (first run)
func (s *Service) CreateDefaultConfig() {
	s.KeyConfig = make(map[command.Command]keyinfo.KeyData)
	s.EnsureDefaultKeyBindings()
	s.sendUIBridgeUpdate()
	logger.Infof(logModule, "created default config")
}

func (s *Service) sendUIBridgeUpdate() {
	select {
	case s.configUpdatedChannel <- s:
	default:
	}
}

// EnsureDefaultKeyBindings adds any missing key binding entries.
// This allows older config files to be upgraded automatically when loaded.
func (s *Service) EnsureDefaultKeyBindings() {
	if s.KeyConfig == nil {
		s.KeyConfig = make(map[command.Command]keyinfo.KeyData)
	}

	defaults := []command.Command{
		command.SPLIT,
		command.UNDO,
		command.SKIP,
		command.PAUSE,
		command.RESET,
		command.COMPARISON_LEFT,
		command.COMPARISON_RIGHT,
	}

	for _, c := range defaults {
		if _, ok := s.KeyConfig[c]; !ok {
			s.KeyConfig[c] = keyinfo.KeyData{}
		}
	}
}
