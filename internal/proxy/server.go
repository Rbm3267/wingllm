package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/Rbm3267/wingllm/internal/discovery"
	"github.com/Rbm3267/wingllm/internal/vram"
)

type Server struct {
	Port         int
	EnginePort   int
	Pair         *discovery.PairedModels
	ReverseProxy *httputil.ReverseProxy
}

func NewServer(port int, enginePort int, pair *discovery.PairedModels) (*Server, error) {
	targetURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", enginePort))
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return &Server{
		Port:         port,
		EnginePort:   enginePort,
		Pair:         pair,
		ReverseProxy: proxy,
	}, nil
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// OpenAI-compatible /v1/models endpoint
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"object": "list",
			"data": []map[string]interface{}{
				{
					"id":       s.Pair.Target.Name,
					"object":   "model",
					"created":  time.Now().Unix(),
					"owned_by": "wingllm",
					"metadata": map[string]interface{}{
						"draft_model": s.Pair.Draft.Name,
						"family":      s.Pair.Target.Family,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Telemetry endpoint
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vramStat, _ := vram.GetVRAMStatus()
		resp := map[string]interface{}{
			"status":   "running",
			"target":   s.Pair.Target.Name,
			"draft":    s.Pair.Draft.Name,
			"pairing":  s.Pair.Reason,
			"vram":     vramStat,
			"endpoint": fmt.Sprintf("http://localhost:%d/v1", s.Port),
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Web Dashboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// Proxy all completion/chat calls to engine
			s.ReverseProxy.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(DashboardHTML(s.Pair)))
	})

	addr := fmt.Sprintf(":%d", s.Port)
	fmt.Printf("\n WingLLM OpenAI API available at: http://localhost:%d/v1\n", s.Port)
	fmt.Printf(" WingLLM Web Dashboard:            http://localhost:%d/\n\n", s.Port)
	return http.ListenAndServe(addr, mux)
}

func DashboardHTML(pair *discovery.PairedModels) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>WingLLM | Local Speculative Acceleration</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; padding: 40px; margin: 0; }
    .card { background: #1e293b; border-radius: 12px; padding: 24px; max-width: 680px; margin: 0 auto; box-shadow: 0 4px 20px rgba(0,0,0,0.4); border: 1px solid #334155; }
    h1 { font-size: 24px; margin-top: 0; color: #38bdf8; display: flex; align-items: center; gap: 8px; }
    .badge { background: #0369a1; color: #e0f2fe; padding: 4px 10px; border-radius: 9999px; font-size: 12px; font-weight: 600; }
    .stat-row { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid #334155; }
    .stat-label { color: #94a3b8; font-size: 14px; }
    .stat-value { font-weight: 600; color: #f1f5f9; font-size: 14px; }
    .highlight { color: #4ade80; }
    .code-box { background: #090d16; padding: 12px; border-radius: 6px; font-family: monospace; font-size: 13px; color: #cbd5e1; margin-top: 16px; overflow-x: auto; }
  </style>
</head>
<body>
  <div class="card">
    <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">
      <svg width="32" height="32" viewBox="0 0 120 120" fill="none" style="filter: drop-shadow(0 2px 8px rgba(56, 189, 248, 0.4));">
        <path d="M60 16 C60.8 16 62 18.5 62.5 24 L63.5 44 L98 64 C99.5 65 99.8 66.8 98.8 67.5 L94.5 70 L63.5 61.5 L63.5 86 L74 95 C74.8 95.8 74.8 97 73.8 97.5 L70.5 98.5 L60 94 L49.5 98.5 L46.2 97.5 C45.2 97 45.2 95.8 46 95 L56.5 86 L56.5 61.5 L25.5 70 L21.2 67.5 C20.2 66.8 20.5 65 22 64 L56.5 44 L57.5 24 C58 18.5 59.2 16 60 16 Z" fill="none" stroke="#38bdf8" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
      <h1 style="margin: 0; font-size: 24px; color: #38bdf8;">WingLLM</h1>
      <span class="badge">ACTIVE (2x Turbo)</span>
    </div>
    <p style="color: #94a3b8; font-size: 14px;">Heterogeneous Speculative Decoding Co-Processor for Low-VRAM GPUs</p>
    
    <div class="stat-row">
      <span class="stat-label">Target Model (CPU System RAM)</span>
      <span class="stat-value">%s</span>
    </div>
    <div class="stat-row">
      <span class="stat-label">Wingman Draft Model (2GB GPU VRAM)</span>
      <span class="stat-value highlight">%s</span>
    </div>
    <div class="stat-row">
      <span class="stat-label">Pairing Strategy</span>
      <span class="stat-value">%s</span>
    </div>
    <div class="stat-row">
      <span class="stat-label">OpenAI Endpoint</span>
      <span class="stat-value">http://localhost:8080/v1</span>
    </div>

    <div style="margin-top: 24px;">
      <h3 style="font-size: 14px; color: #94a3b8; text-transform: uppercase;">Connect to VS Code / Continue / Cline</h3>
      <div class="code-box">
{<br>
&nbsp;&nbsp;"model": "%s",<br>
&nbsp;&nbsp;"apiBase": "http://localhost:8080/v1",<br>
&nbsp;&nbsp;"provider": "openai"<br>
}
      </div>
    </div>
  </div>
</body>
</html>`, pair.Target.Name, pair.Draft.Name, pair.Reason, pair.Target.Name)
}
