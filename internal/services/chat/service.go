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
	ollamaClient  *ai.OllamaClient // Cliente nativo Ollama para melhor suporte a tools
	provider      string           // "openrouter", "groq", "google", "custom", "ollama"
	storage       *storage.Storage
	mu            sync.Mutex
	cancelMu      sync.Mutex // Mutex separado para cancelFunc (evita deadlock)
	chatHistory   []domain.Message
	currentConvID string
	cancelFunc    context.CancelFunc
	excelService  *excel.Service

	// Pending action state (when askBeforeApply pauses execution)
	pendingAction     *ToolCommand
	pendingContextStr string
	pendingOnChunk    func(string) error
}

func NewService(storage *storage.Storage) *Service {
	return &Service{
		client:       ai.NewClient("", "", ""), // API Key, Model, BaseURL set later
		geminiClient: ai.NewGeminiClient("", ""),
		ollamaClient: ai.NewOllamaClient("", ""), // Cliente nativo Ollama
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
	if s.ollamaClient != nil {
		s.ollamaClient.SetBaseURL(url)
	}
}

// SetProvider atualiza o provider e reconfigura os clientes
func (s *Service) SetProvider(provider string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.provider = provider
}

// RefreshConfig recarrega configurações do storage
func (s *Service) RefreshConfig() {
	s.refreshConfig()
}

func (s *Service) SetExcelService(svc *excel.Service) {
	s.excelService = svc
}

// Helper to convert domain messages to AI messages
func (s *Service) toAIMessages(msgs []domain.Message) []ai.Message {
	var result []ai.Message
	for _, m := range msgs {
		aiMsg := ai.Message{
			Role:       string(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}

		// Converter tool calls se existirem
		if m.ToolCalls != nil {
			if tcs, ok := m.ToolCalls.([]ai.ToolCall); ok {
				aiMsg.ToolCalls = tcs
			}
		}

		result = append(result, aiMsg)
	}
	return result
}
