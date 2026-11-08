package bilibili

import (
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errReadCloser 返回读取错误，用于覆盖 io.ReadAll 的错误分支。
type errReadCloser struct{}

func (errReadCloser) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (errReadCloser) Close() error             { return nil }

// rawJSONResponse 构造一个不经过 json.Marshal 的响应，避免 RawMessage 在序列化侧校验。
func rawJSONResponse(t *testing.T, status int, body []byte) *http.Response {
	t.Helper()
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       nopCloser(body),
	}
}

func TestClient_throttle_SleepsWhenRecent(t *testing.T) {
	c := NewClient("")
	c.lastCall = time.Now() // elapsed ≈ 0 < minInterval(300ms)

	start := time.Now()
	c.throttle()
	elapsed := time.Since(start)

	assert.True(t, elapsed >= 250*time.Millisecond, "expected throttle to sleep ~minInterval, got %s", elapsed)
	assert.False(t, c.lastCall.IsZero(), "lastCall should be updated")
}

func TestClient_throttle_NoSleepWhenStale(t *testing.T) {
	c := NewClient("")
	c.lastCall = time.Now().Add(-time.Second) // > minInterval(300ms)

	start := time.Now()
	c.throttle()
	elapsed := time.Since(start)

	assert.True(t, elapsed < 100*time.Millisecond, "expected no sleep for stale lastCall, got %s", elapsed)
	assert.False(t, c.lastCall.IsZero(), "lastCall should be updated")
}

func TestClient_doGet_InvalidURL(t *testing.T) {
	c := NewClient("")
	_, err := c.doGet("\x00") // null byte -> url.Parse error -> http.NewRequest error
	assert.Error(t, err)
}

func TestClient_doGet_ReadBodyError(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       errReadCloser{},
		}, nil
	})

	c := NewClient("")
	_, err := c.doGet("https://api.bilibili.com/x")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read failed")
}

func TestClient_doGet_NoCookieWhenSessdataEmpty(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		_, err := req.Cookie("SESSDATA")
		assert.Error(t, err, "SESSDATA cookie must not be set when sessdata is empty")
		return jsonResponse(t, 200, APIResp{Code: 0, Data: []byte(`{}`)}), nil
	})

	c := NewClient("")
	_, err := c.doGet("https://api.bilibili.com/x")
	require.NoError(t, err)
}

func TestClient_GetView_InvalidJSONBody(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, `not json`), nil
	})

	c := NewClient("")
	_, err := c.GetView("BV1xx")
	assert.Error(t, err)
}

func TestClient_GetView_InvalidDataJSON(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		// data 是合法 JSON（字符串），外层 APIResp 解析成功，但无法反序列化进 ViewInfo。
		return rawJSONResponse(t, 200, []byte(`{"code":0,"data":"oops"}`)), nil
	})

	c := NewClient("")
	_, err := c.GetView("BV1xx")
	assert.Error(t, err)
}

func TestClient_GetView_FullInfoFields(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{
			Code: 0,
			Data: []byte(`{
				"bvid":"BV1full","aid":9,"cid":10,"title":"Full","desc":"d","pic":"https://p/1.jpg",
				"duration":120,"pubdate":1700000000,"tname":"音乐","tag":[{"id":1}],
				"owner":{"mid":7,"name":"owner","face":"https://o/1.jpg"},
				"stat":{"view":1,"danmaku":2,"reply":3,"favorite":4,"coin":5,"share":6,"like":7},
				"pages":[{"cid":10,"page":1,"part":"P1","duration":120,"first_frame":"https://f/1.jpg"}]
			}`),
		}), nil
	})

	c := NewClient("")
	info, err := c.GetView("BV1full")
	require.NoError(t, err)
	assert.Equal(t, "BV1full", info.BVID)
	assert.Equal(t, int64(9), info.AID)
	assert.Equal(t, int64(10), info.CID)
	assert.Equal(t, "Full", info.Title)
	assert.Equal(t, "d", info.Desc)
	assert.Equal(t, "https://p/1.jpg", info.Pic)
	assert.Equal(t, int64(120), info.Duration)
	assert.Equal(t, int64(1700000000), info.Pubdate)
	assert.Equal(t, "音乐", info.TName)
	assert.Equal(t, int64(7), info.Owner.Mid)
	assert.Equal(t, "owner", info.Owner.Name)
	assert.Equal(t, "https://o/1.jpg", info.Owner.Face)
	assert.Equal(t, int64(1), info.Stat.View)
	assert.Equal(t, int64(2), info.Stat.Danmaku)
	assert.Equal(t, int64(3), info.Stat.Reply)
	assert.Equal(t, int64(4), info.Stat.Favorite)
	assert.Equal(t, int64(5), info.Stat.Coin)
	assert.Equal(t, int64(6), info.Stat.Share)
	assert.Equal(t, int64(7), info.Stat.Like)
	require.Len(t, info.Pages, 1)
	assert.Equal(t, int64(10), info.Pages[0].CID)
	assert.Equal(t, 1, info.Pages[0].Page)
	assert.Equal(t, "P1", info.Pages[0].Part)
	assert.Equal(t, int64(120), info.Pages[0].Duration)
	assert.Equal(t, "https://f/1.jpg", info.Pages[0].FirstFrame)
	assert.NotEmpty(t, info.Tag, "tag raw message should be preserved")
	assert.Nil(t, info.Everything, "Everything is json:\"-\" and must stay nil")
}

