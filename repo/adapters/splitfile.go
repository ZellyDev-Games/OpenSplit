package adapters

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/session"
)

const logModule = "adapters"

func DomainSplitFileToDTO(sf session.SplitFile) dto.SplitFile {
	// add personal best if exists
	var PB *dto.Run = nil
	if sf.PB != nil {
		dtoPB := domainRunToDTO(*sf.PB)
		PB = &dtoPB
	}

	return dto.SplitFile{
		ID:           sf.ID.String(),
		GameName:     sf.GameName,
		GameID:       sf.GameID,
		GameCategory: sf.GameCategory,
		CategoryID:   sf.CategoryID,
		Variables:    domainVariablesToDTO(sf.Variables),

		Version: sf.Version,

		SelectedSkin: sf.SelectedSkin,

		Segments: domainSegmentsToDTO(sf.Segments),
		Runs:     domainRunsToDTO(sf.Runs),
		PB:       PB,

		SOB:      sf.SOB.Milliseconds(),
		Attempts: sf.Attempts,
		Offset:   sf.Offset.Milliseconds(),
		Platform: sf.Platform,

		WR: dto.WorldRecord(sf.WR),

		Layout: sf.Layout,

		Windows: dto.SplitterWindows{
			Vertical: dto.SplitterWindow{
				X:      sf.Windows.Vertical.X,
				Y:      sf.Windows.Vertical.Y,
				Width:  sf.Windows.Vertical.Width,
				Height: sf.Windows.Vertical.Height,
			},
			Horizontal: dto.SplitterWindow{
				X:      sf.Windows.Horizontal.X,
				Y:      sf.Windows.Horizontal.Y,
				Width:  sf.Windows.Horizontal.Width,
				Height: sf.Windows.Horizontal.Height,
			},
		},
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
			logger.Error(
				logModule,
				"DTOSplitFileToDomain failed to parse ID from payload",
			)
			return newSplitFile, err
		}
		id = parsedID
	}

	var PB *session.Run = nil
	if payload.PB != nil {
		domainPB, err := dtoRunToDomain(*payload.PB)
		if err != nil {
			logger.Errorf(
				logModule,
				"failed to get PB for split file: %v",
				err,
			)
			PB = nil
		} else {
			PB = &domainPB
		}
	}

	// Self-healing to remove invalid runs caused by a bug present in
	// early versions of OpenSplit.
	fixedRuns := []dto.Run{}
	for _, run := range payload.Runs {
		if run.ID == uuid.Nil.String() {
			logger.Warn(
				logModule,
				"discarding run with invalid UUID",
			)
			continue
		}

		fixedRuns = append(fixedRuns, run)
	}

	newSplitFile.ID = id
	newSplitFile.GameName = payload.GameName
	newSplitFile.GameID = payload.GameID
	newSplitFile.GameCategory = payload.GameCategory
	newSplitFile.CategoryID = payload.CategoryID
	newSplitFile.Variables = dtoVariablesToDomain(payload.Variables)

	newSplitFile.Version = payload.Version

	newSplitFile.SelectedSkin = payload.SelectedSkin

	newSplitFile.Segments = dtoSegmentsToDomain(payload.Segments)
	newSplitFile.Runs = dtoRunsToDomain(fixedRuns)
	newSplitFile.PB = PB

	newSplitFile.SOB = time.Duration(payload.SOB) * time.Millisecond
	newSplitFile.Attempts = payload.Attempts
	newSplitFile.Offset = time.Duration(payload.Offset) * time.Millisecond
	newSplitFile.Platform = payload.Platform

	newSplitFile.WR = session.WorldRecord(payload.WR)

	newSplitFile.Layout = normalizeLayout(payload.Layout)
	newSplitFile.Windows = migrateWindows(payload)

	logger.Debugf(
		logModule,
		"DOMAIN GameID=%q CategoryID=%q",
		newSplitFile.GameID,
		newSplitFile.CategoryID,
	)

	return newSplitFile, nil
}

func normalizeLayout(layout string) string {
	if layout == "horizontal" {
		return "horizontal"
	}

	return "vertical"
}

func defaultVerticalWindow() session.SplitterWindow {
	return session.SplitterWindow{
		X:      100,
		Y:      100,
		Width:  350,
		Height: 550,
	}
}

func defaultHorizontalWindow() session.SplitterWindow {
	return session.SplitterWindow{
		X:      100,
		Y:      100,
		Width:  900,
		Height: 400,
	}
}

func dtoWindowToDomain(window dto.SplitterWindow) session.SplitterWindow {
	return session.SplitterWindow{
		X:      window.X,
		Y:      window.Y,
		Width:  window.Width,
		Height: window.Height,
	}
}

