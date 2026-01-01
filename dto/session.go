package dto

type SessionState byte

type Session struct {
	LoadedSplitFile     *SplitFile   `json:"loaded_split_file"`
	LeafSegments        []Segment    `json:"leaf_segments"`
	CurrentRun          *Run         `json:"current_run"`
	CurrentSegmentIndex int          `json:"current_segment_index"`
	SessionState        SessionState `json:"session_state"`
	Dirty               bool         `json:"dirty"`
}

type Run struct {
	ID               string           `json:"id"`
	SplitFileVersion int              `json:"split_file_version"`
	TotalTime        int64            `json:"total_time"`
	Splits           map[string]Split `json:"splits"`
	LeafSegments     []Segment        `json:"leaf_segments"`
	Completed        bool             `json:"completed"`
}
