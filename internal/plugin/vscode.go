package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ContinueConfig struct {
	Models []map[string]interface{} `json:"models"`
	Rest   map[string]interface{}   `json:"-"`
}

// ConnectVSCode connects WingLLM to VS Code's Continue extension by updating ~/.continue/config.json.
func ConnectVSCode(modelName string) error {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		home, _ := os.UserHomeDir()
		userProfile = home
	}

	continueDir := filepath.Join(userProfile, ".continue")
	if err := os.MkdirAll(continueDir, 0755); err != nil {
		return fmt.Errorf("failed to create .continue directory: %w", err)
	}

	configFile := filepath.Join(continueDir, "config.json")

	wingllmModel := map[string]interface{}{
		"title":    fmt.Sprintf("WingLLM (%s - 2x Turbo)", modelName),
		"provider": "openai",
		"model":    modelName,
		"apiBase":  "http://localhost:8080/v1",
	}

	var root map[string]interface{}

	if data, err := os.ReadFile(configFile); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &root)
	}

	if root == nil {
		root = make(map[string]interface{})
	}

	// Update models array
	var models []interface{}
	if existing, ok := root["models"].([]interface{}); ok {
		models = existing
	}

	// Check if already present
	exists := false
	for _, m := range models {
		if mObj, ok := m.(map[string]interface{}); ok {
			if mObj["apiBase"] == "http://localhost:8080/v1" {
				exists = true
				mObj["title"] = fmt.Sprintf("WingLLM (%s - 2x Turbo)", modelName)
				mObj["model"] = modelName
				break
			}
		}
	}

	if !exists {
		// Prepend to top of models list so it appears first in the dropdown
		models = append([]interface{}{wingllmModel}, models...)
	}

	root["models"] = models

	// Also configure Tab Autocomplete if not present
	if _, ok := root["tabAutocompleteModel"]; !ok {
		root["tabAutocompleteModel"] = map[string]interface{}{
			"title":    "WingLLM Autocomplete",
			"provider": "openai",
			"model":    modelName,
			"apiBase":  "http://localhost:8080/v1",
		}
	}

	output, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write config.json: %w", err)
	}

	fmt.Printf("\n Successfully connected to VS Code (Continue Extension)!\n")
	fmt.Printf(" Config written to: %s\n", configFile)
	fmt.Printf(" Model in dropdown: WingLLM (%s - 2x Turbo)\n\n", modelName)
	return nil
}
