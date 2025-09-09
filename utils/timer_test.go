package utils

import (
	"testing"
	"time"
)

func TestFormatTimeToString(t *testing.T) {
	d := time.Hour*1 + time.Minute*2 + time.Second*3 + time.Millisecond*400
	timeString := FormatTimeToString(d)
	if timeString != "1:02:03.40" {
		t.Errorf("FormatTimeToString() got %s, want %s", timeString, "1:02:03.40")
	}
}

func TestParseStringToTime(t *testing.T) {
	timeString := "1:02:03.40"
	d, _ := ParseStringToTime(timeString)
	w := time.Hour*1 + time.Minute*2 + time.Second*3 + time.Millisecond*400
	if d != w {
		t.Errorf("ParseStringToTime() got %d, want %d", d, w)
	}
}
