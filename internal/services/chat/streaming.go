package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"excel-ai/internal/domain"
	"excel-ai/pkg/ai"
)

// SendMessage envia mensagem para IA (Z.AI) e gerencia o loop autônomo de execução
func (s *Service) SendMessage(message string, contextStr string, askBeforeApply bool, onChunk func(string) error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar se deve usar orquestrador
	if s.useOrchestration {
		return s.orchestrator.OrchestrateMessage(message, contextStr, askBeforeApply, onChunk)
	}

	s.refreshConfig()

	// Verificar API key Z.AI
	if s.zaiClient.GetAPIKey() == "" {
		return "", fmt.Errorf("API key não configurada. Vá em Configurações e configure sua chave de API do Z.AI")
	}

	if s.currentConvID == "" {
		s.currentConvID = s.generateID()
	}

	s.ensureSystemPrompt()

	// Adicionar contexto mínimo (apenas workbook/sheet ativos) se existir
	if contextStr != "" {
		s.chatHistory = append(s.chatHistory, domain.Message{
			Role:      domain.RoleSystem,
			Content:   contextStr,
			Timestamp: time.Now(),
		})
	}

	// Adicionar mensagem do usuário
	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleUser,
		Content:   message,
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
		cancel()
	}()

	// Obter ferramentas Excel para function calling
	tools := ai.GetExcelTools()

	// LOOP AUTÔNOMO (limite alto pois cada ação tem confirmação individual)
	maxSteps := 50
	var finalResponse string

	for step := 0; step < maxSteps; step++ {
		// Verificar cancelamento
		if ctx.Err() != nil {
			return finalResponse, ctx.Err()
		}

		// Converter para AI messages
		aiHistory := s.toAIMessages(s.chatHistory)

		// Chamar Z.AI com tools
		var currentResponse string
		var toolCalls []ai.ToolCall
		var err error

		chunkWrapper := func(chunk string) error {
			currentResponse += chunk
			return onChunk(chunk)
		}

		// Usar cliente nativo Z.AI
		currentResponse, toolCalls, err = s.zaiClient.ChatStreamWithTools(ctx, aiHistory, tools, chunkWrapper)

		if err != nil {
			return finalResponse, err
		}

		// Filtrar tool calls inválidos
		toolCalls = ai.FilterValidToolCalls(toolCalls)

		// Se não recebemos tool calls estruturados válidos, tentar extrair do texto
		if len(toolCalls) == 0 && currentResponse != "" {
			extractedCalls, cleanedResponse := ai.ParseToolCallsFromText(currentResponse)
			if len(extractedCalls) > 0 {
				toolCalls = extractedCalls
				fmt.Printf("[DEBUG] Extracted %d tool call(s) from text response\n", len(extractedCalls))
				
				// Enviar resposta limpa para UI (já que o JSON não foi enviado durante streaming)
				if cleanedResponse != "" && cleanedResponse != currentResponse {
					trimmedClean := strings.TrimSpace(cleanedResponse)
					if trimmedClean != "" {
						onChunk(trimmedClean)
					}
				}
				currentResponse = cleanedResponse
			}
		}

		// Adiciona resposta da IA ao histórico
		s.chatHistory = append(s.chatHistory, domain.Message{
			Role:      domain.RoleAssistant,
			Content:   currentResponse,
			Timestamp: time.Now(),
			ToolCalls: toolCalls,
		})

		finalResponse = currentResponse

		// DEBUG: Log tool calls
		if len(toolCalls) > 0 {
			fmt.Printf("[DEBUG] Received %d tool call(s) from AI\n", len(toolCalls))
			for i, tc := range toolCalls {
				fmt.Printf("[DEBUG] Tool call %d: %s\n", i+1, tc.Function.Name)
			}
		} else {
			fmt.Println("[DEBUG] No tool calls received from AI")
		}

		// Se não há tool calls, terminamos o turno
		if len(toolCalls) == 0 {
			break
		}

		// Executar tool calls
		var executionResults []string
		for _, tc := range toolCalls {
			// Parsear argumentos
			args, parseErr := tc.ParseArguments()
			if parseErr != nil {
				executionResults = append(executionResults, fmt.Sprintf("ERROR parsing %s: %v", tc.Function.Name, parseErr))
				continue
			}

			// Verificar se é ação e precisa de confirmação
			if askBeforeApply && ai.IsActionTool(tc.Function.Name) {
				// Salvar ação pendente
				s.pendingAction = &ToolCommand{
					Type:    ToolTypeAction,
					Content: tc.Function.Arguments,
					Payload: args,
				}
				s.pendingAction.Payload.(map[string]interface{})["_tool_name"] = tc.Function.Name
				s.pendingContextStr = contextStr
				s.pendingOnChunk = onChunk

				pauseMsg := fmt.Sprintf("\n\n🛑 *[Ação Pendente: %s]* Aguardando aprovação do usuário para executar.\n", tc.Function.Name)
				onChunk(pauseMsg)
				finalResponse += pauseMsg

				go s.saveCurrentConversation(contextStr)
				return finalResponse, nil
			}

			// Executar ferramenta
			result, execErr := s.executeToolCall(tc.Function.Name, args)
			if execErr != nil {
				executionResults = append(executionResults, fmt.Sprintf("ERROR %s: %v", tc.Function.Name, execErr))
				onChunk(fmt.Sprintf("\n❌ Erro em %s: %v\n", tc.Function.Name, execErr))
			} else {
				executionResults = append(executionResults, fmt.Sprintf("SUCCESS %s: %s", tc.Function.Name, result))
				onChunk(fmt.Sprintf("\n✅ %s: %s\n", tc.Function.Name, result))
			}
		}

		// Adicionar resultados ao histórico para a IA ver (Hidden = não aparece no chat)
		toolResultMsg := fmt.Sprintf("Resultados das ferramentas executadas:\n\n%s\n\nUse estes dados para responder ao usuário. NÃO execute a mesma ferramenta novamente.", strings.Join(executionResults, "\n\n"))

		s.chatHistory = append(s.chatHistory, domain.Message{
			Role:      domain.RoleUser,
			Content:   toolResultMsg,
			Timestamp: time.Now(),
			Hidden:    true,
		})

		// Throttle para não estourar rate limit
		time.Sleep(2 * time.Second)
	}

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

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel
	defer cancel()

	aiHistory := s.toAIMessages(s.chatHistory)

	var response string
	var err error
	
	response, err = s.zaiClient.ChatStream(ctx, aiHistory, func(c string) error {
		response += c
		return onChunk(c)
	})

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
	// Also clear any pending action
	s.pendingAction = nil
}

