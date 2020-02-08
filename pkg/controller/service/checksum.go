package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Checksum generates a hex representation of the MD5 checksum of the given string
func Checksum(spec interface{}) (string, error) {
	b, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("failed to encode json body: %v", err)
	}

	sum := md5.New()
	sum.Write(b)
	return hex.EncodeToString(sum.Sum(nil)), nil
}
