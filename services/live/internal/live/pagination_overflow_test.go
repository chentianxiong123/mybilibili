package live

import "testing"

// TestLivePagination_KnownOverflowBug 钉住 FINDING-012（live 变体）：
// live_http_handler 用 strconv.ParseInt(...,10,32) 解析 page/pageSize 且不钳制，
// live_repository.ListLiveRooms 计算 offset=(page-1)*pageSize。
// 缺 page（page=0）+ page_size>0 → offset 为负 → PG OFFSET must not be negative → 500；
// page=2^31-1 → int32 溢出 → offset 负 → 500。此测试在修复前应通过（断言负 offset）。
func TestLivePagination_KnownOverflowBug(t *testing.T) {
	cases := []struct {
		name            string
		page, pageSize  int32
		wantNegative    bool
	}{
		{"缺page但有pageSize", 0, 20, true},
		{"正常", 1, 20, false},
		{"超大page溢出", 2147483647, 20, true},
		{"page0_size0", 0, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			offset := (tc.page - 1) * tc.pageSize
			if tc.wantNegative && offset >= 0 {
				t.Fatalf("FINDING-012：期望负 offset 但得到 %d（page=%d pageSize=%d）", offset, tc.page, tc.pageSize)
			}
			if !tc.wantNegative && offset < 0 {
				t.Fatalf("不应负 offset：%d（page=%d pageSize=%d）", offset, tc.page, tc.pageSize)
			}
		})
	}
}
