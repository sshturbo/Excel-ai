package chat

import (
	"excel-ai/internal/domain"
	"excel-ai/internal/dto"
)

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
		if m.Role == domain.RoleSystem {
			continue
		}
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

	go s.saveCurrentConversation("")
	return nil
}

// EditMessageAtVisibleIndex edita uma mensagem usando o índice visível (sem mensagens system)
// e remove todas as mensagens subsequentes (incluindo respostas da IA)
func (s *Service) EditMessageAtVisibleIndex(visibleIndex int, newContent string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Encontrar o índice real no chatHistory baseado no índice visível
	// Índice visível conta apenas mensagens user/assistant (não system)
	currentVisibleIndex := 0
	realIndex := -1

	for i, m := range s.chatHistory {
		if m.Role == domain.RoleSystem {
			continue // Pular mensagens de sistema
		}

		if currentVisibleIndex == visibleIndex {
			realIndex = i
			break
		}
		currentVisibleIndex++
	}

	if realIndex == -1 {
		return nil // Índice não encontrado
	}

	// Cortar o histórico até a mensagem editada (inclusive) e atualizar seu conteúdo
	// Manter mensagens system que vêm antes
	var newHistory []domain.Message
	for i := 0; i <= realIndex; i++ {
		if i == realIndex {
			// Atualizar conteúdo da mensagem editada
			newHistory = append(newHistory, domain.Message{
				Role:      s.chatHistory[i].Role,
				Content:   newContent,
				Timestamp: s.chatHistory[i].Timestamp,
			})
		} else {
			newHistory = append(newHistory, s.chatHistory[i])
		}
	}

	s.chatHistory = newHistory

	// Salvar conversa atualizada
	go s.saveCurrentConversation("")
	return nil
}
