package work

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPool(t *testing.T) (*TranscoderPool, string) {
	t.Helper()
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "transcoders.yaml")
	// 初始配置：1 个软编码节点
	yml := `transcoders:
  - name: transcoder-local-soft
    addr: http://127.0.0.1:8092
    capabilities: [soft, libx264]
    weight: 1
load_balance: round_robin
health_check:
  interval: 1m
  timeout: 3s
  failure_threshold: 3
`
	require.NoError(t, os.WriteFile(cfgFile, []byte(yml), 0o644))

	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	return pool, cfgFile
}

func TestHandleListTranscoders_200(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transcoders", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `"transcoder-local-soft"`)
	assert.Contains(t, body, `"total":1`)
}

func TestHandleCreateTranscoder_200(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	newNode := Node{
		Name:         "transcoder-new",
		Addr:         "http://127.0.0.1:8093",
		Capabilities: []string{"vaapi", "h264_vaapi"},
		Weight:       2,
	}
	body, _ := json.Marshal(newNode)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/transcoders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code, "body=%s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"name":"transcoder-new"`)
	assert.Contains(t, rec.Body.String(), `"weight":2`)

	// 验证持久化
	yml, err := pool.ConfigYAML()
	require.NoError(t, err)
	assert.Contains(t, yml, "transcoder-new")
}

func TestHandleTranscoderHealth_200(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transcoders/health", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `"nodes"`)
	assert.Contains(t, body, `"transcoder-local-soft"`)
}

func TestHandleStats_200(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transcoders/stats", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `"total_jobs"`)
	assert.Contains(t, body, `"total_failed"`)
}

func TestHandleUpdateTranscoder_200(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	upd := Node{Addr: "http://127.0.0.1:9999", Weight: 5}
	body, _ := json.Marshal(upd)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/transcoders/transcoder-local-soft", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())

	n, ok := pool.Get("transcoder-local-soft")
	require.True(t, ok)
	assert.Equal(t, "http://127.0.0.1:9999", n.Addr)
	assert.Equal(t, 5, n.Weight)
}

func TestHandleDeleteTranscoder_200(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/transcoders/transcoder-local-soft", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	_, ok := pool.Get("transcoder-local-soft")
	assert.False(t, ok)
}

func TestHandleDisableEnable(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	// disable
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/transcoders/transcoder-local-soft/disable", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"disabled"`)

	// enable
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/transcoders/transcoder-local-soft/enable", nil)
	rec = httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"healthy"`)
}

func TestHandleConfig_GetPut(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	// GET
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transcoders/config", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, strings.Contains(rec.Body.String(), "transcoder-local-soft"))

	// PUT
	newCfg := `transcoders:
  - name: remote
    addr: http://remote:8092
    capabilities: [nvenc]
    weight: 1
load_balance: round_robin
`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/transcoders/config", strings.NewReader(newCfg))
	rec = httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"saved`)

	// 验证文件被覆盖
	b, err := os.ReadFile(pool.cfgPath)
	require.NoError(t, err)
	assert.Contains(t, string(b), "remote")
}

func TestHandleReload_200(t *testing.T) {
	pool, cfgPath := newTestPool(t)
	api := NewAdminAPI(pool)

	// 直接改 YAML, 然后调 reload
	newCfg := `transcoders:
  - name: updated
    addr: http://updated:8092
    capabilities: [soft]
    weight: 1
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(newCfg), 0o644))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/transcoders/config/reload", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"updated"`)
}

