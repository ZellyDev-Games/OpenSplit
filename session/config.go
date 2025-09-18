package session

// Config holds configuration options so that Service.GetConfig can work for both backend and frontend.
type Config struct {
	SpeedRunAPIBase string `json:"speed_run_API_base"`
}
