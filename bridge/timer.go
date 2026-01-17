package bridge

import (
	"time"

	"github.com/zellydev-games/opensplit/logger"
)

type Timer struct {
	timerEventStopChannel chan any
	timerUpdateChannel    chan time.Duration
	runtimeProvider       RuntimeProvider
}

func NewTimer(timerUpdateChannel chan time.Duration, runtimeProvider RuntimeProvider) *Timer {
	return &Timer{
		timerUpdateChannel: timerUpdateChannel,
		runtimeProvider:    runtimeProvider,
	}
}

func (t *Timer) StartUIPump() {
	// Start the timer event loop. event can be listened to by anything on the frontend, it's primary use is to
	// provide the UI with the current cumulative time with centisecond precision.  Close the channel when we need
	// to terminate this loop.
	t.timerEventStopChannel = make(chan any)
	go func() {
		for {
			select {
			case <-t.timerEventStopChannel:
				return
			case currentTime := <-t.timerUpdateChannel:
				t.runtimeProvider.EventsEmit("timer:update", currentTime.Milliseconds())
			}
		}
	}()
	logger.Debug(logModule, "started timer UI pump")
}
