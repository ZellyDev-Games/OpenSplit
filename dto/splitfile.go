package dto

type Segment struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Gold    int64  `json:"gold"`
	Average int64  `json:"average"`
	PB      int64  `json:"pb"`
}

type Split struct {
	SplitIndex        int    `json:"split_index"`
	SplitSegmentID    string `json:"split_segment_id"`
	CurrentCumulative int64  `json:"current_cumulative"`
	CurrentDuration   int64  `json:"current_duration"`
}

type Run struct {
	ID               string  `json:"id"`
	SplitFileID      string  `json:"splitfile_id"`
	SplitFileVersion int     `json:"splitfile_version"`
	TotalTime        int64   `json:"total_time"`
	Splits           []Split `json:"splits"`
	Completed        bool    `json:"completed"`
}

// SplitFile represents the data and history of a game/category combo.
type SplitFile struct {
	ID           string    `json:"id"`
	Version      int       `json:"version"`
	GameName     string    `json:"game_name"`
	GameCategory string    `json:"game_category"`
	WindowX      int       `json:"window_x"`
	WindowY      int       `json:"window_y"`
	WindowHeight int       `json:"window_height"`
	WindowWidth  int       `json:"window_width"`
	Runs         []Run     `json:"runs"`
	Segments     []Segment `json:"segments"`
	SOB          int64     `json:"sob"`
	PB           *Run      `json:"pb"`
	PBTotalTime  int64     `json:"pb_total_time"`
}
