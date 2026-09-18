package work

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestAdminAPI_Root_GET(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("GET", "/api/v1/admin/transcoders", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAdminAPI_Root_POST_BadJSON(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminAPI_Root_POST_Success(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	body, _ := json.Marshal(Node{Name: "test-node", Addr: "10.0.0.1:9090"})
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 201 {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestAdminAPI_Root_MethodNotAllowed(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("DELETE", "/api/v1/admin/transcoders", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_GET_NotFound(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("GET", "/api/v1/admin/transcoders/nonexistent", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_GET_EmptyName(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("GET", "/api/v1/admin/transcoders/", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_PUT_BadJSON(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("PUT", "/api/v1/admin/transcoders/someone", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_DELETE_Success(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	body, _ := json.Marshal(Node{Name: "to-delete", Addr: "10.0.0.1:9090"})
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)

	req = httptest.NewRequest("DELETE", "/api/v1/admin/transcoders/to-delete", nil)
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_POST_Probe_NotFound(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/nonexistent/probe", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_POST_Enable_NotFound(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/nonexistent/enable", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_POST_Disable_NotFound(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/nonexistent/disable", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_POST_UnknownAction(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/someone/unknown", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_GET_WithAction(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("GET", "/api/v1/admin/transcoders/someone/probe", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_PUT_WithAction(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("PUT", "/api/v1/admin/transcoders/someone/probe", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAdminAPI_Sub_DELETE_WithAction(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("DELETE", "/api/v1/admin/transcoders/someone/probe", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAdminAPI_HealthAll_MethodNotAllowed(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/health", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestAdminAPI_Stats_MethodNotAllowed(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/stats", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestAdminAPI_Config_GET(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("GET", "/api/v1/admin/transcoders/config", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAdminAPI_Config_MethodNotAllowed(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("DELETE", "/api/v1/admin/transcoders/config", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestAdminAPI_ConfigReload_MethodNotAllowed(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("GET", "/api/v1/admin/transcoders/config/reload", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestAdminAPI_TestAddr_MethodNotAllowed(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("GET", "/api/v1/admin/transcoders/test", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestAdminAPI_TestAddr_EmptyAddr(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	body, _ := json.Marshal(map[string]string{"addr": ""})
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/test", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminAPI_TestAddr_BadJSON(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/test", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminAPI_TestAddr_Unreachable(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	body, _ := json.Marshal(map[string]string{"addr": "http://127.0.0.1:1"})
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/test", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if result["ok"] != false {
		t.Errorf("expected ok=false for unreachable addr")
	}
}

func TestAdminAPI_HealthAll_WithNodes(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("GET", "/api/v1/admin/transcoders/health", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAdminAPI_Stats_WithNodes(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("GET", "/api/v1/admin/transcoders/stats", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAdminAPI_ConfigReload_Success(t *testing.T) {
	pool, _ := newTestPool(t)
	a := NewAdminAPI(pool)
	req := httptest.NewRequest("POST", "/api/v1/admin/transcoders/config/reload", nil)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
