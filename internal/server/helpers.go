package server

import (
	"encoding/base64"
)

func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
