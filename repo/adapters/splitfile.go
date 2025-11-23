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
	var PB *dto.Run

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
		var splits = make([]*dto.Split, len(run.Segments))
		for _, split := range run.Splits {
			if split == nil {
				continue
			}
			splits[split.SplitIndex] = &dto.Split{
				SplitIndex:        split.SplitIndex,
				SplitSegmentID:    split.SplitSegmentID.String(),
				CurrentCumulative: split.CurrentCumulative.Milliseconds(),
				CurrentDuration:   split.CurrentDuration.Milliseconds(),
			}
		}

		runs = append(runs, dto.Run{
			ID:               run.ID.String(),
			SplitFileID:      splitFile.ID.String(),
			SplitFileVersion: splitFile.Version,
			TotalTime:        run.TotalTime.Milliseconds(),
			Splits:           splits,
			Completed:        run.Completed,
			Segments:         segments,
		})
	}

	if splitFile.PB != nil {
		phSegments := make([]dto.Segment, len(splitFile.PB.Segments))
		for _, segment := range splitFile.PB.Segments {
			phSegments = append(phSegments, dto.Segment{
				ID:      segment.ID.String(),
				Name:    segment.Name,
				Gold:    segment.Gold.Milliseconds(),
				Average: segment.Average.Milliseconds(),
				PB:      segment.PB.Milliseconds(),
			})
		}

		var splits = make([]*dto.Split, len(splitFile.PB.Segments))
		for _, s := range splitFile.PB.Splits {
			if s == nil {
				continue
			}
			splits[s.SplitIndex] = &dto.Split{
				SplitIndex:        s.SplitIndex,
				SplitSegmentID:    s.SplitSegmentID.String(),
				CurrentCumulative: s.CurrentCumulative.Milliseconds(),
				CurrentDuration:   s.CurrentDuration.Milliseconds(),
			}
		}

		PB = &dto.Run{
			ID:               splitFile.PB.ID.String(),
			SplitFileID:      splitFile.ID.String(),
			SplitFileVersion: splitFile.Version,
			TotalTime:        splitFile.PB.TotalTime.Milliseconds(),
			Splits:           splits,
			Completed:        splitFile.PB.Completed,
			Segments:         phSegments,
		}
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
		PB:           PB,
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
	var PB *session.Run

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
			return nil, err
		}
		var splits = make([]*session.Split, len(run.Segments))
		for _, split := range run.Splits {
			if split == nil {
				continue
			}
			sid, err := uuid.Parse(split.SplitSegmentID)
			if err != nil {
				logger.Error("failed to parse segment ID from split payload")
				return nil, err
			}

			splits[split.SplitIndex] = &session.Split{
				SplitIndex:        split.SplitIndex,
				SplitSegmentID:    sid,
				CurrentCumulative: time.Duration(split.CurrentCumulative) * time.Millisecond,
				CurrentDuration:   time.Duration(split.CurrentDuration) * time.Millisecond,
			}
		}
		runs = append(runs, session.Run{
			ID:               rid,
			TotalTime:        time.Duration(run.TotalTime) * time.Millisecond,
			Splits:           splits,
			Completed:        run.Completed,
			SplitFileVersion: run.SplitFileVersion,
			Segments:         segments,
		})
	}

	if payload.PB != nil {
		pbid, err := uuid.Parse(payload.PB.ID)
		// Self heal older version file format
		if payload.PB.Segments == nil {
			for _, r := range payload.Runs {
				if r.ID == payload.PB.ID {
					payload.PB.Segments = r.Segments
				}
			}
		}

		if err == nil {
			var splits = make([]*session.Split, len(payload.PB.Segments))
			for _, s := range payload.PB.Splits {
				if s == nil {
					continue
				}
				splitSegmentID, err := uuid.Parse(s.SplitSegmentID)
				if err != nil {
					logger.Error("failed to parse split ID from payload payload")
					continue
				}

				splits[s.SplitIndex] = &session.Split{
					SplitIndex:        s.SplitIndex,
					SplitSegmentID:    splitSegmentID,
					CurrentCumulative: time.Duration(s.CurrentCumulative) * time.Millisecond,
					CurrentDuration:   time.Duration(s.CurrentDuration) * time.Millisecond,
				}
			}

			PB = &session.Run{
				ID:               pbid,
				TotalTime:        time.Duration(payload.PB.TotalTime) * time.Millisecond,
				Splits:           splits,
				Completed:        payload.PB.Completed,
				SplitFileVersion: payload.PB.SplitFileVersion,
				Segments:         segments,
			}

		} else {
			logger.Error("failed to parse PB ID from payload payload")
			return nil, err
		}
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
		PB:           PB,
	}, nil
}

// FrontendToSplitFile takes a string input from the frontend (default: JSON) and returns a new *dto.SplitFile
//
// If no ID was provided for the file, or the segments,
// assume this is a new split file or split file with new segments from the SplitEditor and generate new IDs for them.
func FrontendToSplitFile(payload string) (*dto.SplitFile, error) {
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

func SplitFileToFrontEnd(sf *dto.SplitFile) ([]byte, error) {
	return json.Marshal(sf)
}
