package dto

// Segment represents the data of a single split segment.
type Segment struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Gold     int64     `json:"gold"`
	Average  int64     `json:"average"`
	PB       int64     `json:"pb"`
	Icon     string    `json:"icon,omitempty"` // base64 PNG
	Children []Segment `json:"children"`
}

// Split represents the cumulative data of a segment.
type Split struct {
	SplitSegmentID    string `json:"split_segment_id"`
	CurrentCumulative int64  `json:"current_cumulative"`
	CurrentDuration   int64  `json:"current_duration"`
}

type Variable struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	ValueID string `json:"value"`
	Label   string `json:"label"`
}

// SplitterWindow represents the persisted position and size of the
// splitter window for one layout.
type SplitterWindow struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// SplitterWindows contains independent window geometry for each
// splitter layout.
type SplitterWindows struct {
	Vertical   SplitterWindow `json:"vertical"`
	Horizontal SplitterWindow `json:"horizontal"`
}

// SplitFile represents the data and history of a game/category combo.
type SplitFile struct {
	ID           string     `json:"id"`
	GameName     string     `json:"game_name"`
	GameID       string     `json:"speedrun_game_id"`
	GameCategory string     `json:"game_category"`
	CategoryID   string     `json:"speedrun_game_category_id"`
	Variables    []Variable `json:"variables"`

	Version int `json:"version"`

	SelectedSkin string `json:"selected_skin"`

	Segments []Segment `json:"segments"`
	Runs     []Run     `json:"runs"`
	PB       *Run      `json:"pb"`

	SOB      int64  `json:"sob"`
	Attempts int    `json:"attempts"`
	Offset   int64  `json:"offset"`
	Platform string `json:"platform"`

	WR WorldRecord `json:"wr"`

	Layout string `json:"layout"`

	Windows SplitterWindows `json:"windows"`

	// Legacy single-window fields.
	//
	// These are retained only so older split files can be loaded and
	// migrated into Windows. New files should use Windows instead.
	WindowX      int `json:"window_x,omitempty"`
	WindowY      int `json:"window_y,omitempty"`
	WindowWidth  int `json:"window_width,omitempty"`
	WindowHeight int `json:"window_height,omitempty"`
}

// WorldRecord represents the data of the current world record.
type WorldRecord struct {
	Show       bool     `json:"show"`
	RunID      string   `json:"run_id"`
	Players    []string `json:"players"`
	RealTime   float64  `json:"real_time"`
	InGameTime float64  `json:"in_game_time"`
}
