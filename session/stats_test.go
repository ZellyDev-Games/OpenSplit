package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

var rID = uuid.MustParse("9a268f11-1c89-49af-ae00-a9e2246ec82d")
var rID2 = uuid.MustParse("f5ae7f7e-d79d-4418-8933-ba03e171e157")
var rID3 = uuid.MustParse("737b34fd-8e1c-40db-9b17-6a2b0877ba45")

func TestGetStatsPayload(t *testing.T) {
	sf := getSplitFile()

	sf.runs = []Run{{
		id:               rID,
		splitFileVersion: 1,
		totalTime:        time.Second * 35,
		completed:        true,
		splitPayloads: []SplitPayload{{
			SplitIndex:      0,
			SplitSegmentID:  sf.segments[0].id.String(),
			CurrentTime:     "00:00:25.00",
			CurrentDuration: time.Second * 25,
		}, {
			SplitIndex:      1,
			SplitSegmentID:  sf.segments[1].id.String(),
			CurrentTime:     "00:00:35.00",
			CurrentDuration: time.Second * 35,
		}},
	}, {
		id:               rID2,
		splitFileVersion: 1,
		totalTime:        time.Second * 90,
		completed:        true,
		splitPayloads: []SplitPayload{{
			SplitIndex:      0,
			SplitSegmentID:  sf.segments[0].id.String(),
			CurrentTime:     "00:01:00.00",
			CurrentDuration: time.Second * 60,
		}, {
			SplitIndex:      1,
			SplitSegmentID:  sf.segments[1].id.String(),
			CurrentTime:     "00:01:30.00",
			CurrentDuration: time.Second * 90,
		}},
	}, {
		id:               rID3,
		splitFileVersion: 1,
		totalTime:        time.Second * 34,
		completed:        true,
		splitPayloads: []SplitPayload{{
			SplitIndex:      0,
			SplitSegmentID:  sf.segments[0].id.String(),
			CurrentTime:     "00:00:30.00",
			CurrentDuration: time.Second * 30,
		}, {
			SplitIndex:      1,
			SplitSegmentID:  sf.segments[1].id.String(),
			CurrentTime:     "00:00:34.00",
			CurrentDuration: time.Second * 34,
		}},
	}}

	stats := sf.Stats()
	want := (sf.runs[0].splitPayloads[0].CurrentDuration +
		sf.runs[1].splitPayloads[0].CurrentDuration +
		sf.runs[2].splitPayloads[0].CurrentDuration) / 3
	if stats.averages[sf.segments[0].id] != want {
		t.Errorf("segment 1 average time: want %s got %s", want, stats.averages[sf.segments[0].id])
	}

	want = (sf.runs[0].splitPayloads[1].CurrentDuration +
		sf.runs[1].splitPayloads[1].CurrentDuration +
		sf.runs[2].splitPayloads[1].CurrentDuration) / 3
	if stats.averages[sf.segments[1].id] != want {
		t.Errorf("segment 2 average time: want %s got %s", want, stats.averages[sf.segments[1].id])
	}

	want = time.Second * 25
	if stats.golds[sf.segments[0].id] != want {
		t.Errorf("segment 1 gold want: %s got %s", want, stats.golds[sf.segments[0].id])
	}

	want = time.Second * 4
	if stats.golds[sf.segments[1].id] != want {
		t.Errorf("segment 2 gold want: %s got %s", want, stats.golds[sf.segments[1].id])
	}

	if stats.pb.run.ID != rID3 {
		t.Errorf("fastest run (PB) want id %s got %s", rID, stats.pb.run.ID)
	}

	want = time.Second * 29
	if stats.sob != want {
		t.Errorf("sob want %s got %s", want, stats.sob)
	}
}
