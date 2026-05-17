package fsutil

import (
	"fmt"
	"io"
	"os"
)

// CopyFile copies a file from src to dest, preserving source permissions.
// Always overwrites dest. Preservation decisions belong in the caller.
func CopyFile(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("creating dest file: %w", err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return fmt.Errorf("copying data: %w", err)
	}

	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("closing dest file: %w", err)
	}

	return nil
}
