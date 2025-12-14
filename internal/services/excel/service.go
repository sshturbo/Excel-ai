package excel

import (
	"excel-ai/internal/dto"
	"excel-ai/pkg/excel"
	"strings"
	"sync"
)

type Service struct {
	client          *excel.Client
	mu              sync.Mutex
	currentWorkbook string
	currentSheet    string
	previewData     *excel.SheetData
	undoStack       []dto.UndoAction
	currentBatchID  int64
	contextStr      string
}

func NewService() *Service {
	return &Service{
		undoStack: []dto.UndoAction{},
	}
}

// getFirstSheet retorna a primeira aba quando currentSheet contém múltiplas abas
func (s *Service) getFirstSheet() string {
	if strings.Contains(s.currentSheet, ",") {
		return strings.TrimSpace(strings.Split(s.currentSheet, ",")[0])
	}
	return s.currentSheet
}

func (s *Service) Connect() (*dto.ExcelStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		s.client.Close()
	}

	client, err := excel.NewClient()
	if err != nil {
		return &dto.ExcelStatus{Connected: false, Error: err.Error()}, nil
	}

	s.client = client
	workbooks, err := s.client.GetOpenWorkbooks()
	if err != nil {
		// Se falhar ao listar, consideramos que a conexão não foi totalmente bem sucedida
		// para permitir que o usuário tente novamente
		s.client.Close()
		s.client = nil
		return &dto.ExcelStatus{Connected: false, Error: "Conectado ao Excel, mas falha ao listar planilhas: " + err.Error()}, nil
	}

	return &dto.ExcelStatus{Connected: true, Workbooks: workbooks}, nil
}

func (s *Service) RefreshWorkbooks() (*dto.ExcelStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return &dto.ExcelStatus{Connected: false, Error: "Não conectado"}, nil
	}

	workbooks, err := s.client.GetOpenWorkbooks()
	if err != nil {
		return &dto.ExcelStatus{Connected: true, Error: err.Error()}, nil
	}

	return &dto.ExcelStatus{Connected: true, Workbooks: workbooks}, nil
}

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		s.client.Close()
	}
}
