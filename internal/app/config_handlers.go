package app

import (
	"fmt"

	"excel-ai/internal/dto"
	apperrors "excel-ai/pkg/errors"
	"excel-ai/pkg/logger"
	"excel-ai/pkg/storage"
	"excel-ai/pkg/validator"
)

// SetAPIKey configura a chave da API
func (a *App) SetAPIKey(apiKey string) error {
	// Validar API key
	v := validator.NewValidator()
	v.ValidateAPIKey("apiKey", apiKey)
	if v.HasErrors() {
		logger.AppWarn("Validação de API key falhou: " + v.Error().Error())
		return apperrors.InvalidInput(v.Error().Error())
	}

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

		if err := a.storage.SaveConfig(cfg); err != nil {
			logger.AppError("Erro ao salvar API key: " + err.Error())
			return apperrors.Wrap(err, apperrors.ErrCodeStorageError, "erro ao salvar configuração")
		}
		
		logger.AppInfo("API key atualizada com sucesso")
	}
	return nil
}

// SetModel configura o modelo da IA
func (a *App) SetModel(model string) error {
	// Validar nome do modelo
	v := validator.NewValidator()
	v.ValidateModelName("model", model)
	if v.HasErrors() {
		logger.AppWarn("Validação de modelo falhou: " + v.Error().Error())
		return apperrors.InvalidInput(v.Error().Error())
	}

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

		if err := a.storage.SaveConfig(cfg); err != nil {
			logger.AppError("Erro ao salvar modelo: " + err.Error())
			return apperrors.Wrap(err, apperrors.ErrCodeStorageError, "erro ao salvar configuração")
		}
		
		logger.AppInfo("Modelo configurado: " + model)
	}
	return nil
}

// SetToolModel configura o modelo secundário da IA
func (a *App) SetToolModel(toolModel string) error {
	// Validar nome do modelo
	v := validator.NewValidator()
	v.ValidateModelName("toolModel", toolModel)
	if v.HasErrors() {
		logger.AppWarn("Validação de tool model falhou: " + v.Error().Error())
		return apperrors.InvalidInput(v.Error().Error())
	}

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

		if err := a.storage.SaveConfig(cfg); err != nil {
			logger.AppError("Erro ao salvar tool model: " + err.Error())
			return apperrors.Wrap(err, apperrors.ErrCodeStorageError, "erro ao salvar configuração")
		}
		
		logger.AppInfo("Tool model configurado: " + toolModel)
	}
	return nil
}
func (a *App) GetAvailableModels(apiKey, baseURL string) ([]dto.ModelInfo, error) {
	// Validar URL se fornecida
	if baseURL != "" {
		v := validator.NewValidator()
		v.ValidateURL("baseURL", baseURL)
		if v.HasErrors() {
			logger.AppWarn("Validação de baseURL falhou: " + v.Error().Error())
			return nil, apperrors.InvalidInput(v.Error().Error())
		}
	}
	
	logger.AppInfo("Buscando modelos disponíveis")
	return a.chatService.GetAvailableModels(apiKey, baseURL), nil
}

// GetSavedConfig retorna configurações salvas
func (a *App) GetSavedConfig() (*storage.Config, error) {
	if a.storage == nil {
		logger.AppError("Storage não disponível ao carregar configuração")
		return nil, apperrors.StorageError("storage não disponível")
	}
	
	logger.StorageDebug("Carregando configurações salvas")
	return a.storage.LoadConfig()
}

// UpdateConfig atualiza configurações
func (a *App) UpdateConfig(maxRowsContext, maxContextChars, maxRowsPreview int, includeHeaders bool, detailLevel, customPrompt, language, provider, toolModel, baseUrl string) error {
	if a.storage == nil {
		logger.AppError("Storage não disponível ao atualizar configuração")
		return apperrors.StorageError("storage não disponível")
	}
	
	// Validar inputs
	v := validator.NewValidator()
	v.ValidateInteger("maxRowsContext", fmt.Sprintf("%d", maxRowsContext), 1, 100000)
	v.ValidateInteger("maxContextChars", fmt.Sprintf("%d", maxContextChars), 100, 1000000)
	v.ValidateInteger("maxRowsPreview", fmt.Sprintf("%d", maxRowsPreview), 1, 1000)
	v.ValidateEnum("provider", provider, []string{"openrouter", "groq", "zai"}, false)
	
	if baseUrl != "" {
		v.ValidateURL("baseUrl", baseUrl)
	}
	
	if toolModel != "" {
		v.ValidateModelName("toolModel", toolModel)
	}
	
	if customPrompt != "" {
		v.ValidateMaxLength("customPrompt", customPrompt, 2000)
	}
	
	if language != "" {
		v.ValidateEnum("language", language, []string{"en", "pt", "es"}, false)
	}
	
	if v.HasErrors() {
		logger.AppWarn("Validação de configuração falhou: " + v.Error().Error())
		return apperrors.InvalidInput(v.Error().Error())
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

	if err := a.storage.SaveConfig(cfg); err != nil {
		logger.AppError("Erro ao salvar configuração: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeStorageError, "erro ao salvar configuração")
	}
	
	logger.AppInfo("Configuração atualizada com sucesso")
	return nil
}

// SetAskBeforeApply salva a configuração do modo YOLO
func (a *App) SetAskBeforeApply(value bool) error {
	if a.storage == nil {
		logger.AppError("Storage não disponível ao configurar AskBeforeApply")
		return apperrors.StorageError("storage não disponível")
	}
	
	cfg, _ := a.storage.LoadConfig()
	if cfg == nil {
		cfg = &storage.Config{}
	}
	cfg.AskBeforeApply = value
	
	if err := a.storage.SaveConfig(cfg); err != nil {
		logger.AppError("Erro ao salvar AskBeforeApply: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeStorageError, "erro ao salvar configuração")
	}
	
	logger.AppInfo("AskBeforeApply configurado: " + fmt.Sprintf("%v", value))
	return nil
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
	// Validar nome do provedor
	v := validator.NewValidator()
	v.ValidateEnum("providerName", providerName, []string{"openrouter", "groq", "zai"}, true)
	if v.HasErrors() {
		logger.AppWarn("Validação de provider falhou: " + v.Error().Error())
		return nil, apperrors.InvalidInput(v.Error().Error())
	}
	
	if a.storage == nil {
		logger.AppError("Storage não disponível ao trocar provider")
		return nil, apperrors.StorageError("storage não disponível")
	}

	logger.AppInfo("Trocando para provider: " + providerName)
	
	cfg, err := a.storage.SwitchProvider(providerName)
	if err != nil {
		logger.AppError("Erro ao trocar provider: " + err.Error())
		return nil, apperrors.Wrap(err, apperrors.ErrCodeStorageError, "erro ao trocar provider")
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

	logger.AppInfo("Provider alterado com sucesso: " + providerName)
	return cfg, nil
}
