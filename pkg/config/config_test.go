package config

import "testing"

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("Got nil as config")
	}
}
