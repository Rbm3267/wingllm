package engine

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Rbm3267/wingllm/internal/discovery"
)

type Config struct {
	Port         int
	CtxSize      int
	DraftNMax    int
	Threads      int
	BackendPath  string
	UseGPUForDraft bool
}

type Supervisor struct {
	Config  Config
	Pair    *discovery.PairedModels
	Cmd     *exec.Cmd
	Port    int
}

func DefaultConfig() Config {
	return Config{
		Port:         8081, // llama-server listens on internal port, proxy sits on 8080
		CtxSize:      2048,
		DraftNMax:    3,
		Threads:      6,
		UseGPUForDraft: true,
	}
}

// FindLlamaServer checks standard Ollama paths and PATH for llama-server.exe.
func FindLlamaServer() (string, error) {
	userProfile := os.Getenv("USERPROFILE")
	ollamaServer := filepath.Join(userProfile, "AppData", "Local", "Programs", "Ollama", "lib", "ollama", "llama-server.exe")
	if _, err := os.Stat(ollamaServer); err == nil {
		return ollamaServer, nil
	}

	if p, err := exec.LookPath("llama-server"); err == nil {
		return p, nil
	}

	return "", fmt.Errorf("llama-server.exe not found in Ollama directory or system PATH")
}

func NewSupervisor(pair *discovery.PairedModels, cfg Config) *Supervisor {
	return &Supervisor{
		Config: cfg,
		Pair:   pair,
		Port:   cfg.Port,
	}
}

// Start launches the heterogeneous speculative llama-server process.
func (s *Supervisor) Start() error {
	binPath, err := FindLlamaServer()
	if err != nil {
		return err
	}

	args := []string{
		"--port", fmt.Sprintf("%d", s.Port),
		"-c", fmt.Sprintf("%d", s.Config.CtxSize),
		"-t", fmt.Sprintf("%d", s.Config.Threads),
		"--no-warmup",
		"-m", s.Pair.Target.BlobPath,
		"-ngl", "0", // Main target model stays 100% on CPU
		"-md", s.Pair.Draft.BlobPath,
		"--spec-draft-n-max", fmt.Sprintf("%d", s.Config.DraftNMax),
		"--spec-draft-n-min", "1",
	}

	if s.Config.UseGPUForDraft {
		args = append(args, "-ngld", "999") // Offload draft layers to GPU
	}

	cmd := exec.Command(binPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Ensure Ollama library paths are present in PATH for DLL loading
	ollamaLib := filepath.Dir(binPath)
	cudaLib := filepath.Join(ollamaLib, "cuda_v12")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s;%s;%s", cudaLib, ollamaLib, os.Getenv("PATH")),
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start llama-server: %w", err)
	}

	s.Cmd = cmd

	// Wait up to 30s for the server to be ready
	ready := false
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", s.Port)
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		resp, err := http.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			ready = true
			resp.Body.Close()
			break
		}
	}

	if !ready {
		s.Stop()
		return fmt.Errorf("engine server did not become healthy within 30 seconds")
	}

	return nil
}

// Stop cleanly terminates the engine process.
func (s *Supervisor) Stop() error {
	if s.Cmd != nil && s.Cmd.Process != nil {
		return s.Cmd.Process.Kill()
	}
	return nil
}
