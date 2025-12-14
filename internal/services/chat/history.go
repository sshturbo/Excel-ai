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

	go s.saveCurrentConversation("") // Context might be lost here if not stored in service state
	return nil
}
