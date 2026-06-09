package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

// LocalBackend stores objects as regular files under a root directory.
// Directory structure mirrors the key: "saves/42.Civ6Save.gz" becomes
// <root>/saves/42.Civ6Save.gz.
type LocalBackend struct{ root string }

func NewLocal(root string) *LocalBackend { return &LocalBackend{root: root} }

func (l *LocalBackend) abs(key string) string {
	return filepath.Join(l.root, filepath.FromSlash(key))
}

func (l *LocalBackend) Put(_ context.Context, key string, data []byte, _ string) error {
	p := l.abs(key)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func (l *LocalBackend) Get(_ context.Context, key string) ([]byte, error) {
	return os.ReadFile(l.abs(key))
}

func (l *LocalBackend) Delete(_ context.Context, key string) error {
	err := os.Remove(l.abs(key))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (l *LocalBackend) Exists(_ context.Context, key string) (bool, error) {
	_, err := os.Stat(l.abs(key))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}
