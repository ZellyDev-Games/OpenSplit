package bridge

import (
	"github.com/zellydev-games/opensplit/config"
)

type Config struct {
	runtimeProvider      RuntimeProvider
	configUpdatedChannel <-chan *config.Service
}

func NewConfig(configUpdatedChannel <-chan *config.Service, runtimeProvider RuntimeProvider) *Config {
	return &Config{runtimeProvider: runtimeProvider, configUpdatedChannel: configUpdatedChannel}
}

func (c *Config) StartUIPump() {
	go func() {
		for {
			updatedConfig, ok := <-c.configUpdatedChannel
			if !ok {
				return
			}
			c.runtimeProvider.EventsEmit("config:update", updatedConfig)
		}
	}()
}
