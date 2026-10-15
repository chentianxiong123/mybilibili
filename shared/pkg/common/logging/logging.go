// Package logging 提供统一的结构化日志（标准库 log/slog）。
//
// 配置（各服务 main.go 启动时调用 logging.New，也可仅依赖环境变量）:
//   - LOG_DIR    日志目录。未设置: 输出 stdout（不落盘，适合 dev）；
//     设置后: 写 <LOG_DIR>/<Name>.log，并按 10MB 轮转 5 份（适合部署落盘）。
//   - LOG_LEVEL  日志级别: debug/info/warn/error，默认 info，可运行时精简化输出。
//
// 使用:
//
//	logger := logging.New(logging.Options{Name: "core"})
//	defer logger.Close()
//
// 之后全局 log/slog 均可直接使用（New 会 slog.SetDefault）。
package logging

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Options 日志初始化选项。
type Options struct {
	// Name 服务名，落盘时作为 <Name>.log 前缀。
	Name string
	// Level 日志级别（debug/info/warn/error）。空则读 LOG_LEVEL，再缺省 info。
	Level string
	// Dir 日志目录。空则读 LOG_DIR；仍为空则仅输出 stdout。
	Dir string
}

// Writer 输出目标（stdout 或文件），供 Close 时释放。
type Writer struct {
	file *os.File
}

// Close 关闭文件输出（仅落盘模式有实际文件）。
func (w *Writer) Close() {
	if w != nil && w.file != nil {
		_ = w.file.Close()
	}
}

// New 初始化统一 slog Logger，并设为全局默认值。
// 未配置目录: 仅 stdout（开发模式不落盘）；配置目录: stdout + 文件双写（开发也能看控制台）。
func New(o Options) *Writer {
	dir := o.Dir
	if dir == "" {
		dir = os.Getenv("LOG_DIR")
	}
	level := o.Level
	if level == "" {
		level = os.Getenv("LOG_LEVEL")
	}

	var writers []io.Writer
	writers = append(writers, os.Stdout)

	var file *os.File
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			slog.Default().Warn("logging: mkdir failed, stdout only", "dir", dir, "err", err)
		} else {
			path := filepath.Join(dir, logName(o.Name))
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				slog.Default().Warn("logging: open failed, stdout only", "path", path, "err", err)
			} else {
				file = f
				writers = append(writers, &rotator{f: f, path: path, max: 10 << 20, count: 5})
			}
		}
	}

	h := slog.NewTextHandler(io.MultiWriter(writers...), &slog.HandlerOptions{Level: parseLevel(level)})
	logger := slog.New(h)
	slog.SetDefault(logger)

	// 把遗留的标准库 log 也桥接到 slog，避免代码里 log.Printf 旁路成非结构化输出。
	// 以 INFO 级别转发; 若开启了 debug 级别则原样转发（保留调用方已有前缀）。
	bridgeStdLog(logger)

	return &Writer{file: file}
}

func bridgeStdLog(logger *slog.Logger) {
	stdFlags := log.Default().Flags()
	stdPrefix := log.Default().Prefix()
	log.SetOutput(&stdLogBridge{logger: logger})
	log.SetFlags(stdFlags)
	log.SetPrefix(stdPrefix)
}

// stdLogBridge 把标准库 log 输出重定向到 slog。
type stdLogBridge struct {
	logger *slog.Logger
}

func (b *stdLogBridge) Write(p []byte) (int, error) {
	b.logger.Info(strings.TrimSuffix(string(p), "\n"))
	return len(p), nil
}

func logName(name string) string {
	if name == "" {
		return "app.log"
	}
	return name + ".log"
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// rotator 按大小轮转文件: <name>.log -> <name>.log.1 -> ... -> <name>.log.N。
type rotator struct {
	f     *os.File
	path  string
	max   int64
	count int
}

func (r *rotator) Write(p []byte) (int, error) {
	if st, err := r.f.Stat(); err == nil && st.Size()+int64(len(p)) > r.max {
		r.rotate()
	}
	return r.f.Write(p)
}

func (r *rotator) rotate() {
	for i := r.count; i >= 1; i-- {
		src := r.path
		if i > 1 {
			src = fmt.Sprintf("%s.%d", r.path, i-1)
		}
		dst := fmt.Sprintf("%s.%d", r.path, i)
		_ = os.Remove(dst)
		_ = os.Rename(src, dst)
	}
	_ = r.f.Close()
	if nf, err := os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
		r.f = nf
	}
}
