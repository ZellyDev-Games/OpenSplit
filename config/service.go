package config

import (
	"os"
	"runtime"

	"github.com/zellydev-games/opensplit/hotkeys"
	"github.com/zellydev-games/opensplit/logger"
)

// Service holds configuration options so that Service.GetEnvironment can work for both backend and frontend.
type Service struct {
	SpeedRunAPIBase string                     `json:"speed_run_API_base"`
	KeyConfig       map[string]hotkeys.KeyInfo `json:"key_config"`
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

// UpdateKeyBinding changes the KeyInfo for the given command.
func (s *Service) UpdateKeyBinding(command string, newKeyInfo hotkeys.KeyInfo) {
	s.KeyConfig[command] = newKeyInfo
}

func CreateDefaultConfig() *Service {
	keyConfig := map[string]hotkeys.KeyInfo{}
	switch runtime.GOOS {
	case "windows":
		keyConfig["SPLIT"] = hotkeys.KeyInfo{
			KeyCode:    32,
			LocaleName: "SPACE",
		}
	default:
		logger.Warn("OS not yet supported, setting zero value defaults to prevent crash, but hotkeys almost certainly will not work")
		keyConfig["SPLIT"] = hotkeys.KeyInfo{}
	}

	return &Service{
		SpeedRunAPIBase: "https://www.speedrun.com/api/v1",
		KeyConfig:       keyConfig,
	}
}
