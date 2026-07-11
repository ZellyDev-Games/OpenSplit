package dto

type Segment struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Gold     int64     `json:"gold"`
	Average  int64     `json:"average"`
	PB       int64     `json:"pb"`
	Children []Segment `json:"children"`
}

type Split struct {
	SplitSegmentID    string `json:"split_segment_id"`
	CurrentCumulative int64  `json:"current_cumulative"`
	CurrentDuration   int64  `json:"current_duration"`
}

// SplitFile represents the data and history of a game/category combo.
type SplitFile struct {
	ID           string `json:"id"`
	GameName     string `json:"game_name"`
	GameID       string `json:"speedrun_game_id"`
	GameCategory string `json:"game_category"`
	CategoryID   string `json:"speedrun_game_category_id"`
	Version      int    `json:"version"`

	SelectedSkin string `json:"selected_skin"`

	Segments []Segment `json:"segments"`
	Runs     []Run     `json:"runs"`
	PB       *Run      `json:"pb"`

	SOB      int64  `json:"sob"`
	Attempts int    `json:"attempts"`
	Offset   int64  `json:"offset"`
	Platform string `json:"platform"`

	WR WorldRecord `json:"wr"`

	WindowX      int `json:"window_x"`
	WindowY      int `json:"window_y"`
	WindowWidth  int `json:"window_width"`
	WindowHeight int `json:"window_height"`
}

type WorldRecord struct {
	Show       bool     `json:"show"`
	RunID      string   `json:"run_id"`
	Players    []string `json:"players"`
	RealTime   float64  `json:"real_time"`
	InGameTime float64  `json:"in_game_time"`
}
