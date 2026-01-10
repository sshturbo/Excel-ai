package chat

import (
	"context"
	"fmt"
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
	zaiClient     *ai.ZAIClient    // Cliente nativo Z.AI para GLM models
	provider      string           // "openrouter", "groq", "google", "custom", "ollama", "zai"
	storage       *storage.Storage
	mu            sync.Mutex
	cancelMu      sync.Mutex // Mutex separado para cancelFunc (evita deadlock)
	chatHistory   []domain.Message
	currentConvID string
	cancelFunc    context.CancelFunc
	excelService  *excel.Service

	// Orchestrator para execução paralela
	orchestrator     *Orchestrator
	useOrchestration bool // Habilita/desabilita orquestração

	// Pending action state (when askBeforeApply pauses execution)
	pendingAction     *ToolCommand
	pendingContextStr string
	pendingOnChunk    func(string) error
}

func NewService(storage *storage.Storage) *Service {
	svc := &Service{
		client:           ai.NewClient("", "", ""), // API Key, Model, BaseURL set later
		geminiClient:     ai.NewGeminiClient("", ""),
		ollamaClient:     ai.NewOllamaClient("", ""), // Cliente nativo Ollama
		zaiClient:        ai.NewZAIClient("", ""),    // Cliente nativo Z.AI
		provider:         "openrouter",
		storage:          storage,
		chatHistory:      []domain.Message{},
		useOrchestration: false, // Desabilitado por padrão
	}

	// Inicializar orquestrador
	orchestrator, err := NewOrchestrator(svc)
	if err != nil {
		fmt.Printf("[SERVICE] Erro ao inicializar orquestrador: %v\n", err)
		// Continuar mesmo sem orquestrador
	}
	svc.orchestrator = orchestrator

	return svc
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

// SetOrchestration habilita ou desabilita a orquestração
func (s *Service) SetOrchestration(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.useOrchestration = enabled

	if enabled {
		fmt.Println("[SERVICE] Orquestração habilitada")
	} else {
		fmt.Println("[SERVICE] Orquestração desabilitada")
	}
}

// GetOrchestration retorna se a orquestração está habilitada
func (s *Service) GetOrchestration() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.useOrchestration
}

// StartOrchestrator inicia o orquestrador
func (s *Service) StartOrchestrator(ctx context.Context) error {
	return s.orchestrator.Start(ctx)
}

// StopOrchestrator para o orquestrador
func (s *Service) StopOrchestrator() {
	s.orchestrator.Stop()
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

// GetOrchestrator retorna o orquestrador (para acesso externo)
func (s *Service) GetOrchestrator() *Orchestrator {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.orchestrator
}
