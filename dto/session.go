package dto

type SessionState byte

type Session struct {
	LoadedSplitFile     *SplitFile   `json:"loaded_split_file"`
	CurrentRun          *Run         `json:"current_run"`
	CurrentSegmentIndex int          `json:"current_segment_index"`
	SessionState        SessionState `json:"session_state"`
	Dirty               bool         `json:"dirty"`
}
