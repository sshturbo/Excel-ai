package main

import (
	"context"
	"excel-ai/pkg/ai"
	"excel-ai/pkg/excel"
	"excel-ai/pkg/storage"
	"fmt"
	"sync"
	"time"
)

// App struct principal da aplicação
type App struct {
	ctx             context.Context
	excelClient     *excel.Client
	aiClient        *ai.Client
	storage         *storage.Storage
	chatHistory     []ai.Message
	currentConvID   string
	mu              sync.Mutex
	excelContext    string
	currentWorkbook string
	currentSheet    string
	previewData     *excel.SheetData
}

// ExcelStatus status da conexão com Excel
type ExcelStatus struct {
	Connected bool             `json:"connected"`
	Workbooks []excel.Workbook `json:"workbooks"`
	Error     string           `json:"error,omitempty"`
}

// ChatMessage mensagem do chat para frontend
type ChatMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
}

// ConversationInfo informações de conversa para frontend
type ConversationInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Preview   string `json:"preview"`
	UpdatedAt string `json:"updatedAt"`
}

// PreviewData dados para preview no frontend
type PreviewData struct {
	Headers   []string   `json:"headers"`
	Rows      [][]string `json:"rows"`
	TotalRows int        `json:"totalRows"`
	TotalCols int        `json:"totalCols"`
	Workbook  string     `json:"workbook"`
	Sheet     string     `json:"sheet"`
}

// WriteRequest requisição de escrita no Excel
type WriteRequest struct {
	Row     int         `json:"row"`
	Col     int         `json:"col"`
	Value   interface{} `json:"value"`
	Formula string      `json:"formula,omitempty"`
}

// NewApp cria uma nova instância do App
func NewApp() *App {
	stor, _ := storage.NewStorage()

	app := &App{
		chatHistory: []ai.Message{},
		aiClient:    ai.NewClient("", ""),
		storage:     stor,
	}

	// Carregar configurações salvas
	if stor != nil {
		if cfg, err := stor.LoadConfig(); err == nil {
			if cfg.APIKey != "" {
				app.aiClient.SetAPIKey(cfg.APIKey)
			}
			if cfg.Model != "" {
				app.aiClient.SetModel(cfg.Model)
			}
		}
	}

	return app
}

// startup é chamado quando o app inicia
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown é chamado quando o app fecha
func (a *App) shutdown(ctx context.Context) {
	// Salvar conversa atual se existir
	if a.currentConvID != "" && len(a.chatHistory) > 0 {
		a.saveCurrentConversation()
	}

	if a.excelClient != nil {
		a.excelClient.Close()
	}
}

// ConnectExcel tenta conectar a uma instância do Excel
func (a *App) ConnectExcel() ExcelStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.excelClient != nil {
		a.excelClient.Close()
	}

	client, err := excel.NewClient()
	if err != nil {
		return ExcelStatus{
			Connected: false,
			Error:     err.Error(),
		}
	}

	a.excelClient = client

	workbooks, err := client.GetOpenWorkbooks()
	if err != nil {
		return ExcelStatus{
			Connected: true,
			Workbooks: []excel.Workbook{},
			Error:     fmt.Sprintf("Conectado, mas falha ao listar planilhas: %s", err.Error()),
		}
	}

	return ExcelStatus{
		Connected: true,
		Workbooks: workbooks,
	}
}

// RefreshWorkbooks atualiza a lista de pastas de trabalho
func (a *App) RefreshWorkbooks() ExcelStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.excelClient == nil {
		return ExcelStatus{
			Connected: false,
			Error:     "Não conectado ao Excel",
		}
	}

	workbooks, err := a.excelClient.GetOpenWorkbooks()
	if err != nil {
		return ExcelStatus{
			Connected: true,
			Error:     err.Error(),
		}
	}

	return ExcelStatus{
		Connected: true,
		Workbooks: workbooks,
	}
}

