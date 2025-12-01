package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

func (s *SplitFile) BuildStats() {
	if s == nil {
		return
	}

	leafSegs := flattenLeafSegmentsWithPointers(s.Segments)

	// Edge case: no leaf segments
	if len(leafSegs) == 0 {
		s.SOB = 0
		s.PB = nil
		return
	}

	golds, sumMap, countMap := s.perSegmentAggregates(s.Runs)

	// Reset SOB
	var SOB time.Duration

	for _, leaf := range leafSegs {
		id := leaf.ID

		// GOLD
		if gold, ok := golds[id]; ok {
			leaf.Gold = gold
			SOB += gold
		} else {
			leaf.Gold = 0
		}

		// AVERAGE
		if sum, ok := sumMap[id]; ok {
			if cnt := countMap[id]; cnt > 0 {
				leaf.Average = sum / time.Duration(cnt)
			} else {
				leaf.Average = 0
			}
		} else {
			leaf.Average = 0
		}
	}

	PB, _, err := getPB(s.Runs)
	if err != nil {
		s.PB = nil  // no PB available
		s.SOB = SOB // SOB still valid
		return
	}

	s.PB = PB

	if PB != nil && PB.Completed && len(PB.Splits) > 0 {
		for i, leaf := range leafSegs {
			if i < len(PB.Splits) {
				split := PB.Splits[i]
				if split != nil {
					leaf.PB = split.CurrentDuration
				}
			}
		}
	}
	s.SOB = SOB
}

func getPB(runs []Run) (*Run, time.Duration, error) {
	if len(runs) == 0 {
		return nil, 0, errors.New("no runs found")
	}

	var fastestRun *Run = nil
	fastestTotal := time.Duration(0)
	for i, run := range runs {
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

func (s *SplitFile) perSegmentAggregates(runs []Run) (golds map[uuid.UUID]time.Duration, sums map[uuid.UUID]time.Duration, counts map[uuid.UUID]int) {
	golds = make(map[uuid.UUID]time.Duration)
	sums = make(map[uuid.UUID]time.Duration)
	counts = make(map[uuid.UUID]int)

	for _, run := range runs {
		for _, sp := range run.Splits {
			if sp == nil {
				continue
			}
			if cur, ok := golds[sp.SplitSegmentID]; !ok || sp.CurrentDuration < cur {
				golds[sp.SplitSegmentID] = sp.CurrentDuration
			}

			sums[sp.SplitSegmentID] += sp.CurrentDuration
			counts[sp.SplitSegmentID]++
		}
	}

	return golds, sums, counts
}

func flattenLeafSegmentsWithPointers(list []Segment) []*Segment {
	var out []*Segment
	for i := range list {
		seg := &list[i]
		if len(seg.Children) == 0 {
			out = append(out, seg)
			continue
		}
		out = append(out, flattenLeafSegmentsWithPointers(seg.Children)...)
	}
	return out
}
