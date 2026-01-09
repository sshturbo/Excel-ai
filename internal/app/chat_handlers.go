package app

import (
	"context"
	"excel-ai/internal/dto"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SendMessage envia mensagem para o chat
func (a *App) SendMessage(message string, askBeforeApply bool) string {
	fmt.Printf("[APP SendMessage] INÍCIO - Mensagem: %s (len=%d)\n", message[:min(50, len(message))], len(message))

	// Passa apenas o contexto mínimo (workbook/sheet ativos) - dados são obtidos via function calling
	activeContext := a.excelService.GetActiveContext()

	response, err := a.chatService.SendMessage(message, activeContext, askBeforeApply, func(chunk string) error {
		runtime.EventsEmit(a.ctx, "chat:chunk", chunk)
		return nil
	})

	if err != nil {
		fmt.Printf("[APP SendMessage] ERRO: %v\n", err)
		return "Error: " + err.Error()
	}

	fmt.Printf("[APP SendMessage] FIM - Response len: %d\n", len(response))
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
	a.chatService.ClearChat()
}

// CancelChat cancela a requisição de chat em andamento
func (a *App) CancelChat() {
	a.chatService.CancelChat()
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

// EditMessage edita uma mensagem usando o índice visível e remove mensagens subsequentes
func (a *App) EditMessage(visibleIndex int, newContent string) error {
	return a.chatService.EditMessageAtVisibleIndex(visibleIndex, newContent)
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

// HasPendingAction verifica se há ação pendente
func (a *App) HasPendingAction() bool {
	return a.chatService.HasPendingAction()
}

// ConfirmPendingAction confirma e executa a ação pendente, retomando a IA
func (a *App) ConfirmPendingAction() string {
	response, err := a.chatService.ConfirmPendingAction(func(chunk string) error {
		runtime.EventsEmit(a.ctx, "chat:chunk", chunk)
		return nil
	})

	if err != nil {
		return "Error: " + err.Error()
	}
	return response
}

// RejectPendingAction descarta a ação pendente
func (a *App) RejectPendingAction() {
	a.chatService.RejectPendingAction()
}

// SetOrchestration habilita ou desabilita a orquestração paralela
func (a *App) SetOrchestration(enabled bool) {
	a.chatService.SetOrchestration(enabled)
}

// GetOrchestration retorna se a orquestração está habilitada
func (a *App) GetOrchestration() bool {
	return a.chatService.GetOrchestration()
}

// StartOrchestrator inicia o orquestrador
func (a *App) StartOrchestrator() error {
	ctx := context.Background()
	return a.chatService.StartOrchestrator(ctx)
}

// StopOrchestrator para o orquestrador
func (a *App) StopOrchestrator() {
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
	orch := a.chatService.GetOrchestrator()
	if orch != nil {
		orch.ClearCache()
	}
}

// SetOrchestratorCacheTTL define o TTL do cache em minutos
func (a *App) SetOrchestratorCacheTTL(minutes int) {
	orch := a.chatService.GetOrchestrator()
	if orch != nil {
		// Atualizar TTL (precisa de método público no orchestrator)
		// Por enquanto, apenas log
		fmt.Printf("[APP] Cache TTL solicitado: %d minutos\n", minutes)
	}
}

// TriggerOrchestratorRecovery força recovery manual
func (a *App) TriggerOrchestratorRecovery() {
	orch := a.chatService.GetOrchestrator()
	if orch != nil {
		// Forçar verificação de workers
		fmt.Println("[APP] Recovery manual acionado")
	}
}

// GetCurrentConversationID retorna o ID da conversa atual
func (a *App) GetCurrentConversationID() string {
	return a.chatService.GetCurrentConversationID()
}