// HasPendingAction returns true if there's a pending action waiting for confirmation
func (s *Service) HasPendingAction() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pendingAction != nil
}

// ConfirmPendingAction executes pending action and resumes AI loop
func (s *Service) ConfirmPendingAction(onChunk func(string) error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pendingAction == nil {
		return "", fmt.Errorf("no pending action to confirm")
	}

	cmd := s.pendingAction
	contextStr := s.pendingContextStr
	s.pendingAction = nil
	s.pendingContextStr = ""

	// Use provided onChunk or saved one
	if onChunk == nil {
		onChunk = s.pendingOnChunk
	}
	if onChunk == nil {
		onChunk = func(s string) error { return nil } // Fallback no-op
	}

	// Wire conversation ID to excel service for database-backed undo
	if s.excelService != nil {
		s.excelService.SetConversationID(s.currentConvID)
	}

	// Execute pending action usando novo sistema
	onChunk("\n\n✅ *[Executando ação aprovada...]*\n")

	// Extrair nome da ferramenta e argumentos do payload
	var result string
	var err error

	if cmd.Payload != nil {
		payload := cmd.Payload.(map[string]interface{})
		toolName, _ := payload["_tool_name"].(string)

		if toolName != "" {
			// Usar novo sistema executeToolCall
			delete(payload, "_tool_name") // Remover campo especial antes de executar
			result, err = s.executeToolCall(toolName, payload)
		} else {
			// Fallback para sistema antigo (para compatibilidade)
			result, err = s.ExecuteTool(*cmd)
		}
	} else {
		// Fallback para sistema antigo
		result, err = s.ExecuteTool(*cmd)
	}

	var executionResults string
	if err != nil {
		executionResults = fmt.Sprintf("ERROR: %v\n", err)
		onChunk(fmt.Sprintf("\n❌ Erro: %v\n", err))
	} else {
		executionResults = fmt.Sprintf("SUCCESS: %s\n", result)
		onChunk(fmt.Sprintf("\n✅ %s\n", result))
	}

	// Add execution results to chat history (Hidden = não aparece no chat)
	toolMsg := fmt.Sprintf("Resultados das ferramentas executadas:\n%s\n\nAgora responda ao usuário ou execute a próxima ação necessária.", executionResults)
	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleUser,
		Content:   toolMsg,
		Timestamp: time.Now(),
		Hidden:    true,
	})

	// Resume AI loop with function calling
	s.refreshConfig()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelMu.Lock()
	s.cancelFunc = cancel
	s.cancelMu.Unlock()
	defer func() {
		s.cancelMu.Lock()
		s.cancelFunc = nil
		s.cancelMu.Unlock()
		cancel()
	}()

	// Get tools
	tools := ai.GetExcelTools()

	// Convert to AI messages and continue
	aiHistory := s.toAIMessages(s.chatHistory)

	var currentResponse string
	var toolCalls []ai.ToolCall
	chunkWrapper := func(chunk string) error {
		currentResponse += chunk
		return onChunk(chunk)
	}

	currentResponse, toolCalls, err = s.zaiClient.ChatStreamWithTools(ctx, aiHistory, tools, chunkWrapper)

	if err != nil {
		return currentResponse, err
	}

	// Se não recebemos tool calls estruturados, tentar extrair do texto
	if len(toolCalls) == 0 && currentResponse != "" {
		extractedCalls, cleanedResponse := ai.ParseToolCallsFromText(currentResponse)
		if len(extractedCalls) > 0 {
			toolCalls = extractedCalls
			if cleanedResponse != currentResponse {
				currentResponse = cleanedResponse
				fmt.Printf("[DEBUG ResumePending] Extracted %d tool call(s) from text response\n", len(extractedCalls))
			}
		}
	}

	// Add response to history
	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleAssistant,
		Content:   currentResponse,
		Timestamp: time.Now(),
	})

	// Check for more tool calls
	if len(toolCalls) > 0 {
		for _, tc := range toolCalls {
			args, _ := tc.ParseArguments()

			if ai.IsActionTool(tc.Function.Name) {
				// Another action pending - save it
				s.pendingAction = &ToolCommand{
					Type:    ToolTypeAction,
					Content: tc.Function.Arguments,
					Payload: args,
				}
				s.pendingAction.Payload.(map[string]interface{})["_tool_name"] = tc.Function.Name
				s.pendingContextStr = contextStr
				s.pendingOnChunk = onChunk

				pauseMsg := fmt.Sprintf("\n\n🛑 *[Ação Pendente: %s]* Aguardando aprovação do usuário para executar.\n", tc.Function.Name)
				onChunk(pauseMsg)
				currentResponse += pauseMsg
				break
			} else {
				// Execute queries automatically
				queryResult, queryErr := s.executeToolCall(tc.Function.Name, args)
				if queryErr == nil {
					onChunk(fmt.Sprintf("\n📊 %s: %s\n", tc.Function.Name, queryResult))
				}
			}
		}
	}

	go s.saveCurrentConversation(contextStr)

	return currentResponse, nil
}

