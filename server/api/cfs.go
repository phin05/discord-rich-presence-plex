package api

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
)

type customFs struct {
	innerFs fs.FS
}

func (cfs *customFs) Open(path string) (fs.File, error) {
	file, err := cfs.innerFs.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) && path != "index.html" {
			// Serve root index.html in place of non-existent files to support single-page application routing
			return cfs.Open("index.html")
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		indexFilePath := filepath.ToSlash(filepath.Join(path, "index.html"))
		indexFile, err := cfs.innerFs.Open(indexFilePath)
		if err != nil {
			_ = file.Close()
			// Disallow directory indexing by serving root index.html if index.html doesn't exist in the directory
			return cfs.Open("index.html")
		}
		_ = indexFile.Close()
	}
	return file, nil
}
