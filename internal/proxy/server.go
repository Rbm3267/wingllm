package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"runtime"
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
	fmt.Printf(" WingLLM Web Dashboard:            http://localhost:%d/\n", s.Port)
	fmt.Printf(" Opening web chat in your browser...\n\n")

	// Cross-platform auto-open browser
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(fmt.Sprintf("http://localhost:%d/", s.Port))
	}()

	return http.ListenAndServe(addr, mux)
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	case "linux":
		_ = exec.Command("xdg-open", url).Start()
	}
}
