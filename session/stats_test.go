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
		splits: []Split{{
			splitIndex:      0,
			splitSegmentID:  sf.segments[0].id,
			currentDuration: time.Second * 25,
		}, {
			splitIndex:      1,
			splitSegmentID:  sf.segments[1].id,
			currentDuration: time.Second * 35,
		}},
	}, {
		id:               rID2,
		splitFileVersion: 1,
		totalTime:        time.Second * 90,
		completed:        true,
		splits: []Split{{
			splitIndex:      0,
			splitSegmentID:  sf.segments[0].id,
			currentDuration: time.Second * 60,
		}, {
			splitIndex:      1,
			splitSegmentID:  sf.segments[1].id,
			currentDuration: time.Second * 90,
		}},
	}, {
		id:               rID3,
		splitFileVersion: 1,
		totalTime:        time.Second * 34,
		completed:        true,
		splits: []Split{{
			splitIndex:      0,
			splitSegmentID:  sf.segments[0].id,
			currentDuration: time.Second * 30,
		}, {
			splitIndex:      1,
			splitSegmentID:  sf.segments[1].id,
			currentDuration: time.Second * 34,
		}},
	}}

	sf.BuildStats()
	want := (sf.runs[0].splits[0].currentDuration +
		sf.runs[1].splits[0].currentDuration +
		sf.runs[2].splits[0].currentDuration) / 3
	if sf.segments[0].average != want {
		t.Errorf("segment 1 average time: want %s got %s", want, sf.segments[0].average)
	}

	want = (sf.runs[0].splits[1].currentDuration +
		sf.runs[1].splits[1].currentDuration +
		sf.runs[2].splits[1].currentDuration) / 3
	if sf.segments[1].average != want {
		t.Errorf("segment 2 average time: want %s got %s", want, sf.segments[1].average)
	}

	want = time.Second * 25
	if sf.segments[0].gold != want {
		t.Errorf("segment 1 gold want: %s got %s", want, sf.segments[0].gold)
	}

	want = time.Second * 4
	if sf.segments[1].gold != want {
		t.Errorf("segment 2 gold want: %s got %s", want, sf.segments[1].gold)
	}

	want = time.Second * 30
	if sf.segments[0].pb != want {
		t.Errorf("fastest run (PB) split 1 time want %s got %s", want, sf.segments[0].pb)
	}

	want = time.Second * 34
	if sf.segments[1].pb != want {
		t.Errorf("fastest run (PB) split 2 time want %s got %s", want, sf.segments[1].pb)
	}

	want = time.Second * 29
	if sf.sob != want {
		t.Errorf("sob want %s got %s", want, sf.sob)
	}
}
