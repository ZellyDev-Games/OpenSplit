package adapters

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/session"
)

func DomainToSession(svc *session.Service) *dto.Session {
	var dtoSplitFile *dto.SplitFile
	var dtoCurrentRun *dto.Run

	sf := svc.SplitFile()
	if sf != nil {
		var dtoHierSegments []dto.Segment
		for _, seg := range sf.Segments {
			dtoHierSegments = append(dtoHierSegments, domainSegmentToDTO(seg))
		}

		var dtoRuns []dto.Run
		for _, r := range sf.Runs {
			leafDTO := flattenDomainLeafSegmentsDomain(r.Segments)
			splits := make([]*dto.Split, len(leafDTO))
			for _, sp := range r.Splits {
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

			dtoRuns = append(dtoRuns, dto.Run{
				ID:               r.ID.String(),
				SplitFileID:      sf.ID.String(),
				SplitFileVersion: r.SplitFileVersion,
				TotalTime:        r.TotalTime.Milliseconds(),
				Splits:           splits,
				Completed:        r.Completed,
				Segments:         leafDTO,
			})
		}

		dtoSplitFile = &dto.SplitFile{
			ID:           sf.ID.String(),
			Version:      sf.Version,
			GameName:     sf.GameName,
			GameCategory: sf.GameCategory,
			WindowX:      sf.WindowX,
			WindowY:      sf.WindowY,
			WindowHeight: sf.WindowHeight,
			WindowWidth:  sf.WindowWidth,
			Runs:         dtoRuns,
			Segments:     dtoHierSegments,
			SOB:          sf.SOB.Milliseconds(),
		}

		currentRun, ok := svc.Run()
		if ok {
			// Convert domain currentRun.Segments → leaf DTOs
			leafDTO := make([]dto.Segment, len(currentRun.Segments))
			for i, seg := range currentRun.Segments {
				leafDTO[i] = leafSegmentToDTO(seg)
			}

			splits := make([]*dto.Split, len(currentRun.Segments))
			for _, sp := range currentRun.Splits {
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

			dtoCurrentRun = &dto.Run{
				ID:               currentRun.ID.String(),
				SplitFileID:      sf.ID.String(),
				SplitFileVersion: currentRun.SplitFileVersion,
				TotalTime:        currentRun.TotalTime.Milliseconds(),
				Splits:           splits,
				Completed:        currentRun.Completed,
				Segments:         leafDTO, // flat leaf-only
			}
		}
	}

	return &dto.Session{
		LoadedSplitFile:     dtoSplitFile,
		CurrentRun:          dtoCurrentRun,
		CurrentSegmentIndex: svc.Index(),
		SessionState:        dto.SessionState(svc.State()),
		Dirty:               svc.Dirty(),
	}
}
