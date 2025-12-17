package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"excel-ai/internal/domain"
	"excel-ai/pkg/ai"
)

// SendMessage envia mensagem para IA e gerencia o loop autônomo de execução com function calling nativo
func (s *Service) SendMessage(message string, contextStr string, askBeforeApply bool, onChunk func(string) error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refreshConfig()

	// Verificar API key do cliente correto baseado no provider
	if s.provider == "google" {
		if s.geminiClient.GetAPIKey() == "" {
			return "", fmt.Errorf("API key não configurada. Vá em Configurações e configure sua chave de API do Google")
		}
	} else {
		if s.client.GetAPIKey() == "" {
			return "", fmt.Errorf("API key não configurada. Vá em Configurações e configure sua chave de API")
		}
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
	geminiTools := ai.GetGeminiTools()

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

		// Chamar IA com tools
		var currentResponse string
		var toolCalls []ai.ToolCall
		var err error

		chunkWrapper := func(chunk string) error {
			currentResponse += chunk
			return onChunk(chunk)
		}

		if s.provider == "google" {
			currentResponse, toolCalls, err = s.geminiClient.ChatStreamWithTools(ctx, aiHistory, geminiTools, chunkWrapper)
		} else {
			currentResponse, toolCalls, err = s.client.ChatStreamWithTools(ctx, aiHistory, tools, chunkWrapper)
		}

		if err != nil {
			return finalResponse, err
		}

		// Adiciona resposta da IA ao histórico
		s.chatHistory = append(s.chatHistory, domain.Message{
			Role:      domain.RoleAssistant,
			Content:   currentResponse,
			Timestamp: time.Now(),
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

		// Adicionar resultados ao histórico para a IA ver
		resultsJSON, _ := json.Marshal(executionResults)
		toolResultMsg := fmt.Sprintf("TOOL RESULTS:\n%s\nContinue your task based on these results.", string(resultsJSON))

		s.chatHistory = append(s.chatHistory, domain.Message{
			Role:      domain.RoleSystem,
			Content:   toolResultMsg,
			Timestamp: time.Now(),
		})

		// Throttle para não estourar rate limit
		time.Sleep(2 * time.Second)
	}

	go s.saveCurrentConversation(contextStr)

	return finalResponse, nil
}

// executeToolCall executa uma ferramenta baseada no nome e argumentos
func (s *Service) executeToolCall(toolName string, args map[string]interface{}) (string, error) {
	if s.excelService == nil {
		return "", fmt.Errorf("excel service not connected")
	}

	// Mapear nome da ferramenta para operação
	switch toolName {
	// ===== QUERIES =====
	case "list_sheets":
		sheets, err := s.excelService.ListSheets()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Planilhas: %v", sheets), nil

	case "sheet_exists":
		name, _ := args["name"].(string)
		exists, err := s.excelService.SheetExists(name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Planilha '%s' existe: %v", name, exists), nil

	case "get_used_range":
		sheet, _ := args["sheet"].(string)
		rng, err := s.excelService.GetUsedRange(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Intervalo usado: %s", rng), nil

	case "get_headers":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		headers, err := s.excelService.GetHeaders(sheet, rng)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Cabeçalhos: %v", headers), nil

	case "get_range_values":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		values, err := s.excelService.GetRangeValues(sheet, rng)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Valores: %v", values), nil

	case "get_row_count":
		sheet, _ := args["sheet"].(string)
		count, err := s.excelService.GetRowCount(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Linhas: %d", count), nil

	case "get_column_count":
		sheet, _ := args["sheet"].(string)
		count, err := s.excelService.GetColumnCount(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Colunas: %d", count), nil

	case "get_cell_formula":
		sheet, _ := args["sheet"].(string)
		cell, _ := args["cell"].(string)
		formula, err := s.excelService.GetCellFormula(sheet, cell)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Fórmula: %s", formula), nil

	case "get_active_cell":
		cell, err := s.excelService.GetActiveCell()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Célula ativa: %s", cell), nil

	case "has_filter":
		sheet, _ := args["sheet"].(string)
		hasFilter, err := s.excelService.HasFilter(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Tem filtro: %v", hasFilter), nil

	case "list_charts":
		sheet, _ := args["sheet"].(string)
		charts, err := s.excelService.ListCharts(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Gráficos: %v", charts), nil

	case "list_tables":
		sheet, _ := args["sheet"].(string)
		tables, err := s.excelService.ListTables(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Tabelas: %v", tables), nil

	case "list_pivot_tables":
		sheet, _ := args["sheet"].(string)
		pivots, err := s.excelService.ListPivotTables(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Tabelas dinâmicas: %v", pivots), nil

	// ===== ACTIONS =====
	case "write_cell":
		sheet, _ := args["sheet"].(string)
		cell, _ := args["cell"].(string)
		value, _ := args["value"].(string)
		err := s.excelService.UpdateCell("", sheet, cell, value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Escrito '%s' em %s", value, cell), nil

	case "write_range":
		sheet, _ := args["sheet"].(string)
		cell, _ := args["cell"].(string)
		data, _ := args["data"].([]interface{})

		// Converter para [][]interface{}
		batchData := make([][]interface{}, len(data))
		for i, row := range data {
			if rowArr, ok := row.([]interface{}); ok {
				batchData[i] = rowArr
			} else {
				batchData[i] = []interface{}{row}
			}
		}

		err := s.excelService.WriteRange(sheet, cell, batchData)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Escrito %d linhas a partir de %s", len(batchData), cell), nil

	case "create_sheet":
		name, _ := args["name"].(string)
		err := s.excelService.CreateNewSheet(name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Planilha '%s' criada", name), nil

	case "delete_sheet":
		name, _ := args["name"].(string)
		err := s.excelService.DeleteSheet(name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Planilha '%s' excluída", name), nil

	case "rename_sheet":
		oldName, _ := args["old_name"].(string)
		newName, _ := args["new_name"].(string)
		err := s.excelService.RenameSheet(oldName, newName)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Planilha renomeada: %s -> %s", oldName, newName), nil

	case "format_range":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		bold, _ := args["bold"].(bool)
		italic, _ := args["italic"].(bool)
		fontSize := int(getFloat(args["font_size"]))
		fontColor, _ := args["font_color"].(string)
		bgColor, _ := args["bg_color"].(string)

		err := s.excelService.FormatRange(sheet, rng, bold, italic, fontSize, fontColor, bgColor)
		if err != nil {
			return "", err
		}
		return "Formatação aplicada", nil

	case "autofit_columns":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		err := s.excelService.AutoFitColumns(sheet, rng)
		if err != nil {
			return "", err
		}
		return "Colunas ajustadas", nil

	case "clear_range":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		err := s.excelService.ClearRange(sheet, rng)
		if err != nil {
			return "", err
		}
		return "Intervalo limpo", nil

	case "insert_rows":
		sheet, _ := args["sheet"].(string)
		row := int(getFloat(args["row"]))
		count := int(getFloat(args["count"]))
		err := s.excelService.InsertRows(sheet, row, count)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d linhas inseridas", count), nil

	case "delete_rows":
		sheet, _ := args["sheet"].(string)
		row := int(getFloat(args["row"]))
		count := int(getFloat(args["count"]))
		err := s.excelService.DeleteRows(sheet, row, count)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d linhas excluídas", count), nil

	case "merge_cells":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		err := s.excelService.MergeCells(sheet, rng)
		if err != nil {
			return "", err
		}
		return "Células mescladas", nil

	case "unmerge_cells":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		err := s.excelService.UnmergeCells(sheet, rng)
		if err != nil {
			return "", err
		}
		return "Mesclagem desfeita", nil

	case "set_borders":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		style, _ := args["style"].(string)
		err := s.excelService.SetBorders(sheet, rng, style)
		if err != nil {
			return "", err
		}
		return "Bordas aplicadas", nil

	case "set_column_width":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		width := getFloat(args["width"])
		err := s.excelService.SetColumnWidth(sheet, rng, width)
		if err != nil {
			return "", err
		}
		return "Largura definida", nil

	case "set_row_height":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		height := getFloat(args["height"])
		err := s.excelService.SetRowHeight(sheet, rng, height)
		if err != nil {
			return "", err
		}
		return "Altura definida", nil

	case "apply_filter":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		err := s.excelService.ApplyFilter(sheet, rng)
		if err != nil {
			return "", err
		}
		return "Filtro aplicado", nil

	case "clear_filters":
		sheet, _ := args["sheet"].(string)
		err := s.excelService.ClearFilters(sheet)
		if err != nil {
			return "", err
		}
		return "Filtros removidos", nil

	case "sort_range":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		col := int(getFloat(args["column"]))
		asc, _ := args["ascending"].(bool)
		err := s.excelService.SortRange(sheet, rng, col, asc)
		if err != nil {
			return "", err
		}
		return "Dados ordenados", nil

	case "copy_range":
		sheet, _ := args["sheet"].(string)
		src, _ := args["source"].(string)
		dest, _ := args["dest"].(string)
		err := s.excelService.CopyRange(sheet, src, dest)
		if err != nil {
			return "", err
		}
		return "Intervalo copiado", nil

	case "create_chart":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		chartType, _ := args["chart_type"].(string)
		title, _ := args["title"].(string)
		err := s.excelService.CreateChart(sheet, rng, chartType, title)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Gráfico '%s' criado", title), nil

	case "delete_chart":
		sheet, _ := args["sheet"].(string)
		name, _ := args["name"].(string)
		err := s.excelService.DeleteChart(sheet, name)
		if err != nil {
			return "", err
		}
		return "Gráfico excluído", nil

	case "create_pivot_table":
		srcSheet, _ := args["source_sheet"].(string)
		srcRange, _ := args["source_range"].(string)
		destSheet, _ := args["dest_sheet"].(string)
		destCell, _ := args["dest_cell"].(string)
		name, _ := args["name"].(string)
		err := s.excelService.CreatePivotTable(srcSheet, srcRange, destSheet, destCell, name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Tabela dinâmica '%s' criada", name), nil

	case "create_table":
		sheet, _ := args["sheet"].(string)
		rng, _ := args["range"].(string)
		name, _ := args["name"].(string)
		style, _ := args["style"].(string)
		err := s.excelService.CreateTable(sheet, rng, name, style)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Tabela '%s' criada", name), nil

	case "delete_table":
		sheet, _ := args["sheet"].(string)
		name, _ := args["name"].(string)
		err := s.excelService.DeleteTable(sheet, name)
		if err != nil {
			return "", err
		}
		return "Tabela removida", nil

	case "execute_macro":
		// Executar múltiplas ações em sequência
		actions, _ := args["actions"].([]interface{})
		if len(actions) == 0 {
			return "", fmt.Errorf("execute_macro requires 'actions' array")
		}

		s.excelService.StartUndoBatch()
		var results []string

		for i, action := range actions {
			actionMap, ok := action.(map[string]interface{})
			if !ok {
				results = append(results, fmt.Sprintf("Action %d: SKIP (invalid format)", i+1))
				continue
			}

			tool, _ := actionMap["tool"].(string)
			actionArgs, _ := actionMap["args"].(map[string]interface{})

			result, err := s.executeToolCall(tool, actionArgs)
			if err != nil {
				results = append(results, fmt.Sprintf("Action %d (%s): ERROR - %v", i+1, tool, err))
				break // Stop on error
			}
			results = append(results, fmt.Sprintf("Action %d (%s): %s", i+1, tool, result))
		}

		s.excelService.EndUndoBatch()
		return fmt.Sprintf("MACRO completada (%d ações): %v", len(actions), results), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
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
	// Also clear any pending action
	s.pendingAction = nil
}

// HasPendingAction returns true if there's a pending action waiting for confirmation
func (s *Service) HasPendingAction() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pendingAction != nil
}

// ConfirmPendingAction executes the pending action and resumes the AI loop
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

	// Execute the pending action usando novo sistema
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

	// Add execution results to chat history
	toolMsg := fmt.Sprintf("TOOL RESULTS:\n%s\nContinue your task based on these results.", executionResults)
	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleSystem,
		Content:   toolMsg,
		Timestamp: time.Now(),
	})

	// Resume the AI loop with function calling
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
	geminiTools := ai.GetGeminiTools()

	// Convert to AI messages and continue
	aiHistory := s.toAIMessages(s.chatHistory)

	var currentResponse string
	var toolCalls []ai.ToolCall
	chunkWrapper := func(chunk string) error {
		currentResponse += chunk
		return onChunk(chunk)
	}

	if s.provider == "google" {
		currentResponse, toolCalls, err = s.geminiClient.ChatStreamWithTools(ctx, aiHistory, geminiTools, chunkWrapper)
	} else {
		currentResponse, toolCalls, err = s.client.ChatStreamWithTools(ctx, aiHistory, tools, chunkWrapper)
	}

	if err != nil {
		return currentResponse, err
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

// RejectPendingAction discards the pending action
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
			fmt.Printf("[DEBUG refreshConfig] Provider do storage: %s, APIKey presente: %v, Model: %s, BaseURL: %s\n",
				cfg.Provider, cfg.APIKey != "", cfg.Model, cfg.BaseURL)

			s.provider = cfg.Provider
			if cfg.APIKey != "" {
				s.client.SetAPIKey(cfg.APIKey)
			}
			if cfg.Model != "" {
				s.client.SetModel(cfg.Model)
			}
			// Configurar limite de tokens (hardcoded por enquanto, mas aumentado)
			s.client.SetMaxInputTokens(50000) // Default segura para modelos modernos (GPT-4o, Claude)

			if cfg.BaseURL != "" {
				s.client.SetBaseURL(cfg.BaseURL)
			} else if cfg.Provider == "groq" {
				// Groq limits
				s.client.SetBaseURL("https://api.groq.com/openai/v1")
				s.client.SetMaxInputTokens(8000) // Llama 3 limit safe
			}

			if cfg.Provider == "google" {
				s.geminiClient.SetAPIKey(cfg.APIKey)
				s.geminiClient.SetModel(cfg.Model)
				// Gemini Flash has huge context.
				s.geminiClient.SetMaxInputTokens(100000) // 100k tokens safe for Flash

				if cfg.BaseURL != "" {
					s.geminiClient.SetBaseURL(cfg.BaseURL)
				}
			}
		}
	}
}

func (s *Service) ensureSystemPrompt() {
	systemPrompt := `Você é um AGENTE Excel inteligente com acesso a ferramentas para consultar e modificar planilhas do Microsoft Excel.

IDIOMA: SEMPRE responda em Português do Brasil.

MODO DE TRABALHO:
1. PRIMEIRO PASSO CRÍTICO: Antes de qualquer ação, use "list_sheets" para verificar se o Excel está conectado e quais planilhas existem. Se não houver planilhas, avise o usuário para abrir um arquivo Excel.
2. CONSULTE antes de AGIR: Use ferramentas de consulta (get_headers, get_used_range, etc.) para entender os dados antes de modificá-los.
3. Use "execute_macro" para múltiplas ações em sequência (criar planilha + escrever + formatar).

DICAS IMPORTANTES:
- Use autofit_columns após inserir dados para melhor visualização
- Use format_range para destacar cabeçalhos com negrito e cores
- Para fórmulas, use sintaxe PT-BR (SOMA, MÉDIA, PROCV) com ponto-e-vírgula como separador
- Em write_range, use array 2D: [["Col1", "Col2"], ["Val1", "Val2"]]
- Sempre especifique o parâmetro "sheet" em operações de escrita

MODO DE RACIOCÍNIO:
Ao fazer tarefas complexas, explique seu raciocínio passo a passo antes de executar.

NÃO gere código VBA. Use apenas as ferramentas disponíveis.`

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
