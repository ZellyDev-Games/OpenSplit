package adapters

import (
	"encoding/json"

	"github.com/zellydev-games/opensplit/keyinfo"
)

func PayloadToKeyData(payload []byte) (keyinfo.KeyData, error) {
	var keyInfo keyinfo.KeyData
	err := json.Unmarshal(payload, &keyInfo)
	return keyInfo, err
}

func KeyDataToPayload(keyData keyinfo.KeyData) ([]byte, error) {
	return json.Marshal(keyData)
}
