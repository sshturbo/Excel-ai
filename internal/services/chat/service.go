package chat

import (
	"context"
	"sync"

	"excel-ai/internal/domain"
	"excel-ai/internal/services/excel"
	"excel-ai/pkg/ai"
	"excel-ai/pkg/logger"
	"excel-ai/pkg/storage"
)

type Service struct {
	zaiClient     *ai.ZAIClient    // Cliente nativo Z.AI para GLM models
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
	logger.ChatInfo("Inicializando chat service com Z.AI")
	
	svc := &Service{
		zaiClient:        ai.NewZAIClient("", ""),    // Cliente nativo Z.AI
		storage:          storage,
		chatHistory:      []domain.Message{},
		useOrchestration: false, // Desabilitado por padrão
	}

	// Inicializar orquestrador
	orchestrator, err := NewOrchestrator(svc)
	if err != nil {
		logger.ChatError("Erro ao inicializar orquestrador: " + err.Error())
		// Continuar mesmo sem orquestrador
	} else {
		logger.ChatInfo("Orquestrador inicializado com sucesso")
	}
	svc.orchestrator = orchestrator

	return svc
}

func (s *Service) SetAPIKey(apiKey string) {
	logger.ChatInfo("Atualizando API key Z.AI")
	s.mu.Lock()
	defer s.mu.Unlock()
	s.zaiClient.SetAPIKey(apiKey)
}

func (s *Service) SetModel(modelID string) {
	logger.ChatInfo("Atualizando modelo Z.AI: " + modelID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.zaiClient.SetModel(modelID)
}

func (s *Service) SetBaseURL(url string) {
	logger.ChatInfo("Atualizando base URL Z.AI: " + url)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.zaiClient.SetBaseURL(url)
}

// RefreshConfig recarrega configurações do storage
func (s *Service) RefreshConfig() {
	logger.ChatInfo("Recarregando configurações")
	s.refreshConfig()
}

func (s *Service) SetExcelService(svc *excel.Service) {
	logger.ChatInfo("Configurando Excel service")
	s.excelService = svc
}

// SetOrchestration habilita ou desabilita a orquestração
func (s *Service) SetOrchestration(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.useOrchestration = enabled

	if enabled {
		logger.ChatInfo("Orquestração habilitada")
	} else {
		logger.ChatInfo("Orquestração desabilitada")
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
	logger.ChatInfo("Iniciando orquestrador")
	err := s.orchestrator.Start(ctx)
	if err != nil {
		logger.ChatError("Erro ao iniciar orquestrador: " + err.Error())
	}
	return err
}

// StopOrchestrator para o orquestrador
func (s *Service) StopOrchestrator() {
	logger.ChatInfo("Parando orquestrador")
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
