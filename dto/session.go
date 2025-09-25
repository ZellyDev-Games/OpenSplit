package dto

import (
	"github.com/zellydev-games/opensplit/session"
)

type Session struct {
	LoadedSplitFile     *SplitFile    `json:"loaded_split_file"`
	CurrentRun          *Run          `json:"current_run"`
	CurrentSegmentIndex int           `json:"current_segment_index"`
	SessionState        session.State `json:"session_state"`
	Dirty               bool          `json:"dirty"`
}
