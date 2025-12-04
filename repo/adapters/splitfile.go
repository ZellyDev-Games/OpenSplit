package adapters

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/session"
)

func DomainToSplitFile(sf *session.SplitFile) *dto.SplitFile {
	var hierarchicalSegments []dto.Segment
	for _, seg := range sf.Segments {
		hierarchicalSegments = append(hierarchicalSegments, domainSegmentToDTO(seg))
	}

	var runs []dto.Run
	for _, run := range sf.Runs {
		// Splits
		splits := make([]*dto.Split, len(run.Segments))
		for _, sp := range run.Splits {
			if sp == nil {
				continue
			}
			splits[sp.SplitIndex] = &dto.Split{
				SplitIndex:        sp.SplitIndex,
				SplitSegmentID:    sp.SplitSegmentID.String(),
				CurrentCumulative: sp.CurrentCumulative.Milliseconds(),
				CurrentDuration:   sp.CurrentDuration.Milliseconds(),
			}
		}

		runs = append(runs, dto.Run{
			ID:               run.ID.String(),
			SplitFileID:      sf.ID.String(),
			SplitFileVersion: sf.Version,
			TotalTime:        run.TotalTime.Milliseconds(),
			Splits:           splits,
			Completed:        run.Completed,
			Segments:         flattenDomainLeafSegmentsDomain(run.Segments),
		})
	}

	// add personal best if exists
	var PB *dto.Run
	if sf.PB != nil {
		// Splits
		splits := make([]*dto.Split, len(sf.PB.Segments))
		for _, sp := range sf.PB.Splits {
			if sp == nil {
				continue
			}
			splits[sp.SplitIndex] = &dto.Split{
				SplitIndex:        sp.SplitIndex,
				SplitSegmentID:    sp.SplitSegmentID.String(),
				CurrentCumulative: sp.CurrentCumulative.Milliseconds(),
				CurrentDuration:   sp.CurrentDuration.Milliseconds(),
			}
		}

		PB = &dto.Run{
			ID:               sf.PB.ID.String(),
			SplitFileID:      sf.ID.String(),
			SplitFileVersion: sf.Version,
			TotalTime:        sf.PB.TotalTime.Milliseconds(),
			Splits:           splits,
			Completed:        sf.PB.Completed,
			Segments:         flattenDomainLeafSegmentsDomain(sf.PB.Segments),
		}
	}

	return &dto.SplitFile{
		ID:           sf.ID.String(),
		GameName:     sf.GameName,
		GameCategory: sf.GameCategory,
		Version:      sf.Version,
		Segments:     hierarchicalSegments,
		Runs:         runs,
		PB:           PB,
		SOB:          sf.SOB.Milliseconds(),
		WindowX:      sf.WindowX,
		WindowY:      sf.WindowY,
		WindowWidth:  sf.WindowWidth,
		WindowHeight: sf.WindowHeight,
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
		segments = append(segments, dtoSegmentToDomain(segment))
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

func domainSegmentToDTO(seg session.Segment) dto.Segment {
	out := dto.Segment{
		ID:       seg.ID.String(),
		Name:     seg.Name,
		Gold:     seg.Gold.Milliseconds(),
		Average:  seg.Average.Milliseconds(),
		PB:       seg.PB.Milliseconds(),
		Children: []dto.Segment{},
	}

	for _, c := range seg.Children {
		out.Children = append(out.Children, domainSegmentToDTO(c))
	}

	return out
}

func leafSegmentToDTO(seg session.Segment) dto.Segment {
	return dto.Segment{
		ID:       seg.ID.String(),
		Name:     seg.Name,
		Gold:     seg.Gold.Milliseconds(),
		Average:  seg.Average.Milliseconds(),
		PB:       seg.PB.Milliseconds(),
		Children: []dto.Segment{},
	}
}

func dtoSegmentToDomain(dtoSeg dto.Segment) session.Segment {
	seg := session.Segment{
		ID:      uuid.MustParse(dtoSeg.ID),
		Name:    dtoSeg.Name,
		Gold:    time.Duration(dtoSeg.Gold) * time.Millisecond,
		Average: time.Duration(dtoSeg.Average) * time.Millisecond,
		PB:      time.Duration(dtoSeg.PB) * time.Millisecond,
	}

	// recursively convert children
	for _, child := range dtoSeg.Children {
		seg.Children = append(seg.Children, dtoSegmentToDomain(child))
	}

	return seg
}

func flattenDomainLeafSegmentsDomain(list []session.Segment) []dto.Segment {
	out := []dto.Segment{}

	for _, seg := range list {
		if len(seg.Children) == 0 {
			out = append(out, leafSegmentToDTO(seg))
		} else {
			out = append(out, flattenDomainLeafSegmentsDomain(seg.Children)...)
		}
	}

	return out
}
