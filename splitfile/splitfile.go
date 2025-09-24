package splitfile

// WindowParams stores the last size and position the user set splitter window while this file was loaded
type WindowParams struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

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
	ID               string    `json:"id"`
	SplitFileID      string    `json:"splitfile_id"`
	SplitFileVersion int       `json:"splitfile_version"`
	SOB              int64     `json:"sob"`
	TotalTime        int64     `json:"total_time"`
	Segments         []Segment `json:"segments"`
	Splits           []Split   `json:"splits"`
}

// SplitFile represents the data and history of a game/category combo.
type SplitFile struct {
	ID           string       `json:"id"`
	Version      int          `json:"version"`
	GameName     string       `json:"game_name"`
	GameCategory string       `json:"game_category"`
	WindowParams WindowParams `json:"window_params"`
	Runs         []Run        `json:"runs"`
	Segments     []Segment    `json:"segments"`
	SOB          int64        `json:"SOB"`
}
