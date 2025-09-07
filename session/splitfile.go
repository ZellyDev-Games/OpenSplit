package session

type SplitFilePayload struct {
	GameName     string           `json:"game_name"`
	GameCategory string           `json:"game_category"`
	Segments     []SegmentPayload `json:"segments"`
	Attempts     int              `json:"attempts"`
}

type SplitFile struct {
	gameName     string
	gameCategory string
	segments     []Segment
	attempts     int
}

func NewSplitFile(gameName string, gameCategory string, segments []Segment, attempts int) *SplitFile {
	return &SplitFile{
		gameName:     gameName,
		gameCategory: gameCategory,
		segments:     segments,
		attempts:     attempts,
	}
}

func (s *SplitFile) NewAttempt() {
	s.attempts++
}

func (s *SplitFile) SetAttempts(attempts int) {
	s.attempts = attempts
}

func (s *SplitFile) GetPayload() SplitFilePayload {
	var segmentPayloads []SegmentPayload
	for _, segment := range s.segments {
		segmentPayloads = append(segmentPayloads, segment.GetPayload())
	}
	return SplitFilePayload{
		GameName:     s.gameName,
		GameCategory: s.gameCategory,
		Segments:     segmentPayloads,
		Attempts:     s.attempts,
	}
}

func newFromPayload(payload SplitFilePayload) (*SplitFile, error) {
	var segments []Segment
	for _, segment := range payload.Segments {
		newSegment, err := NewFromPayload(segment)
		if err != nil {
			return nil, err
		}
		segments = append(segments, newSegment)
	}

	return &SplitFile{
		gameName:     payload.GameName,
		gameCategory: payload.GameCategory,
		attempts:     payload.Attempts,
		segments:     segments,
	}, nil
}
