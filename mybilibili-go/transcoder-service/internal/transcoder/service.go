package transcoder

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"mybilibili/pkg/abstraction"
)

// Request 转码请求（MinIO 对象引用，不直接传文件内容）。
// SourceKey: bucket 内的源视频对象 key（如 manuscripts/10/videos/25/source/video.mp4）。
type Request struct {
	Bucket       string   `json:"bucket"`
	SourceKey    string   `json:"source_key"`
	ManuscriptID int64    `json:"manuscript_id"`
	VideoID      int64    `json:"video_id"`
	Qualities    []string `json:"qualities"`
	ExtractAudio bool     `json:"extract_audio"`
}

// PlayURLs 转码产物播放地址（对齐 videos 表 play_url_hd/sd/ld）。
type Result struct {
	PlayURLs map[string]string `json:"play_urls"`
	AudioKey string            `json:"audio_key,omitempty"`
	// IsVertical 视频方向：0=横屏 1=竖屏 -1=未知
	IsVertical int32 `json:"is_vertical"`
}

// Service 转码服务：从 MinIO 读源，本地 ffmpeg 处理，产物写回 MinIO。
// 硬编类型由编译时 build tag 决定（hw_nvenc.go / hw_vaapi.go / hw_soft.go），
// 不做运行时探测：编译 nvenc 版 → NVIDIA，编译 vaapi 版 → AMD，默认 → 软编。
type Service struct {
	storage abstraction.StorageService
	ffmpeg  FFmpegRunner
	ffprobe FFprobeRunner
	vaapi   string // VAAPI 设备路径（仅 vaapi 版用到）
}

func NewService(storage abstraction.StorageService, _ string) *Service {
	dev := os.Getenv("VAAPI_DEVICE")
	if dev == "" {
		dev = "/dev/dri/renderD128"
	}
	s := &Service{
		storage: storage,
		ffmpeg:  systemFFmpeg{},
		ffprobe: systemFFprobe{},
		vaapi:   dev,
	}
	log.Printf("transcoder compiled hw: %s (vaapiDev=%s)", hwName(), s.vaapi)
	return s
}

// GetVideoSize 探测源视频宽高，用于判断横竖屏（高>宽为竖屏）。
func (s *Service) GetVideoSize(ctx context.Context, srcFile string) (width, height int, err error) {
	out, err := s.ffprobe.Run(ctx,
		"-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0", srcFile)
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid probe output")
	}
	fmt.Sscanf(parts[0], "%d", &width)
	fmt.Sscanf(parts[1], "%d", &height)
	return width, height, nil
}

// Process 执行一次转码任务：下载源 → 可选转码三档 → 可选提音频 → 上传产物 → 清理。
func (s *Service) Process(ctx context.Context, req Request) (*Result, error) {
	if req.Bucket == "" {
		req.Bucket = "mybilibili"
	}
	if len(req.Qualities) == 0 {
		req.Qualities = []string{"1080p", "720p", "480p"}
	}
	if req.SourceKey == "" {
		return nil, fmt.Errorf("source_key required")
	}

	workDir, err := os.MkdirTemp("", "transcode-*")
	if err != nil {
		return nil, fmt.Errorf("mk temp: %w", err)
	}
	defer os.RemoveAll(workDir)

	// 1. 从 MinIO 下载源到本地
	src := filepath.Join(workDir, "source.mp4")
	if err := s.downloadFromMinIO(ctx, req.Bucket, req.SourceKey, src); err != nil {
		return nil, fmt.Errorf("download source: %w", err)
	}

	res := &Result{PlayURLs: map[string]string{}, IsVertical: -1}

	// 探测方向（高>宽即竖屏），供 videos.is_vertical 回写
	if w, h, err := s.GetVideoSize(ctx, src); err == nil && w > 0 && h > 0 {
		if h > w {
			res.IsVertical = 1
		} else {
			res.IsVertical = 0
		}
	}

	// 2. 转码（多档 HLS）
	for _, q := range req.Qualities {
		qDir := filepath.Join(workDir, q)
		if err := s.transcode(ctx, src, qDir, q); err != nil {
			// 单档失败不整体失败，跳过（保持与 work 现状一致）
			log.Printf("transcode %s warning: %v", q, err)
			continue
		}
		if err := s.uploadDir(ctx, req.Bucket,
			fmt.Sprintf("manuscripts/%d/videos/%d/transcoded/%s", req.ManuscriptID, req.VideoID, q),
			qDir); err != nil {
			return nil, fmt.Errorf("upload %s: %w", q, err)
		}
		res.PlayURLs[q] = fmt.Sprintf("/uploads/manuscripts/%d/videos/%d/transcoded/%s/playlist.m3u8",
			req.ManuscriptID, req.VideoID, q)
	}

	// 3. 提音频（字幕/whisper 用）
	if req.ExtractAudio {
		outDir := filepath.Join(workDir, "audio")
		if err := s.extractAudio(ctx, src, outDir); err != nil {
			return nil, fmt.Errorf("extract audio: %w", err)
		}
		audioKey := fmt.Sprintf("manuscripts/%d/videos/%d/audio/audio.mp3", req.ManuscriptID, req.VideoID)
		if f, err := os.Open(filepath.Join(outDir, "audio.mp3")); err == nil {
			if err := s.storage.Put(ctx, req.Bucket, audioKey, f, "audio/mpeg"); err != nil {
				f.Close()
				return nil, fmt.Errorf("upload audio: %w", err)
			}
			f.Close()
			res.AudioKey = audioKey
		}
	}

	return res, nil
}

