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
				{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", ContextLength: 1000000},
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

	// OpenAI-compatible providers (OpenRouter, Groq, Ollama, custom)
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

		// Ollama fallback models
		if strings.Contains(currentBaseURL, "ollama") || strings.Contains(currentBaseURL, "11434") {
			return []dto.ModelInfo{
				{
					ID:            "llama3.2:latest",
					Name:          "Llama 3.2 (3B)",
					Description:   "Modelo compacto e rápido com suporte a tools",
					ContextLength: 128000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "llama3.1:latest",
					Name:          "Llama 3.1 (8B)",
					Description:   "Modelo versátil com suporte a function calling",
					ContextLength: 128000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "qwen2.5-coder:latest",
					Name:          "Qwen 2.5 Coder",
					Description:   "Especializado em código com suporte a tools",
					ContextLength: 32000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "mistral:latest",
					Name:          "Mistral 7B",
					Description:   "Modelo rápido com suporte a function calling",
					ContextLength: 32000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "llama3.1:70b",
					Name:          "Llama 3.1 (70B)",
					Description:   "Modelo grande e poderoso com function calling",
					ContextLength: 128000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
			}
		}

		if strings.Contains(currentBaseURL, "groq.com") {
			// Groq: todos os modelos suportam function calling
			// Lista atualizada com modelos recomendados pela documentação
			return []dto.ModelInfo{
				{
					ID:            "llama-3.3-70b-versatile",
					Name:          "Llama 3.3 70B Versatile",
					Description:   "Modelo versátil recomendado para uso geral (Groq)",
					ContextLength: 128000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "llama-3.1-8b-instant",
					Name:          "Llama 3.1 8B Instant",
					Description:   "Modelo rápido para respostas instantâneas (Groq)",
					ContextLength: 128000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "openai/gpt-oss-120b",
					Name:          "GPT-OSS 120B",
					Description:   "Modelo open source grande com function calling (Groq)",
					ContextLength: 128000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "openai/gpt-oss-20b",
					Name:          "GPT-OSS 20B",
					Description:   "Modelo open source eficiente com function calling (Groq)",
					ContextLength: 128000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
				{
					ID:            "groq/compound",
					Name:          "Groq Compound",
					Description:   "Sistema agêntico com search e code execution (Groq)",
					ContextLength: 128000,
					PricePrompt:   "0",
					PriceComplete: "0",
				},
			}
		}

		// Fallback to hardcoded list if API fails (apenas modelos com suporte a function calling)
		return []dto.ModelInfo{
			{
				ID:            "openai/gpt-4o",
				Name:          "GPT-4o",
				Description:   "Modelo flagship da OpenAI com function calling",
				ContextLength: 128000,
				PricePrompt:   "$2.5/1M",
				PriceComplete: "$10/1M",
			},
			{
				ID:            "openai/gpt-4o-mini",
				Name:          "GPT-4o Mini",
				Description:   "Modelo rápido e econômico com function calling",
				ContextLength: 128000,
				PricePrompt:   "$0.15/1M",
				PriceComplete: "$0.6/1M",
			},
			{
				ID:            "anthropic/claude-3.5-sonnet",
				Name:          "Claude 3.5 Sonnet",
				Description:   "Alta inteligência com function calling",
				ContextLength: 200000,
				PricePrompt:   "$3/1M",
				PriceComplete: "$15/1M",
			},
			{
				ID:            "google/gemini-2.0-flash-001",
				Name:          "Gemini 2.0 Flash",
				Description:   "Modelo rápido do Google com function calling",
				ContextLength: 1000000,
				PricePrompt:   "$0.1/1M",
				PriceComplete: "$0.4/1M",
			},
			{
				ID:            "google/gemini-1.5-pro",
				Name:          "Gemini 1.5 Pro",
				Description:   "Modelo avançado do Google com function calling",
				ContextLength: 2000000,
				PricePrompt:   "$1.25/1M",
				PriceComplete: "$5/1M",
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
