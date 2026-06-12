// Package storage provides an abstract file-storage backend.
// Swap STORAGE_BACKEND=local (default) for STORAGE_BACKEND=s3 when
// Hetzner Object Storage credentials are available.
package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Backend is the single interface every storage implementation must satisfy.
type Backend interface {
	// Put writes data under the given key (forward-slash path, e.g. "saves/42.Civ6Save.gz").
	Put(ctx context.Context, key string, data []byte, contentType string) error
	// Get retrieves the data stored under key.
	Get(ctx context.Context, key string) ([]byte, error)
	// Delete removes the object at key; no-op if it does not exist.
	Delete(ctx context.Context, key string) error
	// Exists reports whether the key is present without reading the data.
	Exists(ctx context.Context, key string) (bool, error)
}

// New creates a Backend from environment-style configuration.
// backend: "local" (default) | "s3"
// root:    base path for the local backend (ignored for S3).
func New(backend, root string) (Backend, error) {
	switch backend {
	case "local", "":
		if root == "" {
			// Prefer the server's state directory when present (production),
			// then fall back to the XDG user data dir so the server works
			// without root permissions in development.
			if fi, err := os.Stat("/var/lib/civ6"); err == nil && fi.IsDir() {
				root = "/var/lib/civ6"
			} else if home, err := os.UserHomeDir(); err == nil {
				root = filepath.Join(home, ".local", "share", "civ6")
			} else {
				root = "/var/lib/civ6"
			}
		}
		if err := os.MkdirAll(root, 0o755); err != nil {
			return nil, fmt.Errorf("storage: create root %q: %w", root, err)
		}
		return NewLocal(root), nil

	case "s3":
		return nil, errors.New("storage: S3 backend is not yet configured; set STORAGE_BACKEND=local")

	default:
		return nil, fmt.Errorf("storage: unknown backend %q", backend)
	}
}
