package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// collectGoFiles returns all .go files in source, or the file itself if not a directory
func collectGoFiles(source string) (files []string, isDir bool, err error) {
	info, err := os.Stat(source)
	if err != nil {
		return nil, false, fmt.Errorf("stat %s: %w", source, err)
	}

	if !info.IsDir() {
		return []string{source}, false, nil
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return nil, true, fmt.Errorf("read directory %s: %w", source, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			files = append(files, filepath.Join(source, entry.Name()))
		}
	}

	return files, true, nil
}
