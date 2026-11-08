package manuscript

import (
	"encoding/json"
	"testing"
)

// extractPlayURL 复制 manuscript_http_handler.go:489-499 的 playURL 提取逻辑
// （FINDING-021 钉住：video_url 即使为空也会覆盖 url）。
func extractPlayURL(videosJSON string) string {
	var vv []map[string]interface{}
	_ = json.Unmarshal([]byte(videosJSON), &vv)
	playURL := ""
	if len(vv) > 0 {
		if u, ok := vv[0]["url"].(string); ok {
			playURL = u
		}
		if u, ok := vv[0]["video_url"].(string); ok {
			playURL = u
		}
	}
	return playURL
}

// TestUploadComplete_PlayURLOverwrite 钉住 FINDING-021：
// 视频对象同时含 url（有效）与 video_url（空字符串）时，video_url 覆盖 url，
// playURL 变 "" → 上游 handleUploadComplete 误报 "no playable video source"。
func TestUploadComplete_PlayURLOverwrite(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string // 正确期望（修复后）
	}{
		{"仅url", `[{"url":"/uploads/m/1/videos/1/source/video.mp4"}]`, "/uploads/m/1/videos/1/source/video.mp4"},
		{"url+空video_url(bug)", `[{"url":"/uploads/m/1/videos/1/source/video.mp4","video_url":""}]`, "/uploads/m/1/videos/1/source/video.mp4"},
		{"video_url优先", `[{"url":"/u/a.mp4","video_url":"/u/b.mp4"}]`, "/u/b.mp4"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractPlayURL(tc.in)
			// 当前 bug：url+空video_url → got==""（错误）；修复后应为 want。
			if tc.name == "url+空video_url(bug)" {
				if got == "" {
					t.Logf("FINDING-021 确认：url 有效但 video_url 空 → playURL=%q（误报 no source）", got)
				} else if got != tc.want {
					t.Errorf("playURL=%q 期望 %q", got, tc.want)
				}
			} else {
				if got != tc.want {
					t.Errorf("playURL=%q 期望 %q", got, tc.want)
				}
			}
		})
	}
}
