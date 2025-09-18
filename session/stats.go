package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
)

type StatTime struct {
	Raw       int64  `json:"raw"`
	Formatted string `json:"formatted"`
}

type PBStatsPayload struct {
	Run   *RunPayload `json:"run"`
	Total StatTime    `json:"total"`
}

type PBStats struct {
	run   *Run
	total time.Duration
}

func (s *SplitFile) BuildStats() {
	golds, sumMap, countMap := s.perSegmentAggregates(s.runs)
	var sob time.Duration
	for sid, t := range golds {
		for i, seg := range s.segments {
			if seg.id == sid {
				seg.gold = golds[sid]
				s.segments[i] = seg
			}
		}
		sob += t
	}

	for i, seg := range s.segments {
		if sum, ok := sumMap[seg.id]; ok {
			if cnt := countMap[seg.id]; cnt > 0 {
				seg.average = sum / time.Duration(countMap[seg.id])
				s.segments[i] = seg
			}
		}
	}

	pb, err := getPB(s.runs)
	if err != nil {
		logger.Debug(fmt.Sprintf("No pb found: %s", err))
	} else {
		for _, split := range pb.run.splits {
			for i, seg := range s.segments {
				if split.splitSegmentID == seg.id {
					seg.pb = split.currentDuration
					s.segments[i] = seg
				}
			}
		}
	}

	s.sob = sob
}

func getPB(runs []Run) (*PBStats, error) {
	if len(runs) == 0 {
		return nil, errors.New("no runs found")
	}

	var fastestRun *Run = nil
	fastestTotal := time.Duration(0)
	for i, run := range runs {
		if !run.completed || len(run.splits) == 0 {
			continue
		}

		total := run.splits[len(run.splits)-1].currentDuration
		if fastestRun == nil || total < fastestTotal {
			fastestRun = &runs[i]
			fastestTotal = total
		}
	}

	if fastestRun == nil {
		return nil, errors.New("no completed runs found")
	}

	return &PBStats{
		run:   fastestRun,
		total: fastestTotal,
	}, nil
}

func (s *SplitFile) perSegmentAggregates(runs []Run) (golds map[uuid.UUID]time.Duration, sums map[uuid.UUID]time.Duration, counts map[uuid.UUID]int) {
	golds = make(map[uuid.UUID]time.Duration)
	sums = make(map[uuid.UUID]time.Duration)
	counts = make(map[uuid.UUID]int)

	for _, run := range runs {
		var last time.Duration
		for i, sp := range run.splits {
			segmentDuration := sp.currentDuration - last
			if segmentDuration < 0 {
				logger.Warn(fmt.Sprintf("non-monotonic cumulative at split %d", i))
				continue
			}

			last = sp.currentDuration
			if cur, ok := golds[sp.splitSegmentID]; !ok || segmentDuration < cur {
				golds[sp.splitSegmentID] = segmentDuration
			}

			sums[sp.splitSegmentID] += sp.currentDuration
			counts[sp.splitSegmentID]++
		}
	}

	return golds, sums, counts
}
