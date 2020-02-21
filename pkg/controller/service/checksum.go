package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// checksum generates a hex representation of the MD5 checksum of the given string
func checksum(spec interface{}) (string, error) {
	b, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("failed to encode json body: %v", err)
	}

	sum := md5.New()
	_, err = sum.Write(b)
	if err != nil {
		return "", fmt.Errorf("failed to write content to checksum: %v", err)
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}
