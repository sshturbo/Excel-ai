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
	// Carregar configurações do storage
	cfg, err := a.storage.LoadConfig()
	if err != nil {
		// Usar valores padrão se não conseguir carregar
		return a.excelService.SetContextWithConfig(workbookName, sheetName, 50, true)
	}

	// Usar configurações do usuário
	maxRows := cfg.MaxRowsContext
	if maxRows <= 0 {
		maxRows = 50 // Padrão
	}

	return a.excelService.SetContextWithConfig(workbookName, sheetName, maxRows, cfg.IncludeHeaders)
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

// ConfigurePivotFields configura os campos de uma tabela dinâmica
func (a *App) ConfigurePivotFields(sheetName, tableName string, rowFields []string, dataFields []map[string]string) error {
	return a.excelService.ConfigurePivotFields(sheetName, tableName, rowFields, dataFields)
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

// FormatRange aplica formatação a um range
func (a *App) FormatRange(sheet, rangeAddr string, bold, italic bool, fontSize int, fontColor, bgColor string) error {
	return a.excelService.FormatRange(sheet, rangeAddr, bold, italic, fontSize, fontColor, bgColor)
}

// DeleteSheet exclui uma aba
func (a *App) DeleteSheet(sheetName string) error {
	return a.excelService.DeleteSheet(sheetName)
}

// RenameSheet renomeia uma aba
func (a *App) RenameSheet(oldName, newName string) error {
	return a.excelService.RenameSheet(oldName, newName)
}

// ClearRange limpa o conteúdo de um range
func (a *App) ClearRange(sheet, rangeAddr string) error {
	return a.excelService.ClearRange(sheet, rangeAddr)
}

// AutoFitColumns ajusta largura das colunas
func (a *App) AutoFitColumns(sheet, rangeAddr string) error {
	return a.excelService.AutoFitColumns(sheet, rangeAddr)
}

// InsertRows insere linhas
func (a *App) InsertRows(sheet string, rowNumber, count int) error {
	return a.excelService.InsertRows(sheet, rowNumber, count)
}

// DeleteRows exclui linhas
func (a *App) DeleteRows(sheet string, rowNumber, count int) error {
	return a.excelService.DeleteRows(sheet, rowNumber, count)
}

// MergeCells mescla células
func (a *App) MergeCells(sheet, rangeAddr string) error {
	return a.excelService.MergeCells(sheet, rangeAddr)
}

// UnmergeCells desfaz mesclagem
func (a *App) UnmergeCells(sheet, rangeAddr string) error {
	return a.excelService.UnmergeCells(sheet, rangeAddr)
}

// SetBorders adiciona bordas
func (a *App) SetBorders(sheet, rangeAddr, style string) error {
	return a.excelService.SetBorders(sheet, rangeAddr, style)
}

// SetColumnWidth define largura
func (a *App) SetColumnWidth(sheet, rangeAddr string, width float64) error {
	return a.excelService.SetColumnWidth(sheet, rangeAddr, width)
}

// SetRowHeight define altura
func (a *App) SetRowHeight(sheet, rangeAddr string, height float64) error {
	return a.excelService.SetRowHeight(sheet, rangeAddr, height)
}

// ApplyFilter aplica filtro
func (a *App) ApplyFilter(sheet, rangeAddr string) error {
	return a.excelService.ApplyFilter(sheet, rangeAddr)
}

// ClearFilters limpa filtros
func (a *App) ClearFilters(sheet string) error {
	return a.excelService.ClearFilters(sheet)
}

// SortRange ordena dados
func (a *App) SortRange(sheet, rangeAddr string, column int, ascending bool) error {
	return a.excelService.SortRange(sheet, rangeAddr, column, ascending)
}

// CopyRange copia range
func (a *App) CopyRange(sheet, sourceRange, destRange string) error {
	return a.excelService.CopyRange(sheet, sourceRange, destRange)
}

// ListCharts lista gráficos
func (a *App) ListCharts(sheet string) ([]string, error) {
	return a.excelService.ListCharts(sheet)
}

// DeleteChartByName exclui gráfico
func (a *App) DeleteChartByName(sheet, chartName string) error {
	return a.excelService.DeleteChart(sheet, chartName)
}

// CreateTable cria tabela formatada
func (a *App) CreateTable(sheet, rangeAddr, tableName, style string) error {
	return a.excelService.CreateTable(sheet, rangeAddr, tableName, style)
}

// ListTables lista tabelas
func (a *App) ListTables(sheet string) ([]string, error) {
	return a.excelService.ListTables(sheet)
}

// DeleteTable remove tabela
func (a *App) DeleteTable(sheet, tableName string) error {
	return a.excelService.DeleteTable(sheet, tableName)
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
	if cfg.ProviderConfigs == nil {
		cfg.ProviderConfigs = make(map[string]storage.ProviderConfig)
	}

	// Atualizar configurações gerais
	cfg.MaxRowsContext = maxRowsContext
	cfg.MaxRowsPreview = maxRowsPreview
	cfg.IncludeHeaders = includeHeaders
	cfg.DetailLevel = detailLevel
	cfg.CustomPrompt = customPrompt
	cfg.Language = language
	cfg.Provider = provider
	cfg.BaseURL = baseUrl

	// Salvar configurações do provedor atual no mapa de providers
	cfg.ProviderConfigs[provider] = storage.ProviderConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: baseUrl,
	}

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

// SwitchProvider troca para outro provedor, carregando suas configurações salvas
func (a *App) SwitchProvider(providerName string) (*storage.Config, error) {
	if a.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}

	cfg, err := a.storage.SwitchProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Atualizar serviço de chat com as novas configurações
	if cfg.APIKey != "" {
		a.chatService.SetAPIKey(cfg.APIKey)
	}
	if cfg.Model != "" {
		a.chatService.SetModel(cfg.Model)
	}
	if cfg.BaseURL != "" {
		a.chatService.SetBaseURL(cfg.BaseURL)
	}

	return cfg, nil
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

// SendErrorFeedback envia erro para a IA corrigir
func (a *App) SendErrorFeedback(errorMessage string) (string, error) {
	return a.chatService.SendErrorFeedback(errorMessage, func(chunk string) error {
		runtime.EventsEmit(a.ctx, "chat:chunk", chunk)
		return nil
	})
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

// ========== OPERAÇÕES DE CONSULTA (QUERY) ==========

// ListSheets lista todas as abas do workbook atual
func (a *App) ListSheets() ([]string, error) {
	return a.excelService.ListSheets()
}

// SheetExists verifica se uma aba existe
func (a *App) SheetExists(sheetName string) (bool, error) {
	return a.excelService.SheetExists(sheetName)
}

// ListPivotTables lista tabelas dinâmicas em uma aba
func (a *App) ListPivotTables(sheetName string) ([]string, error) {
	return a.excelService.ListPivotTables(sheetName)
}

// GetHeaders retorna os cabeçalhos de um range
func (a *App) GetHeaders(sheetName, rangeAddr string) ([]string, error) {
	return a.excelService.GetHeaders(sheetName, rangeAddr)
}

// GetUsedRange retorna o range utilizado de uma aba
func (a *App) GetUsedRange(sheetName string) (string, error) {
	return a.excelService.GetUsedRange(sheetName)
}

// QueryResult resultado de uma query
type QueryResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// QueryExcel executa uma query genérica no Excel
func (a *App) QueryExcel(queryType string, params map[string]string) QueryResult {
	switch queryType {
	case "list-sheets":
		sheets, err := a.ListSheets()
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: sheets}

	case "sheet-exists":
		exists, err := a.SheetExists(params["name"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: exists}

	case "list-pivot-tables":
		pivots, err := a.ListPivotTables(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: pivots}

	case "get-headers":
		headers, err := a.GetHeaders(params["sheet"], params["range"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: headers}

	case "get-used-range":
		usedRange, err := a.GetUsedRange(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: usedRange}

	case "get-row-count":
		count, err := a.excelService.GetRowCount(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: count}

	case "get-column-count":
		count, err := a.excelService.GetColumnCount(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: count}

	case "get-cell-formula":
		formula, err := a.excelService.GetCellFormula(params["sheet"], params["cell"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: formula}

	case "has-filter":
		hasFilter, err := a.excelService.HasFilter(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: hasFilter}

	case "get-active-cell":
		cell, err := a.excelService.GetActiveCell()
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: cell}

	case "get-range-values":
		values, err := a.excelService.GetRangeValues(params["sheet"], params["range"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: values}

	case "list-charts":
		charts, err := a.excelService.ListCharts(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: charts}

	case "list-tables":
		tables, err := a.excelService.ListTables(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: tables}

	default:
		return QueryResult{Success: false, Error: fmt.Sprintf("query type '%s' não reconhecido", queryType)}
	}
}
