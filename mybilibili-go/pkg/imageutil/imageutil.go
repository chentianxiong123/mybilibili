package imageutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Compression quality (1-100, lower = smaller file)
// 30 is "狠狠压缩" -> ~80-90% size reduction from JPEG
const webpQuality = 30

// Max dimension for longest edge (pixels)
const maxDimension = 1920

// CompressToWebP converts an image to WebP using ffmpeg.
// It replaces the original file with the .webp version and returns the new path.
// Supported inputs: jpg, jpeg, png, gif, webp (skipped if already webp)
func CompressToWebP(srcPath string) (string, error) {
	if srcPath == "" {
		return "", fmt.Errorf("imageutil: empty path")
	}

	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext == "" {
		return "", fmt.Errorf("imageutil: no extension: %s", srcPath)
	}

	// Skip if already webp
	if ext == ".webp" {
		return srcPath, nil
	}

	// Only process image formats
	supported := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".tiff": true, ".tif": true}
	if !supported[ext] {
		return srcPath, nil
	}

	webpPath := strings.TrimSuffix(srcPath, ext) + ".webp"

	scaleFilter := "scale='min(1920,iw)':-2"

	cmd := exec.Command("ffmpeg",
		"-y", "-i", srcPath,
		"-vf", scaleFilter,
		"-c:v", "libwebp",
		"-quality", fmt.Sprintf("%d", webpQuality),
		"-compression_level", "6",
		"-preset", "picture",
		"-loop", "0",
		"-an",
		"-threads", "2",
		webpPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("imageutil: ffmpeg failed: %w\n%s", err, string(output))
	}

	// Verify output exists
	if _, err := os.Stat(webpPath); err != nil {
		return "", fmt.Errorf("imageutil: output not created: %w", err)
	}

	// Remove original file
	os.Remove(srcPath)

	return webpPath, nil
}

// CompressAndReplace reads a file, compresses it to WebP in the same directory,
// and returns the new filename (.webp). Returns empty string if input is not an image.
func CompressAndReplace(srcPath string) (string, error) {
	webpPath, err := CompressToWebP(srcPath)
	if err != nil {
		return "", err
	}
	if webpPath == srcPath {
		return srcPath, nil
	}
	return filepath.Base(webpPath), nil
}