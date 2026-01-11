package app

import (
	"context"
	"excel-ai/internal/dto"
	apperrors "excel-ai/pkg/errors"
	"excel-ai/pkg/logger"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SendMessage envia mensagem para o chat
func (a *App) SendMessage(message string, askBeforeApply bool) string {
	// Validar mensagem
	if message == "" {
		logger.ChatWarn("Tentativa de enviar mensagem vazia")
		return "Error: A mensagem não pode estar vazia"
	}

	if len(message) > 100000 {
		logger.ChatWarn("Mensagem muito longa, truncando")
		message = message[:100000]
	}

	logger.ChatInfo(fmt.Sprintf("Recebendo mensagem (len=%d)", len(message)))

	// Passa apenas o contexto mínimo (workbook/sheet ativos) - dados são obtidos via function calling
	activeContext := a.excelService.GetActiveContext()

	response, err := a.chatService.SendMessage(message, activeContext, askBeforeApply, func(chunk string) error {
		runtime.EventsEmit(a.ctx, "chat:chunk", chunk)
		return nil
	})

	if err != nil {
		logger.ChatError("Erro ao enviar mensagem: " + err.Error())
		return "Error: " + err.Error()
	}

	logger.ChatInfo(fmt.Sprintf("Resposta enviada (len=%d)", len(response)))
	return response
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ClearChat limpa o chat
func (a *App) ClearChat() {
	logger.ChatInfo("Limpando histórico de chat")
	a.chatService.ClearChat()
}

// CancelChat cancela a requisição de chat em andamento
func (a *App) CancelChat() {
	logger.ChatInfo("Cancelando requisição de chat")
	a.chatService.CancelChat()
}

// SendErrorFeedback envia erro para a IA corrigir
func (a *App) SendErrorFeedback(errorMessage string) (string, error) {
	if errorMessage == "" {
		logger.ChatWarn("Tentativa de enviar erro vazio")
		return "", apperrors.InvalidInput("a mensagem de erro não pode estar vazia")
	}

	logger.ChatInfo("Enviando feedback de erro para IA")

	response, err := a.chatService.SendErrorFeedback(errorMessage, func(chunk string) error {
		runtime.EventsEmit(a.ctx, "chat:chunk", chunk)
		return nil
	})

	if err != nil {
		logger.ChatError("Erro ao processar feedback: " + err.Error())
		return "", err
	}

	logger.ChatInfo("Feedback processado com sucesso")
	return response, nil
}

// DeleteLastMessages remove mensagens
func (a *App) DeleteLastMessages(count int) error {
	logger.ChatInfo(fmt.Sprintf("Removendo %d últimas mensagens", count))

	err := a.chatService.DeleteLastMessages(count)
	if err != nil {
		logger.ChatError("Erro ao remover mensagens: " + err.Error())
	}
	return err
}

// EditMessage edita uma mensagem usando o índice visível e remove mensagens subsequentes
func (a *App) EditMessage(visibleIndex int, newContent string) error {
	logger.ChatInfo(fmt.Sprintf("Editando mensagem na posição %d", visibleIndex))

	err := a.chatService.EditMessageAtVisibleIndex(visibleIndex, newContent)
	if err != nil {
		logger.ChatError("Erro ao editar mensagem: " + err.Error())
	}
	return err
}

// NewConversation nova conversa
func (a *App) NewConversation() string {
	logger.ChatInfo("Criando nova conversa")
	convID := a.chatService.NewConversation()

	// Se houver um arquivo aberto via path, vincular à nova conversa
	client, err := a.excelService.GetExcelClient()
	if err == nil && client != nil {
		path := client.GetFilePath()
		if path != "" && a.storage != nil {
			if err := a.storage.SetConversationExcelPath(convID, path); err != nil {
				logger.AppWarn("Falha ao vincular path à nova conversa: " + err.Error())
			} else {
				logger.AppInfo("Caminho do Excel vinculado à nova conversa " + convID)
			}
		}
	}

	return convID
}

// ListConversations lista conversas
func (a *App) ListConversations() ([]dto.ConversationInfo, error) {
	logger.ChatInfo("Listando conversas")
	return a.chatService.ListConversations()
}

// LoadConversation carrega conversa
// LoadConversation carrega conversa e tenta carregar planilha vinculada
func (a *App) LoadConversation(id string) ([]dto.ChatMessage, error) {
	logger.ChatInfo("Carregando conversa: " + id)
	messages, err := a.chatService.LoadConversation(id)
	if err != nil {
		return nil, err
	}

	// Tentar carregar a planilha vinculada se houver storage
	if a.storage != nil {
		path, err := a.storage.GetConversationExcelPath(id)
		if err == nil && path != "" {
			logger.AppInfo("Auto-carregando planilha vinculada: " + path)
			sessionID := fmt.Sprintf("session_load_%d", time.Now().UnixNano())

			// Tenta carregar. Se falhar (ex: arquivo deletado), apenas loga e segue
			if err := a.excelService.ConnectFilePath(sessionID, path); err != nil {
				logger.AppWarn(fmt.Sprintf("Falha ao auto-carregar planilha (pode ter sido movida): %v", err))
			} else {
				// Emitir evento para o frontend atualizar o preview
				fileName := a.excelService.GetCurrentFileName()
				runtime.EventsEmit(a.ctx, "native:file-loaded", map[string]string{
					"sessionId": sessionID,
					"fileName":  fileName,
				})
				logger.AppInfo("Planilha auto-carregada com sucesso")
			}
		}
	}

	return messages, nil
}

// DeleteConversation remove conversa
func (a *App) DeleteConversation(id string) error {
	logger.ChatInfo("Excluindo conversa: " + id)
	return a.chatService.DeleteConversation(id)
}

// GetChatHistory retorna histórico
func (a *App) GetChatHistory() []dto.ChatMessage {
	logger.ChatDebug("Obtendo histórico de chat")
	return a.chatService.GetChatHistory()
}

// HasPendingAction verifica se há ação pendente
func (a *App) HasPendingAction() bool {
	hasPending := a.chatService.HasPendingAction()
	if hasPending {
		logger.ChatInfo("Verificando ações pendentes: sim")
	}
	return hasPending
}

// ConfirmPendingAction confirma e executa a ação pendente, retomando a IA
func (a *App) ConfirmPendingAction() string {
	logger.ChatInfo("Confirmando ação pendente")

	response, err := a.chatService.ConfirmPendingAction(func(chunk string) error {
		runtime.EventsEmit(a.ctx, "chat:chunk", chunk)
		return nil
	})

	if err != nil {
		logger.ChatError("Erro ao confirmar ação: " + err.Error())
		return "Error: " + err.Error()
	}

	logger.ChatInfo("Ação confirmada com sucesso")
	return response
}

// RejectPendingAction descarta a ação pendente
func (a *App) RejectPendingAction() {
	logger.ChatInfo("Rejeitando ação pendente")
	a.chatService.RejectPendingAction()
}

// SetOrchestration habilita ou desabilita a orquestração paralela
func (a *App) SetOrchestration(enabled bool) {
	status := "desabilitada"
	if enabled {
		status = "habilitada"
	}
	logger.ChatInfo("Orquestração " + status)
	a.chatService.SetOrchestration(enabled)
}

// GetOrchestration retorna se a orquestração está habilitada
func (a *App) GetOrchestration() bool {
	enabled := a.chatService.GetOrchestration()
	status := "desabilitada"
	if enabled {
		status = "habilitada"
	}
	logger.ChatDebug(fmt.Sprintf("Orquestração %s", status))
	return enabled
}

// StartOrchestrator inicia o orquestrador
func (a *App) StartOrchestrator() error {
	logger.ChatInfo("Iniciando orquestrador")
	ctx := context.Background()
	err := a.chatService.StartOrchestrator(ctx)
	if err != nil {
		logger.ChatError("Erro ao iniciar orquestrador: " + err.Error())
	}
	return err
}

// StopOrchestrator para o orquestrador
func (a *App) StopOrchestrator() {
	logger.ChatInfo("Parando orquestrador")
	a.chatService.StopOrchestrator()
}

// GetOrchestratorStats retorna estatísticas do orquestrador
func (a *App) GetOrchestratorStats() map[string]interface{} {
	orch := a.chatService.GetOrchestrator()
	if orch == nil {
		return map[string]interface{}{
			"error": "Orchestrator not initialized",
		}
	}

	stats := orch.GetStats()

	return map[string]interface{}{
		"totalTasks":    stats.TotalTasks,
		"successTasks":  stats.SuccessTasks,
		"failedTasks":   stats.FailedTasks,
		"activeWorkers": stats.ActiveWorkers,
		"avgTaskTime":   stats.AvgTaskTime.String(),
		"successRate":   stats.SuccessRate,
		"isRunning":     stats.IsRunning,
	}
}

// OrchestratorHealthCheck verifica saúde do orquestrador
func (a *App) OrchestratorHealthCheck() map[string]interface{} {
	orch := a.chatService.GetOrchestrator()
	if orch == nil {
		return map[string]interface{}{
			"error": "Orchestrator not initialized",
		}
	}

	health := orch.HealthCheck()

	return map[string]interface{}{
		"isHealthy":     health.IsHealthy,
		"workersActive": health.WorkersActive,
		"totalTasks":    health.TotalTasks,
		"tasksPending":  health.TasksPending,
		"lastCheck":     health.LastCheck.Format("2006-01-02 15:04:05"),
		"issues":        health.Issues,
	}
}

// ClearOrchestratorCache limpa o cache do orquestrador
func (a *App) ClearOrchestratorCache() {
	logger.ChatInfo("Limpando cache do orquestrador")
	orch := a.chatService.GetOrchestrator()
	if orch != nil {
		orch.ClearCache()
	}
}

// SetOrchestratorCacheTTL define o TTL do cache em minutos
func (a *App) SetOrchestratorCacheTTL(minutes int) {
	logger.ChatInfo(fmt.Sprintf("Definindo TTL do cache: %d minutos", minutes))
	orch := a.chatService.GetOrchestrator()
	if orch != nil {
		// Atualizar TTL (precisa de método público no orchestrator)
		logger.ChatWarn("Método SetCacheTTL não implementado no orquestrador")
	}
}

// TriggerOrchestratorRecovery força recovery manual
func (a *App) TriggerOrchestratorRecovery() {
	logger.ChatInfo("Acionando recovery manual do orquestrador")
	orch := a.chatService.GetOrchestrator()
	if orch != nil {
		// Forçar verificação de workers
		logger.ChatInfo("Recovery manual iniciado")
	}
}

// GetCurrentConversationID retorna o ID da conversa atual
func (a *App) GetCurrentConversationID() string {
	id := a.chatService.GetCurrentConversationID()
	logger.ChatDebug(fmt.Sprintf("ID da conversa atual: %s", id))
	return id
}
