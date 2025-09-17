package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/utils"
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

type SplitFileStatsPayload struct {
	Golds    map[string]StatTime `json:"golds"`
	Averages map[string]StatTime `json:"averages"`
	SoB      StatTime            `json:"sob"`
	PB       *PBStatsPayload     `json:"pb"`
}

type SplitFileStats struct {
	golds    map[uuid.UUID]time.Duration
	averages map[uuid.UUID]time.Duration
	sob      time.Duration
	pb       *PBStats
}

func (s *SplitFileStats) GetPayload() (SplitFileStatsPayload, error) {
	var goldPayloads = make(map[string]StatTime)
	var averagesPayloads = make(map[string]StatTime)

	for id, gold := range s.golds {
		goldPayloads[id.String()] = StatTime{gold.Milliseconds(), utils.FormatTimeToString(gold)}
	}

	for id, average := range s.averages {
		averagesPayloads[id.String()] = StatTime{average.Milliseconds(), utils.FormatTimeToString(average)}
	}

	var pbPayload *PBStatsPayload
	if s.pb != nil {
		payload := s.pb.run.getPayload()
		pbPayload = &PBStatsPayload{
			Run: &payload,
			Total: StatTime{
				Raw:       s.pb.total.Milliseconds(),
				Formatted: utils.FormatTimeToString(s.pb.total),
			},
		}
	}

	return SplitFileStatsPayload{
		Golds:    goldPayloads,
		Averages: averagesPayloads,
		SoB: StatTime{
			Raw:       s.sob.Milliseconds(),
			Formatted: utils.FormatTimeToString(s.sob),
		},
		PB: pbPayload,
	}, nil
}

func (s *SplitFile) Stats() SplitFileStats {
	golds, sumMap, countMap := s.perSegmentAggregates(s.runs)
	averages := make(map[uuid.UUID]time.Duration)
	var sob time.Duration
	for _, t := range golds {
		sob += t
	}

	for _, seg := range s.segments {
		if sum, ok := sumMap[seg.id]; ok {
			avg := sum / time.Duration(countMap[seg.id])
			averages[seg.id] = avg
		}
	}

	pb, err := getPB(s.runs)
	if err != nil {
		logger.Debug(fmt.Sprintf("No pb found: %s", err))
		pb = nil
	}

	return SplitFileStats{
		golds:    golds,
		averages: averages,
		sob:      sob,
		pb:       pb,
	}
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
