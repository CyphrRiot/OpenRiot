package lock

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"openriot/logger"
	"openriot/screen"
)

// BuildCache converts all lock-screen JPGs to screen-resolution PNGs
// in ~/.local/share/openriot/Locked/.cache/{w}x{h}/, mirroring the
// source directory layout (default/, stealth/).
func BuildCache() error {
	home, _ := os.UserHomeDir()
	lockDir := filepath.Join(home, ".local/share/openriot/Locked")

	w, h := screen.GetResolution()
	res := fmt.Sprintf("%dx%d", w, h)

	// e.g. ~/.local/share/openriot/Locked/.cache/1920x1080/
	cacheRoot := filepath.Join(lockDir, ".cache", res)

	logger.Info("Building Lock Screens...")

	// Purge old cache for this resolution
	if err := os.RemoveAll(cacheRoot); err != nil {
		return fmt.Errorf("failed to clear old cache: %w", err)
	}

	dirs := []string{"default", "stealth"}

	// Detect CPU count for parallelisation
	ncpu := 4
	if out, err := exec.Command("sysctl", "-n", "hw.ncpu").Output(); err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil && n > 0 {
			ncpu = n
		}
	}

	// Prefer ffmpeg (much faster), fall back to ImageMagick convert
	useFFmpeg := false
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		useFFmpeg = true
	}

	var total, converted int
	for _, dir := range dirs {
		matches, _ := filepath.Glob(filepath.Join(lockDir, dir, "*.jpg"))
		total += len(matches)
	}

	for _, dir := range dirs {
		srcDir := filepath.Join(lockDir, dir)
		dstDir := filepath.Join(cacheRoot, dir)

		matches, _ := filepath.Glob(filepath.Join(srcDir, "*.jpg"))
		if len(matches) == 0 {
			continue
		}

		if err := os.MkdirAll(dstDir, 0755); err != nil {
			logger.Warn(fmt.Sprintf("Failed to create %s: %v", dstDir, err))
			continue
		}

		if useFFmpeg {
			// Parallel ffmpeg: ~2× faster per image, plus parallelisation
			filter := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2", w, h, w, h)
			script := fmt.Sprintf(
				`for f in %s/*.jpg; do
					bn=$(basename "$f" .jpg)
					echo "$f" "%s/${bn}.png"
				done | xargs -P %d -n2 sh -c 'ffmpeg -y -i "$1" -vf "%s" "$2" 2>/dev/null' _`,
				srcDir, dstDir, ncpu, filter)
			cmd := exec.Command("/bin/sh", "-c", script)
			if out, err := cmd.CombinedOutput(); err != nil {
				logger.Warn(fmt.Sprintf("ffmpeg batch failed: %v\n%s", err, string(out)))
				useFFmpeg = false // fall through to convert
			} else {
				// Count successes
				pngs, _ := filepath.Glob(filepath.Join(dstDir, "*.png"))
				converted += len(pngs)
				continue
			}
		}

		// Fallback: parallel ImageMagick convert
		script := fmt.Sprintf(
			`find %s -name '*.jpg' | xargs -P %d -I{} sh -c 'f="{}"; bn=$(basename "$f" .jpg); convert "$f" -resize "%s^" -gravity center -extent "%s" "%s/${bn}.png"'`,
			srcDir, ncpu, res, res, dstDir)
		cmd := exec.Command("/bin/sh", "-c", script)
		if out, err := cmd.CombinedOutput(); err != nil {
			logger.Warn(fmt.Sprintf("convert batch failed: %v\n%s", err, string(out)))
			continue
		}
		pngs, _ := filepath.Glob(filepath.Join(dstDir, "*.png"))
		converted += len(pngs)
	}

	logger.Done("[" + strconv.Itoa(converted) + "/" + strconv.Itoa(total) + "] Lock screen images completed")
	return nil
}
