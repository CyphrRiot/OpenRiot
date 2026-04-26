package fsutil

import (
	"bytes"
	"fmt"
	"os"
)

// CopyFile copies a file from src to dest, preserving source permissions.
// Skips the write if dest already exists with identical content.
func CopyFile(src, dest string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}
	if existing, err := os.ReadFile(dest); err == nil && bytes.Equal(existing, sourceData) {
		return nil
	}
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	if err := os.WriteFile(dest, sourceData, info.Mode()); err != nil {
		return fmt.Errorf("writing dest file: %w", err)
	}
	return nil
}
