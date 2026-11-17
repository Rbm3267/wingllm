package discovery

import (
	"fmt"
	"strings"
)

type PairedModels struct {
	Target       ModelInfo `json:"target"`
	Draft        ModelInfo `json:"draft"`
	RecommendedN int       `json:"recommended_n"`
	Reason       string    `json:"reason"`
}

// FindBestPair finds the most suitable draft model for a requested target model.
func FindBestPair(targetName string, available []ModelInfo) (*PairedModels, error) {
	var target *ModelInfo
	for i := range available {
		if strings.EqualFold(available[i].Name, targetName) || strings.Contains(strings.ToLower(available[i].Name), strings.ToLower(targetName)) {
			target = &available[i]
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("target model '%s' not found locally in Ollama", targetName)
	}

	// Search for compatible draft model in the same architectural family
	var bestDraft *ModelInfo
	for i := range available {
		m := &available[i]
		if !m.IsDraft {
			continue
		}

		// Exact family alignment
		if target.Family != "unknown" && m.Family == target.Family {
			// Prefer 0.5B for low VRAM cards (< 600MB)
			if strings.Contains(strings.ToLower(m.Name), "0.5b") {
				bestDraft = m
				break
			}
			if bestDraft == nil || m.Size < bestDraft.Size {
				bestDraft = m
			}
		}
	}

	if bestDraft == nil {
		return nil, fmt.Errorf("no compatible draft model found for %s. Suggest pulling 'ollama pull qwen2.5-coder:0.5b'", target.Name)
	}

	return &PairedModels{
		Target:       *target,
		Draft:        *bestDraft,
		RecommendedN: 3,
		Reason: fmt.Sprintf("Paired %s (%.1f GB Target) with %s (%.0f MB Draft) via shared %s vocabulary",
			target.Name, float64(target.Size)/(1024*1024*1024),
			bestDraft.Name, float64(bestDraft.Size)/(1024*1024),
			target.Family,
		),
	}, nil
}
