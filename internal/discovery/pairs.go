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
		defaultDraft := GetDefaultDraftName(target.Family)
		return nil, fmt.Errorf("no compatible draft model found for %s. Run 'ollama pull %s'", target.Name, defaultDraft)
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

// PairWithCustomDraft manually pairs a target model with an explicit draft model chosen by the user.
func PairWithCustomDraft(targetName string, draftName string, available []ModelInfo) (*PairedModels, error) {
	var target *ModelInfo
	var draft *ModelInfo

	for i := range available {
		if strings.EqualFold(available[i].Name, targetName) || strings.Contains(strings.ToLower(available[i].Name), strings.ToLower(targetName)) {
			target = &available[i]
		}
		if strings.EqualFold(available[i].Name, draftName) || strings.Contains(strings.ToLower(available[i].Name), strings.ToLower(draftName)) {
			draft = &available[i]
		}
	}

	if target == nil {
		return nil, fmt.Errorf("target model '%s' not found locally in Ollama", targetName)
	}
	if draft == nil {
		return nil, fmt.Errorf("custom draft model '%s' not found locally in Ollama", draftName)
	}

	return &PairedModels{
		Target:       *target,
		Draft:        *draft,
		RecommendedN: 3,
		Reason: fmt.Sprintf("User specified custom pair: %s (Target) + %s (Draft)",
			target.Name, draft.Name,
		),
	}, nil
}
func GetDefaultDraftName(family string) string {
	switch family {
	case "qwen2.5":
		return "qwen2.5-coder:0.5b"
	case "llama3":
		return "llama3.2:1b"
	default:
		return "qwen2.5-coder:0.5b"
	}
}
