package session

// Package session contains runtime state and statistical calculations for split files.
//
// statistics.go contains incremental and full rebuild helpers for Gold, Personal Best,
// Rolling Average, and Sum of Best values.

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
)

func (s *SplitFile) AddRun(run *Run, rollingWindow int) {
	if run == nil {
		return
	}

	s.Runs = append(s.Runs, *run)

	if rollingWindow > 0 {
		s.RollingAverageRuns = rollingWindow
	}

	// Legacy split files (version 0) accept every run.
	eligible := s.Version == 0 || run.SplitFileVersion == s.Version
	if !eligible {
		return
	}

	leafSegments := getLeafSegments(s.Segments, nil)

	// ---------------------------------------------------------------------
	// PB
	// ---------------------------------------------------------------------

	if s.PB == nil || run.TotalTime < s.PB.TotalTime {
		copyRun := deepCopyRun(*run)
		s.PB = &copyRun

		for _, seg := range leafSegments {
			if split, ok := run.Splits[seg.ID]; ok {
				seg.PB = split.CurrentDuration
			}
		}

		logger.Debugf(
			logModule,
			"new PB %dms",
			run.TotalTime.Milliseconds(),
		)
	}

	// ---------------------------------------------------------------------
	// Golds + SOB
	// ---------------------------------------------------------------------

	sob := time.Duration(0)

	for _, seg := range leafSegments {
		if split, ok := run.Splits[seg.ID]; ok {
			if seg.Gold == 0 || split.CurrentDuration < seg.Gold {
				seg.Gold = split.CurrentDuration
			}
		}

		sob += seg.Gold
	}

	s.SOB = sob

	// ---------------------------------------------------------------------
	// Rolling Average
	// ---------------------------------------------------------------------

	averages := s.computeRollingAverages()

	for _, seg := range leafSegments {
		if avg, ok := averages[seg.ID]; ok {
			seg.Average = avg
		}
	}

	logger.Debugf(
		logModule,
		"statistics updated incrementally SOB=%d PB=%v",
		s.SOB.Milliseconds(),
		s.PB != nil,
	)
}

// RebuildStatistics performs a complete rebuild from run history.
//
// This should only be used after migration, importing, or bulk editing of run
// history.
func (s *SplitFile) RebuildStatistics() {
	if s == nil {
		return
	}

	logger.Debugf(
		logModule,
		"rebuilding statistics (%d runs)",
		len(s.Runs),
	)

	leafSegments := getLeafSegments(s.Segments, nil)

	if len(leafSegments) == 0 {
		s.SOB = 0
		s.PB = nil
		return
	}

	golds := s.computeGolds()
	averages := s.computeRollingAverages()

	sob := time.Duration(0)

	for _, seg := range leafSegments {
		if gold, ok := golds[seg.ID]; ok {
			seg.Gold = gold
			sob += gold
		}

		if avg, ok := averages[seg.ID]; ok {
			seg.Average = avg
		}
	}

	pb, err := s.computePB()
	if err == nil {
		s.PB = pb

		for _, seg := range leafSegments {
			if split, ok := pb.Splits[seg.ID]; ok {
				seg.PB = split.CurrentDuration
			}
		}
	} else {
		s.PB = nil
	}

	s.SOB = sob

	logger.Debugf(
		logModule,
		"statistics rebuilt SOB=%d PB=%v",
		s.SOB.Milliseconds(),
		s.PB != nil,
	)
}

func (s *SplitFile) computePB() (*Run, error) {
	var fastest *Run

	for i := range s.Runs {
		run := &s.Runs[i]

		if !run.Completed {
			continue
		}

		// Legacy split files accept all runs.
		if s.Version != 0 &&
			run.SplitFileVersion != s.Version {
			continue
		}

		if fastest == nil || run.TotalTime < fastest.TotalTime {
			fastest = run
		}
	}

	if fastest == nil {
		return nil, errors.New("no completed runs found")
	}

	copy := deepCopyRun(*fastest)
	return &copy, nil
}

func (s *SplitFile) computeGolds() map[uuid.UUID]time.Duration {
	golds := make(map[uuid.UUID]time.Duration)

	for _, run := range s.Runs {
		// Legacy split files accept every run.
		if s.Version != 0 &&
			run.SplitFileVersion != s.Version {
			continue
		}

		for id, split := range run.Splits {
			cur, ok := golds[id]
			if !ok || split.CurrentDuration < cur {
				golds[id] = split.CurrentDuration
			}
		}
	}

	return golds
}

func (s *SplitFile) computeRollingAverages() map[uuid.UUID]time.Duration {
	window := s.RollingAverageRuns
	if window <= 0 {
		window = 10
	}

	// Keep the most recent observations for each segment.
	observations := make(map[uuid.UUID][]time.Duration)

	for _, run := range s.Runs {
		// Legacy split files (version 0) accept every run.
		if s.Version != 0 &&
			run.SplitFileVersion != s.Version {
			continue
		}

		for id, split := range run.Splits {
			observations[id] = append(observations[id], split.CurrentDuration)
		}
	}

	averages := make(map[uuid.UUID]time.Duration)

	for id, times := range observations {
		if len(times) == 0 {
			continue
		}

		start := 0
		if len(times) > window {
			start = len(times) - window
		}

		var sum time.Duration
		for _, t := range times[start:] {
			sum += t
		}

		averages[id] = sum / time.Duration(len(times[start:]))
	}

	return averages
}