func TestClient_GetPlayURL_InvalidJSONBody(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, `{bad`), nil
	})

	c := NewClient("")
	_, err := c.GetPlayURL("BV1xx", 1, 80)
	assert.Error(t, err)
}

func TestClient_GetPlayURL_InvalidDataJSON(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		// data 是合法 JSON（数字），外层 APIResp 解析成功，但无法反序列化进 PlayURLData。
		return rawJSONResponse(t, 200, []byte(`{"code":0,"data":123}`)), nil
	})

	c := NewClient("")
	_, err := c.GetPlayURL("BV1xx", 1, 80)
	assert.Error(t, err)
}

func TestClient_GetPlayURL_HTTPError(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 500, Body: nopCloser([]byte("")), Header: make(http.Header)}, nil
	})

	c := NewClient("")
	_, err := c.GetPlayURL("BV1xx", 1, 80)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bilibili http 500")
}

func TestClient_GetPlayURL_RequestError(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})

	c := NewClient("")
	_, err := c.GetPlayURL("BV1xx", 1, 80)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestClient_GetPlayURL_DashMapping(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{
			Code: 0,
			Data: []byte(`{
				"format":"dash","quality":64,"accept_quality":[32,64],
				"dash":{
					"video":[{"id":16,"base_url":"https://cdn/v.mp4","baseUrl":"https://cdn/v.mp4","bandwidth":100,"mime_type":"video/mp4","codecs":"avc1"}],
					"audio":[{"id":1,"base_url":"https://cdn/a.m4a","baseUrl":"https://cdn/a.m4a","bandwidth":50,"mime_type":"audio/mp4","codecs":"mp4a"}]
				}
			}`),
		}), nil
	})

	c := NewClient("")
	pu, err := c.GetPlayURL("BV1xx", 1, 64)
	require.NoError(t, err)
	assert.Equal(t, "dash", pu.Format)
	assert.Equal(t, 64, pu.Quality)
	assert.Equal(t, []int{32, 64}, pu.AcceptQuality)
	require.NotNil(t, pu.Dash)
	require.Len(t, pu.Dash.Video, 1)
	assert.Equal(t, 16, pu.Dash.Video[0].ID)
	assert.Equal(t, "https://cdn/v.mp4", pu.Dash.Video[0].BaseURL)
	assert.Equal(t, "https://cdn/v.mp4", pu.Dash.Video[0].BaseURLC)
	assert.Equal(t, int64(100), pu.Dash.Video[0].Bandwidth)
	assert.Equal(t, "video/mp4", pu.Dash.Video[0].MimeType)
	assert.Equal(t, "avc1", pu.Dash.Video[0].Codecs)
	require.Len(t, pu.Dash.Audio, 1)
	assert.Equal(t, "https://cdn/a.m4a", pu.Dash.Audio[0].BaseURL)
}

func TestClient_ResolveStream_DashNoVideoOnlyAudio(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{
			Code: 0,
			Data: []byte(`{"dash":{"audio":[{"id":1,"base_url":"https://cdn/a.m4a"}]}}`),
		}), nil
	})

	c := NewClient("")
	_, _, err := c.ResolveStream("BV1xx", 1, 80)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no playable stream")
}

func TestClient_ResolveStream_DashNil(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{Code: 0, Data: []byte(`{"durl":[],"dash":null}`)}), nil
	})

	c := NewClient("")
	_, _, err := c.ResolveStream("BV1xx", 1, 80)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no playable stream")
}

func TestClient_GetView_Concurrent(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{Code: 0, Data: []byte(`{"bvid":"BV1xx"}`)}), nil
	})

	// 注意：同一 Client 共享 lastCall 是非并发安全的（api.go:39-43 存在数据竞争），
	// 因此这里每个 goroutine 使用独立 Client，仍并发走共享的包级 httpClient。
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := NewClient("sess")
			info, err := c.GetView("BV1xx")
			require.NoError(t, err)
			assert.Equal(t, "BV1xx", info.BVID)
		}()
	}
	wg.Wait()
}
