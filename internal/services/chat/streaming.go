package chat

import (
	"context"
	"fmt"
	"time"

	"excel-ai/internal/domain"
)

// SendMessage envia mensagem para IA e gerencia o loop autônomo de execução
func (s *Service) SendMessage(message string, contextStr string, onChunk func(string) error) (string, error) {
	s.mu.Lock()
	// Lock é perigoso se o loop demorar muito e bloquear outras leituras,
	// mas necessário para proteger s.chatHistory.
	// O ideal seria travar apenas nas modificações de histórico, mas vamos manter assim por segurança.
	defer s.mu.Unlock()

	s.refreshConfig()

	if s.client.GetAPIKey() == "" {
		return "", fmt.Errorf("API key não configurada. Vá em Configurações e configure sua chave de API")
	}

	if s.currentConvID == "" {
		s.currentConvID = s.generateID()
	}

	if len(s.chatHistory) == 0 {
		s.ensureSystemPrompt()
	} else {
		s.ensureSystemPrompt() // Garante injecão mesmo se histórico não estiver vazio
	}

	// 1. Adicionar mensagem do usuário
	fullContent := message
	if contextStr != "" {
		fullContent = fmt.Sprintf("Contexto do Excel (Atualizado):\n%s\n\nPergunta do usuário: %s", contextStr, message)
	}

	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleUser,
		Content:   fullContent,
		Timestamp: time.Now(),
	})

	// Criar context cancelável
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelMu.Lock()
	s.cancelFunc = cancel
	s.cancelMu.Unlock()
	defer func() {
		s.cancelMu.Lock()
		s.cancelFunc = nil
		s.cancelMu.Unlock()
		cancel() // Garante limpeza
	}()

	// LOOP AUTÔNOMO (Max 5 passos para economizar quota no tier gratuito)
	maxSteps := 5
	var finalResponse string

	for step := 0; step < maxSteps; step++ {
		// Verificar cancelamento
		if ctx.Err() != nil {
			return finalResponse, ctx.Err()
		}

		// Converter para AI messages
		aiHistory := s.toAIMessages(s.chatHistory)

		// Call AI
		var currentResponse string
		var err error

		// Wrapper para onChunk acumular resposta atual
		chunkWrapper := func(chunk string) error {
			currentResponse += chunk
			return onChunk(chunk) // Passa pro frontend
		}

		if s.provider == "google" {
			_, err = s.geminiClient.ChatStream(ctx, aiHistory, chunkWrapper)
		} else {
			_, err = s.client.ChatStream(ctx, aiHistory, chunkWrapper)
		}

		if err != nil {
			// Se erro, removemos a última mensagem do usuário se for o primeiro passo?
			// Melhor não, apenas retornamos erro.
			return finalResponse, err
		}

		// Adiciona resposta da IA ao histórico
		s.chatHistory = append(s.chatHistory, domain.Message{
			Role:      domain.RoleAssistant,
			Content:   currentResponse,
			Timestamp: time.Now(),
		})

		// Atualiza resposta final (acumulativa ou última? Geralmente a última conversa é o que importa)
		finalResponse = currentResponse

		// PARSE COMMANDS
		commands := s.ParseToolCommands(currentResponse)

		// DEBUG: Log para ver se comandos foram parseados
		if len(commands) > 0 {
			fmt.Printf("[DEBUG] Parsed %d command(s) from AI response\n", len(commands))
			for i, cmd := range commands {
				fmt.Printf("[DEBUG] Command %d: Type=%s\n", i+1, cmd.Type)
			}
		} else {
			fmt.Println("[DEBUG] No commands parsed from AI response")
		}

		if len(commands) == 0 {
			// Sem comandos, terminamos o turno
			break
		}

		// Notificar usuário sobre progresso do passo
		stepMsg := fmt.Sprintf("\n\n🔄 *[Passo %d/%d] Executando %d ação(ões)...*\n\n", step+1, maxSteps, len(commands))
		onChunk(stepMsg)

		// Executar Comandos
		var executionResults string
		for _, cmd := range commands {
			result, err := s.ExecuteTool(cmd)
			if err != nil {
				executionResults += fmt.Sprintf("ERROR Executing %s: %v\n", cmd.Content, err)
			} else {
				executionResults += fmt.Sprintf("SUCCESS: %s\n", result)
			}
		}

		// Adicionar resultados ao histórico como System Message para a IA ver
		// Isso alimenta o próximo passo do loop
		toolMsg := fmt.Sprintf("TOOL RESULTS:\n%s\nContinue your task based on these results.", executionResults)

		s.chatHistory = append(s.chatHistory, domain.Message{
			Role:      domain.RoleUser, // OpenAI usa 'function' role, mas 'user' funciona bem para modelos genéricos
			Content:   toolMsg,
			Timestamp: time.Now(),
		})

		// Verificar se atingimos o limite de passos
		if step == maxSteps-1 {
			pauseMsg := "\n\n⚠️ *[Limite de Passos Atingido]* O agente atingiu o máximo de 5 passos por turno.\n\n:::agent-paused:::\n"
			onChunk(pauseMsg)
			finalResponse += pauseMsg // Incluir na resposta final para detecção
		}

		// THROTTLE: Aguardar para não estourar o Rate Limit da API (RPM)
		// Aumentado para 6s para melhor compatibilidade com tier gratuito
		time.Sleep(6 * time.Second)

		// Loop continua...
	}

	// Verificar se saímos por limite de passos (não por falta de comandos)
	// Se o loop rodou todas as iterações possíveis, emitir marcador de pausa
	// Note: Este código é alcançado apenas pelo for loop normal, não pelo break

	go s.saveCurrentConversation(contextStr)

	return finalResponse, nil
}

