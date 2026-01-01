package adapters

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/session"
)

func DomainToDTO(svc *session.Service) *dto.Session {
	var dtoSplitFile *dto.SplitFile
	sf, loaded := svc.SplitFile()
	if loaded {
		dtoSF := DomainSplitFileToDTO(sf)
		dtoSplitFile = &dtoSF
	}
	var dtoRun *dto.Run = nil
	currentRun, loaded := svc.Run()
	if loaded {
		r := domainRunToDTO(currentRun, sf.ID, sf.Version)
		dtoRun = &r
	}

	return &dto.Session{
		LoadedSplitFile:     dtoSplitFile,
		LeafSegments:        domainSegmentsToDTO(sf.DeepCopyLeafSegments()),
		CurrentRun:          dtoRun,
		CurrentSegmentIndex: svc.Index(),
		SessionState:        dto.SessionState(svc.State()),
		Dirty:               svc.Dirty(),
	}
}