func migrateWindows(payload dto.SplitFile) session.SplitterWindows {
	windows := session.SplitterWindows{
		Vertical:   defaultVerticalWindow(),
		Horizontal: defaultHorizontalWindow(),
	}

	// New layout-specific window data takes precedence when present.
	if payload.Windows.Vertical.Width > 0 {
		windows.Vertical = dtoWindowToDomain(payload.Windows.Vertical)
	}

	if payload.Windows.Horizontal.Width > 0 {
		windows.Horizontal = dtoWindowToDomain(payload.Windows.Horizontal)
	}

	// Migrate the old single-window rectangle into the active layout.
	if payload.WindowWidth > 0 && payload.WindowHeight > 0 {
		legacy := session.SplitterWindow{
			X:      payload.WindowX,
			Y:      payload.WindowY,
			Width:  payload.WindowWidth,
			Height: payload.WindowHeight,
		}

		switch normalizeLayout(payload.Layout) {
		case "horizontal":
			if payload.Windows.Horizontal.Width <= 0 {
				windows.Horizontal = legacy
			}
		default:
			if payload.Windows.Vertical.Width <= 0 {
				windows.Vertical = legacy
			}
		}
	}

	return windows
}

// JSONSplitFileToDTO takes a string input from the frontend (default: JSON)
// and returns a new *dto.SplitFile.
//
// If no ID was provided for the file, or the segments, assume this is a new
// split file or split file with new segments from the SplitEditor and
// generate new IDs for them.
func JSONSplitFileToDTO(payload string) (dto.SplitFile, error) {
	var sf dto.SplitFile

	err := json.Unmarshal([]byte(payload), &sf)
	logger.Debugf(
		logModule,
		"DTO GameID=%q CategoryID=%q",
		sf.GameID,
		sf.CategoryID,
	)

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
			logger.Debug(
				logModule,
				"generated segment UUID",
			)
		}

		if len(seg.Children) > 0 {
			checkSegmentIDs(seg.Children)
		}
	}
}

func domainSegmentsToDTO(segments []session.Segment) []dto.Segment {
	out := make([]dto.Segment, len(segments))
	for i := range segments {
		out[i] = domainSegmentToDTO(segments[i])
	}
	return out
}

func domainSegmentToDTO(s session.Segment) dto.Segment {
	dtoSeg := dto.Segment{
		ID:       s.ID.String(),
		Name:     s.Name,
		Gold:     s.Gold.Milliseconds(),
		Average:  s.Average.Milliseconds(),
		PB:       s.PB.Milliseconds(),
		Icon:     s.Icon,
		Children: []dto.Segment{},
	}

	for _, c := range s.Children {
		dtoSeg.Children = append(
			dtoSeg.Children,
			domainSegmentToDTO(c),
		)
	}

	return dtoSeg
}

func dtoSegmentsToDomain(segments []dto.Segment) []session.Segment {
	out := make([]session.Segment, len(segments))
	for i, s := range segments {
		out[i] = dtoSegmentToDomain(s)
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
		Icon:    dtoSeg.Icon,
	}

	// recursively convert children
	for _, child := range dtoSeg.Children {
		seg.Children = append(
			seg.Children,
			dtoSegmentToDomain(child),
		)
	}

	return seg
}

func domainRunsToDTO(runs []session.Run) []dto.Run {
	out := make([]dto.Run, len(runs))

	for i, r := range runs {
		out[i] = domainRunToDTO(r)
	}

	return out
}

func domainRunToDTO(run session.Run) dto.Run {
	return dto.Run{
		ID:               run.ID.String(),
		SplitFileVersion: run.SplitFileVersion,
		TotalTime:        run.TotalTime.Milliseconds(),
		Splits:           domainSplitsToDTO(run.Splits),
		LeafSegments:     nil,
		Completed:        run.Completed,
	}
}

func dtoRunsToDomain(runs []dto.Run) []session.Run {
	out := make([]session.Run, len(runs))

	for i, r := range runs {
		r, err := dtoRunToDomain(r)
		if err != nil {
			logger.Errorf(
				logModule,
				"failed to get run from DTO splitfile: %s\n",
				err.Error(),
			)
			continue
		}

		out[i] = r
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

func domainSplitsToDTO(
	splits map[uuid.UUID]session.Split,
) map[string]dto.Split {
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

func dtoSplitsToDomain(
	splits map[string]dto.Split,
) map[uuid.UUID]session.Split {
	out := map[uuid.UUID]session.Split{}

	for segmentID, split := range splits {
		uid, err := uuid.Parse(segmentID)
		if err != nil {
			logger.Error(
				logModule,
				"failed to parse split ID from splits payload",
			)
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

func domainVariablesToDTO(
	vars []session.Variable,
) []dto.Variable {
	out := make([]dto.Variable, len(vars))

	for i, v := range vars {
		out[i] = dto.Variable{
			ID:      v.ID,
			Name:    v.Name,
			ValueID: v.ValueID,
			Label:   v.Label,
		}
	}

	return out
}

func dtoVariablesToDomain(
	vars []dto.Variable,
) []session.Variable {
	out := make([]session.Variable, len(vars))

	for i, v := range vars {
		out[i] = session.Variable{
			ID:      v.ID,
			Name:    v.Name,
			ValueID: v.ValueID,
			Label:   v.Label,
		}
	}

	return out
}
