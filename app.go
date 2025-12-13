package main

import (
	"context"
	"excel-ai/internal/dto"
	chatService "excel-ai/internal/services/chat"
	excelService "excel-ai/internal/services/excel"
	"excel-ai/pkg/excel"
	"excel-ai/pkg/storage"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct principal da aplicação
type App struct {
	ctx          context.Context
	excelService *excelService.Service
	chatService  *chatService.Service
	storage      *storage.Storage
}

// NewApp cria uma nova instância do App
func NewApp() *App {
	stor, _ := storage.NewStorage()

	// Inicializar serviços
	excelSvc := excelService.NewService()
	chatSvc := chatService.NewService(stor)

	// Carregar configurações salvas
	if stor != nil {
		if cfg, err := stor.LoadConfig(); err == nil {
			if cfg.APIKey != "" {
				chatSvc.SetAPIKey(cfg.APIKey)
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

	return &App{
		excelService: excelSvc,
		chatService:  chatSvc,
		storage:      stor,
	}
}

// startup é chamado quando o app inicia
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown é chamado quando o app fecha
func (a *App) shutdown(ctx context.Context) {
	a.excelService.Close()
}

// ConnectExcel tenta conectar a uma instância do Excel
func (a *App) ConnectExcel() (*dto.ExcelStatus, error) {
	return a.excelService.Connect()
}

// RefreshWorkbooks atualiza a lista de pastas de trabalho
func (a *App) RefreshWorkbooks() (*dto.ExcelStatus, error) {
	return a.excelService.RefreshWorkbooks()
}

// GetPreviewData obtém preview dos dados antes de enviar para IA
func (a *App) GetPreviewData(workbookName, sheetName string) (*dto.PreviewData, error) {
	return a.excelService.GetPreviewData(workbookName, sheetName)
}

// SetExcelContext define o contexto do Excel para uso no chat
func (a *App) SetExcelContext(workbookName, sheetName string) (string, error) {
	return a.excelService.SetContext(workbookName, sheetName)
}

// GetActiveSelection obtém a seleção atual do Excel
func (a *App) GetActiveSelection() (*excel.SheetData, error) {
	return a.excelService.GetActiveSelection()
}

// UpdateExcelCell atualiza uma célula no Excel
func (a *App) UpdateExcelCell(workbook, sheet, cell, value string) error {
	return a.excelService.UpdateCell(workbook, sheet, cell, value)
}

// CreateNewWorkbook cria uma nova pasta de trabalho
func (a *App) CreateNewWorkbook() (string, error) {
	return a.excelService.CreateNewWorkbook()
}

// CreateChart cria um gráfico
func (a *App) CreateChart(sheet, rangeAddr, chartType, title string) error {
	return a.excelService.CreateChart(sheet, rangeAddr, chartType, title)
}

// CreatePivotTable cria uma tabela dinâmica
func (a *App) CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName string) error {
	return a.excelService.CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName)
}

// UndoLastChange desfaz a última alteração
func (a *App) UndoLastChange() error {
	return a.excelService.UndoLastChange()
}

// StartUndoBatch inicia um lote de alterações
func (a *App) StartUndoBatch() {
	a.excelService.StartUndoBatch()
}

// EndUndoBatch finaliza o lote de alterações
func (a *App) EndUndoBatch() {
	a.excelService.EndUndoBatch()
}

// SetAPIKey configura a chave da API
func (a *App) SetAPIKey(apiKey string) error {
	a.chatService.SetAPIKey(apiKey)
	// Save to storage
	if a.storage != nil {
		cfg, _ := a.storage.LoadConfig()
		if cfg == nil {
			cfg = &storage.Config{}
		}
		cfg.APIKey = apiKey
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
		cfg.Model = model
		return a.storage.SaveConfig(cfg)
	}
	return nil
}

// GetAvailableModels retorna modelos disponíveis
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
func (a *App) UpdateConfig(maxRowsContext, maxRowsPreview int, includeHeaders bool, detailLevel, customPrompt, language, provider, baseUrl string) error {
	if a.storage == nil {
		return fmt.Errorf("storage não disponível")
	}
	cfg, _ := a.storage.LoadConfig()
	if cfg == nil {
		cfg = &storage.Config{}
	}
	cfg.MaxRowsContext = maxRowsContext
	cfg.MaxRowsPreview = maxRowsPreview
	cfg.IncludeHeaders = includeHeaders
	cfg.DetailLevel = detailLevel
	cfg.CustomPrompt = customPrompt
	cfg.Language = language
	cfg.Provider = provider
	cfg.BaseURL = baseUrl

	// Atualizar serviço
	if baseUrl != "" {
		a.chatService.SetBaseURL(baseUrl)
	} else if provider == "groq" {
		a.chatService.SetBaseURL("https://api.groq.com/openai/v1")
	} else {
		a.chatService.SetBaseURL("https://openrouter.ai/api/v1")
	}

	return a.storage.SaveConfig(cfg)
}

// SendMessage envia mensagem para IA
func (a *App) SendMessage(message string) (string, error) {
	contextStr := a.excelService.GetContextString()
	return a.chatService.SendMessage(message, contextStr, func(chunk string) error {
		runtime.EventsEmit(a.ctx, "chat:chunk", chunk)
		return nil
	})
}

// ClearChat limpa o chat
func (a *App) ClearChat() {
	a.chatService.ClearChat()
}

// DeleteLastMessages remove mensagens
func (a *App) DeleteLastMessages(count int) error {
	return a.chatService.DeleteLastMessages(count)
}

// NewConversation nova conversa
func (a *App) NewConversation() string {
	return a.chatService.NewConversation()
}

// ListConversations lista conversas
func (a *App) ListConversations() ([]dto.ConversationInfo, error) {
	return a.chatService.ListConversations()
}

// LoadConversation carrega conversa
func (a *App) LoadConversation(id string) ([]dto.ChatMessage, error) {
	return a.chatService.LoadConversation(id)
}

// DeleteConversation remove conversa
func (a *App) DeleteConversation(id string) error {
	return a.chatService.DeleteConversation(id)
}

// GetChatHistory retorna histórico
func (a *App) GetChatHistory() []dto.ChatMessage {
	return a.chatService.GetChatHistory()
}

// WriteToExcel escreve valor
func (a *App) WriteToExcel(row, col int, value string) error {
	return a.excelService.WriteToExcel(row, col, value)
}

// ApplyFormula aplica fórmula
func (a *App) ApplyFormula(row, col int, formula string) error {
	return a.excelService.ApplyFormula(row, col, formula)
}

// CreateNewSheet cria nova aba
func (a *App) CreateNewSheet(name string) error {
	return a.excelService.CreateNewSheet(name)
}
