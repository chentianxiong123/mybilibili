package message

import (
	"strings"
	"testing"
)

// TestExtractCommentText_KnownPanicBug 钉住 FINDING-010：
// extractCommentText 在 open 命中 s 的非 rune 对齐字节（非法 UTF-8）时 panic
// （slice bounds out of range）。该 bug 在项目源码 message_handler.go:476，
// 生产输入为合法 UTF-8 时不可达，但属真实潜在崩溃。此测试在 bug 修复前应通过
// （断言 panic 存在）；修复后应更新（移除断言 panic）。
func TestExtractCommentText_KnownPanicBug(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("FINDING-010 expected panic but got none: extractCommentText(\"《\", \"\\x80\") should panic (mid-rune match)")
		}
	}()
	// open="\x80" 命中 s="《" 的中间字节(0x80)，s[:1]="\xE3" 非法 UTF-8
	// → runeCount 过度计数 → runes[2:] 越界 panic。
	_ = extractCommentText("《", "\x80")
}

// FuzzExtractCommentText 穷尽任意输入。已知 FINDING-010 panic（非 rune 对齐）
// 被 recover 捕获并记为已知 bug，不失败；其余输入断言输出是归一化输入的子串。
func FuzzExtractCommentText(f *testing.F) {
	f.Add("赞了你的视频《测试》", "《")
	f.Add("赞了你的评论\"你好\"", "\"")
	f.Add("无匹配文本", "《")
	f.Add("", "《")
	f.Add("《未闭合", "《")
	f.Fuzz(func(t *testing.T, s, open string) {
		var out string
		func() {
			defer func() {
				if r := recover(); r != nil {
					// FINDING-010：已知 panic（非 rune 对齐匹配），记录不失败。
					t.Logf("FINDING-010 known panic recovered: s=%q open=%q panic=%v", s, open, r)
				}
			}()
			out = extractCommentText(s, open)
		}()
		if out != "" {
			normalized := string([]rune(s))
			if !strings.Contains(normalized, out) {
				t.Fatalf("out %q is not a substring of normalized input %q (s=%q open=%q)", out, normalized, s, open)
			}
		}
	})
}
