package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ModelInfo struct {
	Name      string `json:"name"`
	Family    string `json:"family"`
	Size      int64  `json:"size"`
	BlobPath  string `json:"blob_path"`
	IsDraft   bool   `json:"is_draft"`
}

type Manifest struct {
	Layers []struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
	} `json:"layers"`
}

// GetOllamaDir returns the path to the Ollama data directory.
func GetOllamaDir() string {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".ollama")
	}
	return filepath.Join(userProfile, ".ollama")
}

// ScanOllamaModels scans the local Ollama directory and maps model tags to GGUF blobs.
func ScanOllamaModels() ([]ModelInfo, error) {
	ollamaDir := GetOllamaDir()
	manifestsDir := filepath.Join(ollamaDir, "models", "manifests")
	blobsDir := filepath.Join(ollamaDir, "models", "blobs")

	if _, err := os.Stat(manifestsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("ollama manifests directory not found at: %s", manifestsDir)
	}

	var models []ModelInfo

	err := filepath.Walk(manifestsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Read manifest JSON
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var manifest Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil
		}

		// Derive model name from relative path
		rel, err := filepath.Rel(manifestsDir, path)
		if err != nil {
			return nil
		}

		parts := strings.Split(filepath.ToSlash(rel), "/")
		var modelName string
		if len(parts) >= 3 && parts[0] == "registry.ollama.ai" && parts[1] == "library" {
			modelName = strings.Join(parts[2:], ":")
		} else {
			modelName = strings.Join(parts, "/")
		}

		// Find model blob layer
		for _, layer := range manifest.Layers {
			if strings.Contains(layer.MediaType, "model") {
				cleanDigest := strings.Replace(layer.Digest, ":", "-", 1)
				blobFile := filepath.Join(blobsDir, cleanDigest)

				m := ModelInfo{
					Name:     modelName,
					Size:     layer.Size,
					BlobPath: blobFile,
					Family:   detectFamily(modelName),
					IsDraft:  detectIsDraft(modelName, layer.Size),
				}
				models = append(models, m)
				break
			}
		}

		return nil
	})

	return models, err
}

func detectFamily(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "qwen2.5-coder") || strings.Contains(lower, "qwen2.5") {
		return "qwen2.5"
	}
	if strings.Contains(lower, "llama-3") || strings.Contains(lower, "llama3") {
		return "llama3"
	}
	if strings.Contains(lower, "deepseek") {
		return "deepseek"
	}
	return "unknown"
}

func detectIsDraft(name string, size int64) bool {
	lower := strings.ToLower(name)
	// Under 1.5GB or explicitly named 0.5b/1b
	if size < 1500*1024*1024 || strings.Contains(lower, "0.5b") || strings.Contains(lower, "1b") || strings.Contains(lower, "1.5b") {
		return true
	}
	return false
}