// transcode 单档转码：scale + 硬编/软编 → HLS。
// 硬编参数由 build tag 选定的 hwBuildArgs 提供（nvenc/vaapi/soft）。
func (s *Service) transcode(ctx context.Context, srcFile, outDir, quality string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	crf := "32"
	scale := "min(1280\\,iw):-2"
	switch quality {
	case "1080p":
		crf = "28"
		scale = "min(1920\\,iw):-2"
	case "720p":
		crf = "30"
		scale = "min(1280\\,iw):-2"
	default:
		crf = "32"
		scale = "min(854\\,iw):-2"
	}

	segPattern := filepath.Join(outDir, "segment%03d.ts")
	playlist := filepath.Join(outDir, "playlist.m3u8")
	audio := []string{"-c:a", "aac", "-b:a", "128k"}
	hls := []string{
		"-f", "hls", "-hls_time", "6", "-hls_list_size", "0",
		"-hls_segment_filename", segPattern,
		playlist,
	}

	args := hwBuildArgs(srcFile, scale, crf, s.vaapi, audio, hls)
	_, err := s.ffmpeg.Run(ctx, args...)
	if err == nil || !hwAvailable() {
		return err
	}

	// 硬编失败自动回退软编
	log.Printf("[%s] %s encode failed (%v), fallback to libx265", quality, hwName(), err)
	args = softBuildArgs(srcFile, scale, crf, audio, hls)
	_, retryErr := s.ffmpeg.Run(ctx, args...)
	return retryErr
}

// extractAudio 提取音频为 MP3（whisper 兼容）。
func (s *Service) extractAudio(ctx context.Context, srcFile, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	out := filepath.Join(outDir, "audio.mp3")
	_, err := s.ffmpeg.Run(ctx,
		"-i", srcFile, "-vn", "-acodec", "libmp3lame",
		"-b:a", "64k", "-ar", "44100", "-ac", "1", out)
	return err
}

// downloadFromMinIO 从 MinIO 下载对象到本地文件。
func (s *Service) downloadFromMinIO(ctx context.Context, bucket, key, dest string) error {
	rc, err := s.storage.Get(ctx, bucket, key)
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

// uploadDir 将本地目录内全部非目录文件上传到 MinIO prefix。
func (s *Service) uploadDir(ctx context.Context, bucket, prefix, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		key := prefix + "/" + e.Name()
		ct := "application/octet-stream"
		if strings.HasSuffix(e.Name(), ".m3u8") {
			ct = "application/vnd.apple.mpegurl"
		} else if strings.HasSuffix(e.Name(), ".ts") {
			ct = "video/mp2t"
		}
		if err := s.storage.Put(ctx, bucket, key, f, ct); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}