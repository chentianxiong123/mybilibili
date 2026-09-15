package transcoder

import (
	"context"
	"os/exec"
)

// FFmpegRunner abstracts the ffmpeg binary execution so it can be stubbed in tests.
type FFmpegRunner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

type systemFFmpeg struct{}

func (systemFFmpeg) Run(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput()
}

// FFprobeRunner abstracts the ffprobe binary execution.
type FFprobeRunner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

type systemFFprobe struct{}

func (systemFFprobe) Run(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "ffprobe", args...).CombinedOutput()
}

var _ FFmpegRunner = systemFFmpeg{}
var _ FFprobeRunner = systemFFprobe{}