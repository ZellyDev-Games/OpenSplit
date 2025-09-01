package splits

type SplitFilePayload struct {
	GameName        string           `json:"game_name"`
	GameCategory    string           `json:"game_category"`
	Segments        []Segment        `json:"-"`
	SegmentPayloads []SegmentPayload `json:"segments"`
	Attempts        int              `json:"attempts"`
}

type SplitFile struct {
	gameName     string
	gameCategory string
	segments     []Segment
	attempts     int
}

func NewSplitFile(gameName string, gameCategory string, segments []Segment) *SplitFile {
	return &SplitFile{
		gameName:     gameName,
		gameCategory: gameCategory,
		segments:     segments,
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
		GameName:        s.gameName,
		GameCategory:    s.gameCategory,
		Segments:        s.segments,
		SegmentPayloads: segmentPayloads,
		Attempts:        s.attempts,
	}
}
