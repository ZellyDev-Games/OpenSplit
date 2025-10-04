package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
)

func (s *SplitFile) BuildStats() {
	golds, sumMap, countMap := s.perSegmentAggregates(s.Runs)
	var SOB time.Duration
	for sid, t := range golds {
		for i, seg := range s.Segments {
			if seg.ID == sid {
				seg.Gold = golds[sid]
				s.Segments[i] = seg
			}
		}
		SOB += t
	}

	for i, seg := range s.Segments {
		if sum, ok := sumMap[seg.ID]; ok {
			if cnt := countMap[seg.ID]; cnt > 0 {
				seg.Average = sum / time.Duration(countMap[seg.ID])
				s.Segments[i] = seg
			}
		}
	}

	PB, _, err := getPB(s.Runs)
	if err != nil {
		logger.Debug(fmt.Sprintf("No PB found: %s", err))
	}

	for _, pbSplit := range PB.Splits {
		if pbSplit == nil {
			continue
		}
		for i, seg := range s.Segments {
			if pbSplit.SplitSegmentID == seg.ID {
				seg.PB = pbSplit.CurrentDuration
				s.Segments[i] = seg
			}
		}
	}

	s.SOB = SOB
	s.PB = PB
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
