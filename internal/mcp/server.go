package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Rbm3267/wingllm/internal/discovery"
	"github.com/Rbm3267/wingllm/internal/vram"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// RunMCPServer starts a Model Context Protocol stdio server.
func RunMCPServer(enginePort int) error {
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			var req JSONRPCRequest
			if jsonErr := json.Unmarshal([]byte(line), &req); jsonErr == nil {
				handleRequest(&req, enginePort)
			}
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func handleRequest(req *JSONRPCRequest, enginePort int) {
	switch req.Method {
	case "initialize":
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "wingllm-mcp",
					"version": "0.1.0",
				},
				"capabilities": map[string]interface{}{
					"tools": map[string]bool{
						"listChanged": false,
					},
				},
			},
		}
		sendResponse(resp)

	case "notifications/initialized":
		// No response required for notification

	case "tools/list":
		tools := []Tool{
			{
				Name:        "wingllm_vram",
				Description: "Query local GPU VRAM status, used memory, and available safety budget for local AI models.",
				InputSchema: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			{
				Name:        "wingllm_models",
				Description: "List local Ollama models and find auto-recommended 2x turbo draft pairs.",
				InputSchema: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			{
				Name:        "wingllm_status",
				Description: "Check if WingLLM speculative turbo engine is currently active and query current throughput.",
				InputSchema: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		}

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": tools,
			},
		}
		sendResponse(resp)

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		_ = json.Unmarshal(req.Params, &params)

		var textResult string
		switch params.Name {
		case "wingllm_vram":
			status, err := vram.GetVRAMStatus()
			if err != nil {
				textResult = fmt.Sprintf("Error checking VRAM: %v", err)
			} else {
				data, _ := json.MarshalIndent(status, "", "  ")
				textResult = string(data)
			}

		case "wingllm_models":
			models, err := discovery.ScanOllamaModels()
			if err != nil {
				textResult = fmt.Sprintf("Error scanning models: %v", err)
			} else {
				var summaries []map[string]interface{}
				for _, m := range models {
					summaries = append(summaries, map[string]interface{}{
						"name":    m.Name,
						"size_gb": fmt.Sprintf("%.2f GB", float64(m.Size)/(1024*1024*1024)),
						"family":  m.Family,
						"is_draft": m.IsDraft,
					})
				}
				data, _ := json.MarshalIndent(summaries, "", "  ")
				textResult = string(data)
			}

		case "wingllm_status":
			healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", enginePort)
			resp, err := http.Get(healthURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				textResult = fmt.Sprintf("WingLLM Turbo is ACTIVE and healthy on internal port %d (OpenAI endpoint: http://localhost:8080/v1)", enginePort)
				resp.Body.Close()
			} else {
				textResult = "WingLLM Turbo engine is currently idle/stopped."
			}

		default:
			textResult = fmt.Sprintf("Unknown tool '%s'", params.Name)
		}

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]string{
					{
						"type": "text",
						"text": textResult,
					},
				},
			},
		}
		sendResponse(resp)

	default:
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
			},
		}
		sendResponse(resp)
	}
}

func sendResponse(resp JSONRPCResponse) {
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}
