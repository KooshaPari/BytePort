package lib

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/zalando/go-keyring"
)

// fileKeyringMu serializes read-modify-write of the fallback key file.
var fileKeyringMu sync.Mutex

// fileKeyringPath is where signing keys live when the OS keyring cannot be
// used. BYTEPORT_KEYRING_FILE overrides the location (tests); otherwise the
// file is $XDG_CONFIG_HOME/byteport/keyring.json via os.UserConfigDir.
func fileKeyringPath() (string, error) {
	if p := os.Getenv("BYTEPORT_KEYRING_FILE"); p != "" {
		return p, nil
	}
	config, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(config, "byteport", "keyring.json"), nil
}

func readFileKeyring() (map[string]map[string]string, error) {
	path, err := fileKeyringPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return map[string]map[string]string{}, nil
	}
	store := map[string]map[string]string{}
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return store, nil
}

func writeFileKeyring(store map[string]map[string]string) error {
	path, err := fileKeyringPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func fileKeyringGet(service, user string) (string, error) {
	fileKeyringMu.Lock()
	defer fileKeyringMu.Unlock()
	store, err := readFileKeyring()
	if err != nil {
		return "", err
	}
	if users, ok := store[service]; ok {
		if secret, ok := users[user]; ok {
			return secret, nil
		}
	}
	return "", keyring.ErrNotFound
}

func fileKeyringSet(service, user, secret string) error {
	fileKeyringMu.Lock()
	defer fileKeyringMu.Unlock()
	store, err := readFileKeyring()
	if err != nil {
		return err
	}
	if store[service] == nil {
		store[service] = map[string]string{}
	}
	store[service][user] = secret
	return writeFileKeyring(store)
}
