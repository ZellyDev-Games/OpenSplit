package bridge

type RuntimeProvider interface {
	EventsEmit(string, ...any)
}
