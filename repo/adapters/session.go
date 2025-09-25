package adapters

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/session"
)

func DomainToSession(session *session.Service) *dto.Session {
	var dtoSplitFile *dto.SplitFile
	var dtoCurrentRun *dto.Run

	sessionSplitFile := session.SplitFile()
	if sessionSplitFile != nil {
		var segments []dto.Segment
		var runs []dto.Run

		for _, s := range sessionSplitFile.Segments {
			segments = append(segments, dto.Segment{
				ID:      s.ID.String(),
				Name:    s.Name,
				Gold:    s.Gold.Milliseconds(),
				Average: s.Average.Milliseconds(),
				PB:      s.PB.Milliseconds(),
			})
		}

		for _, r := range sessionSplitFile.Runs {
			var splits []dto.Split
			for _, s := range r.Splits {
				splits = append(splits, dto.Split{
					SplitIndex:        s.SplitIndex,
					SplitSegmentID:    s.SplitSegmentID.String(),
					CurrentCumulative: s.CurrentCumulative.Milliseconds(),
					CurrentDuration:   s.CurrentDuration.Milliseconds(),
				})
			}

			runs = append(runs, dto.Run{
				ID:               r.ID.String(),
				SplitFileID:      sessionSplitFile.ID.String(),
				SplitFileVersion: r.SplitFileVersion,
				TotalTime:        r.TotalTime.Milliseconds(),
				Splits:           splits,
				Completed:        r.Completed,
			})
		}

		dtoSplitFile = &dto.SplitFile{
			ID:           sessionSplitFile.ID.String(),
			Version:      sessionSplitFile.Version,
			GameName:     sessionSplitFile.GameName,
			GameCategory: sessionSplitFile.GameCategory,
			WindowX:      sessionSplitFile.WindowX,
			WindowY:      sessionSplitFile.WindowY,
			WindowHeight: sessionSplitFile.WindowHeight,
			WindowWidth:  sessionSplitFile.WindowWidth,
			Runs:         runs,
			Segments:     segments,
			SOB:          sessionSplitFile.SOB.Milliseconds(),
		}

		currentRun, ok := session.Run()
		if ok {
			var splits []dto.Split
			for _, s := range currentRun.Splits {
				splits = append(splits, dto.Split{
					SplitIndex:        s.SplitIndex,
					SplitSegmentID:    s.SplitSegmentID.String(),
					CurrentCumulative: s.CurrentCumulative.Milliseconds(),
					CurrentDuration:   s.CurrentDuration.Milliseconds(),
				})
			}

			dtoCurrentRun = &dto.Run{
				ID:               currentRun.ID.String(),
				SplitFileID:      sessionSplitFile.ID.String(),
				SplitFileVersion: currentRun.SplitFileVersion,
				TotalTime:        currentRun.TotalTime.Milliseconds(),
				Splits:           splits,
				Completed:        currentRun.Completed,
			}
		}
	}
	return &dto.Session{
		LoadedSplitFile:     dtoSplitFile,
		CurrentRun:          dtoCurrentRun,
		CurrentSegmentIndex: session.Index(),
		SessionState:        session.State(),
		Dirty:               session.Dirty(),
	}
}
