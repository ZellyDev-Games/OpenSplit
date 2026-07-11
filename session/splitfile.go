package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
)

type SplitFile struct {
	ID           uuid.UUID
	GameName     string
	GameID       string
	GameCategory string
	CategoryID   string
	Version      int

	SelectedSkin string

	Segments []Segment
	Runs     []Run
	PB       *Run

	SOB      time.Duration
	Attempts int
	Offset   time.Duration
	Platform string

	WR WorldRecord

	WindowX      int
	WindowY      int
	WindowHeight int
	WindowWidth  int
}

type WorldRecord struct {
	Show       bool
	RunID      string
	Players    []string
	RealTime   float64
	InGameTime float64
}

func (s *SplitFile) DeepCopyLeafSegments() []Segment {
	leafPointers := getLeafSegments(s.Segments, nil)
	out := make([]Segment, 0, len(leafPointers))
	for i := range leafPointers {
		out = append(out, *leafPointers[i])
	}
	return deepCopySegments(out)
}

func (s *SplitFile) BuildStats() {
	if s == nil {
		return
	}

	leafSegments := getLeafSegments(s.Segments, nil)
	logger.Infof("stats", "building stats for %d segments", len(leafSegments))

	// edge case: no leaf segments
	if len(leafSegments) == 0 {
		s.SOB = 0
		s.PB = nil
		logger.Warn("stats", "no leaf segments found")
		return
	}

	golds, sumMap, countMap := s.perSegmentAggregates(s.Runs)

	// Reset SOB
	var SOB time.Duration

	for _, leaf := range leafSegments {
		id := leaf.ID

		// GOLD
		if gold, ok := golds[id]; ok {
			leaf.Gold = gold
			SOB += gold
		} else {
			leaf.Gold = -1
		}

		// AVERAGE
		if sum, ok := sumMap[id]; ok {
			if cnt := countMap[id]; cnt > 0 {
				leaf.Average = sum / time.Duration(cnt)
			} else {
				leaf.Average = -1
			}
		} else {
			leaf.Average = -1
		}
	}

	PB, _, err := getPB(s.Runs)
	if err != nil {
		s.PB = nil  // no PB available
		s.SOB = SOB // SOB still valid
		return
	}

	for i, leaf := range leafSegments {
		for _, split := range PB.Splits {
			if split.SplitSegmentID == leaf.ID {
				leaf.PB = split.CurrentDuration
				leafSegments[i] = leaf
				break
			}
		}
	}

	s.PB = PB
	s.SOB = SOB
	logger.Infof("stats", "stats built: PB: %f SOB:%f", s.PB.TotalTime.Seconds(), s.SOB.Seconds())
}

func (s *SplitFile) perSegmentAggregates(runs []Run) (golds map[uuid.UUID]time.Duration, sums map[uuid.UUID]time.Duration, counts map[uuid.UUID]int) {
	golds = make(map[uuid.UUID]time.Duration)
	sums = make(map[uuid.UUID]time.Duration)
	counts = make(map[uuid.UUID]int)

	for _, run := range runs {
		for segmentID, sp := range run.Splits {
			if cur, ok := golds[segmentID]; !ok || sp.CurrentDuration < cur {
				golds[segmentID] = sp.CurrentDuration
			}

			sums[segmentID] += sp.CurrentDuration
			counts[segmentID]++
		}
	}

	return golds, sums, counts
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
		Version:      inFile.Version,

		SelectedSkin: inFile.SelectedSkin,

		Segments: segments,
		Runs:     runs,
		PB:       pbRun,

		SOB:      inFile.SOB,
		Attempts: inFile.Attempts,
		Offset:   inFile.Offset,
		Platform: inFile.Platform,

		WR: inFile.WR,

		WindowX:      inFile.WindowX,
		WindowY:      inFile.WindowY,
		WindowWidth:  inFile.WindowWidth,
		WindowHeight: inFile.WindowHeight,
	}
}

func getPB(runs []Run) (*Run, time.Duration, error) {
	if len(runs) == 0 {
		return nil, 0, errors.New("no runs found")
	}

	var fastestRun *Run = nil
	fastestTotal := time.Duration(0)
	for i, run := range runs {
		if !run.Completed {
			continue
		}
		if fastestRun == nil || run.TotalTime < fastestTotal {
			fastestRun = &runs[i]
			fastestTotal = run.TotalTime
		}
	}

	if fastestRun == nil {
		return nil, time.Duration(0), errors.New("no completed runs found")
	}

	return fastestRun, fastestTotal, nil
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
			Children: deepCopySegments(s.Children),
		}
	}
	return out
}
