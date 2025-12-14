package chat

import (
	"context"
	"sync"

	"excel-ai/internal/domain"
	"excel-ai/internal/services/excel"
	"excel-ai/pkg/ai"
	"excel-ai/pkg/storage"
)

type Service struct {
	client        *ai.Client
	geminiClient  *ai.GeminiClient
	provider      string // "openrouter", "groq", "google", "custom"
	storage       *storage.Storage
	mu            sync.Mutex
	chatHistory   []domain.Message
	currentConvID string
	cancelFunc    context.CancelFunc
	excelService  *excel.Service
}

func NewService(storage *storage.Storage) *Service {
	return &Service{
		client:       ai.NewClient("", "", ""), // API Key, Model, BaseURL set later
		geminiClient: ai.NewGeminiClient("", ""),
		provider:     "openrouter",
		storage:      storage,
		chatHistory:  []domain.Message{},
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

func (s *Service) SetBaseURL(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client.SetBaseURL(url)
}

func (s *Service) SetExcelService(svc *excel.Service) {
	s.excelService = svc
}

// Helper to convert domain messages to AI messages
func (s *Service) toAIMessages(msgs []domain.Message) []ai.Message {
	var result []ai.Message
	for _, m := range msgs {
		result = append(result, ai.Message{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}
	return result
}
