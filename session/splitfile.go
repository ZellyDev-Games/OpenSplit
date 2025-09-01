package session

type SplitFile struct {
	gameName     string
	gameCategory string
	segments     []Segment
}

func NewSplitFile(gameName string, gameCategory string, segments []Segment) *SplitFile {
	return &SplitFile{
		gameName:     gameName,
		gameCategory: gameCategory,
		segments:     segments,
	}
}