func TestHandleTranscoderNotFound_404(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transcoders/ghost", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleMethodNotAllowed_405(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/transcoders", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestHandleHealthAll_200(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transcoders/health", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	nodes, ok := resp["nodes"].([]any)
	require.True(t, ok)
	assert.Len(t, nodes, 1)
	node := nodes[0].(map[string]any)
	assert.Equal(t, "transcoder-local-soft", node["name"])
	assert.Equal(t, "healthy", node["status"])
}

func TestHandleHealthAll_Degraded(t *testing.T) {
	pool, _ := newTestPool(t)
	// Add a second node, then disable it to simulate degraded state
	require.NoError(t, pool.Add(Node{Name: "transcoder-remote", Addr: "http://10.0.0.1:8092", Capabilities: []string{"nvenc"}, Weight: 1}))
	require.NoError(t, pool.Disable("transcoder-remote"))

	api := NewAdminAPI(pool)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transcoders/health", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	nodes := resp["nodes"].([]any)
	assert.Len(t, nodes, 2)

	// One healthy, one disabled
	statuses := map[string]string{}
	for _, n := range nodes {
		nn := n.(map[string]any)
		statuses[nn["name"].(string)] = nn["status"].(string)
	}
	assert.Equal(t, "healthy", statuses["transcoder-local-soft"])
	assert.Equal(t, "disabled", statuses["transcoder-remote"])
}

func TestHandleHealthAll_MethodNotAllowed(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/transcoders/health", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestHandleTestAddr_200(t *testing.T) {
	// Mock a transcoder that responds to /api/v1/capabilities
	capSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/capabilities", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"encoder":      "libx264",
			"capabilities": []string{"soft", "libx264"},
		})
	}))
	defer capSrv.Close()

	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	body, _ := json.Marshal(map[string]string{"addr": capSrv.URL})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/transcoders/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["ok"])
	assert.Equal(t, "libx264", resp["encoder"])
	caps := resp["capabilities"].([]any)
	assert.Contains(t, caps, "soft")
}

func TestHandleTestAddr_400_NoAddr(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	// Missing addr field
	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/transcoders/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "addr required")
}

func TestHandleTestAddr_400_BadJSON(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/transcoders/test", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "bad json")
}

func TestHandleTestAddr_MethodNotAllowed(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transcoders/test", nil)
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestHandleTestAddr_ConnectFail(t *testing.T) {
	pool, _ := newTestPool(t)
	api := NewAdminAPI(pool)

	body, _ := json.Marshal(map[string]string{"addr": "http://127.0.0.1:19999"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/transcoders/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["ok"])
	assert.NotEmpty(t, resp["error"])
}

func TestNode_AvgLatencyMs(t *testing.T) {
	n := &Node{}
	assert.Equal(t, int64(0), n.AvgLatencyMs(), "zero jobs should return 0")

	n.recordResult(true, 100*time.Millisecond)
	n.recordResult(true, 300*time.Millisecond)
	assert.Equal(t, int64(200), n.AvgLatencyMs(), "avg of 100ms + 300ms = 200ms")

	n.recordResult(false, 500*time.Millisecond)
	assert.Equal(t, int64(300), n.AvgLatencyMs(), "avg of 100+300+500 = 300ms")
}

func TestTranscoderPool_Probe(t *testing.T) {
	// Mock health endpoint
	healthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthSrv.Close()

	// Mock capabilities endpoint for ProbeAddr
	capSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"encoder":      "libx264",
			"capabilities": []string{"soft"},
		})
	}))
	defer capSrv.Close()

	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "t1", Addr: healthSrv.URL, Weight: 1}))

	// Probe existing node
	err = pool.Probe("t1")
	assert.NoError(t, err)

	n, ok := pool.Get("t1")
	require.True(t, ok)
	assert.Equal(t, StatusHealthy, n.Status)

	// ProbeAddr
	enc, caps, err := pool.ProbeAddr(capSrv.URL)
	assert.NoError(t, err)
	assert.Equal(t, "libx264", enc)
	assert.Contains(t, caps, "soft")

	// Probe non-existent node
	err = pool.Probe("ghost")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTranscoderPool_StartStop(t *testing.T) {
	healthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthSrv.Close()

	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "t1", Addr: healthSrv.URL, Weight: 1}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)
	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)
	pool.Stop()
}

func TestTranscoderPool_ProbeOne_Unhealthy(t *testing.T) {
	// Server that returns non-200
	badSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badSrv.Close()

	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "t1", Addr: badSrv.URL, Weight: 1}))

	// Probe multiple times to exceed threshold
	for i := 0; i < 4; i++ {
		pool.Probe("t1")
	}

	n, ok := pool.Get("t1")
	require.True(t, ok)
	assert.Equal(t, StatusUnhealthy, n.Status)
	assert.NotEmpty(t, n.LastError)
}