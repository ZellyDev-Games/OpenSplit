package keyinfo

// KeyData is the Go-friendly struct to capture key code and key name data from the OS
type KeyData struct {
	KeyCode    int    `json:"key_code"`
	LocaleName string `json:"locale_name"`
}
