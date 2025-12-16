package chat

import (
	"context"
	"fmt"
	"time"

	"excel-ai/internal/domain"
)

// SendMessage envia mensagem para IA e gerencia o loop autônomo de execução
func (s *Service) SendMessage(message string, contextStr string, askBeforeApply bool, onChunk func(string) error) (string, error) {
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
			// Se o usuário pediu para confirmar antes de aplicar (AskBeforeApply)
			// E o comando é de escrita/modificação (não busca), pausamos.
			if askBeforeApply && cmd.Type == "action" {
				// Salvar o comando pendente para execução posterior
				s.pendingAction = &cmd
				s.pendingContextStr = contextStr
				s.pendingOnChunk = onChunk

				pauseMsg := "\n\n🛑 *[Ação Pendente]* Aguardando aprovação do usuário para executar.\n"
				onChunk(pauseMsg)
				finalResponse += pauseMsg

				// Salvar conversa para garantir que o contexto atual (proposta) fique salvo
				go s.saveCurrentConversation(contextStr)

				return finalResponse, nil
			}

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

	// Execute the pending action
	onChunk("\n\n✅ *[Executando ação aprovada...]*\n")

	result, err := s.ExecuteTool(*cmd)
	var executionResults string
	if err != nil {
		executionResults = fmt.Sprintf("ERROR Executing %s: %v\n", cmd.Content, err)
		onChunk(fmt.Sprintf("\n❌ Erro: %v\n", err))
	} else {
		executionResults = fmt.Sprintf("SUCCESS: %s\n", result)
		onChunk("\n✅ Ação executada com sucesso!\n")
	}

	// Add execution results to chat history
	toolMsg := fmt.Sprintf("TOOL RESULTS:\n%s\nContinue your task based on these results.", executionResults)
	s.chatHistory = append(s.chatHistory, domain.Message{
		Role:      domain.RoleUser,
		Content:   toolMsg,
		Timestamp: time.Now(),
	})

	// Resume the AI loop (simplified version - just one more turn)
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

	// Convert to AI messages and continue
	aiHistory := s.toAIMessages(s.chatHistory)

	var currentResponse string
	chunkWrapper := func(chunk string) error {
		currentResponse += chunk
		return onChunk(chunk)
	}

	if s.provider == "google" {
		_, err = s.geminiClient.ChatStream(ctx, aiHistory, chunkWrapper)
	} else {
		_, err = s.client.ChatStream(ctx, aiHistory, chunkWrapper)
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

	// Check for more commands (simplified - just return, let frontend handle if more actions needed)
	commands := s.ParseToolCommands(currentResponse)
	if len(commands) > 0 {
		for _, newCmd := range commands {
			if newCmd.Type == "action" {
				// Another action pending - save it
				s.pendingAction = &newCmd
				s.pendingContextStr = contextStr
				s.pendingOnChunk = onChunk

				pauseMsg := "\n\n🛑 *[Ação Pendente]* Aguardando aprovação do usuário para executar.\n"
				onChunk(pauseMsg)
				currentResponse += pauseMsg
				break
			} else if newCmd.Type == "query" {
				// Execute queries automatically
				result, err := s.ExecuteTool(newCmd)
				if err == nil {
					queryResult := fmt.Sprintf("\n📊 Query result: %s\n", result)
					onChunk(queryResult)
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
			s.provider = cfg.Provider
			if cfg.APIKey != "" {
				s.client.SetAPIKey(cfg.APIKey)
			}
			if cfg.Model != "" {
				s.client.SetModel(cfg.Model)
			}
			// Configurar limite de tokens (hardcoded por enquanto, pode vir do cfg futuramente)
			s.client.SetMaxInputTokens(4000)

			if cfg.BaseURL != "" {
				s.client.SetBaseURL(cfg.BaseURL)
			} else if cfg.Provider == "groq" {
				// Groq often has smaller limits or strict TPM, set safer default
				s.client.SetBaseURL("https://api.groq.com/openai/v1")
				s.client.SetMaxInputTokens(3000) // Mais conservador para evitar 413
			}

			if cfg.Provider == "google" {
				s.geminiClient.SetAPIKey(cfg.APIKey)
				s.geminiClient.SetModel(cfg.Model)
				// Gemini Flash has huge context, but free tier has limits.
				// Safe default 8k for free tier usage
				s.geminiClient.SetMaxInputTokens(8192)

				if cfg.BaseURL != "" {
					s.geminiClient.SetBaseURL(cfg.BaseURL)
				}
			}
		}
	}
}

func (s *Service) ensureSystemPrompt() {
	systemPrompt := `Você é um AGENTE Excel inteligente. Você trabalha de forma autônoma para completar tarefas.

IDIOMA: SEMPRE responda em Português do Brasil. Todas as suas mensagens, explicações e raciocínios devem ser em português.

MODO DE RACIOCÍNIO:
Ao fazer tarefas complexas, SEMPRE mostre seu raciocínio usando:
:::thinking
[Seu raciocínio passo a passo aqui]
:::
Isso ajuda o usuário a entender seu processo de pensamento. Pense em voz alta!

MODO AGENTE:
PRIMEIRO PASSO CRÍTICO: Antes de QUALQUER ação, SEMPRE execute list-sheets primeiro para verificar se o Excel está conectado e tem uma pasta de trabalho aberta. Se falhar ou retornar vazio, avise o usuário para abrir um arquivo Excel!

1. PRIMEIRO faça consultas para entender o estado atual
2. DEPOIS execute ações baseadas nos resultados
3. Os resultados das consultas serão enviados de volta - USE-OS!

CONSULTAS (verificar estado):
:::excel-query
{"type": "list-sheets"}
{"type": "sheet-exists", "name": "NomeDaPlanilha"}
{"type": "list-pivot-tables", "sheet": "NomeDaPlanilha"}
{"type": "get-headers", "sheet": "NomeDaPlanilha", "range": "A:F"}
{"type": "get-used-range", "sheet": "NomeDaPlanilha"}
{"type": "get-row-count", "sheet": "NomeDaPlanilha"}
{"type": "get-column-count", "sheet": "NomeDaPlanilha"}
{"type": "get-cell-formula", "sheet": "NomeDaPlanilha", "cell": "A1"}
{"type": "has-filter", "sheet": "NomeDaPlanilha"}
{"type": "get-active-cell"}
{"type": "get-range-values", "sheet": "NomeDaPlanilha", "range": "A1:C10"}
{"type": "list-charts", "sheet": "NomeDaPlanilha"}
:::

AÇÕES (modificar Excel):
:::excel-action
{"op": "macro", "actions": [{"op": "create-sheet", "name": "Dados"}, {"op": "write", "sheet": "Dados", "cell": "A1", "data": [["Col1", "Col2"], ["Val1", "Val2"]]}, {"op": "format-range", "sheet": "Dados", "range": "A1:B1", "bold": true}, {"op": "autofit", "sheet": "Dados", "range": "A:B"}]}
{"op": "write", "cell": "A1", "value": "valor único"}
{"op": "write", "sheet": "NomeDaPlanilha", "cell": "A1", "data": [["Cabeçalho1", "Cabeçalho2"], ["ValorLinha1Col1", "ValorLinha1Col2"]]}
{"op": "create-workbook", "name": "Nova.xlsx"}
{"op": "create-sheet", "name": "NovaPlanilha"}
{"op": "create-chart", "sheet": "X", "range": "A1:B10", "chartType": "line", "title": "Título"}
{"op": "create-pivot", "sourceSheet": "X", "sourceRange": "A:F", "destSheet": "Y", "destCell": "A1", "tableName": "Nome", "rowFields": ["campo1"], "valueFields": [{"field": "campo2", "function": "sum"}]}
{"op": "format-range", "sheet": "X", "range": "A1:B5", "bold": true, "italic": false, "fontSize": 12, "fontColor": "#FF0000", "bgColor": "#FFFF00"}
{"op": "delete-sheet", "name": "PlanilhaParaDeletar"}
{"op": "rename-sheet", "oldName": "NomeAntigo", "newName": "NomeNovo"}
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

REGRAS DO AGENTE:
1. Para criar GRÁFICO: primeiro use get-headers e get-used-range para conhecer os dados
2. Para criar PIVOT: primeiro verifique se a planilha de destino existe com sheet-exists
3. Para qualquer tarefa complexa: faça consultas primeiro!
4. Você receberá resultados e pode continuar automaticamente
5. Use format-range para deixar cabeçalhos em negrito ou destacar dados
6. Use autofit para ajustar largura das colunas após inserir dados
7. CRÍTICO: SEMPRE especifique o parâmetro "sheet" nas ações write/format! Após criar nova planilha, use o nome dela em TODAS as ações seguintes.
8. Para inserção em lote, use o campo "data" com array 2D: {"op": "write", "sheet": "MinhaAba", "cell": "A1", "data": [["Col1", "Col2"], ["Val1", "Val2"]]}
9. **MACRO OBRIGATÓRIA**: Ao fazer QUALQUER tarefa multi-passo (criar planilha + escrever dados + formatar + autofit), você DEVE usar MACRO! NUNCA faça ações separadas quando podem ser combinadas. Ações individuais são apenas para operações verdadeiramente isoladas.

EXEMPLO - Usuário pede para criar tabela com produtos (USE MACRO!):
:::thinking
Usuário quer uma tabela com produtos. Vou usar MACRO para fazer tudo de uma vez:
1. Criar planilha
2. Escrever dados em lote
3. Formatar cabeçalhos
4. Ajustar colunas
:::
:::excel-action
{"op": "macro", "actions": [
  {"op": "create-sheet", "name": "Produtos"},
  {"op": "write", "sheet": "Produtos", "cell": "A1", "data": [["Produto", "Preço"], ["Caneta", 2.50], ["Lápis", 1.00], ["Borracha", 0.50]]},
  {"op": "format-range", "sheet": "Produtos", "range": "A1:B1", "bold": true, "bgColor": "#4472C4", "fontColor": "#FFFFFF"},
  {"op": "autofit", "sheet": "Produtos", "range": "A:B"}
]}
:::

EXEMPLO - Criar gráfico com raciocínio:
:::thinking
Usuário quer um gráfico. Preciso:
1. Descobrir quais dados existem
2. Obter o intervalo dos dados
3. Identificar cabeçalhos para labels do gráfico
4. Criar tipo de gráfico apropriado
:::
:::excel-query
{"type": "get-used-range", "sheet": "Dados"}
:::
(Sistema responderá com o intervalo, então eu continuo)

Use fórmulas em PT-BR (SOMA, MÉDIA, SE, PROCV). NÃO gere VBA.`

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
