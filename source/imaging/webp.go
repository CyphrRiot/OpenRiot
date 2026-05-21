package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ConvertPngToWebP converts all .png files in backgrounds/ and Locked/ to .webp
func ConvertPngToWebP() error {
	dirs := []string{"backgrounds", "Locked"}
	converted := 0
	for _, dir := range dirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.png"))
		if err != nil {
			return fmt.Errorf("glob %s: %w", dir, err)
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
			converted++
			fmt.Printf("converted: %s -> %s\n", f, out)
		}
	}
	fmt.Printf("done: %d files converted\n", converted)
	return nil
}
