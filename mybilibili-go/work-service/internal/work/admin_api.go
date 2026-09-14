package work

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// AdminAPI 提供 transcoder 池的 HTTP 管理端点。
// 路由前缀: /api/v1/admin/transcoders
// 鉴权: 由 traefik / 上游网关处理; 容器内只做内部管理。
type AdminAPI struct {
	pool *TranscoderPool
	mux  *http.ServeMux
}

func NewAdminAPI(pool *TranscoderPool) *AdminAPI {
	a := &AdminAPI{
		pool: pool,
		mux:  http.NewServeMux(),
	}
	a.register()
	return a
}

func (a *AdminAPI) Handler() http.Handler { return a.mux }

func (a *AdminAPI) register() {
	a.mux.HandleFunc("/api/v1/admin/transcoders", a.handleRoot)
	a.mux.HandleFunc("/api/v1/admin/transcoders/", a.handleSub)
	a.mux.HandleFunc("/api/v1/admin/transcoders/health", a.handleHealthAll)
	a.mux.HandleFunc("/api/v1/admin/transcoders/stats", a.handleStats)
	a.mux.HandleFunc("/api/v1/admin/transcoders/config", a.handleConfig)
	a.mux.HandleFunc("/api/v1/admin/transcoders/config/reload", a.handleConfigReload)
	a.mux.HandleFunc("/api/v1/admin/transcoders/test", a.handleTestAddr)
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

// --- 路由分发 ---

// handleSub 处理 /api/v1/admin/transcoders/{name}[/action]
func (a *AdminAPI) handleSub(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/transcoders/")
	parts := strings.SplitN(path, "/", 2)
	name := parts[0]
	if name == "" {
		writeErr(w, 400, "name required")
		return
	}
	action := ""
	if len(parts) == 2 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodGet:
		if action != "" {
			writeErr(w, 404, "no such action")
			return
		}
		n, ok := a.pool.Get(name)
		if !ok {
			writeErr(w, 404, "node not found")
			return
		}
		writeJSON(w, 200, n)
	case http.MethodPut:
		if action != "" {
			writeErr(w, 404, "no such action")
			return
		}
		var n Node
		if err := readJSON(r, &n); err != nil {
			writeErr(w, 400, "bad json: "+err.Error())
			return
		}
		if err := a.pool.Update(name, n); err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		n2, _ := a.pool.Get(name)
		writeJSON(w, 200, n2)
	case http.MethodDelete:
		if action != "" {
			writeErr(w, 404, "no such action")
			return
		}
		if err := a.pool.Delete(name); err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, map[string]string{"status": "deleted"})
	case http.MethodPost:
		switch action {
		case "probe":
			if err := a.pool.Probe(name); err != nil {
				writeErr(w, 400, err.Error())
				return
			}
			n, _ := a.pool.Get(name)
			writeJSON(w, 200, n)
		case "enable":
			if err := a.pool.Enable(name); err != nil {
				writeErr(w, 400, err.Error())
				return
			}
			n, _ := a.pool.Get(name)
			writeJSON(w, 200, n)
		case "disable":
			if err := a.pool.Disable(name); err != nil {
				writeErr(w, 400, err.Error())
				return
			}
			n, _ := a.pool.Get(name)
			writeJSON(w, 200, n)
		default:
			writeErr(w, 404, "unknown action: "+action)
		}
	default:
		writeErr(w, 405, "method not allowed")
	}
}

// handleRoot 处理 /api/v1/admin/transcoders (list / create)
func (a *AdminAPI) handleRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		nodes := a.pool.List()
		writeJSON(w, 200, map[string]any{"nodes": nodes, "total": len(nodes)})
	case http.MethodPost:
		var n Node
		if err := readJSON(r, &n); err != nil {
			writeErr(w, 400, "bad json: "+err.Error())
			return
		}
		if err := a.pool.Add(n); err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		added, _ := a.pool.Get(n.Name)
		writeJSON(w, 201, added)
	default:
		writeErr(w, 405, "method not allowed")
	}
}

// handleHealthAll 全部节点健康快照。
func (a *AdminAPI) handleHealthAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	nodes := a.pool.List()
	writeJSON(w, 200, map[string]any{"nodes": nodes})
}

// handleStats 统计聚合 (跟 List 同源, 这里只是语义别名)。
func (a *AdminAPI) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	nodes := a.pool.List()
	var totalJobs, totalFailed int64
	for _, n := range nodes {
		totalJobs += n.TotalJobs
		totalFailed += n.FailedJobs
	}
	writeJSON(w, 200, map[string]any{
		"nodes":        nodes,
		"total_jobs":   totalJobs,
		"total_failed": totalFailed,
	})
}

// handleConfig 配置读写 (YAML)。
func (a *AdminAPI) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		yml, err := a.pool.ConfigYAML()
		if err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		_, _ = w.Write([]byte(yml))
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeErr(w, 400, "read body: "+err.Error())
			return
		}
		defer r.Body.Close()
		// 直接覆盖配置文件; 下次 SIGHUP 或调用 reload 时应用
		if err := writeFileAtomic(a.pool.cfgPath, body); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, map[string]string{"status": "saved, call /config/reload to apply"})
	default:
		writeErr(w, 405, "method not allowed")
	}
}

// handleConfigReload 重读 YAML。
func (a *AdminAPI) handleConfigReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	if err := a.pool.Reload(); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"nodes": a.pool.List()})
}

// handleTestAddr 测一个未保存的地址连通性 + 拉 capabilities (新增节点前的"测试"按钮)。
func (a *AdminAPI) handleTestAddr(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var req struct {
		Addr string `json:"addr"`
	}
	if err := readJSON(r, &req); err != nil {
		writeErr(w, 400, "bad json: "+err.Error())
		return
	}
	if req.Addr == "" {
		writeErr(w, 400, "addr required")
		return
	}
	encoder, caps, err := a.pool.ProbeAddr(req.Addr)
	if err != nil {
		writeJSON(w, 200, map[string]any{
			"addr":   req.Addr,
			"ok":     false,
			"error":  err.Error(),
		})
		return
	}
	writeJSON(w, 200, map[string]any{
		"addr":         req.Addr,
		"ok":           true,
		"encoder":      encoder,
		"capabilities": caps,
	})
}
