package whisper

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"mybilibili/pkg/abstraction"
)

// Request 转写请求（MinIO 对象引用，不直接传文件内容）。
type Request struct {
	Bucket       string `json:"bucket"`
	SourceKey    string `json:"source_key"`
	Language     string `json:"language"`
	Model        string `json:"model"`
	WhisperModel string `json:"whisper_model,omitempty"` // GGUF 模型文件名（如 ggml-base.bin）
}

// Segment 转写片段（对齐 Cloudflare Whisper 响应格式）。
type Segment struct {
	Index int     `json:"index"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// Result 转写结果。
type Result struct {
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
}

// Service 本地 whisper.cpp 转写服务：从 MinIO 读音频 → whisper-cli 转写 → 返回结果。
type Service struct {
	storage      abstraction.StorageService
	cliPath      string
	modelDir     string
	defaultModel string
	language     string
	threads      int
}

func NewService(storage abstraction.StorageService) *Service {
	cliPath := os.Getenv("WHISPER_CLI_PATH")
	if cliPath == "" {
		cliPath = "whisper-cli"
	}
	modelDir := os.Getenv("WHISPER_MODEL_DIR")
	if modelDir == "" {
		modelDir = "/models/whisper"
	}
	defaultModel := os.Getenv("WHISPER_DEFAULT_MODEL")
	if defaultModel == "" {
		defaultModel = "base"
	}
	lang := os.Getenv("WHISPER_LANGUAGE")
	if lang == "" {
		lang = "zh"
	}
	threads := 4
	if v := os.Getenv("WHISPER_THREADS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			threads = n
		}
	}
	log.Printf("whisper-local: cli=%s modelDir=%s defaultModel=%s lang=%s threads=%d",
		cliPath, modelDir, defaultModel, lang, threads)
	return &Service{
		storage:      storage,
		cliPath:      cliPath,
		modelDir:     modelDir,
		defaultModel: defaultModel,
		language:     lang,
		threads:      threads,
	}
}

// Transcribe 从 MinIO 下载音频 → whisper-cli 转写 → 返回结果。
func (s *Service) Transcribe(ctx context.Context, req Request) (*Result, error) {
	if req.Bucket == "" {
		req.Bucket = "mybilibili"
	}
	if req.Language == "" {
		req.Language = s.language
	}
	if req.Model == "" {
		req.Model = s.defaultModel
	}
	if req.SourceKey == "" {
		return nil, fmt.Errorf("source_key required")
	}

	// 1. 从 MinIO 下载音频到临时文件
	workDir, err := os.MkdirTemp("", "whisper-*")
	if err != nil {
		return nil, fmt.Errorf("mk temp: %w", err)
	}
	defer os.RemoveAll(workDir)

	audioFile := filepath.Join(workDir, "audio.mp3")
	if err := s.downloadFromMinIO(ctx, req.Bucket, req.SourceKey, audioFile); err != nil {
		return nil, fmt.Errorf("download audio: %w", err)
	}

	// 2. 构建 whisper-cli 参数
	modelPath := filepath.Join(s.modelDir, req.WhisperModel)
	if req.WhisperModel == "" {
		modelPath = filepath.Join(s.modelDir, fmt.Sprintf("ggml-%s.bin", req.Model))
	}

	args := []string{
		"-m", modelPath,
		"-l", req.Language,
		"--threads", strconv.Itoa(s.threads),
		"--fp16", "false",
		"-otxt",
		"-o", workDir,
		audioFile,
	}

	// 3. 执行 whisper-cli
	log.Printf("whisper-local: exec %s %s", s.cliPath, strings.Join(args, " "))
	out, err := exec.CommandContext(ctx, s.cliPath, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper-cli exec: %w\noutput: %s", err, string(out))
	}

	// 4. 读取输出 .txt 文件
	txtPath := audioFile + ".txt"
	txtData, err := os.ReadFile(txtPath)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}

	text := strings.TrimSpace(string(txtData))
	if text == "" {
		return &Result{Text: "", Segments: []Segment{}}, nil
	}

	// 5. 简单分段（whisper-cli -otxt 只输出纯文本，无时间戳）
	segments := splitTextToSegments(text)

	return &Result{
		Text:     text,
		Segments: segments,
	}, nil
}

// TranscribeSRT 带时间戳的转写（输出 SRT 格式，解析为 segments）。
func (s *Service) TranscribeSRT(ctx context.Context, req Request) (*Result, error) {
	if req.Bucket == "" {
		req.Bucket = "mybilibili"
	}
	if req.Language == "" {
		req.Language = s.language
	}
	if req.Model == "" {
		req.Model = s.defaultModel
	}
	if req.SourceKey == "" {
		return nil, fmt.Errorf("source_key required")
	}

	workDir, err := os.MkdirTemp("", "whisper-*")
	if err != nil {
		return nil, fmt.Errorf("mk temp: %w", err)
	}
	defer os.RemoveAll(workDir)

	audioFile := filepath.Join(workDir, "audio.mp3")
	if err := s.downloadFromMinIO(ctx, req.Bucket, req.SourceKey, audioFile); err != nil {
		return nil, fmt.Errorf("download audio: %w", err)
	}

	modelPath := filepath.Join(s.modelDir, req.WhisperModel)
	if req.WhisperModel == "" {
		modelPath = filepath.Join(s.modelDir, fmt.Sprintf("ggml-%s.bin", req.Model))
	}

	args := []string{
		"-m", modelPath,
		"-l", req.Language,
		"--threads", strconv.Itoa(s.threads),
		"--fp16", "false",
		"-osrt",
		"-o", workDir,
		audioFile,
	}

	log.Printf("whisper-local: exec %s %s (SRT mode)", s.cliPath, strings.Join(args, " "))
	out, err := exec.CommandContext(ctx, s.cliPath, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper-cli exec: %w\noutput: %s", err, string(out))
	}

	srtPath := audioFile + ".srt"
	srtData, err := os.ReadFile(srtPath)
	if err != nil {
		return nil, fmt.Errorf("read srt output: %w", err)
	}

	segments, text := parseSRT(string(srtData))
	return &Result{
		Text:     text,
		Segments: segments,
	}, nil
}

// srtTimeRe 匹配 SRT 时间码 00:00:00,000 --> 00:00:02,500
var srtTimeRe = regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2})[,.](\d{3})\s*-->\s*(\d{2}):(\d{2}):(\d{2})[,.](\d{3})`)

