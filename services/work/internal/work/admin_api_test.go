package work

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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