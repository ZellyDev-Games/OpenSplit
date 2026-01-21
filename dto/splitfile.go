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
	ID               string    `json:"id"`
	Version          int       `json:"version"`
	Attempts         int       `json:"attempts"`
	GameName         string    `json:"game_name"`
	GameCategory     string    `json:"game_category"`
	WindowX          int       `json:"window_x"`
	WindowY          int       `json:"window_y"`
	WindowHeight     int       `json:"window_height"`
	WindowWidth      int       `json:"window_width"`
	Runs             []Run     `json:"runs"`
	Segments         []Segment `json:"segments"`
	SOB              int64     `json:"sob"`
	PB               *Run      `json:"pb"`
	Offset           int64     `json:"offset"`
	AutosplitterFile string    `json:"autosplitter_file"`
}
