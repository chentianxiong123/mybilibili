package ai

import (
	"context"
	"testing"
)

// TestLangchaingoCaller_NotConfigured: 未配置 base_url/api_key 时 Call 必须返回
// 显式错误（FINDING-083: 不占位/不假装成功），而非返回占位符。
func TestLangchaingoCaller_NotConfigured(t *testing.T) {
	caller, err := NewLangchaingoCaller(LangchaingoConfig{})
	if err != nil {
		t.Fatalf("NewLangchaingoCaller should construct even unconfigured: %v", err)
	}
	ctx := context.Background()

	// Summary：未配置 → 报错，绝不返回占位符。
	var summaryResp struct{ Summary string }
	if err := caller.Call(ctx, "ai", "Summary", map[string]interface{}{"video_id": float64(1)}, &summaryResp); err == nil {
		t.Fatalf("FINDING-083: 未配置 LLM 应返回错误，实际 err=nil")
	}
	if summaryResp.Summary != "" {
		t.Fatalf("FINDING-083: 未配置 LLM 不应返回占位摘要，实际=%q", summaryResp.Summary)
	}

	// CustomerChat：未配置 → 报错。
	var chatResp struct{ Reply string }
	if err := caller.Call(ctx, "ai", "CustomerChat", map[string]interface{}{"content": "hello"}, &chatResp); err == nil {
		t.Fatalf("FINDING-083: 未配置 LLM 应返回错误，实际 err=nil")
	}

	// Moderate：未配置 → 报错（失败关闭，不再 passed=true）。
	if err := caller.Call(ctx, "ai", "Moderate", map[string]interface{}{"content": "x", "scene": "COMMENT"}, &map[string]interface{}{}); err == nil {
		t.Fatalf("FINDING-083: 未配置 LLM 应返回错误，实际 err=nil")
	}

	// SummaryStream：未配置 → 报错。
	if _, err := caller.CallStream(ctx, "ai", "SummaryStream", map[string]interface{}{"video_id": float64(1)}); err == nil {
		t.Fatalf("FINDING-083: 未配置 LLM 流式应返回错误，实际 err=nil")
	}
}

// TestLangchaingoCaller_NonLLMMethods: 非 LLM 方法（CheckSummary/CustomerHistory/CustomerTransfer）
// 不经 LLM，应正常返回且不报错。
func TestLangchaingoCaller_NonLLMMethods(t *testing.T) {
	caller, _ := NewLangchaingoCaller(LangchaingoConfig{})
	ctx := context.Background()

	var checkResp struct{ HasSummary bool }
	if err := caller.Call(ctx, "ai", "CheckSummary", map[string]interface{}{"video_id": float64(1)}, &checkResp); err != nil {
		t.Fatalf("CheckSummary should not error: %v", err)
	}

	if err := caller.Call(ctx, "ai", "CustomerHistory", map[string]interface{}{"user_id": float64(1)}, &[]map[string]interface{}{}); err != nil {
		t.Fatalf("CustomerHistory should not error: %v", err)
	}

	if err := caller.Call(ctx, "ai", "CustomerTransfer", map[string]interface{}{}, &map[string]interface{}{}); err != nil {
		t.Fatalf("CustomerTransfer should not error: %v", err)
	}
}

// TestLangchaingoCaller_ExtractJSON: extractJSON 从散文中提取 JSON 对象。
func TestLangchaingoCaller_ExtractJSON(t *testing.T) {
	cases := map[string]string{
		`{"passed":true}`:              `{"passed":true}`,
		"```json\n{\"passed\":false}\n```": `{"passed":false}`,
		"前置文本 {\"passed\":true} 后置":      `{"passed":true}`,
	}
	for in, want := range cases {
		if got := extractJSON(in); got != want {
			t.Errorf("extractJSON(%q)=%q, want %q", in, got, want)
		}
	}
}
