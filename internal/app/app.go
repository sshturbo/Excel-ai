package app

import (
	"context"

	chatService "excel-ai/internal/services/chat"
	excelService "excel-ai/internal/services/excel"
	"excel-ai/pkg/logger"
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
	excelSvc.SetStorage(stor) // Wire storage para undo no banco de dados
	chatSvc := chatService.NewService(stor)
	chatSvc.SetExcelService(excelSvc)

	// Carregar configurações salvas
	if stor != nil {
		if cfg, err := stor.LoadConfig(); err == nil && cfg != nil {
			if cfg.APIKey != "" {
				chatSvc.SetAPIKey(cfg.APIKey)
				logger.AppInfo("API key configurada pelo usuário")
			} else {
				logger.AppWarn("API key não configurada - usuário deve configurar nas configurações")
			}
			if cfg.Model != "" {
				chatSvc.SetModel(cfg.Model)
			}
			if cfg.BaseURL != "" {
				chatSvc.SetBaseURL(cfg.BaseURL)
			}
		} else {
			logger.AppWarn("Nenhuma configuração encontrada - usuário deve configurar API key")
		}
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
		logger.AppInfo("Licença válida: " + msg)
	} else {
		logger.AppError("Licença inválida: " + msg)
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
