package config

import "os"

// Config holds configuration options so that Service.GetConfig can work for both backend and frontend.
type Config struct {
	SpeedRunAPIBase string `json:"speed_run_API_base"`
}

// GetConfig is designed to expose configuration options from the environment or other sources (config files) to the
// frontend.  Go services can just read the environment, but the frontend has no reliable way to do so, so this func
// is bound to the app in main which generates a typescript function for the frontend.
func (c *Config) GetConfig() *Config {
	speedRunBase := os.Getenv("SPEEDRUN_API_BASE")
	if speedRunBase == "" {
		speedRunBase = "https://www.speedrun.com/api/v1"
	}
	return &Config{
		SpeedRunAPIBase: speedRunBase,
	}
}
