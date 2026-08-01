package session

import (
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
)

// MergeStatistics copies statistics from an older split file into a newer one
// when segment IDs match.
func MergeStatistics(oldFile, newFile *SplitFile) {
	if oldFile == nil || newFile == nil {
		logger.Warn(logModule, "MergeStatistics called with nil split file")
		return
	}

	oldSegments := make(map[uuid.UUID]*Segment)
	indexSegments(oldFile.Segments, oldSegments)

	mergeSegments(newFile.Segments, oldSegments)
}

func indexSegments(segments []Segment, out map[uuid.UUID]*Segment) {
	for i := range segments {
		out[segments[i].ID] = &segments[i]

		if len(segments[i].Children) > 0 {
			indexSegments(segments[i].Children, out)
		}
	}
}

func mergeSegments(newSegments []Segment, old map[uuid.UUID]*Segment) {
	for i := range newSegments {
		newSeg := &newSegments[i]

		if oldSeg, ok := old[newSeg.ID]; ok {
			mergeStatistic(&newSeg.Gold, oldSeg.Gold)
			mergeStatistic(&newSeg.Average, oldSeg.Average)
			mergeStatistic(&newSeg.PB, oldSeg.PB)
		}

		if len(newSeg.Children) > 0 {
			mergeSegments(newSeg.Children, old)
		}
	}
}

// UpgradeSplitFile migrates run history and statistics from an older split file
// to a newer version.
func UpgradeSplitFile(oldFile, newFile *SplitFile) {
	MergeStatistics(oldFile, newFile)

	newFile.Runs = deepCopyRuns(oldFile.Runs)
	newFile.Attempts = oldFile.Attempts

	// PB and SOB are derived.
	newFile.PB = nil
	newFile.SOB = 0

	newFile.RebuildStatistics()
}

// mergeStatistic implements the migration rules.
//
// destination == 0
//
//	inherit previous value
//
// destination == -1
//
//	intentionally do not inherit
//
// destination > 0
//
//	preserve manual value
func mergeStatistic(dst *time.Duration, src time.Duration) {
	switch *dst {
	// Explicitly do not inherit.
	case -1:
		*dst = 0

	// Inherit previous value.
	case 0:
		if src > 0 {
			*dst = src
		}

	// Manual value.
	default:
		// Preserve.
	}
}