// SendErrorFeedback mantém a lógica simples de 1 turno
func (s *Service) SendErrorFeedback(errorMessage string, onChunk func(string) error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	feedbackMsg := fmt.Sprintf("Feedback de Erro: %s", errorMessage)

	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleUser,
		Content:   feedbackMsg,
		Timestamp: time.Now(),
	})

	// Reutiliza SendMessage logic? Não, pois SendMessage adiciona User msg.
	// Vamos simplificar e copiar a chamada simples, ou refatorar para usar o loop também.
	// Por enquanto, chamada direta simples para não complicar.

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel
	defer cancel()

	aiHistory := s.toAIMessages(s.chatHistory)

	var response string
	var err error
	if s.provider == "google" {
		response, err = s.geminiClient.ChatStream(ctx, aiHistory, func(c string) error {
			response += c
			return onChunk(c)
		})
	} else {
		response, err = s.client.ChatStream(ctx, aiHistory, func(c string) error {
			response += c
			return onChunk(c)
		})
	}

	if err == nil {
		s.chatHistory = append(s.chatHistory, domain.Message{
			Role:      domain.RoleAssistant,
			Content:   response,
			Timestamp: time.Now(),
		})
		go s.saveCurrentConversation("")
	}

	return response, err
}

func (s *Service) CancelChat() {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
}

