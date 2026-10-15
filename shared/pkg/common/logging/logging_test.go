package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"":        slog.LevelInfo,
		"debug":   slog.LevelDebug,
		"DEBUG":   slog.LevelDebug,
		"warn":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
		"unk":     slog.LevelInfo,
	}
	for in, want := range cases {
		if got := parseLevel(in); got != want {
			t.Errorf("parseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNewStdoutOnly(t *testing.T) {
	t.Setenv("LOG_DIR", "")
	t.Setenv("LOG_LEVEL", "")
	w := New(Options{Name: "t"})
	defer w.Close()
	if w.file != nil {
		t.Fatal("expected stdout-only writer, got file")
	}
}

func TestNewFileMode(t *testing.T) {
	dir := t.TempDir()
	w := New(Options{Name: "svc", Dir: dir})
	defer w.Close()
	if w.file == nil {
		t.Fatal("expected file writer")
	}
	_, err := os.Stat(filepath.Join(dir, "svc.log"))
	if err != nil {
		t.Fatalf("svc.log not created: %v", err)
	}
}

func TestRotate(t *testing.T) {
	dir := t.TempDir()
	r := &rotator{f: mustOpen(t, filepath.Join(dir, "a.log")), path: filepath.Join(dir, "a.log"), max: 10, count: 3}
	r.Write([]byte("01234567890123456789"))
	if _, err := os.Stat(filepath.Join(dir, "a.log.1")); err != nil {
		t.Fatalf("rotation not happened: %v", err)
	}
	r.f.Close()
}

func mustOpen(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	return f
}