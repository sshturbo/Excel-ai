package chat

import (
	"strings"

	"excel-ai/internal/dto"
	"excel-ai/pkg/ai"
)

func (s *Service) GetAvailableModels(apiKey, baseURL string) []dto.ModelInfo {
	// Check if this is a Google/Gemini API URL
	if strings.Contains(baseURL, "generativelanguage.googleapis.com") {
		geminiClient := ai.NewGeminiClient(apiKey, "")
		if baseURL != "" {
			geminiClient.SetBaseURL(baseURL)
		}
		models, err := geminiClient.GetAvailableModels()
		if err != nil {
			// Return fallback Gemini models
			return []dto.ModelInfo{
				{ID: "gemini-2.0-flash-exp", Name: "Gemini 2.0 Flash (Experimental)", ContextLength: 1000000},
				{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", ContextLength: 1000000},
				{ID: "gemini-1.5-flash-8b", Name: "Gemini 1.5 Flash 8B", ContextLength: 1000000},
				{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", ContextLength: 2000000},
			}
		}
		var result []dto.ModelInfo
		for _, m := range models {
			result = append(result, dto.ModelInfo{
				ID:            m.ID,
				Name:          m.Name,
				ContextLength: m.ContextLength,
			})
		}
		return result
	}

	// OpenAI-compatible providers (OpenRouter, Groq, custom)
	var client *ai.Client
	// Se uma URL base for fornecida, cria um cliente temporário para buscar os modelos
	if baseURL != "" {
		client = ai.NewClient(apiKey, "", baseURL)
	} else {
		// Caso contrário usa o cliente configurado
		client = s.client
	}

	models, err := client.GetAvailableModels()
	if err != nil {
		// Check BaseURL to determine fallback
		currentBaseURL := client.GetBaseURL()
		if strings.Contains(currentBaseURL, "groq.com") {
			return []dto.ModelInfo{
				{
					ID:            "llama3-8b-8192",
					Name:          "Llama 3 8B",
					Description:   "Modelo rápido e eficiente da Meta (Groq)",
					ContextLength: 8192,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "llama3-70b-8192",
					Name:          "Llama 3 70B",
					Description:   "Modelo de alta capacidade da Meta (Groq)",
					ContextLength: 8192,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "mixtral-8x7b-32768",
					Name:          "Mixtral 8x7B",
					Description:   "Modelo MoE de alta performance (Groq)",
					ContextLength: 32768,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "gemma-7b-it",
					Name:          "Gemma 7B IT",
					Description:   "Modelo do Google (Groq)",
					ContextLength: 8192,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
			}
		}

		// Fallback to hardcoded list if API fails
		return []dto.ModelInfo{
			{
				ID:            "google/gemini-2.0-flash-exp:free",
				Name:          "Gemini 2.0 Flash (Free)",
				Description:   "Modelo experimental rápido e gratuito do Google",
				ContextLength: 1000000,
				PricePrompt:   "0",
				PriceComplete: "0",
			},
			{
				ID:            "google/gemini-exp-1206:free",
				Name:          "Gemini Exp 1206 (Free)",
				Description:   "Modelo experimental atualizado",
				ContextLength: 1000000,
				PricePrompt:   "0",
				PriceComplete: "0",
			},
			{
				ID:            "meta-llama/llama-3.2-90b-vision-instruct:free",
				Name:          "Llama 3.2 90B (Free)",
				Description:   "Modelo open source poderoso da Meta",
				ContextLength: 128000,
				PricePrompt:   "0",
				PriceComplete: "0",
			},
			{
				ID:            "microsoft/phi-3-medium-128k-instruct:free",
				Name:          "Phi-3 Medium (Free)",
				Description:   "Modelo eficiente da Microsoft",
				ContextLength: 128000,
				PricePrompt:   "0",
				PriceComplete: "0",
			},
			{
				ID:            "anthropic/claude-3.5-sonnet",
				Name:          "Claude 3.5 Sonnet",
				Description:   "Alta inteligência e capacidade de codificação",
				ContextLength: 200000,
				PricePrompt:   "$3/1M",
				PriceComplete: "$15/1M",
			},
			{
				ID:            "openai/gpt-4o",
				Name:          "GPT-4o",
				Description:   "Modelo flagship da OpenAI",
				ContextLength: 128000,
				PricePrompt:   "$5/1M",
				PriceComplete: "$15/1M",
			},
		}
	}

	var result []dto.ModelInfo
	for _, m := range models {
		// Map pricing from strings to strings (could use domain.ModelPricing here too if updated)
		result = append(result, dto.ModelInfo{
			ID:            m.ID,
			Name:          m.Name,
			Description:   m.Description,
			ContextLength: m.ContextLength,
			PricePrompt:   m.Pricing.Prompt,
			PriceComplete: m.Pricing.Completion,
		})
	}

	return result
}