// RejectPendingAction discards pending action
func (s *Service) RejectPendingAction() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingAction = nil
	s.pendingContextStr = ""
	s.pendingOnChunk = nil
}

func (s *Service) refreshConfig() {
	if s.storage != nil {
		if cfg, err := s.storage.LoadConfig(); err == nil && cfg != nil {
			fmt.Printf("[DEBUG refreshConfig] Provider: zai, APIKey presente: %v, Model: %s, ToolModel: %s, BaseURL: %s\n",
				cfg.APIKey != "", cfg.Model, cfg.ToolModel, cfg.BaseURL)

			// Z.AI (GLM Models) - Usar cliente nativo com Coding API
			if cfg.APIKey != "" {
				s.zaiClient.SetAPIKey(cfg.APIKey)
			}
			if cfg.ToolModel != "" {
				s.zaiClient.SetModel(cfg.ToolModel)
			} else if cfg.Model != "" {
				s.zaiClient.SetModel(cfg.Model)
			}

			// SEMPRE usar Coding API
			baseURL := cfg.BaseURL
			if baseURL == "" || baseURL == "https://api.z.ai/api/paas/v4" || baseURL == "https://api.z.ai/api/paas/v4/" {
				baseURL = "https://api.z.ai/api/coding/paas/v4/"
			}
			// Garantir que a URL termina com barra
			if !strings.HasSuffix(baseURL, "/") {
				baseURL += "/"
			}
			s.zaiClient.SetBaseURL(baseURL)
			s.zaiClient.SetMaxInputTokens(128000) // GLM models have 128k context

			// Log detalhado para debug
			apiKeyLen := 0
			if cfg.APIKey != "" {
				apiKeyLen = len(cfg.APIKey)
			}
			fmt.Printf("[Z.AI] Configurado com URL: %s, Context: 128k tokens\n", baseURL)
			fmt.Printf("[Z.AI] API Key presente: %v, Comprimento: %d, Model: %s\n", cfg.APIKey != "", apiKeyLen, cfg.Model)
		}
	}
}

