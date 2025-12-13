package chat

import (
	"excel-ai/internal/dto"
	"excel-ai/pkg/ai"
	"excel-ai/pkg/storage"
	"fmt"
	"sync"
	"time"
)

type Service struct {
	client        *ai.Client
	storage       *storage.Storage
	mu            sync.Mutex
	chatHistory   []ai.Message
	currentConvID string
}

func NewService(storage *storage.Storage) *Service {
	return &Service{
		client:      ai.NewClient("", ""), // API Key set later
		storage:     storage,
		chatHistory: []ai.Message{},
	}
}

func (s *Service) SetAPIKey(apiKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client.SetAPIKey(apiKey)
}

func (s *Service) SetModel(modelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client.SetModel(modelID)
}

func (s *Service) GetAvailableModels() []dto.ModelInfo {
	models, err := s.client.GetAvailableModels()
	if err != nil {
		// Fallback to hardcoded list if API fails
		return []dto.ModelInfo{
			{
				ID:            "google/gemini-2.0-flash-exp:free",
				Name:          "Gemini 2.0 Flash (Free)",
				Description:   "Modelo experimental rápido e gratuito do Google",
				ContextLength: 1000000,
				PricePrompt:   "0",
				PriceComplete: "0",
			},
			{
				ID:            "google/gemini-exp-1206:free",
				Name:          "Gemini Exp 1206 (Free)",
				Description:   "Modelo experimental atualizado",
				ContextLength: 1000000,
				PricePrompt:   "0",
				PriceComplete: "0",
			},
			{
				ID:            "meta-llama/llama-3.2-90b-vision-instruct:free",
				Name:          "Llama 3.2 90B (Free)",
				Description:   "Modelo open source poderoso da Meta",
				ContextLength: 128000,
				PricePrompt:   "0",
				PriceComplete: "0",
			},
			{
				ID:            "microsoft/phi-3-medium-128k-instruct:free",
				Name:          "Phi-3 Medium (Free)",
				Description:   "Modelo eficiente da Microsoft",
				ContextLength: 128000,
				PricePrompt:   "0",
				PriceComplete: "0",
			},
			{
				ID:            "anthropic/claude-3.5-sonnet",
				Name:          "Claude 3.5 Sonnet",
				Description:   "Alta inteligência e capacidade de codificação",
				ContextLength: 200000,
				PricePrompt:   "$3/1M",
				PriceComplete: "$15/1M",
			},
			{
				ID:            "openai/gpt-4o",
				Name:          "GPT-4o",
				Description:   "Modelo flagship da OpenAI",
				ContextLength: 128000,
				PricePrompt:   "$5/1M",
				PriceComplete: "$15/1M",
			},
		}
	}

	var result []dto.ModelInfo
	for _, m := range models {
		result = append(result, dto.ModelInfo{
			ID:            m.ID,
			Name:          m.Name,
			Description:   m.Description,
			ContextLength: m.ContextLength,
			PricePrompt:   m.Pricing.Prompt,
			PriceComplete: m.Pricing.Completion,
		})
	}

	return result
}

func (s *Service) SendMessage(message string, contextStr string, onChunk func(string) error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add context if it's the first message or context changed
	// For simplicity, we append context to the system message or the first user message
	// Here we just ensure the history has the context if needed.
	// Actually, let's just append the user message.

	// If context is provided, we might want to prepend it to the user message or as a system message.
	// Let's prepend to user message for now if it's not empty.
	fullContent := message
	if contextStr != "" {
		fullContent = fmt.Sprintf("Contexto do Excel:\n%s\n\nPergunta do usuário: %s", contextStr, message)
	}

	s.chatHistory = append(s.chatHistory, ai.Message{
		Role:    "user",
		Content: fullContent,
	})

	// Call AI
	response, err := s.client.ChatStream(s.chatHistory, onChunk)
	if err != nil {
		// Remove user message on error
		s.chatHistory = s.chatHistory[:len(s.chatHistory)-1]
		return "", err
	}

	s.chatHistory = append(s.chatHistory, ai.Message{
		Role:    "assistant",
		Content: response,
	})

	go s.saveCurrentConversation(contextStr)

	return response, nil
}

func (s *Service) ClearChat() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatHistory = []ai.Message{}
}

func (s *Service) GetChatHistory() []dto.ChatMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []dto.ChatMessage
	for _, m := range s.chatHistory {
		result = append(result, dto.ChatMessage{
			Role:    m.Role,
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
		s.chatHistory = []ai.Message{}
	} else {
		s.chatHistory = s.chatHistory[:len(s.chatHistory)-count]
	}

	go s.saveCurrentConversation("") // Context might be lost here if not stored in service state
	return nil
}

func (s *Service) NewConversation() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentConvID != "" && len(s.chatHistory) > 0 {
		s.saveCurrentConversation("")
	}

	s.chatHistory = []ai.Message{}
	s.currentConvID = storage.GenerateID()
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
	s.chatHistory = []ai.Message{}

	var result []dto.ChatMessage
	for _, m := range conv.Messages {
		s.chatHistory = append(s.chatHistory, ai.Message{
			Role:    m.Role,
			Content: m.Content,
		})
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
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: time.Now(),
		})
	}

	conv := &storage.Conversation{
		ID:       s.currentConvID,
		Messages: msgs,
		Context:  contextStr,
	}

	s.storage.SaveConversation(conv)
}
