package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"excel-ai/internal/domain"
	"excel-ai/pkg/ai"
)

// SendMessage envia mensagem para IA e gerencia o loop autônomo de execução com function calling nativo
func (s *Service) SendMessage(message string, contextStr string, askBeforeApply bool, onChunk func(string) error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar se deve usar orquestrador
	if s.useOrchestration {
		return s.orchestrator.OrchestrateMessage(message, contextStr, askBeforeApply, onChunk)
	}

	s.refreshConfig()

	// Verificar API key do cliente correto baseado no provider
	switch s.provider {
	case "google":
		if s.geminiClient.GetAPIKey() == "" {
			return "", fmt.Errorf("API key não configurada. Vá em Configurações e configure sua chave de API do Google")
		}
	case "zai":
		if s.zaiClient.GetBaseURL() == "" {
			return "", fmt.Errorf("API key não configurada. Vá em Configurações e configure sua chave de API do Z.AI")
		}
	default:
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

	// Detectar se é provider que pode retornar tool calls como texto (apenas custom agora)
	// Ollama agora usa cliente nativo que processa tool calls corretamente
	isTextToolCallProvider := s.provider == "custom"

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

		// Buffer para acumular resposta antes de enviar (para providers que retornam tool calls como texto)
		var responseBuffer strings.Builder
		var lastSentLength int

		chunkWrapper := func(chunk string) error {
			if isTextToolCallProvider {
				// Acumular no buffer e enviar apenas texto que não parece ser JSON de tool call
				responseBuffer.WriteString(chunk)
				fullText := responseBuffer.String()

				// Verificar se estamos no meio de um JSON (detectar { sem } correspondente)
				if !ai.IsPartialToolCallJSON(fullText) {
					// Enviar apenas a parte nova que não foi enviada ainda
					newContent := fullText[lastSentLength:]
					if newContent != "" {
						lastSentLength = len(fullText)
						return onChunk(newContent)
					}
				}
				return nil
			}
			// Provider normal - enviar direto
			currentResponse += chunk
			return onChunk(chunk)
		}

		// Chamar IA com tools baseado no provider
		switch s.provider {
		case "google":
			currentResponse, toolCalls, err = s.geminiClient.ChatStreamWithTools(ctx, aiHistory, geminiTools, chunkWrapper)
		case "zai":
			// Usar cliente nativo Z.AI
			currentResponse, toolCalls, err = s.zaiClient.ChatStreamWithTools(ctx, aiHistory, tools, chunkWrapper)
		case "ollama":
			// Usar cliente nativo Ollama para melhor suporte a tools
			currentResponse, toolCalls, err = s.ollamaClient.ChatStreamWithTools(ctx, aiHistory, tools, chunkWrapper)
		default:
			// OpenRouter, Groq, Custom
			currentResponse, toolCalls, err = s.client.ChatStreamWithTools(ctx, aiHistory, tools, chunkWrapper)
		}

		// Para providers de texto (apenas custom agora, Ollama usa cliente nativo)
		if s.provider == "custom" && isTextToolCallProvider {
			currentResponse = responseBuffer.String()
		}

		if err != nil {
			return finalResponse, err
		}

		// Filtrar tool calls inválidos (nome vazio ou null) que podem vir do Ollama
		toolCalls = ai.FilterValidToolCalls(toolCalls)

		// Se não recebemos tool calls estruturados válidos, tentar extrair do texto
		// Isso é necessário para Ollama e outros provedores que retornam tool calls como JSON no texto
		if len(toolCalls) == 0 && currentResponse != "" {
			extractedCalls, cleanedResponse := ai.ParseToolCallsFromText(currentResponse)
			if len(extractedCalls) > 0 {
				toolCalls = extractedCalls
				fmt.Printf("[DEBUG] Extracted %d tool call(s) from text response\n", len(extractedCalls))

				// Enviar resposta limpa para UI (já que o JSON não foi enviado durante streaming)
				if cleanedResponse != "" && cleanedResponse != currentResponse {
					// Enviar apenas a parte de texto limpa
					trimmedClean := strings.TrimSpace(cleanedResponse)
					if trimmedClean != "" {
						onChunk(trimmedClean)
					}
				}
				currentResponse = cleanedResponse
			} else if isTextToolCallProvider {
				// Limpar JSONs malformados do texto mesmo se não extraímos tool calls válidos
				_, cleanedResponse := ai.ParseToolCallsFromText(currentResponse)
				if cleanedResponse != currentResponse {
					currentResponse = cleanedResponse
				}
				// Enviar qualquer texto restante que não foi enviado durante streaming
				if lastSentLength < len(currentResponse) {
					remaining := currentResponse[lastSentLength:]
					if remaining != "" {
						onChunk(remaining)
					}
				}
			}
		}

		// Limpeza final para Ollama - remover qualquer JSON residual da resposta
		if s.provider == "ollama" && currentResponse != "" {
			_, cleanedResponse := ai.ParseToolCallsFromText(currentResponse)
			currentResponse = strings.TrimSpace(cleanedResponse)
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

		// Aplicar max_rows se especificado
		if maxRows, ok := args["max_rows"].(float64); ok && maxRows > 0 {
			limit := int(maxRows)
			if len(values) > limit {
				values = values[:limit]
			}
		}

		// Aplicar filtro se especificado
		if filterCol, ok := args["filter_column"].(string); ok && filterCol != "" {
			if filterVal, ok := args["filter_value"].(string); ok && filterVal != "" {
				values = s.filterValues(values, filterCol, filterVal)
			}
		}

		return fmt.Sprintf("Valores (%d linhas): %v", len(values), values), nil

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

	case "query_batch":
		sheet, _ := args["sheet"].(string)
		queries, _ := args["queries"].([]interface{})
		sampleRows := 5
		if sr, ok := args["sample_rows"].(float64); ok {
			sampleRows = int(sr)
		}

		var results []string
		for _, q := range queries {
			queryName, _ := q.(string)
			switch queryName {
			case "headers":
				if h, err := s.excelService.GetHeaders(sheet, "A:Z"); err == nil {
					results = append(results, fmt.Sprintf("Cabeçalhos: %v", h))
				}
			case "row_count":
				if count, err := s.excelService.GetRowCount(sheet); err == nil {
					results = append(results, fmt.Sprintf("Linhas: %d", count))
				}
			case "used_range":
				if rng, err := s.excelService.GetUsedRange(sheet); err == nil {
					results = append(results, fmt.Sprintf("Intervalo: %s", rng))
				}
			case "column_count":
				if count, err := s.excelService.GetColumnCount(sheet); err == nil {
					results = append(results, fmt.Sprintf("Colunas: %d", count))
				}
			case "sample_data":
				rng := fmt.Sprintf("A1:Z%d", sampleRows+1)
				if vals, err := s.excelService.GetRangeValues(sheet, rng); err == nil {
					results = append(results, fmt.Sprintf("Amostra (%d linhas): %v", len(vals), vals))
				}
			case "has_filter":
				if hasFilter, err := s.excelService.HasFilter(sheet); err == nil {
					results = append(results, fmt.Sprintf("Tem filtro: %v", hasFilter))
				}
			case "charts":
				if charts, err := s.excelService.ListCharts(sheet); err == nil {
					results = append(results, fmt.Sprintf("Gráficos: %v", charts))
				}
			case "tables":
				if tables, err := s.excelService.ListTables(sheet); err == nil {
					results = append(results, fmt.Sprintf("Tabelas: %v", tables))
				}
			}
		}
		return strings.Join(results, "\n"), nil

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

	// Add execution results to chat history (Hidden = não aparece no chat)
	toolMsg := fmt.Sprintf("Resultados das ferramentas executadas:\n%s\n\nAgora responda ao usuário ou execute a próxima ação necessária.", executionResults)
	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleUser,
		Content:   toolMsg,
		Timestamp: time.Now(),
		Hidden:    true,
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

	switch s.provider {
	case "google":
		currentResponse, toolCalls, err = s.geminiClient.ChatStreamWithTools(ctx, aiHistory, geminiTools, chunkWrapper)
	case "zai":
		currentResponse, toolCalls, err = s.zaiClient.ChatStreamWithTools(ctx, aiHistory, tools, chunkWrapper)
	default:
		currentResponse, toolCalls, err = s.client.ChatStreamWithTools(ctx, aiHistory, tools, chunkWrapper)
	}

	if err != nil {
		return currentResponse, err
	}

	// Se não recebemos tool calls estruturados, tentar extrair do texto
	// Isso é necessário para Ollama e outros provedores que retornam tool calls como JSON no texto
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
			fmt.Printf("[DEBUG refreshConfig] Provider do storage: %s, APIKey presente: %v, Model: %s, ToolModel: %s, BaseURL: %s\n",
				cfg.Provider, cfg.APIKey != "", cfg.Model, cfg.ToolModel, cfg.BaseURL)

			s.provider = cfg.Provider
			if cfg.APIKey != "" {
				s.client.SetAPIKey(cfg.APIKey)
			}
			if cfg.Model != "" {
				s.client.SetModel(cfg.Model)
			}
			// Configurar limite de tokens (hardcoded por enquanto, mas aumentado)
			s.client.SetMaxInputTokens(50000) // Default segura para modelos modernos (GPT-4o, Claude)

			// Configuração específica por provider
			if cfg.Provider == "ollama" {
				// Ollama - usar cliente nativo para melhor suporte a tools
				baseURL := cfg.BaseURL
				if baseURL == "" {
					baseURL = "http://localhost:11434"
				}
				s.ollamaClient.SetBaseURL(baseURL)
				s.ollamaClient.SetModel(cfg.Model)

				// Também configurar client OpenAI-compatible como fallback
				s.client.SetBaseURL(baseURL + "/v1")
				s.client.SetMaxInputTokens(32000)
				if cfg.APIKey == "" {
					s.client.SetAPIKey("ollama")
				}

				// Verificar se modelo é adequado para tools
				suitable, warning := ai.IsModelSuitableForTools(cfg.Model)
				if !suitable {
					fmt.Printf("[OLLAMA WARNING] %s\n", warning)
				} else if warning != "" {
					fmt.Printf("[OLLAMA INFO] %s\n", warning)
				}
			} else if cfg.Provider == "google" {
				if cfg.APIKey != "" {
					s.geminiClient.SetAPIKey(cfg.APIKey)
				}
				if cfg.Model != "" {
					s.geminiClient.SetModel(cfg.Model)
				}
				// Gemini Flash has huge context.
				s.geminiClient.SetMaxInputTokens(100000) // 100k tokens safe for Flash

				if cfg.BaseURL != "" {
					s.geminiClient.SetBaseURL(cfg.BaseURL)
				}
			} else if cfg.Provider == "zai" {
				// Z.AI (GLM Models) - Usar cliente nativo com Coding API
				if cfg.APIKey != "" {
					s.zaiClient.SetAPIKey(cfg.APIKey)
				}
				if cfg.ToolModel != "" {
					s.zaiClient.SetModel(cfg.ToolModel)
				} else if cfg.Model != "" {
					s.zaiClient.SetModel(cfg.Model)
				}

				// SEMPRE usar Coding API - ignorar BaseURL do storage se for a URL antiga
				baseURL := cfg.BaseURL
				if baseURL == "" || baseURL == "https://api.z.ai/api/paas/v4" || baseURL == "https://api.z.ai/api/paas/v4/" {
					baseURL = "https://api.z.ai/api/coding/paas/v4/"
				}
				// Garantir que a URL termina com barra (requerido pela API Z.AI)
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
			} else if cfg.Provider == "groq" {
				s.client.SetBaseURL("https://api.groq.com/openai/v1")
				s.client.SetMaxInputTokens(8000) // Llama 3 limit safe
			} else if cfg.BaseURL != "" {
				// Custom ou outros providers com BaseURL
				s.client.SetBaseURL(cfg.BaseURL)
			}
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

	// Se for Ollama, simplificar ainda mais
	if s.provider == "ollama" {
		systemPrompt = `Você é um AGENTE Excel profissional. 
Responda em Português do Brasil de forma direta e objetiva.
NUNCA use blocos de "thinking" ou explicações desnecessárias.
Use ferramentas apenas quando solicitado.
Para saudações, seja breve e educado.`
	}

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
