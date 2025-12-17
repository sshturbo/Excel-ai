package app

import (
	"excel-ai/internal/dto"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SendMessage envia mensagem para o chat
func (a *App) SendMessage(message string, askBeforeApply bool) string {
	// Contexto do Excel agora é obtido sob demanda via function calling
	response, err := a.chatService.SendMessage(message, "", askBeforeApply, func(chunk string) error {
		runtime.EventsEmit(a.ctx, "chat:chunk", chunk)
		return nil
	})

	if err != nil {
		return "Error: " + err.Error()
	}
	return response
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

// GetCurrentConversationID retorna o ID da conversa atual
func (a *App) GetCurrentConversationID() string {
	return a.chatService.GetCurrentConversationID()
}
