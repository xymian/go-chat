package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type MediaStore interface {
	Save(key string, r io.Reader) error
}

type LocalMediaStore struct {
	BaseDir string
}

func (localStore *LocalMediaStore) Save(key string, r io.Reader) error {
	fullPath := filepath.Join(localStore.BaseDir, key)

	absBase, err := filepath.Abs(localStore.BaseDir)
	if err != nil {
		return fmt.Errorf("storage: failed to resolve path: %w", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("storage: failed to resolve target path: %w", err)
	}

	if !strings.HasPrefix(absPath, absBase) {
		return fmt.Errorf("storage: illegal path traversal attempt: %s", key)
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("storage: failed to create directory: %w", dir, err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("storage: failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, r); err != nil {
		return fmt.Errorf("storage: failed to caopy data: %w", err)
	}

	if err := dst.Sync(); err != nil {
		return fmt.Errorf("storage: failed to sync data to disk: %w", err)
	}

	if err := dst.Close(); err != nil {
		return fmt.Errorf("storage: failed to close file: %w", err)
	}
	return nil
}