// parseSRT 解析 SRT 格式为 segments。
func parseSRT(srt string) ([]Segment, string) {
	var segments []Segment
	var fullText strings.Builder

	blocks := strings.Split(srt, "\n\n")
	for _, block := range blocks {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) < 3 {
			continue
		}

		matches := srtTimeRe.FindStringSubmatch(lines[1])
		if matches == nil {
			continue
		}

		start := srtTimeToFloat(matches[1], matches[2], matches[3], matches[4])
		end := srtTimeToFloat(matches[5], matches[6], matches[7], matches[8])

		text := strings.TrimSpace(strings.Join(lines[2:], " "))
		if text == "" {
			continue
		}

		segments = append(segments, Segment{
			Index: len(segments) + 1,
			Start: start,
			End:   end,
			Text:  text,
		})

		if fullText.Len() > 0 {
			fullText.WriteString("\n")
		}
		fullText.WriteString(text)
	}

	return segments, fullText.String()
}

func srtTimeToFloat(h, m, s, ms string) float64 {
	hh, _ := strconv.Atoi(h)
	mm, _ := strconv.Atoi(m)
	ss, _ := strconv.Atoi(s)
	msi, _ := strconv.Atoi(ms)
	return float64(hh*3600+mm*60+ss) + float64(msi)/1000.0
}

// splitTextToSegments 将纯文本按句分割为 segments（无精确时间戳）。
func splitTextToSegments(text string) []Segment {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanLines)

	var segments []Segment
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		segments = append(segments, Segment{
			Index: len(segments) + 1,
			Start: 0,
			End:   0,
			Text:  line,
		})
	}
	return segments
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
