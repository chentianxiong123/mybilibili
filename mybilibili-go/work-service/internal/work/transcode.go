package work

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"mybilibili/pkg/abstraction"
)

// FFmpegRunner abstracts the ffmpeg binary execution so it can be stubbed in tests.
type FFmpegRunner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

type systemFFmpeg struct{}

func (systemFFmpeg) Run(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput()
}

type FFprobeRunner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

type systemFFprobe struct{}

func (systemFFprobe) Run(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "ffprobe", args...).CombinedOutput()
}

type TranscodeWorker struct {
	ffmpeg  FFmpegRunner
	ffprobe FFprobeRunner
	storage abstraction.StorageService
}

func NewTranscodeWorker(storage abstraction.StorageService) *TranscodeWorker {
	return &TranscodeWorker{
		ffmpeg:  systemFFmpeg{},
		ffprobe: systemFFprobe{},
		storage: storage,
	}
}

func (w *TranscodeWorker) GetDuration(ctx context.Context, srcFile string) (int, error) {
	out, err := w.ffprobe.Run(ctx,
		"-v", "error", "-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", srcFile)
	if err != nil {
		return 0, err
	}
	var seconds int
	fmt.Sscanf(string(out), "%d", &seconds)
	return seconds, nil
}

// Transcode converts source video to HLS segments at the given quality.
func (w *TranscodeWorker) Transcode(ctx context.Context, srcFile, outDir string, quality string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	playlist := filepath.Join(outDir, "index.m3u8")
	crf := "23"
	scale := "1280:-2"
	switch quality {
	case "1080p":
		crf = "20"
		scale = "1920:-2"
	case "720p":
		crf = "22"
		scale = "1280:-2"
	default:
		crf = "23"
		scale = "854:-2"
	}
	args := []string{
		"-i", srcFile,
		"-vf", "scale=" + scale,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", crf,
		"-c:a", "aac", "-b:a", "128k",
		"-f", "hls", "-hls_time", "6", "-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(outDir, "segment%03d.ts"),
		playlist,
	}
	_, err := w.ffmpeg.Run(ctx, args...)
	return err
}

// ExtractAudio extracts the audio track into WAV format for Whisper.
func (w *TranscodeWorker) ExtractAudio(ctx context.Context, srcFile, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	out := filepath.Join(outDir, "audio.wav")
	_, err := w.ffmpeg.Run(ctx,
		"-i", srcFile, "-vn", "-acodec", "pcm_s16le",
		"-ar", "16000", "-ac", "1", out)
	return err
}

// Cleanup removes the local working directory after processing.
func (w *TranscodeWorker) Cleanup(dir string) error {
	return os.RemoveAll(dir)
}
