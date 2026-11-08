package abstraction

import "testing"

// FuzzExtractJSON 穷尽任意输入，确认 extractJSON 永不 panic 且返回子串。
func FuzzExtractJSON(f *testing.F) {
	f.Add(`{"a":1}`)
	f.Add(`前缀 {"a":1} 后缀`)
	f.Add(`无花括号文本`)
	f.Add(`{a}b}`)
	f.Add(`}开头{` )
	f.Fuzz(func(t *testing.T, s string) {
		out := extractJSON(s)
		if len(out) > len(s) {
			t.Fatalf("extractJSON returned longer (%d) than input (%d)", len(out), len(s))
		}
	})
}
