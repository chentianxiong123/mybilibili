package bilibili

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nopCloser(b []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(b))
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type mockTransport struct {
	handler func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.handler(req)
}

func jsonResponse(t *testing.T, status int, body any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       nopCloser(b),
	}
}

func installMockClient(t *testing.T, handler func(req *http.Request) (*http.Response, error)) {
	t.Helper()
	orig := httpClient
	httpClient = &http.Client{Transport: &mockTransport{handler: handler}}
	t.Cleanup(func() { httpClient = orig })
}

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("sess")
	assert.Equal(t, "sess", c.sessdata)
	assert.Equal(t, playURLPrefix, c.PlayPrefix())

	c.SetPlayPrefix("/custom/prefix/")
	assert.Equal(t, "/custom/prefix/", c.PlayPrefix())
}

func TestClient_GetView_Success(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "https://api.bilibili.com/x/web-interface/view", req.URL.Scheme+"://"+req.URL.Host+req.URL.Path)
		assert.Equal(t, "BV1xx", req.URL.Query().Get("bvid"))
		assert.Equal(t, userAgent, req.Header.Get("User-Agent"))
		assert.Equal(t, Referer, req.Header.Get("Referer"))
		return jsonResponse(t, 200, APIResp{
			Code:    0,
			Message: "0",
			Data:    []byte(`{"bvid":"BV1xx","aid":123,"cid":456,"title":"T","owner":{"mid":1,"name":"up"},"stat":{"view":100}}`),
		}), nil
	})

	c := NewClient("")
	info, err := c.GetView("BV1xx")
	require.NoError(t, err)
	assert.Equal(t, "BV1xx", info.BVID)
	assert.Equal(t, int64(123), info.AID)
	assert.Equal(t, int64(456), info.CID)
	assert.Equal(t, "T", info.Title)
	assert.Equal(t, "up", info.Owner.Name)
	assert.Equal(t, int64(100), info.Stat.View)
}

func TestClient_GetView_APICode(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{Code: -400, Message: "请求错误"}), nil
	})

	c := NewClient("")
	_, err := c.GetView("BV1xx")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code=-400")
}

func TestClient_GetView_HTTPError(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 500, Body: nopCloser([]byte("")), Header: make(http.Header)}, nil
	})

	c := NewClient("")
	_, err := c.GetView("BV1xx")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bilibili http 500")
}

func TestClient_GetView_RequestError(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("connection refused")
	})

	c := NewClient("")
	_, err := c.GetView("BV1xx")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestClient_GetPlayURL_Durl(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "/x/player/playurl", req.URL.Path)
		q := req.URL.Query()
		assert.Equal(t, "BV1xx", q.Get("bvid"))
		assert.Equal(t, "456", q.Get("cid"))
		assert.Equal(t, "80", q.Get("qn"))
		assert.Equal(t, "html5", q.Get("platform"))
		return jsonResponse(t, 200, APIResp{
			Code: 0,
			Data: []byte(`{"format":"mp4","quality":80,"durl":[{"url":"https://cdn/1.mp4","size":100,"length":10}]}`),
		}), nil
	})

	c := NewClient("")
	pu, err := c.GetPlayURL("BV1xx", 456, 80)
	require.NoError(t, err)
	require.Len(t, pu.DURL, 1)
	assert.Equal(t, "https://cdn/1.mp4", pu.DURL[0].URL)
	assert.Equal(t, 80, pu.Quality)
}

func TestClient_GetPlayURL_Code(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{Code: -404, Message: "啥都木有"}), nil
	})

	c := NewClient("")
	_, err := c.GetPlayURL("BV1xx", 1, 80)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "playurl code=-404")
}

func TestClient_ResolveStream_DurlPriority(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{
			Code: 0,
			Data: []byte(`{"durl":[{"url":"https://cdn/a.mp4"}]}`),
		}), nil
	})

	c := NewClient("")
	url, ct, err := c.ResolveStream("BV1xx", 1, 80)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn/a.mp4", url)
	assert.Equal(t, "video/mp4", ct)
}

func TestClient_ResolveStream_DashFallback(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{
			Code: 0,
			Data: []byte(`{"dash":{"video":[{"id":16,"base_url":"https://cdn/dash.mp4"}]}}`),
		}), nil
	})

	c := NewClient("")
	url, ct, err := c.ResolveStream("BV1xx", 1, 80)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn/dash.mp4", url)
	assert.Equal(t, "video/mp4", ct)
}

func TestClient_ResolveStream_NoPlayable(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{Code: 0, Data: []byte(`{}`)}), nil
	})

	c := NewClient("")
	_, _, err := c.ResolveStream("BV1xx", 1, 80)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no playable stream")
}

func TestClient_ResolveStream_UpstreamError(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, 200, APIResp{Code: -1, Message: "err"}), nil
	})

	c := NewClient("")
	_, _, err := c.ResolveStream("BV1xx", 1, 80)
	assert.Error(t, err)
}

func TestClient_GetView_WithSessdataCookie(t *testing.T) {
	installMockClient(t, func(req *http.Request) (*http.Response, error) {
		ck, err := req.Cookie("SESSDATA")
		require.NoError(t, err)
		assert.Equal(t, "s3cret", ck.Value)
		return jsonResponse(t, 200, APIResp{Code: 0, Data: []byte(`{"bvid":"BV1xx"}`)}), nil
	})

	c := NewClient("s3cret")
	_, err := c.GetView("BV1xx")
	require.NoError(t, err)
}