func (s *Service) ensureSystemPrompt() {
	systemPrompt := `Você é um AGENTE Excel profissional e direto. Seu objetivo é ajudar o usuário com planilhas de forma eficiente.

IDIOMA: Português do Brasil.

COMPORTAMENTO:
- Seja direto e profissional. Evite explicações longas desnecessárias.
- Responda apenas o que foi solicitado.
- Você pode usar raciocínio interno quando necessário para tarefas complexas.

QUANDO USAR FERRAMENTAS:
- Use ferramentas apenas quando o usuário pedir explicitamente ou quando for absolutamente necessário.
- NÃO use ferramentas para perguntas simples ou conversação.
- Antes de usar uma ferramenta, certifique-se de que é realmente necessária.

FERRAMENTAS DISPONÍVEIS:
1. Use query_batch para consultas e execute_macro para ações.
2. NUNCA retorne JSON ou definições de ferramentas no texto.
3. Se precisar de mais informações, pergunte de forma clara.`

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

// filterValues filtra valores baseado em uma coluna e valor
func (s *Service) filterValues(values [][]string, filterCol string, filterVal string) [][]string {
	if len(values) < 2 {
		return values
	}

	// Encontrar índice da coluna de filtro
	headers := values[0]
	colIndex := -1
	for i, h := range headers {
		if h == filterCol {
			colIndex = i
			break
		}
	}

	// Se não encontrou por nome, tentar como letra de coluna (A, B, C...)
	if colIndex == -1 && len(filterCol) == 1 {
		colIndex = int(filterCol[0] - 'A')
	}

	if colIndex == -1 || colIndex >= len(headers) {
		return values
	}

	// Filtrar linhas
	filtered := [][]string{headers}
	for i := 1; i < len(values); i++ {
		if colIndex < len(values[i]) {
			if values[i][colIndex] == filterVal {
				filtered = append(filtered, values[i])
			}
		}
	}

	return filtered
}
