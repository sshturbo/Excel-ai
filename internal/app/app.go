package app

import (
	"context"
	"fmt"

	chatService "excel-ai/internal/services/chat"
	excelService "excel-ai/internal/services/excel"
	"excel-ai/pkg/storage"
)

// App struct principal da aplicação
type App struct {
	ctx                context.Context
	excelService       *excelService.Service
	chatService        *chatService.Service
	storage            *storage.Storage
	watcherCancel      context.CancelFunc
	lastWorkbooksState string
	licenseValid       bool
	licenseMessage     string
}

// NewApp cria uma nova instância do App
func NewApp() *App {
	stor, _ := storage.NewStorage()

	// Inicializar serviços
	excelSvc := excelService.NewService()
	chatSvc := chatService.NewService(stor)
	chatSvc.SetExcelService(excelSvc)

	// Configurações padrão (Groq)
	defaultAPIKey := "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"
	defaultModel := "openai/gpt-oss-120b"
	defaultBaseURL := "https://api.groq.com/openai/v1"

	// Carregar configurações salvas ou usar padrão
	configLoaded := false
	if stor != nil {
		if cfg, err := stor.LoadConfig(); err == nil && cfg != nil {
			if cfg.APIKey != "" {
				chatSvc.SetAPIKey(cfg.APIKey)
				configLoaded = true
			}
			if cfg.Model != "" {
				chatSvc.SetModel(cfg.Model)
			}
			if cfg.BaseURL != "" {
				chatSvc.SetBaseURL(cfg.BaseURL)
			} else if cfg.Provider == "groq" {
				chatSvc.SetBaseURL("https://api.groq.com/openai/v1")
			}
		}
	}

	// Se não carregou configuração do usuário, usar padrão Groq
	if !configLoaded {
		fmt.Println("[DEBUG] Usando configuração padrão Groq")
		chatSvc.SetAPIKey(defaultAPIKey)
		chatSvc.SetModel(defaultModel)
		chatSvc.SetBaseURL(defaultBaseURL)
	} else {
		fmt.Println("[DEBUG] Configuração do usuário carregada")
	}

	return &App{
		excelService: excelSvc,
		chatService:  chatSvc,
		storage:      stor,
	}
}

// Startup é chamado quando o app inicia
// (Mudado para Exportado Capitalized para ser visível no main se necessário, mas Wails usa reflect, pode ser minúsculo se passado via options.
// No main.go original era app.startup referenciado. Se estiver em outro pacote, precisa ser Exportado Startup).
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Validar licença no startup
	valid, msg := a.CheckLicense()
	a.licenseValid = valid
	a.licenseMessage = msg

	if valid {
		fmt.Println("[LICENSE] ✅ Licença válida:", msg)
	} else {
		fmt.Println("[LICENSE] ❌ Licença inválida:", msg)
	}
}

// Shutdown é chamado quando o app fecha
func (a *App) Shutdown(ctx context.Context) {
	a.StopWorkbookWatcher()
	a.excelService.Close()
}

// Getter para contexto (útil se precisar em outros lugares, mas por enquanto interno no package)
func (a *App) Context() context.Context {
	return a.ctx
}
