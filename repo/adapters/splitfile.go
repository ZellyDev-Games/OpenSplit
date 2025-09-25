package adapters

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/session"
)

func DomainToSplitFile(splitFile *session.SplitFile) *dto.SplitFile {
	var segments []dto.Segment
	var runs []dto.Run

	for _, segment := range splitFile.Segments {
		segments = append(segments, dto.Segment{
			ID:      segment.ID.String(),
			Name:    segment.Name,
			Gold:    segment.Gold.Milliseconds(),
			Average: segment.Average.Milliseconds(),
			PB:      segment.PB.Milliseconds(),
		})
	}

	for _, run := range splitFile.Runs {
		var splits []dto.Split
		for _, split := range run.Splits {
			splits = append(splits, dto.Split{
				SplitIndex:        split.SplitIndex,
				SplitSegmentID:    split.SplitSegmentID.String(),
				CurrentCumulative: split.CurrentCumulative.Milliseconds(),
				CurrentDuration:   split.CurrentDuration.Milliseconds(),
			})
		}

		runs = append(runs, dto.Run{
			ID:               splitFile.ID.String(),
			SplitFileID:      splitFile.ID.String(),
			SplitFileVersion: splitFile.Version,
			TotalTime:        run.TotalTime.Milliseconds(),
			Splits:           splits,
			Completed:        run.Completed,
		})
	}

	return &dto.SplitFile{
		ID:           splitFile.ID.String(),
		GameName:     splitFile.GameName,
		GameCategory: splitFile.GameCategory,
		Runs:         runs,
		Segments:     segments,
		SOB:          splitFile.SOB.Milliseconds(),
		WindowWidth:  splitFile.WindowWidth,
		WindowHeight: splitFile.WindowHeight,
		WindowX:      splitFile.WindowX,
		WindowY:      splitFile.WindowY,
		Version:      splitFile.Version,
	}
}

func SplitFileToDomain(payload *dto.SplitFile) (*session.SplitFile, error) {
	var id uuid.UUID
	if payload.ID == "" {
		id = uuid.New()
	} else {
		parsedID, err := uuid.Parse(payload.ID)
		if err != nil {
			logger.Error("SplitFileToDomain failed to parse ID from payload")
			return nil, err
		}
		id = parsedID
	}

	var runs []session.Run
	var segments []session.Segment

	for _, segment := range payload.Segments {
		sid, err := uuid.Parse(segment.ID)
		if err != nil {
			logger.Error("failed to parse segment ID from payload payload")
			return nil, err
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
		rid, err := uuid.Parse(run.ID)
		if err != nil {
			logger.Error("failed to parse run ID from payload payload")
			continue
		}
		var splits []session.Split
		for _, split := range run.Splits {
			sid, err := uuid.Parse(split.SplitSegmentID)
			if err != nil {
				logger.Error("failed to parse segment ID from split payload")
				return nil, err
			}

			splits = append(splits, session.Split{
				SplitIndex:        split.SplitIndex,
				SplitSegmentID:    sid,
				CurrentCumulative: time.Duration(split.CurrentCumulative) * time.Millisecond,
				CurrentDuration:   time.Duration(split.CurrentDuration) * time.Millisecond,
			})
		}
		runs = append(runs, session.Run{
			ID:               rid,
			TotalTime:        time.Duration(run.TotalTime) * time.Millisecond,
			Splits:           splits,
			Completed:        run.Completed,
			SplitFileVersion: run.SplitFileVersion,
		})
	}

	return &session.SplitFile{
		ID:           id,
		Version:      payload.Version,
		GameName:     payload.GameName,
		GameCategory: payload.GameCategory,
		Segments:     segments,
		SOB:          time.Duration(payload.SOB) * time.Millisecond,
		WindowWidth:  payload.WindowWidth,
		WindowHeight: payload.WindowHeight,
		WindowX:      payload.WindowX,
		WindowY:      payload.WindowY,
		Runs:         runs,
	}, nil
}

func JsonToSplitFile(payload string) (*dto.SplitFile, error) {
	var sf dto.SplitFile
	err := json.Unmarshal([]byte(payload), &sf)
	if err != nil {
		return nil, err
	}
	if sf.ID == "" {
		sf.ID = uuid.New().String()
	}

	for i, seg := range sf.Segments {
		if seg.ID == "" {
			seg.ID = uuid.New().String()
			sf.Segments[i] = seg
		}
	}

	return &sf, nil
}

func SplitFileToJson(sf *dto.SplitFile) ([]byte, error) {
	return json.Marshal(sf)
}
