package chat

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"excel-ai/internal/domain"
	"excel-ai/internal/dto"
	"excel-ai/pkg/ai"
	"excel-ai/pkg/storage"
)

type Service struct {
	client        *ai.Client
	geminiClient  *ai.GeminiClient
	provider      string // "openrouter", "groq", "google", "custom"
	storage       *storage.Storage
	mu            sync.Mutex
	chatHistory   []domain.Message
	currentConvID string
	cancelFunc    context.CancelFunc
}

func NewService(storage *storage.Storage) *Service {
	return &Service{
		client:       ai.NewClient("", "", ""), // API Key, Model, BaseURL set later
		geminiClient: ai.NewGeminiClient("", ""),
		provider:     "openrouter",
		storage:      storage,
		chatHistory:  []domain.Message{},
	}
}

func (s *Service) SetAPIKey(apiKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client.SetAPIKey(apiKey)
}

func (s *Service) SetModel(modelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client.SetModel(modelID)
}

func (s *Service) SetBaseURL(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client.SetBaseURL(url)
}

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

func (s *Service) SendMessage(message string, contextStr string, onChunk func(string) error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Recarregar configurações do storage para garantir API key atualizada
	if s.storage != nil {
		if cfg, err := s.storage.LoadConfig(); err == nil && cfg != nil {
			// Update provider
			s.provider = cfg.Provider

			// Configure OpenAI-compatible client
			if cfg.APIKey != "" {
				s.client.SetAPIKey(cfg.APIKey)
			}
			if cfg.Model != "" {
				s.client.SetModel(cfg.Model)
			}
			if cfg.BaseURL != "" {
				s.client.SetBaseURL(cfg.BaseURL)
			} else if cfg.Provider == "groq" {
				s.client.SetBaseURL("https://api.groq.com/openai/v1")
			}

			// Configure Gemini client
			if cfg.Provider == "google" {
				s.geminiClient.SetAPIKey(cfg.APIKey)
				s.geminiClient.SetModel(cfg.Model)
				if cfg.BaseURL != "" {
					s.geminiClient.SetBaseURL(cfg.BaseURL)
				}
			}
		}
	}

	// Verificar se API key está configurada
	if s.client.GetAPIKey() == "" {
		return "", fmt.Errorf("API key não configurada. Vá em Configurações e configure sua chave de API")
	}

	// Gerar ID de conversa se não existir
	if s.currentConvID == "" {
		s.currentConvID = storage.GenerateID()
	}

	// Garantir que temos um system prompt se o histórico estiver vazio
	if len(s.chatHistory) == 0 {
		systemPrompt := `You are an intelligent Excel AGENT. You work autonomously to complete tasks.

THINKING MODE:
When doing complex tasks, ALWAYS show your reasoning using:
:::thinking
[Your step-by-step reasoning here]
:::
This helps the user understand your thought process. Think out loud!

AGENT MODE:
1. FIRST make queries to understand the current state
2. THEN execute actions based on the results
3. Query results will be sent back to you - USE THEM!

QUERIES (check state):
:::excel-query
{"type": "list-sheets"}
{"type": "sheet-exists", "name": "SheetName"}
{"type": "list-pivot-tables", "sheet": "SheetName"}
{"type": "get-headers", "sheet": "SheetName", "range": "A:F"}
{"type": "get-used-range", "sheet": "SheetName"}
{"type": "get-row-count", "sheet": "SheetName"}
{"type": "get-column-count", "sheet": "SheetName"}
{"type": "get-cell-formula", "sheet": "SheetName", "cell": "A1"}
{"type": "has-filter", "sheet": "SheetName"}
{"type": "get-active-cell"}
{"type": "get-range-values", "sheet": "SheetName", "range": "A1:C10"}
{"type": "list-charts", "sheet": "SheetName"}
:::

ACTIONS (modify Excel):
:::excel-action
{"op": "write", "cell": "A1", "value": "value"}
{"op": "create-workbook", "name": "New.xlsx"}
{"op": "create-sheet", "name": "NewSheet"}
{"op": "create-chart", "range": "A1:B10", "chartType": "line", "title": "Title"}
{"op": "create-pivot", "sourceSheet": "X", "sourceRange": "A:F", "destSheet": "Y", "destCell": "A1", "tableName": "Name", "rowFields": ["field1"], "valueFields": [{"field": "field2", "function": "sum"}]}
{"op": "format-range", "range": "A1:B5", "bold": true, "italic": false, "fontSize": 12, "fontColor": "#FF0000", "bgColor": "#FFFF00"}
{"op": "delete-sheet", "name": "SheetToDelete"}
{"op": "rename-sheet", "oldName": "OldName", "newName": "NewName"}
{"op": "clear-range", "range": "A1:C10"}
{"op": "autofit", "range": "A:D"}
{"op": "insert-rows", "row": 5, "count": 3}
{"op": "delete-rows", "row": 5, "count": 2}
{"op": "merge-cells", "range": "A1:C1"}
{"op": "unmerge-cells", "range": "A1:C1"}
{"op": "set-borders", "range": "A1:D10", "style": "thin"}
{"op": "set-column-width", "range": "A:B", "width": 20}
{"op": "set-row-height", "range": "1:5", "height": 25}
{"op": "apply-filter", "range": "A1:D100"}
{"op": "clear-filters"}
{"op": "sort", "range": "A1:D100", "column": 1, "ascending": true}
{"op": "copy-range", "source": "A1:B10", "dest": "D1"}
{"op": "list-charts"}
{"op": "delete-chart", "name": "Chart1"}
{"op": "create-table", "range": "A1:D10", "name": "MinhaTabela", "style": "TableStyleMedium2"}
{"op": "delete-table", "name": "MinhaTabela"}
:::

AGENT RULES:
1. To create CHART: first use get-headers and get-used-range to know the data
2. To create PIVOT: first check if destination sheet exists with sheet-exists
3. For any complex task: make queries first!
4. You will receive results and can continue automatically
5. Use format-range to make headers bold or highlight data
6. Use autofit to adjust column widths after inserting data

EXAMPLE - Create chart with thinking:
:::thinking
User wants a chart. I need to:
1. Find out what data exists
2. Get the range of data
3. Identify column headers for chart labels
4. Create appropriate chart type
:::
:::excel-query
{"type": "get-used-range", "sheet": "Data"}
:::
(System will respond with range, then I continue)

Use formulas in PT-BR (SOMA, MÉDIA, SE, PROCV). DO NOT generate VBA.`

		s.chatHistory = append(s.chatHistory, domain.Message{
			Role:      domain.RoleSystem,
			Content:   systemPrompt,
			Timestamp: time.Now(),
		})
	}

	// Se context for fornecido, adiciona como mensagem do usuário
	fullContent := message
	if contextStr != "" {
		fullContent = fmt.Sprintf("Contexto do Excel:\n%s\n\nPergunta do usuário: %s", contextStr, message)
	}

	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleUser,
		Content:   fullContent,
		Timestamp: time.Now(),
	})

	// Criar context cancelável
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Converter domain history para AI history
	aiHistory := s.toAIMessages(s.chatHistory)

	// Call AI (use correct client based on provider)
	var response string
	var err error
	if s.provider == "google" {
		response, err = s.geminiClient.ChatStream(ctx, aiHistory, onChunk)
	} else {
		response, err = s.client.ChatStream(ctx, aiHistory, onChunk)
	}
	s.cancelFunc = nil
	if err != nil {
		// Remove user message on error (exceto se foi cancelado)
		if ctx.Err() != context.Canceled {
			s.chatHistory = s.chatHistory[:len(s.chatHistory)-1]
		}
		return "", err
	}

	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleAssistant,
		Content:   response,
		Timestamp: time.Now(),
	})

	go s.saveCurrentConversation(contextStr)

	return response, nil
}

