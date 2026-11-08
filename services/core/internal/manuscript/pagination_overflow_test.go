package manuscript

import "testing"

// TestManuscriptPagination_KnownOverflowBug 钉住 FINDING-012：
// handleMyCollections/Likes/Coins 的分页 start := int((page-1)*size) 为 int32 乘法，
// 当 page 极大（ParsePageParams 允许到 2^31-1）时溢出为负，导致 ids[start:end] 切片
// panic（slice bounds out of range）。此测试在 bug 修复前应通过（断言溢出+panic）；
// 修复后应更新（改为断言正确分页）。
func TestManuscriptPagination_KnownOverflowBug(t *testing.T) {
	// 复刻 handler 的分页计算（int32 乘法 + 切片），验证 FINDING-012 溢出。
	page := int32(2147483647) // ParsePageParams 允许的最大值
	size := int32(50)         // 默认/最大 page_size
	start := int((page - 1) * size)
	if start >= 0 {
		t.Fatalf("未复现 FINDING-012：start=%d（期望为负，int32 溢出）", start)
	}
	t.Logf("FINDING-012 确认：page=%d size=%d -> start=%d（int32 溢出为负）", page, size, start)

	// 复刻切片（handler 的 ids[start:end]），断言 panic。
	ids := []int64{1, 2, 3}
	end := start + int(size)
	if start > len(ids) {
		start = len(ids)
	}
	if end > len(ids) {
		end = len(ids)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("FINDING-012 期望 panic（负 start 切片）但未发生")
		}
	}()
	_ = ids[start:end]
}
