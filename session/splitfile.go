package session

import (
	"time"

	"github.com/google/uuid"
)

// SplitFile contains a game's split definitions together with run history.
type SplitFile struct {
	ID           uuid.UUID
	GameName     string
	GameID       string
	GameCategory string
	CategoryID   string
	Variables    []Variable

	Version int

	SelectedSkin string

	Segments []Segment
	Runs     []Run
	PB       *Run

	SOB                time.Duration
	Attempts           int
	Offset             time.Duration
	Platform           string
	RollingAverageRuns int

	WR WorldRecord

	Layout string

	Windows SplitterWindows
}

type SplitterWindow struct {
	X      int
	Y      int
	Width  int
	Height int
}

type SplitterWindows struct {
	Vertical   SplitterWindow
	Horizontal SplitterWindow
}

type Variable struct {
	ID      string
	Name    string
	ValueID string // speedrun.com value ID
	Label   string // display label
}

// WorldRecord stores optional world record information displayed in the UI.
type WorldRecord struct {
	Show       bool
	RunID      string
	Players    []string
	RealTime   float64
	InGameTime float64
}

func (s *SplitFile) WindowForLayout(layout string) *SplitterWindow {
	switch layout {
	case "horizontal":
		return &s.Windows.Horizontal
	case "vertical":
		fallthrough
	default:
		return &s.Windows.Vertical
	}
}

func (s *SplitFile) DeepCopyLeafSegments() []Segment {
	leafPointers := getLeafSegments(s.Segments, nil)
	out := make([]Segment, 0, len(leafPointers))
	for i := range leafPointers {
		out = append(out, *leafPointers[i])
	}
	return deepCopySegments(out)
}

func DeepCopySplitFile(inFile *SplitFile) SplitFile {
	segments := deepCopySegments(inFile.Segments)
	runs := deepCopyRuns(inFile.Runs)

	var pbRun *Run
	if inFile.PB != nil {
		pbCopy := deepCopyRun(*inFile.PB)
		pbRun = &pbCopy
	}

	return SplitFile{
		ID:           inFile.ID,
		GameName:     inFile.GameName,
		GameID:       inFile.GameID,
		GameCategory: inFile.GameCategory,
		CategoryID:   inFile.CategoryID,
		Variables:    inFile.Variables,

		Version: inFile.Version,

		SelectedSkin: inFile.SelectedSkin,

		Segments: segments,
		Runs:     runs,
		PB:       pbRun,

		SOB:                inFile.SOB,
		Attempts:           inFile.Attempts,
		Offset:             inFile.Offset,
		Platform:           inFile.Platform,
		RollingAverageRuns: inFile.RollingAverageRuns,

		WR: inFile.WR,

		Layout: inFile.Layout,

		Windows: inFile.Windows,
	}
}

func getLeafSegments(segments []Segment, out []*Segment) []*Segment {
	for i := range segments {
		if len(segments[i].Children) == 0 {
			out = append(out, &segments[i])
		} else {
			out = getLeafSegments(segments[i].Children, out)
		}
	}
	return out
}

func deepCopyRuns(inRuns []Run) []Run {
	runs := make([]Run, len(inRuns))
	for i := range inRuns {
		runs[i] = deepCopyRun(inRuns[i])
	}
	return runs
}

func deepCopyRun(run Run) Run {
	segments := deepCopySegments(run.LeafSegments)
	splits := deepCopySplits(run.Splits)

	return Run{
		ID:               run.ID,
		SplitFileVersion: run.SplitFileVersion,
		TotalTime:        run.TotalTime,
		Splits:           splits,
		Completed:        run.Completed,
		LeafSegments:     segments,
	}
}

func deepCopySplits(inSplits map[uuid.UUID]Split) map[uuid.UUID]Split {
	splits := map[uuid.UUID]Split{}
	for segmentID, split := range inSplits {
		splits[segmentID] = Split{
			SplitSegmentID:    split.SplitSegmentID,
			CurrentCumulative: split.CurrentCumulative,
			CurrentDuration:   split.CurrentDuration,
		}
	}
	return splits
}

func deepCopySegments(list []Segment) []Segment {
	out := make([]Segment, len(list))
	for i, s := range list {
		out[i] = Segment{
			ID:       s.ID,
			Name:     s.Name,
			Gold:     s.Gold,
			Average:  s.Average,
			PB:       s.PB,
			Icon:     s.Icon,
			Children: deepCopySegments(s.Children),
		}
	}
	return out
}
