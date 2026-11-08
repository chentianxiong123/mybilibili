package bilibili

import (
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_GetView_ConcurrentShared_Race 真正钉住 FINDING-004：
// bilibili.Client.throttle() 读写 c.lastCall 于 api.go:39-43 无同步。
// 与 TestClient_GetView_Concurrent 不同，本测试使用【同一共享 Client】并发调用
// GetView → 多 goroutine 并发 throttle → -race 应报告 lastCall 读写竞争。
func TestClient_GetView_ConcurrentShared_Race(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{Code: 0, Data: []byte(`{"bvid":"BV1xx"}`)}), nil
	})

	// 共享一个 Client（FINDING-004：lastCall 非并发安全）
	c := NewClient("sess")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			info, err := c.GetView("BV1xx")
			require.NoError(t, err)
			assert.Equal(t, "BV1xx", info.BVID)
		}()
	}
	wg.Wait()
}