// GetPreviewData obtém preview dos dados antes de enviar para IA
func (a *App) GetPreviewData(workbookName, sheetName string) (*PreviewData, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.excelClient == nil {
		return nil, fmt.Errorf("não conectado ao Excel")
	}

	data, err := a.excelClient.GetSheetData(workbookName, sheetName, 100)
	if err != nil {
		return nil, err
	}

	a.previewData = data
	a.currentWorkbook = workbookName
	a.currentSheet = sheetName

	// Converter para formato simples
	var rows [][]string
	for _, row := range data.Rows {
		var rowStrings []string
		for _, cell := range row {
			rowStrings = append(rowStrings, cell.Text)
		}
		rows = append(rows, rowStrings)
	}

	return &PreviewData{
		Headers:   data.Headers,
		Rows:      rows,
		TotalRows: len(data.Rows) + 1, // +1 para header
		TotalCols: len(data.Headers),
		Workbook:  workbookName,
		Sheet:     sheetName,
	}, nil
}

// SetExcelContext define o contexto do Excel para uso no chat
func (a *App) SetExcelContext(workbookName, sheetName string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.excelClient == nil {
		return "", fmt.Errorf("não conectado ao Excel")
	}

	data, err := a.excelClient.GetSheetData(workbookName, sheetName, 50)
	if err != nil {
		return "", err
	}

	a.previewData = data
	a.currentWorkbook = workbookName
	a.currentSheet = sheetName

	// Converter para formato de texto para context
	var rows [][]string
	for _, row := range data.Rows {
		var rowStrings []string
		for _, cell := range row {
			rowStrings = append(rowStrings, cell.Text)
		}
		rows = append(rows, rowStrings)
	}

	a.excelContext = ai.BuildExcelContext(data.Headers, rows, 30)
	return fmt.Sprintf("Contexto carregado: %s - %s (%d linhas)", workbookName, sheetName, len(data.Rows)), nil
}

// GetActiveSelection obtém a seleção atual do Excel
func (a *App) GetActiveSelection() (*excel.SheetData, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.excelClient == nil {
		return nil, fmt.Errorf("não conectado ao Excel")
	}

	return a.excelClient.GetActiveSelection()
}

// SetAPIKey configura a chave da API OpenRouter
func (a *App) SetAPIKey(apiKey string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.aiClient.SetAPIKey(apiKey)

	// Salvar no storage
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
	a.mu.Lock()
	defer a.mu.Unlock()

	a.aiClient.SetModel(model)

	// Salvar no storage
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

// GetSavedConfig retorna configurações salvas
func (a *App) GetSavedConfig() (*storage.Config, error) {
	if a.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}
	return a.storage.LoadConfig()
}

// UpdateConfig atualiza todas as configurações
func (a *App) UpdateConfig(maxRowsContext, maxRowsPreview int, includeHeaders bool, detailLevel, customPrompt, language string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

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

	return a.storage.SaveConfig(cfg)
}

// SendMessage envia uma mensagem para a IA
func (a *App) SendMessage(message string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Criar nova conversa se não existir
	if a.currentConvID == "" {
		a.currentConvID = storage.GenerateID()
	}

	// Adicionar mensagem do usuário ao histórico
	a.chatHistory = append(a.chatHistory, ai.Message{
		Role:    "user",
		Content: message,
	})

	// Sistema prompt aprimorado com contexto do Excel
	systemContent := `Você é um assistente de análise de dados especializado em Excel. 
Ajude o usuário a entender, analisar e manipular dados de planilhas.
Seja claro, conciso e forneça respostas práticas.
Quando apropriado, sugira fórmulas do Excel ou passos para resolver problemas.
Se sugerir uma fórmula ou valor para escrever em uma célula, formate assim:
[ESCREVER A1: =SOMA(B1:B10)] ou [ESCREVER B2: Texto aqui]
Responda em português.`

	// Incluir contexto do Excel no system prompt se disponível
	if a.excelContext != "" {
		systemContent = fmt.Sprintf(`%s

=== DADOS DA PLANILHA ATUAL ===
Pasta de Trabalho: %s
Aba: %s

%s
=== FIM DOS DADOS ===

Use esses dados para responder às perguntas do usuário. Quando o usuário perguntar sobre os dados, analise a tabela acima.`,
			systemContent, a.currentWorkbook, a.currentSheet, a.excelContext)
	}

	systemMsg := ai.Message{
		Role:    "system",
		Content: systemContent,
	}

	messages := append([]ai.Message{systemMsg}, a.chatHistory...)

	response, err := a.aiClient.Chat(messages)
	if err != nil {
		// Remover a mensagem do usuário se houve erro
		a.chatHistory = a.chatHistory[:len(a.chatHistory)-1]
		return "", err
	}

	a.chatHistory = append(a.chatHistory, ai.Message{
		Role:    "assistant",
		Content: response,
	})

	// Salvar conversa automaticamente
	go a.saveCurrentConversation()

	return response, nil
}

