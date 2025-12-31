package adapters

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/session"
)

func DomainSplitFileToDTO(sf session.SplitFile) dto.SplitFile {
	// add personal best if exists
	var PB *dto.Run = nil
	if sf.PB != nil {
		dtoPB := domainRunToDTO(*sf.PB, sf.ID, sf.Version)
		PB = &dtoPB
	}

	return dto.SplitFile{
		ID:           sf.ID.String(),
		GameName:     sf.GameName,
		GameCategory: sf.GameCategory,
		Version:      sf.Version,
		Segments:     domainSegmentsToDTO(sf.Segments),
		Runs:         domainRunsToDTO(sf.Runs, sf.ID, sf.Version),
		PB:           PB,
		SOB:          sf.SOB.Milliseconds(),
		WindowX:      sf.WindowX,
		WindowY:      sf.WindowY,
		WindowWidth:  sf.WindowWidth,
		WindowHeight: sf.WindowHeight,
		Offset:       sf.Offset.Milliseconds(),
	}
}

func DTOSplitFileToDomain(payload dto.SplitFile) (session.SplitFile, error) {
	newSplitFile := session.SplitFile{}
	var id uuid.UUID
	if payload.ID == "" {
		id = uuid.New()
	} else {
		parsedID, err := uuid.Parse(payload.ID)
		if err != nil {
			logger.Error("DTOSplitFileToDomain failed to parse ID from payload")
			return newSplitFile, err
		}
		id = parsedID
	}
	var PB *session.Run = nil
	if payload.PB != nil {
		domainPB, err := dtoRunToDomain(*payload.PB)
		if err != nil {
			logger.Error("failed to get PB for split file")
			PB = nil
		} else {
			PB = &domainPB
		}
	}

	newSplitFile.ID = id
	newSplitFile.Version = payload.Version
	newSplitFile.GameName = payload.GameName
	newSplitFile.GameCategory = payload.GameCategory
	newSplitFile.Segments = dtoSegmentsToDomain(payload.Segments)
	newSplitFile.SOB = time.Duration(payload.SOB) * time.Millisecond
	newSplitFile.WindowWidth = payload.WindowWidth
	newSplitFile.WindowHeight = payload.WindowHeight
	newSplitFile.WindowX = payload.WindowX
	newSplitFile.WindowY = payload.WindowY
	newSplitFile.Runs = dtoRunsToDomain(payload.Runs)
	newSplitFile.PB = PB
	newSplitFile.Offset = time.Duration(payload.Offset) * time.Millisecond
	return newSplitFile, nil
}

// JSONSplitFileToDTO takes a string input from the frontend (default: JSON) and returns a new *dto.SplitFile
//
// If no ID was provided for the file, or the segments,
// assume this is a new split file or split file with new segments from the SplitEditor and generate new IDs for them.
func JSONSplitFileToDTO(payload string) (dto.SplitFile, error) {
	var sf dto.SplitFile
	err := json.Unmarshal([]byte(payload), &sf)
	if err != nil {
		return sf, err
	}
	if sf.ID == "" {
		sf.ID = uuid.New().String()
	}

	checkSegmentIDs(sf.Segments)
	return sf, nil
}

func SplitFileToFrontEnd(sf dto.SplitFile) ([]byte, error) {
	return json.Marshal(sf)
}

func checkSegmentIDs(segments []dto.Segment) {
	for i, seg := range segments {
		if seg.ID == "" {
			seg.ID = uuid.New().String()
			segments[i] = seg
		}

		if len(seg.Children) > 0 {
			checkSegmentIDs(seg.Children)
		}
	}
}

func domainSegmentsToDTO(segs []session.Segment) []dto.Segment {
	out := make([]dto.Segment, len(segs))
	for _, s := range segs {
		out = append(out, domainSegmentToDTO(s, true))
	}
	return out
}

func domainSegmentToDTO(s session.Segment, includeChildren bool) dto.Segment {
	dtoSeg := dto.Segment{
		ID:       s.ID.String(),
		Name:     s.Name,
		Gold:     s.Gold.Milliseconds(),
		Average:  s.Average.Milliseconds(),
		PB:       s.PB.Milliseconds(),
		Children: nil,
	}

	if includeChildren {
		for _, c := range s.Children {
			dtoSeg.Children = append(dtoSeg.Children, domainSegmentToDTO(c, true))
		}
	}

	return dtoSeg
}

func dtoSegmentsToDomain(segs []dto.Segment) []session.Segment {
	out := make([]session.Segment, len(segs))
	for _, s := range segs {
		out = append(out, dtoSegmentToDomain(s))
	}
	return out
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

func domainRunsToDTO(runs []session.Run, splitFileID uuid.UUID, splitFileVersion int) []dto.Run {
	out := make([]dto.Run, len(runs))
	for i, r := range runs {
		out[i] = domainRunToDTO(r, splitFileID, splitFileVersion)
	}
	return out
}

func domainRunToDTO(run session.Run, splitFileID uuid.UUID, splitFileVersion int) dto.Run {
	return dto.Run{
		ID:               run.ID.String(),
		SplitFileID:      splitFileID.String(),
		SplitFileVersion: splitFileVersion,
		TotalTime:        run.TotalTime.Milliseconds(),
		Splits:           domainSplitsToDTO(run.Splits),
		LeafSegments:     nil,
		Completed:        false,
	}
}

func dtoRunsToDomain(runs []dto.Run) []session.Run {
	out := make([]session.Run, len(runs))
	for _, r := range runs {
		r, err := dtoRunToDomain(r)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get run from DTO splitfile: %s\n", err.Error()))
			continue
		}
		out = append(out, r)
	}
	return out
}

func dtoRunToDomain(run dto.Run) (session.Run, error) {
	uid, err := uuid.Parse(run.ID)
	if err != nil {
		return session.Run{}, err
	}

	return session.Run{
		ID:               uid,
		TotalTime:        time.Duration(run.TotalTime) * time.Millisecond,
		Splits:           dtoSplitsToDomain(run.Splits),
		LeafSegments:     dtoSegmentsToDomain(run.LeafSegments),
		Completed:        run.Completed,
		SplitFileVersion: run.SplitFileVersion,
	}, nil
}

func domainSplitsToDTO(splits map[uuid.UUID]session.Split) map[string]dto.Split {
	out := map[string]dto.Split{}
	for segmentID, split := range splits {
		out[segmentID.String()] = dto.Split{
			SplitSegmentID:    split.SplitSegmentID.String(),
			CurrentCumulative: split.CurrentCumulative.Milliseconds(),
			CurrentDuration:   split.CurrentDuration.Milliseconds(),
		}
	}
	return out
}

func dtoSplitsToDomain(splits map[string]dto.Split) map[uuid.UUID]session.Split {
	out := map[uuid.UUID]session.Split{}
	for segmentID, split := range splits {
		uid, err := uuid.Parse(segmentID)
		if err != nil {
			logger.Error("failed to parse split ID from splits payload")
			continue
		}
		out[uid] = session.Split{
			SplitSegmentID:    uid,
			CurrentCumulative: time.Duration(split.CurrentCumulative) * time.Millisecond,
			CurrentDuration:   time.Duration(split.CurrentDuration) * time.Millisecond,
		}
	}
	return out
}
