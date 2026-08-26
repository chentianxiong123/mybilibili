package work

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// EncoderKind 标识转码编码方式。auto 默认：Linux 探测 VAAPI 硬编（AMD/Intel），无 GPU 回退软编。
type TranscodeWorker struct {
	ffmpeg   FFmpegRunner
	ffprobe  FFprobeRunner
	storage  abstraction.StorageService
	encoder  string // auto | vaapi | x264
	vaapiDev string
	useHW    bool
}

// NewTranscodeWorker 创建转码 worker。
// encoder 取值：
//   - auto（默认）：探测 /dev/dri/renderD128 + ffmpeg 是否支持 h264_vaapi，支持则硬编，否则软编；
//   - vaapi：强制 VAAPI 硬编（Linux AMD/Intel）；
//   - x264：强制软件编码（任何机器都能跑，永不依赖驱动）。
func NewTranscodeWorker(storage abstraction.StorageService, encoder string) *TranscodeWorker {
	if encoder == "" {
		encoder = "auto"
	}
	dev := os.Getenv("VAAPI_DEVICE")
	if dev == "" {
		dev = "/dev/dri/renderD128"
	}
	w := &TranscodeWorker{
		ffmpeg:   systemFFmpeg{},
		ffprobe:  systemFFprobe{},
		storage:  storage,
		encoder:  encoder,
		vaapiDev: dev,
	}
	w.useHW = w.detectHW()
	log.Printf("transcode encoder: %s (useHW=%v, device=%s)", w.encoder, w.useHW, w.vaapiDev)
	return w
}

// detectHW 探测当前机器是否可用 VAAPI 硬编。auto/vaapi 模式下启用，x264 强制禁用。
// 探测条件：
//  1. /dev/dri/renderD128 设备节点存在（Linux GPU/核显驱动已加载）；
//  2. 本机 ffmpeg 编译了 h264_vaapi 编码器。
// 满足两者才走硬编；换电脑缺其一则自动回退软编，转码永不失败。
func (w *TranscodeWorker) detectHW() bool {
	if w.encoder == "x264" {
		return false
	}
	if _, err := os.Stat(w.vaapiDev); err != nil {
		return false
	}
	out, err := w.ffmpeg.Run(context.Background(), "-hide_banner", "-encoders")
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "h264_vaapi")
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

// GetVideoSize 用 ffprobe 探测源视频的宽高，用于判断横竖屏（高>宽为竖屏）。
func (w *TranscodeWorker) GetVideoSize(ctx context.Context, srcFile string) (width, height int, err error) {
	out, err := w.ffprobe.Run(ctx,
		"-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0", srcFile)
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "x")
	if len(parts) == 2 {
		fmt.Sscanf(parts[0], "%d", &width)
		fmt.Sscanf(parts[1], "%d", &height)
	}
	return width, height, nil
}

// transcodeArgs 组装 ffmpeg 参数。
// 硬编：h264_vaapi CQP（快、CPU 近零，体积略大）；软编：libx264 CRF（体积优，吃 CPU）。
func (w *TranscodeWorker) transcodeArgs(scale, crf, srcFile, outDir string) []string {
	segPattern := filepath.Join(outDir, "segment%03d.ts")
	playlist := filepath.Join(outDir, "playlist.m3u8")
	audio := []string{"-c:a", "aac", "-b:a", "128k"}
	hls := []string{
		"-f", "hls", "-hls_time", "6", "-hls_list_size", "0",
		"-hls_segment_filename", segPattern,
		playlist,
	}
	if w.useHW {
		return append([]string{
			"-i", srcFile,
			"-vf", "scale=" + scale + ",format=nv12,hwupload",
			"-vaapi_device", w.vaapiDev,
			"-c:v", "h264_vaapi", "-rc_mode", "CQP", "-qp", crf,
		}, append(audio, hls...)...)
	}
	return append([]string{
		"-i", srcFile,
		"-vf", "scale=" + scale,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", crf,
	}, append(audio, hls...)...)
}

// Transcode converts source video to HLS segments at the given quality.
// 硬编经 useHW 探测保证可用；万一探测通过但运行失败，自动回退软编再试一次，保证转码不中断。
func (w *TranscodeWorker) Transcode(ctx context.Context, srcFile, outDir string, quality string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	crf := "23"
	scale := "min(1280,iw):-2"
	switch quality {
	case "1080p":
		crf = "20"
		scale = "min(1920,iw):-2"
	case "720p":
		crf = "22"
		scale = "min(1280,iw):-2"
	default:
		crf = "23"
		scale = "min(854,iw):-2"
	}

	args := w.transcodeArgs(scale, crf, srcFile, outDir)
	_, err := w.ffmpeg.Run(ctx, args...)
	if err == nil || !w.useHW {
		return err
	}
	// VAAPI 运行失败 → 自动回退软编重试，保证换电脑/无 GPU 也能转码
	log.Printf("[%s] VAAPI encode failed (%v), falling back to libx264", quality, err)
	tmp := w.useHW
	w.useHW = false
	args = w.transcodeArgs(scale, crf, srcFile, outDir)
	_, retryErr := w.ffmpeg.Run(ctx, args...)
	w.useHW = tmp
	return retryErr
}

// ExtractAudio extracts the audio track into MP3 format (Cloudflare Whisper compatible).
func (w *TranscodeWorker) ExtractAudio(ctx context.Context, srcFile, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	out := filepath.Join(outDir, "audio.mp3")
	_, err := w.ffmpeg.Run(ctx,
		"-i", srcFile, "-vn", "-acodec", "libmp3lame",
		"-b:a", "64k", "-ar", "44100", "-ac", "1", out)
	return err
}

// Cleanup removes the local working directory after processing.
func (w *TranscodeWorker) Cleanup(dir string) error {
	return os.RemoveAll(dir)
}