// SendErrorFeedback envia um erro de execução para a IA e pede uma correção
func (s *Service) SendErrorFeedback(errorMessage string, onChunk func(string) error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Adiciona mensagem de erro como feedback
	feedbackMsg := fmt.Sprintf(`ERRO NA EXECUÇÃO DO COMANDO: %s

IMPORTANTE: Para corrigir este erro, você DEVE enviar DOIS comandos separados :::excel-action em sequência:
1. PRIMEIRO: Crie a aba com create-sheet
2. SEGUNDO: Crie a tabela dinâmica com create-pivot

Exemplo de resposta correta:
:::excel-action
{"op": "create-sheet", "name": "NOME_DA_ABA"}
:::

:::excel-action
{"op": "create-pivot", "sourceSheet": "...", "sourceRange": "...", "destSheet": "NOME_DA_ABA", "destCell": "A1", "tableName": "..."}
:::

Por favor, envie os comandos corrigidos agora.`, errorMessage)

	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleUser,
		Content:   feedbackMsg,
		Timestamp: time.Now(),
	})

	// Criar context cancelável
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Converter domain history para AI history
	aiHistory := s.toAIMessages(s.chatHistory)

	// Call AI para obter correção (use correct client based on provider)
	var response string
	var err error
	if s.provider == "google" {
		response, err = s.geminiClient.ChatStream(ctx, aiHistory, onChunk)
	} else {
		response, err = s.client.ChatStream(ctx, aiHistory, onChunk)
	}
	s.cancelFunc = nil
	if err != nil {
		// Remove mensagem de erro on failure
		if ctx.Err() != context.Canceled {
			s.chatHistory = s.chatHistory[:len(s.chatHistory)-1]
		}
		return "", err
	}

	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleAssistant,
		Content:   response,
		Timestamp: time.Now(),
	})

	go s.saveCurrentConversation("")

	return response, nil
}

