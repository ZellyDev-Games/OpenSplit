package adapters

import (
	"encoding/json"

	"github.com/zellydev-games/opensplit/config"
)

func ConfigToFrontEnd(configService *config.Service) ([]byte, error) {
	return json.Marshal(configService)
}

func FrontEndToConfig(configServiceBytes []byte) (*config.Service, error) {
	var configService config.Service
	err := json.Unmarshal(configServiceBytes, &configService)
	return &configService, err
}
