package adapters

import (
	"encoding/json"

	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/hotkeys"
)

func ConfigToFrontEnd(configService *config.Service) ([]byte, error) {
	return json.Marshal(configService)
}

func FrontEndToConfig(configServiceBytes []byte) (*config.Service, error) {
	var configService config.Service
	err := json.Unmarshal(configServiceBytes, &configService)
	return &configService, err
}

func KeyInfoToPayload(info hotkeys.KeyInfo) ([]byte, error) {
	return json.Marshal(info)
}

func PayloadToConfigKeyInfo(payload []byte) (hotkeys.KeyInfo, error) {
	var keyInfo hotkeys.KeyInfo
	err := json.Unmarshal(payload, &keyInfo)
	return keyInfo, err
}
