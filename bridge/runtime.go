package bridge

const logModule = "bridge"

type RuntimeProvider interface {
	EventsEmit(string, ...any)
}
