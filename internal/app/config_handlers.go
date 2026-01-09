package app

import (
	"fmt"

	"excel-ai/internal/dto"
	"excel-ai/pkg/storage"
)

// SetAPIKey configura a chave da API
func (a *App) SetAPIKey(apiKey string) error {
	a.chatService.SetAPIKey(apiKey)
	// Save to storage
	if a.storage != nil {
		cfg, _ := a.storage.LoadConfig()
		if cfg == nil {
			cfg = &storage.Config{}
		}
		if cfg.ProviderConfigs == nil {
			cfg.ProviderConfigs = make(map[string]storage.ProviderConfig)
		}
		cfg.APIKey = apiKey

		// Também salvar no mapa do provedor atual
		provider := cfg.Provider
		if provider == "" {
			provider = "openrouter"
		}
		providerCfg := cfg.ProviderConfigs[provider]
		providerCfg.APIKey = apiKey
		cfg.ProviderConfigs[provider] = providerCfg

		return a.storage.SaveConfig(cfg)
	}
	return nil
}

// SetModel configura o modelo da IA
func (a *App) SetModel(model string) error {
	a.chatService.SetModel(model)
	// Save to storage
	if a.storage != nil {
		cfg, _ := a.storage.LoadConfig()
		if cfg == nil {
			cfg = &storage.Config{}
		}
		if cfg.ProviderConfigs == nil {
			cfg.ProviderConfigs = make(map[string]storage.ProviderConfig)
		}
		cfg.Model = model

		// Também salvar no mapa do provedor atual
		provider := cfg.Provider
		if provider == "" {
			provider = "openrouter"
		}
		providerCfg := cfg.ProviderConfigs[provider]
		providerCfg.Model = model
		cfg.ProviderConfigs[provider] = providerCfg

		return a.storage.SaveConfig(cfg)
	}
	return nil
}

// SetToolModel configura o modelo secundário da IA
func (a *App) SetToolModel(toolModel string) error {
	// a.chatService.SetToolModel(toolModel) // Implementar se necessário no service
	// Save to storage
	if a.storage != nil {
		cfg, _ := a.storage.LoadConfig()
		if cfg == nil {
			cfg = &storage.Config{}
		}
		if cfg.ProviderConfigs == nil {
			cfg.ProviderConfigs = make(map[string]storage.ProviderConfig)
		}
		cfg.ToolModel = toolModel

		// Também salvar no mapa do provedor atual
		provider := cfg.Provider
		if provider == "" {
			provider = "openrouter"
		}
		providerCfg := cfg.ProviderConfigs[provider]
		providerCfg.ToolModel = toolModel
		cfg.ProviderConfigs[provider] = providerCfg

		return a.storage.SaveConfig(cfg)
	}
	return nil
}
func (a *App) GetAvailableModels(apiKey, baseURL string) ([]dto.ModelInfo, error) {
	return a.chatService.GetAvailableModels(apiKey, baseURL), nil
}

// GetSavedConfig retorna configurações salvas
func (a *App) GetSavedConfig() (*storage.Config, error) {
	if a.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}
	return a.storage.LoadConfig()
}

// UpdateConfig atualiza configurações
func (a *App) UpdateConfig(maxRowsContext, maxContextChars, maxRowsPreview int, includeHeaders bool, detailLevel, customPrompt, language, provider, toolModel, baseUrl string) error {
	if a.storage == nil {
		return fmt.Errorf("storage não disponível")
	}
	cfg, _ := a.storage.LoadConfig()
	if cfg == nil {
		cfg = &storage.Config{}
	}
	if cfg.ProviderConfigs == nil {
		cfg.ProviderConfigs = make(map[string]storage.ProviderConfig)
	}

	// Atualizar configurações gerais
	cfg.MaxRowsContext = maxRowsContext
	cfg.MaxContextChars = maxContextChars
	cfg.MaxRowsPreview = maxRowsPreview
	cfg.IncludeHeaders = includeHeaders
	cfg.DetailLevel = detailLevel
	cfg.CustomPrompt = customPrompt
	cfg.Language = language
	cfg.Provider = provider
	cfg.BaseURL = baseUrl
	cfg.ToolModel = toolModel

	// Salvar configurações do provedor atual no mapa de providers
	cfg.ProviderConfigs[provider] = storage.ProviderConfig{
		APIKey:    cfg.APIKey,
		Model:     cfg.Model,
		ToolModel: toolModel,
		BaseURL:   baseUrl,
	}

	// Atualizar serviço
	if baseUrl != "" {
		a.chatService.SetBaseURL(baseUrl)
	} else if provider == "groq" {
		a.chatService.SetBaseURL("https://api.groq.com/openai/v1")
	} else if provider == "zai" {
		a.chatService.SetBaseURL("https://api.z.ai/api/paas/v4")
	} else {
		a.chatService.SetBaseURL("https://openrouter.ai/api/v1")
	}

	return a.storage.SaveConfig(cfg)
}

// SetAskBeforeApply salva a configuração do modo YOLO
func (a *App) SetAskBeforeApply(value bool) error {
	if a.storage == nil {
		return fmt.Errorf("storage não disponível")
	}
	cfg, _ := a.storage.LoadConfig()
	if cfg == nil {
		cfg = &storage.Config{}
	}
	cfg.AskBeforeApply = value
	return a.storage.SaveConfig(cfg)
}

// GetAskBeforeApply retorna a configuração do modo YOLO
func (a *App) GetAskBeforeApply() (bool, error) {
	if a.storage == nil {
		return true, nil // Default seguro
	}
	cfg, err := a.storage.LoadConfig()
	if err != nil {
		return true, nil
	}
	return cfg.AskBeforeApply, nil
}

// SwitchProvider troca para outro provedor, carregando suas configurações salvas
func (a *App) SwitchProvider(providerName string) (*storage.Config, error) {
	if a.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}

	cfg, err := a.storage.SwitchProvider(providerName)
	if err != nil {
		return nil, err
	}

	// IMPORTANTE: Atualizar o provider no serviço de chat PRIMEIRO
	a.chatService.SetProvider(providerName)

	// Atualizar serviço de chat com as novas configurações
	if cfg.APIKey != "" {
		a.chatService.SetAPIKey(cfg.APIKey)
	}
	if cfg.Model != "" {
		a.chatService.SetModel(cfg.Model)
	}
	// ToolModel é usado apenas internamente na execução de tools, mas podemos setar se houver método
	// a.chatService.SetToolModel(cfg.ToolModel)
	if cfg.BaseURL != "" {
		a.chatService.SetBaseURL(cfg.BaseURL)
	}

	// Recarregar configurações completas
	a.chatService.RefreshConfig()

	return cfg, nil
}
