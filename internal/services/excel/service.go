package excel

import (
	"excel-ai/internal/dto"
	"excel-ai/pkg/excel"
	"excel-ai/pkg/logger"
	"excel-ai/pkg/storage"
	"fmt"
	"strings"
	"sync"
)

type Service struct {
	client              *excel.Client
	mu                  sync.Mutex
	currentWorkbook     string
	currentSheet        string
	previewData         *excel.SheetData
	undoStack           []dto.UndoAction // Mantido para fallback
	currentBatchID      int64
	lastExecutedBatchID int64
	contextStr          string
	storage             *storage.Storage // Banco de dados para undo
	currentConvID       string           // ID da conversa atual
}

func NewService() *Service {
	return &Service{
		undoStack: []dto.UndoAction{},
	}
}

// SetStorage configura o storage para persistência de undo
func (s *Service) SetStorage(store *storage.Storage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage = store
	logger.ExcelInfo("Storage configurado")
}

// SetConversationID define a conversa atual para vincular ações de undo
func (s *Service) SetConversationID(convID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentConvID = convID
	logger.ExcelDebug("ID da conversa definido: " + convID)
}

// GetConversationID retorna a conversa atual
func (s *Service) GetConversationID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentConvID
}

// getFirstSheet retorna a primeira aba quando currentSheet contém múltiplas abas
func (s *Service) getFirstSheet() string {
	if strings.Contains(s.currentSheet, ",") {
		return strings.TrimSpace(strings.Split(s.currentSheet, ",")[0])
	}
	return s.currentSheet
}

func (s *Service) Connect() (*dto.ExcelStatus, error) {
	logger.ExcelInfo("Tentando conectar ao Excel")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		s.client.Close()
	}

	client, err := excel.NewClient()
	if err != nil {
		logger.ExcelError("Erro ao criar cliente Excel: " + err.Error())
		return &dto.ExcelStatus{Connected: false, Error: err.Error()}, nil
	}

	s.client = client
	workbooks, err := s.client.GetOpenWorkbooks()
	if err != nil {
		// Se falhar ao listar, consideramos que a conexão não foi totalmente bem sucedida
		// para permitir que o usuário tente novamente
		s.client.Close()
		s.client = nil
		logger.ExcelError("Erro ao listar planilhas: " + err.Error())
		return &dto.ExcelStatus{Connected: false, Error: "Conectado ao Excel, mas falha ao listar planilhas: " + err.Error()}, nil
	}

	logger.ExcelInfo("Conectado ao Excel com sucesso")
	return &dto.ExcelStatus{Connected: true, Workbooks: workbooks}, nil
}

func (s *Service) RefreshWorkbooks() (*dto.ExcelStatus, error) {
	logger.ExcelDebug("Atualizando lista de workbooks")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		logger.ExcelWarn("Excel não conectado")
		return &dto.ExcelStatus{Connected: false, Error: "Não conectado"}, nil
	}

	workbooks, err := s.client.GetOpenWorkbooks()
	if err != nil {
		logger.ExcelError("Erro ao atualizar workbooks: " + err.Error())
		return &dto.ExcelStatus{Connected: true, Error: err.Error()}, nil
	}

	logger.ExcelInfo("Workbooks atualizados com sucesso")
	return &dto.ExcelStatus{Connected: true, Workbooks: workbooks}, nil
}

// GetActiveWorkbookName retorna o nome da pasta de trabalho ativa
func (s *Service) GetActiveWorkbookName() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return "", fmt.Errorf("não conectado")
	}
	return s.client.GetActiveWorkbookName()
}

func (s *Service) Close() {
	logger.ExcelInfo("Fechando conexão com Excel")
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		s.client.Close()
		logger.ExcelInfo("Conexão fechada")
	}
}

// SaveUndoAction salva uma ação de undo no banco de dados
func (s *Service) SaveUndoAction(opType, workbook, sheet, cell, oldValue, undoData string) error {
	logger.ExcelDebug(fmt.Sprintf("Salvando ação de undo: %s", opType))
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.storage != nil && s.currentConvID != "" && s.currentBatchID != 0 {
		err := s.storage.SaveUndoActionFull(s.currentConvID, s.currentBatchID, opType, workbook, sheet, cell, oldValue, undoData)
		if err != nil {
			logger.ExcelError("Erro ao salvar ação de undo: " + err.Error())
		} else {
			logger.ExcelDebug("Ação de undo salva com sucesso")
		}
		return err
	}
	return nil
}