// saveCurrentConversation salva a conversa atual
func (a *App) saveCurrentConversation() {
	if a.storage == nil || a.currentConvID == "" {
		return
	}

	var msgs []storage.Message
	for _, m := range a.chatHistory {
		msgs = append(msgs, storage.Message{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: time.Now(),
		})
	}

	conv := &storage.Conversation{
		ID:       a.currentConvID,
		Messages: msgs,
		Context:  a.excelContext,
	}

	a.storage.SaveConversation(conv)
}

// ClearChat limpa o histórico do chat
func (a *App) ClearChat() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Salvar antes de limpar
	if a.currentConvID != "" && len(a.chatHistory) > 0 {
		a.saveCurrentConversation()
	}

	a.chatHistory = []ai.Message{}
	a.excelContext = ""
	a.currentConvID = ""
	a.previewData = nil
}

// NewConversation inicia uma nova conversa
func (a *App) NewConversation() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Salvar conversa anterior
	if a.currentConvID != "" && len(a.chatHistory) > 0 {
		a.saveCurrentConversation()
	}

	a.chatHistory = []ai.Message{}
	a.excelContext = ""
	a.currentConvID = storage.GenerateID()

	return a.currentConvID
}

// ListConversations lista conversas salvas
func (a *App) ListConversations() ([]ConversationInfo, error) {
	if a.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}

	summaries, err := a.storage.ListConversations()
	if err != nil {
		return nil, err
	}

	var result []ConversationInfo
	for _, s := range summaries {
		result = append(result, ConversationInfo{
			ID:        s.ID,
			Title:     s.Title,
			Preview:   s.Preview,
			UpdatedAt: s.UpdatedAt.Format("02/01/2006 15:04"),
		})
	}

	return result, nil
}

// LoadConversation carrega uma conversa salva
func (a *App) LoadConversation(id string) ([]ChatMessage, error) {
	if a.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Salvar conversa atual primeiro
	if a.currentConvID != "" && len(a.chatHistory) > 0 {
		a.saveCurrentConversation()
	}

	conv, err := a.storage.LoadConversation(id)
	if err != nil {
		return nil, err
	}

	a.currentConvID = id
	a.excelContext = conv.Context
	a.chatHistory = []ai.Message{}

	var messages []ChatMessage
	for _, m := range conv.Messages {
		a.chatHistory = append(a.chatHistory, ai.Message{
			Role:    m.Role,
			Content: m.Content,
		})
		messages = append(messages, ChatMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp.Format("15:04"),
		})
	}

	return messages, nil
}

// DeleteConversation remove uma conversa
func (a *App) DeleteConversation(id string) error {
	if a.storage == nil {
		return fmt.Errorf("storage não disponível")
	}
	return a.storage.DeleteConversation(id)
}

// GetChatHistory retorna o histórico do chat
func (a *App) GetChatHistory() []ChatMessage {
	a.mu.Lock()
	defer a.mu.Unlock()

	var messages []ChatMessage
	for _, msg := range a.chatHistory {
		messages = append(messages, ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return messages
}

// WriteToExcel escreve um valor em uma célula
func (a *App) WriteToExcel(row, col int, value string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.excelClient == nil {
		return fmt.Errorf("não conectado ao Excel")
	}

	if a.currentWorkbook == "" || a.currentSheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	return a.excelClient.WriteCell(a.currentWorkbook, a.currentSheet, row, col, value)
}

// ApplyFormula aplica uma fórmula em uma célula
func (a *App) ApplyFormula(row, col int, formula string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.excelClient == nil {
		return fmt.Errorf("não conectado ao Excel")
	}

	if a.currentWorkbook == "" || a.currentSheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	return a.excelClient.ApplyFormula(a.currentWorkbook, a.currentSheet, row, col, formula)
}

// CreateNewSheet cria uma nova aba
func (a *App) CreateNewSheet(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.excelClient == nil {
		return fmt.Errorf("não conectado ao Excel")
	}

	if a.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}

	return a.excelClient.InsertNewSheet(a.currentWorkbook, name)
}
