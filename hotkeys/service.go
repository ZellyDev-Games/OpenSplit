package hotkeys

type KeyInfo struct {
	KeyCode    int
	LocaleName string
}

type Service struct {
	HotkeyChannel chan KeyInfo
}

func NewService(keyInfoChannel chan KeyInfo) *Service {
	return &Service{keyInfoChannel}
}
