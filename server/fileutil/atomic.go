package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

const dirPerm = 0o700

func AtomicWrite(filePath string, perm os.FileMode, write func(file *os.File) error) error {
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, dirPerm); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	tmpFile, err := os.CreateTemp(dirPath, filepath.Base(filePath)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	tmpFilePath := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpFilePath)
		}
	}()
	if err := tmpFile.Chmod(perm); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("chmod: %w", err)
	}
	if err := write(tmpFile); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("sync: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	if err := os.Rename(tmpFilePath, filePath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	cleanup = false
	return nil
}
