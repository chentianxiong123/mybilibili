package comment

import "testing"

// FuzzParseMentionNames 穷尽任意输入，确认 parseMentionNames 永不 panic 且结果去重。
func FuzzParseMentionNames(f *testing.F) {
	f.Add("@a,@b  @c 测试 @d。")
	f.Add("@@@")
	f.Add("@")
	f.Add("@,.;:!?，。；：！？、）】")
	f.Add("普通文本无提及")
	f.Add("@a@b@c")
	f.Fuzz(func(t *testing.T, s string) {
		out := parseMentionNames(s)
		// 结果不得含空串，且去重。
		seen := map[string]bool{}
		for _, n := range out {
			if n == "" {
				t.Fatalf("empty mention for %q", s)
			}
			if seen[n] {
				t.Fatalf("duplicate mention %q for %q", n, s)
			}
			seen[n] = true
		}
	})
}