// CancelChat cancela a requisição de chat em andamento
func (s *Service) CancelChat() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
}

func (s *Service) ClearChat() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatHistory = []domain.Message{}
}

func (s *Service) GetChatHistory() []dto.ChatMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []dto.ChatMessage
	for _, m := range s.chatHistory {
		result = append(result, dto.ChatMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}
	return result
}

func (s *Service) DeleteLastMessages(count int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if count <= 0 {
		return nil
	}

	if count > len(s.chatHistory) {
		s.chatHistory = []domain.Message{}
	} else {
		s.chatHistory = s.chatHistory[:len(s.chatHistory)-count]
	}

	go s.saveCurrentConversation("") // Context might be lost here if not stored in service state
	return nil
}

func (s *Service) NewConversation() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentConvID != "" && len(s.chatHistory) > 0 {
		s.saveCurrentConversation("")
	}

	s.chatHistory = []domain.Message{}
	s.currentConvID = storage.GenerateID()
	return s.currentConvID
}

func (s *Service) ListConversations() ([]dto.ConversationInfo, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}

	summaries, err := s.storage.ListConversations()
	if err != nil {
		return nil, err
	}

	var result []dto.ConversationInfo
	for _, summary := range summaries {
		result = append(result, dto.ConversationInfo{
			ID:        summary.ID,
			Title:     summary.Title,
			Preview:   summary.Preview,
			UpdatedAt: summary.UpdatedAt.Format("02/01/2006 15:04"),
		})
	}
	return result, nil
}

func (s *Service) LoadConversation(id string) ([]dto.ChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}

	conv, err := s.storage.LoadConversation(id)
	if err != nil {
		return nil, err
	}

	s.currentConvID = conv.ID
	s.chatHistory = []domain.Message{}

	var result []dto.ChatMessage
	for _, m := range conv.Messages {
		domainMsg := domain.Message{
			Role:      domain.MessageRole(m.Role),
			Content:   m.Content,
			Timestamp: m.Timestamp,
		}
		s.chatHistory = append(s.chatHistory, domainMsg)

		result = append(result, dto.ChatMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp.Format(time.RFC3339),
		})
	}

	return result, nil
}

func (s *Service) DeleteConversation(id string) error {
	if s.storage == nil {
		return fmt.Errorf("storage não disponível")
	}
	return s.storage.DeleteConversation(id)
}

func (s *Service) saveCurrentConversation(contextStr string) {
	if s.storage == nil || s.currentConvID == "" {
		return
	}

	var msgs []storage.Message
	for _, m := range s.chatHistory {
		msgs = append(msgs, storage.Message{
			Role:      string(m.Role),
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}

	conv := &storage.Conversation{
		ID:       s.currentConvID,
		Messages: msgs,
		Context:  contextStr,
	}

	s.storage.SaveConversation(conv)
}

// Helper to convert domain messages to AI messages
func (s *Service) toAIMessages(msgs []domain.Message) []ai.Message {
	var result []ai.Message
	for _, m := range msgs {
		result = append(result, ai.Message{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}
	return result
}
