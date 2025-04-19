package utils

import (
	"encoding/base64"
)

// EncodeBytes encodes a byte array to a base64 string
func EncodeBytes(data []byte) string {
	if data == nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBytes decodes a base64 string to a byte array
func DecodeBytes(data string) ([]byte, error) {
	if data == "" {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(data)
}
