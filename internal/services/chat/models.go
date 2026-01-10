package chat

import (
	"excel-ai/internal/dto"
)

func (s *Service) GetAvailableModels(apiKey, baseURL string) []dto.ModelInfo {
	// Apenas Z.AI é suportado
	return []dto.ModelInfo{
		{ID: "glm-4.7", Name: "GLM-4.7", Description: "Latest flagship model - optimized for coding", ContextLength: 128000, PricePrompt: "¥2.5/1M", PriceComplete: "¥10/1M"},
		{ID: "glm-4.6v", Name: "GLM-4.6V", Description: "Vision model with multimodal capabilities", ContextLength: 128000, PricePrompt: "¥2/1M", PriceComplete: "¥8/1M"},
		{ID: "glm-4.6", Name: "GLM-4.6", Description: "Balanced model with native function calling", ContextLength: 128000, PricePrompt: "¥1.5/1M", PriceComplete: "¥6/1M"},
		{ID: "glm-4.5", Name: "GLM-4.5", Description: "Cost-effective model with function calling", ContextLength: 128000, PricePrompt: "¥1/1M", PriceComplete: "¥4/1M"},
		{ID: "glm-4.5-air", Name: "GLM-4.5 Air", Description: "Lightweight model for fast responses", ContextLength: 128000, PricePrompt: "¥0.5/1M", PriceComplete: "¥2/1M"},
	}
}