func (s *Service) refreshConfig() {
	if s.storage != nil {
		if cfg, err := s.storage.LoadConfig(); err == nil && cfg != nil {
			s.provider = cfg.Provider
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

func (s *Service) ensureSystemPrompt() {
	systemPrompt := `You are an intelligent Excel AGENT. You work autonomously to complete tasks.

THINKING MODE:
When doing complex tasks, ALWAYS show your reasoning using:
:::thinking
[Your step-by-step reasoning here]
:::
This helps the user understand your thought process. Think out loud!

AGENT MODE:
CRITICAL FIRST STEP: Before ANY action, ALWAYS run list-sheets first to verify Excel is connected and has an open workbook. If it fails or returns empty, tell user to open an Excel file!

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
{"op": "macro", "actions": [{"op": "create-sheet", "name": "Dados"}, {"op": "write", "sheet": "Dados", "cell": "A1", "data": [["Col1", "Col2"], ["Val1", "Val2"]]}, {"op": "format-range", "sheet": "Dados", "range": "A1:B1", "bold": true}, {"op": "autofit", "sheet": "Dados", "range": "A:B"}]}
{"op": "write", "cell": "A1", "value": "single value"}
{"op": "write", "sheet": "SheetName", "cell": "A1", "data": [["Header1", "Header2"], ["Row1Val1", "Row1Val2"]]}
{"op": "create-workbook", "name": "New.xlsx"}
{"op": "create-sheet", "name": "NewSheet"}
{"op": "create-chart", "sheet": "X", "range": "A1:B10", "chartType": "line", "title": "Title"}
{"op": "create-pivot", "sourceSheet": "X", "sourceRange": "A:F", "destSheet": "Y", "destCell": "A1", "tableName": "Name", "rowFields": ["field1"], "valueFields": [{"field": "field2", "function": "sum"}]}
{"op": "format-range", "sheet": "X", "range": "A1:B5", "bold": true, "italic": false, "fontSize": 12, "fontColor": "#FF0000", "bgColor": "#FFFF00"}
{"op": "delete-sheet", "name": "SheetToDelete"}
{"op": "rename-sheet", "oldName": "OldName", "newName": "NewName"}
{"op": "clear-range", "sheet": "X", "range": "A1:C10"}
{"op": "autofit", "sheet": "X", "range": "A:D"}
{"op": "insert-rows", "sheet": "X", "row": 5, "count": 3}
{"op": "delete-rows", "sheet": "X", "row": 5, "count": 2}
{"op": "merge-cells", "sheet": "X", "range": "A1:C1"}
{"op": "unmerge-cells", "sheet": "X", "range": "A1:C1"}
{"op": "set-borders", "sheet": "X", "range": "A1:D10", "style": "thin"}
{"op": "set-column-width", "sheet": "X", "range": "A:B", "width": 20}
{"op": "set-row-height", "sheet": "X", "range": "1:5", "height": 25}
{"op": "apply-filter", "sheet": "X", "range": "A1:D100"}
{"op": "clear-filters", "sheet": "X"}
{"op": "sort", "sheet": "X", "range": "A1:D100", "column": 1, "ascending": true}
{"op": "copy-range", "sheet": "X", "source": "A1:B10", "dest": "D1"}
{"op": "list-charts", "sheet": "X"}
{"op": "delete-chart", "sheet": "X", "name": "Chart1"}
{"op": "create-table", "sheet": "X", "range": "A1:D10", "name": "MinhaTabela", "style": "TableStyleMedium2"}
{"op": "delete-table", "sheet": "X", "name": "MinhaTabela"}
:::

AGENT RULES:
1. To create CHART: first use get-headers and get-used-range to know the data
2. To create PIVOT: first check if destination sheet exists with sheet-exists
3. For any complex task: make queries first!
4. You will receive results and can continue automatically
5. Use format-range to make headers bold or highlight data
6. Use autofit to adjust column widths after inserting data
7. CRITICAL: ALWAYS specify "sheet" parameter in write/format actions! After creating a new sheet, use that sheet name in ALL following actions.
8. For batch data insert, use the "data" field with a 2D array: {"op": "write", "sheet": "MinhaAba", "cell": "A1", "data": [["Col1", "Col2"], ["Val1", "Val2"]]}
9. PREFER MACRO: When you need to do 2+ actions together (create sheet + write + format), use MACRO to run them all at once. This is faster and more efficient!

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

	if len(s.chatHistory) > 0 {
		if s.chatHistory[0].Role == domain.RoleSystem {
			// Update existing system prompt
			s.chatHistory[0].Content = systemPrompt
			return
		}
	}

	// Prepend system prompt
	sysMsg := domain.Message{
		Role:      domain.RoleSystem,
		Content:   systemPrompt,
		Timestamp: time.Now(),
	}
	s.chatHistory = append([]domain.Message{sysMsg}, s.chatHistory...)
}
