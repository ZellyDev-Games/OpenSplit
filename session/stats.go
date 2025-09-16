package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/utils"
)

type PBStatsPayload struct {
	Run   *RunPayload `json:"run"`
	Total string      `json:"total"`
}

type PBStats struct {
	run   *RunPayload
	total time.Duration
}

type SplitFileStatsPayload struct {
	Golds    map[string]string `json:"golds"`
	Averages map[string]string `json:"averages"`
	SoB      string            `json:"sob"`
	PB       *PBStatsPayload   `json:"pb"`
}

type SplitFileStats struct {
	golds    map[uuid.UUID]time.Duration
	averages map[uuid.UUID]time.Duration
	sob      time.Duration
	pb       *PBStats
}

func (s *SplitFileStats) GetPayload() (SplitFileStatsPayload, error) {
	var goldPayloads = make(map[string]string)
	var averagesPayloads = make(map[string]string)

	for id, gold := range s.golds {
		goldPayloads[id.String()] = utils.FormatTimeToString(gold)
	}

	for id, average := range s.averages {
		goldPayloads[id.String()] = utils.FormatTimeToString(average)
	}

	var pbPayload *PBStatsPayload
	if s.pb != nil {
		pbPayload = &PBStatsPayload{
			Run:   s.pb.run,
			Total: utils.FormatTimeToString(s.pb.total),
		}
	}

	return SplitFileStatsPayload{
		Golds:    goldPayloads,
		Averages: averagesPayloads,
		SoB:      utils.FormatTimeToString(s.sob),
		PB:       pbPayload,
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
		if !run.completed || len(run.splitPayloads) == 0 {
			continue
		}

		total := run.splitPayloads[len(run.splitPayloads)-1].CurrentDuration
		if fastestRun == nil || total < fastestTotal {
			fastestRun = &runs[i]
			fastestTotal = total
		}
	}

	if fastestRun == nil {
		return nil, errors.New("no completed runs found")
	}

	segsInBestRun := make([]SplitPayload, len(fastestRun.splitPayloads))
	copy(segsInBestRun, fastestRun.splitPayloads)

	payload := fastestRun.getPayload()
	return &PBStats{
		run:   &payload,
		total: fastestTotal,
	}, nil
}

func (s *SplitFile) perSegmentAggregates(runs []Run) (golds map[uuid.UUID]time.Duration, sums map[uuid.UUID]time.Duration, counts map[uuid.UUID]int) {
	golds = make(map[uuid.UUID]time.Duration)
	sums = make(map[uuid.UUID]time.Duration)
	counts = make(map[uuid.UUID]int)

	for _, run := range runs {
		var last time.Duration
		for i, sp := range run.splitPayloads {
			id, err := uuid.Parse(sp.SplitSegmentID)
			if err != nil {
				logger.Error(fmt.Sprintf("failed to parse uuid for split payload in perSegmentAggregates: %s", err))
				continue
			}

			segmentDuration := sp.CurrentDuration - last
			if segmentDuration < 0 {
				logger.Warn(fmt.Sprintf("non-monotonic cumulative at split %d", i))
				continue
			}

			last = sp.CurrentDuration
			if cur, ok := golds[id]; !ok || segmentDuration < cur {
				golds[id] = segmentDuration
			}

			sums[id] += sp.CurrentDuration
			counts[id]++
		}
	}

	return golds, sums, counts
}
