package chat

import (
	"context"
	"fmt"
	"time"

	"excel-ai/internal/domain"
)

func (s *Service) SendMessage(message string, contextStr string, onChunk func(string) error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Recarregar configurações do storage para garantir API key atualizada
	s.refreshConfig()

	// Verificar se API key está configurada
	if s.client.GetAPIKey() == "" {
		return "", fmt.Errorf("API key não configurada. Vá em Configurações e configure sua chave de API")
	}

	// Gerar ID de conversa se não existir
	if s.currentConvID == "" {
		s.currentConvID = s.generateID()
	}

	// Garantir que temos um system prompt se o histórico estiver vazio
	if len(s.chatHistory) == 0 {
		s.addSystemPrompt()
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

// Helper methods refactored out of main body
func (s *Service) refreshConfig() {
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
}

func (s *Service) addSystemPrompt() {
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
