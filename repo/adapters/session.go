package adapters

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/session"
)

func DomainToDTO(svc *session.Service) *dto.Session {
	var dtoSplitFile *dto.SplitFile

	sf, splitFileLoaded := svc.SplitFile()
	if splitFileLoaded {
		dtoSF := DomainSplitFileToDTO(sf)
		dtoSplitFile = &dtoSF
	}

	var dtoRun *dto.Run

	currentRun, runLoaded := svc.Run()
	if runLoaded {
		r := domainRunToDTO(currentRun)
		dtoRun = &r
	}

	var leafSegments []dto.Segment

	if splitFileLoaded {
		leafSegments = domainSegmentsToDTO(sf.DeepCopyLeafSegments())
	}

	return &dto.Session{
		LoadedSplitFile:     dtoSplitFile,
		LeafSegments:        leafSegments,
		CurrentRun:          dtoRun,
		CurrentSegmentIndex: svc.Index(),
		SessionState:        dto.SessionState(svc.State()),
		Dirty:               svc.Dirty(),
	}
}

func CleanSplitFile(dtoSplitFile dto.SplitFile) (dto.SplitFile, error) {
	splitFile, err := DTOSplitFileToDomain(dtoSplitFile)
	if err != nil {
		return dto.SplitFile{}, err
	}

	sf := session.DeepCopySplitFile(&splitFile)

	sf.Attempts = 0
	sf.SOB = 0
	sf.Runs = []session.Run{}
	sf.PB = nil

	for i := 0; i < len(sf.Segments); i++ {
		clearSegmentRecursive(&sf.Segments[i])
	}

	return DomainSplitFileToDTO(sf), nil
}

func clearSegmentRecursive(segment *session.Segment) {
	segment.PB = -1
	segment.Gold = -1
	segment.Average = -1

	for i := 0; i < len(segment.Children); i++ {
		clearSegmentRecursive(&segment.Children[i])
	}
}
