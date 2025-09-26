package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

var rID = uuid.MustParse("9a268f11-1c89-49af-ae00-a9e2246ec82d")
var rID2 = uuid.MustParse("f5ae7f7e-d79d-4418-8933-ba03e171e157")
var rID3 = uuid.MustParse("737b34fd-8e1c-40db-9b17-6a2b0877ba45")

func TestBuildStats(t *testing.T) {
	sf := getSplitFile()

	sf.Runs = []Run{{
		ID:               rID,
		SplitFileVersion: 1,
		TotalTime:        time.Second * 35,
		Completed:        true,
		Splits: []Split{{
			SplitIndex:        0,
			SplitSegmentID:    sf.Segments[0].ID,
			CurrentCumulative: time.Second * 25,
			CurrentDuration:   time.Second * 25,
		}, {
			SplitIndex:        1,
			SplitSegmentID:    sf.Segments[1].ID,
			CurrentCumulative: time.Second * 35,
			CurrentDuration:   time.Second * 10,
		}},
	}, {
		ID:               rID2,
		SplitFileVersion: 1,
		TotalTime:        time.Second * 90,
		Completed:        true,
		Splits: []Split{{
			SplitIndex:        0,
			SplitSegmentID:    sf.Segments[0].ID,
			CurrentCumulative: time.Second * 60,
			CurrentDuration:   time.Second * 60,
		}, {
			SplitIndex:        1,
			SplitSegmentID:    sf.Segments[1].ID,
			CurrentCumulative: time.Second * 90,
			CurrentDuration:   time.Second * 30,
		}},
	}, {
		ID:               rID3,
		SplitFileVersion: 1,
		TotalTime:        time.Second * 34,
		Completed:        true,
		Splits: []Split{{
			SplitIndex:        0,
			SplitSegmentID:    sf.Segments[0].ID,
			CurrentCumulative: time.Second * 30,
			CurrentDuration:   time.Second * 30,
		}, {
			SplitIndex:        1,
			SplitSegmentID:    sf.Segments[1].ID,
			CurrentCumulative: time.Second * 34,
			CurrentDuration:   time.Second * 4,
		}},
	}}

	sf.BuildStats()
	want := (sf.Runs[0].Splits[0].CurrentDuration +
		sf.Runs[1].Splits[0].CurrentDuration +
		sf.Runs[2].Splits[0].CurrentDuration) / 3
	if sf.Segments[0].Average != want {
		t.Errorf("segment 1 Average time: want %s got %s", want, sf.Segments[0].Average)
	}

	want = (sf.Runs[0].Splits[1].CurrentDuration +
		sf.Runs[1].Splits[1].CurrentDuration +
		sf.Runs[2].Splits[1].CurrentDuration) / 3
	if sf.Segments[1].Average != want {
		t.Errorf("segment 2 Average time: want %s got %s", want, sf.Segments[1].Average)
	}

	want = time.Second * 25
	if sf.Segments[0].Gold != want {
		t.Errorf("segment 1 Gold want: %s got %s", want, sf.Segments[0].Gold)
	}

	want = time.Second * 4
	if sf.Segments[1].Gold != want {
		t.Errorf("segment 2 Gold want: %s got %s", want, sf.Segments[1].Gold)
	}

	want = time.Second * 30
	if sf.Segments[0].PB != want {
		t.Errorf("fastest run (PB) split 1 time want %s got %s", want, sf.Segments[0].PB)
	}

	want = time.Second * 4
	if sf.Segments[1].PB != want {
		t.Errorf("fastest run (PB) split 2 time want %s got %s", want, sf.Segments[1].PB)
	}

	want = time.Second * 29
	if sf.SOB != want {
		t.Errorf("SOB want %s got %s", want, sf.SOB)
	}
}
