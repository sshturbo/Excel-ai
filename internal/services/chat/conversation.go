package chat

import (
	"fmt"
	"strings"
	"time"

	"excel-ai/internal/domain"
	"excel-ai/internal/dto"
	"excel-ai/pkg/storage"
)

func (s *Service) NewConversation() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Printf("[DEBUG] NewConversation chamado. Histórico atual: %d mensagens\n", len(s.chatHistory))

	if s.currentConvID != "" && len(s.chatHistory) > 0 {
		s.saveCurrentConversation("")
	}

	s.chatHistory = []domain.Message{}
	s.currentConvID = s.generateID()

	fmt.Printf("[DEBUG] Histórico LIMPO. Novo ID: %s, mensagens: %d\n", s.currentConvID, len(s.chatHistory))

	return s.currentConvID
}

// GetCurrentConversationID retorna o ID da conversa atual
func (s *Service) GetCurrentConversationID() string {
	return s.currentConvID
}

func (s *Service) ListConversations() ([]dto.ConversationInfo, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}

	summaries, err := s.storage.ListConversations()
	if err != nil {
		return nil, err
	}

	var result []dto.ConversationInfo
	for _, summary := range summaries {
		result = append(result, dto.ConversationInfo{
			ID:        summary.ID,
			Title:     summary.Title,
			Preview:   summary.Preview,
			UpdatedAt: summary.UpdatedAt.Format("02/01/2006 15:04"),
		})
	}
	return result, nil
}

func (s *Service) LoadConversation(id string) ([]dto.ChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.storage == nil {
		return nil, fmt.Errorf("storage não disponível")
	}

	conv, err := s.storage.LoadConversation(id)
	if err != nil {
		return nil, err
	}

	s.currentConvID = conv.ID
	s.chatHistory = []domain.Message{}

	var result []dto.ChatMessage
	for _, m := range conv.Messages {
		domainMsg := domain.Message{
			Role:      domain.MessageRole(m.Role),
			Content:   m.Content,
			Timestamp: m.Timestamp,
			Hidden:    m.Hidden,
		}
		s.chatHistory = append(s.chatHistory, domainMsg)

		// Filter out system messages, hidden messages, and tool results from the UI
		if domain.MessageRole(m.Role) == domain.RoleSystem {
			continue
		}
		if m.Hidden {
			continue
		}
		if strings.HasPrefix(m.Content, "Resultados das ferramentas") {
			continue
		}

		result = append(result, dto.ChatMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp.Format(time.RFC3339),
		})
	}

	return result, nil
}

func (s *Service) DeleteConversation(id string) error {
	if s.storage == nil {
		return fmt.Errorf("storage não disponível")
	}
	return s.storage.DeleteConversation(id)
}

func (s *Service) saveCurrentConversation(contextStr string) {
	if s.storage == nil || s.currentConvID == "" {
		return
	}

	var msgs []storage.Message
	for _, m := range s.chatHistory {
		msgs = append(msgs, storage.Message{
			Role:      string(m.Role),
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}

	conv := &storage.Conversation{
		ID:       s.currentConvID,
		Messages: msgs,
		Context:  contextStr,
	}

	s.storage.SaveConversation(conv)
}

func (s *Service) generateID() string {
	return storage.GenerateID()
}
