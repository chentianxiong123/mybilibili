package ai

import "testing"

// FuzzExtractSummaryContent 穷尽任意输入，确认 extractSummaryContent 永不 panic
// 且输出是输入的子串（或为空）。
func FuzzExtractSummaryContent(f *testing.F) {
	f.Add("=====\n视频标题:x\n\n【视频摘要】正文")
	f.Add("### 视频摘要\n内容")
	f.Add("无标记正文")
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) {
		out := extractSummaryContent(s)
		if out != "" && len(out) > len(s) {
			t.Fatalf("output longer than input")
		}
	})
}
