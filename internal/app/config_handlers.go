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
		cfg.APIKey = apiKey
		cfg.Provider = "zai" // Sempre Z.ai

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
		cfg.Model = model

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

	// Save to storage
	if a.storage != nil {
		cfg, _ := a.storage.LoadConfig()
		if cfg == nil {
			cfg = &storage.Config{}
		}
		cfg.ToolModel = toolModel

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

	// Atualizar configurações gerais
	cfg.MaxRowsContext = maxRowsContext
	cfg.MaxContextChars = maxContextChars
	cfg.MaxRowsPreview = maxRowsPreview
	cfg.IncludeHeaders = includeHeaders
	cfg.DetailLevel = detailLevel
	cfg.CustomPrompt = customPrompt
	cfg.Language = language
	cfg.Provider = "zai" // Sempre Z.ai
	cfg.BaseURL = baseUrl
	cfg.ToolModel = toolModel

	// Atualizar serviço
	if baseUrl != "" {
		a.chatService.SetBaseURL(baseUrl)
	} else {
		// Usar Coding API por padrão
		a.chatService.SetBaseURL("https://api.z.ai/api/coding/paas/v4/")
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
