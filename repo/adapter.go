package repo

import (
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/session"
	"github.com/zellydev-games/opensplit/splitfile"
)

func PayloadFromSplitFile(splitFile session.SplitFile) splitfile.SplitFile {
	var segments []splitfile.Segment
	var runs []splitfile.Run

	for _, segment := range splitFile.Segments {
		segments = append(segments, splitfile.Segment{
			ID:      segment.ID.String(),
			Name:    segment.Name,
			Gold:    segment.Gold.Milliseconds(),
			Average: segment.Average.Milliseconds(),
			PB:      segment.PB.Milliseconds(),
		})
	}

	for _, run := range splitFile.Runs {
		var splits []splitfile.Split
		for _, split := range run.Splits {
			splits = append(splits, splitfile.Split{
				SplitIndex:        split.SplitIndex,
				SplitSegmentID:    split.SplitSegmentID.String(),
				CurrentCumulative: split.CurrentCumulative.Milliseconds(),
				CurrentDuration:   split.CurrentDuration.Milliseconds(),
			})
		}

		runs = append(runs, splitfile.Run{
			ID:               splitFile.ID.String(),
			SplitFileID:      splitFile.ID.String(),
			SplitFileVersion: splitFile.Version,
			SOB:              splitFile.SOB.Milliseconds(),
			TotalTime:        run.TotalTime.Milliseconds(),
			Segments:         segments,
			Splits:           splits,
		})
	}

	return splitfile.SplitFile{
		ID:           splitFile.ID.String(),
		GameName:     splitFile.GameName,
		GameCategory: splitFile.GameCategory,
		Runs:         runs,
		Segments:     segments,
	}
}

func SplitFileFromPayload(payload splitfile.SplitFile) (session.SplitFile, error) {
	id, err := uuid.Parse(payload.ID)
	if err != nil {
		logger.Error("failed to parse ID from payload payload")
		return session.SplitFile{}, err
	}

	var segments []session.Segment
	var splits []session.Split

	for _, segment := range payload.Segments {
		sid, err := uuid.Parse(segment.ID)
		if err != nil {
			logger.Error("failed to parse segment ID from payload payload")
			return session.SplitFile{}, err
		}

		segments = append(segments, session.Segment{
			ID:      sid,
			Name:    segment.Name,
			Gold:    time.Duration(segment.Gold) * time.Millisecond,
			Average: time.Duration(segment.Average) * time.Millisecond,
			PB:      time.Duration(segment.PB) * time.Millisecond,
		})
	}

	for _, run := range payload.Runs {
		for _, split := range run.Splits {
			sid, err := uuid.Parse(split.SplitSegmentID)
			if err != nil {
				logger.Error("failed to parse segment ID from split payload")
				return session.SplitFile{}, err
			}

			splits = append(splits, session.Split{
				SplitIndex:        split.SplitIndex,
				SplitSegmentID:    sid,
				CurrentCumulative: time.Duration(split.CurrentCumulative) * time.Millisecond,
				CurrentDuration:   time.Duration(split.CurrentDuration) * time.Millisecond,
			})
		}
	}

	return session.SplitFile{
		ID:           id,
		Version:      payload.Version,
		GameName:     payload.GameName,
		GameCategory: payload.GameCategory,
		Segments:     segments,
		SOB:          time.Duration(payload.SOB) * time.Millisecond,
	}, nil
}
