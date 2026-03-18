package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type storedToken struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func tokenPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "hyperseer", "token.json")
}

func saveToken(t storedToken) error {
	path := tokenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadToken() string {
	data, err := os.ReadFile(tokenPath())
	if err != nil {
		return ""
	}
	var t storedToken
	if err := json.Unmarshal(data, &t); err != nil {
		return ""
	}
	if time.Now().After(t.ExpiresAt) {
		return ""
	}
	return t.AccessToken
}
