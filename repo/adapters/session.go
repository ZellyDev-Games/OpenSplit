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

func ExportSessionSplitfile(splitFile *session.SplitFile) {
	sf := session.DeepCopySplitFile(splitFile)
	sf.WindowY = 100
	sf.WindowX = 100
	sf.Attempts = 0
	sf.SOB = 0
	sf.Runs = []session.Run{}
	sf.PB = nil

	for i := 0; i < len(sf.Segments); i++ {
		clearSegmentRecursive(&sf.Segments[i])
	}
}

func clearSegmentRecursive(segment *session.Segment) {
	segment.PB = 0
	segment.Gold = 0
	segment.Average = 0

	for i := 0; i < len(segment.Children); i++ {
		clearSegmentRecursive(&segment.Children[i])
	}
}
