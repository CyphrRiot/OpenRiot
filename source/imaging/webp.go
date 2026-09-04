package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ConvertPngToWebP converts all .png files in backgrounds/, Locked/ and assets/ to .webp
// and removes the original .png files.
func ConvertPngToWebP() error {
	dirs := []string{"backgrounds", "Locked", "assets"}
	converted := 0
	for _, dir := range dirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.png"))
		if err != nil {
			return fmt.Errorf("glob %s: %w", dir, err)
		}
		if len(files) == 0 {
			continue
		}

		webps, _ := filepath.Glob(filepath.Join(dir, "*.webp"))
		for _, w := range webps {
			src := strings.TrimSuffix(w, ".webp") + ".png"
			if _, err := os.Stat(src); os.IsNotExist(err) {
				if err := os.Remove(w); err != nil {
					fmt.Fprintf(os.Stderr, "remove stale %s: %v\n", w, err)
				} else {
					fmt.Printf("removed stale: %s\n", w)
				}
			}
		}

		for _, f := range files {
			out := strings.TrimSuffix(f, ".png") + ".webp"
			cmd := exec.Command("cwebp", f, "-o", out)
			if err := cmd.Run(); err != nil {
				// try convert (ImageMagick) as fallback
				cmd2 := exec.Command("convert", f, out)
				if err2 := cmd2.Run(); err2 != nil {
					fmt.Fprintf(os.Stderr, "skip %s: %v\n", f, err)
					continue
				}
			}
			if err := os.Remove(f); err != nil {
				fmt.Fprintf(os.Stderr, "remove %s: %v\n", f, err)
			}
			converted++
			fmt.Printf("converted: %s -> %s\n", f, out)
		}
	}
	fmt.Printf("done: %d files converted\n", converted)
	return nil
}
