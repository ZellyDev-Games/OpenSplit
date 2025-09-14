package utils

import (
	"fmt"
	"time"
)

// FormatTimeToString takes a time.Duration and returns a string designed to be worked with by the frontend.
//
// Inverse of ParseStringToTime
func FormatTimeToString(d time.Duration) string {
	sign := ""
	if d < 0 {
		sign = "-"
		d = -d
	}
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	cs := (d - s*time.Second) / (10 * time.Millisecond) // centiseconds

	return fmt.Sprintf("%s%d:%02d:%02d.%02d", sign, h, m, s, cs)
}

// ParseStringToTime unserializes a string, usually from the frontend, into a time.Duration.
//
// Inverse of FormatTimeToString.
func ParseStringToTime(s string) (time.Duration, error) {
	var neg bool
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}

	var h, m, sec, cs int
	_, err := fmt.Sscanf(s, "%02d:%02d:%02d.%02d", &h, &m, &sec, &cs)
	if err != nil {
		return 0, err
	}

	d := (time.Duration(h) * time.Hour) +
		(time.Duration(m) * time.Minute) +
		(time.Duration(sec) * time.Second) +
		(time.Duration(cs) * 10 * time.Millisecond)

	if neg {
		d = -d
	}
	return d, nil
}
