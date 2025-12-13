package excel

import (
	"excel-ai/internal/dto"
	"excel-ai/pkg/excel"
	"fmt"
	"sync"
	"time"
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
		return &dto.ExcelStatus{Connected: true, Error: "Erro ao listar planilhas: " + err.Error()}, nil
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

func (s *Service) SetContext(workbook, sheet string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return "", fmt.Errorf("não conectado ao Excel")
	}

	data, err := s.client.GetSheetData(workbook, sheet, 50) // Ler primeiras 50 linhas para contexto
	if err != nil {
		return "", err
	}

	s.currentWorkbook = workbook
	s.currentSheet = sheet

	// Construir representação em string para a IA
	contextStr := fmt.Sprintf("Planilha: %s\nAba: %s\n\nDados (amostra):\n", workbook, sheet)

	// Cabeçalhos
	for i, h := range data.Headers {
		if i > 0 {
			contextStr += " | "
		}
		contextStr += h
	}
	contextStr += "\n"

	// Linhas
	for _, row := range data.Rows {
		for i, cell := range row {
			if i > 0 {
				contextStr += " | "
			}
			contextStr += cell.Text
		}
		contextStr += "\n"
	}

	s.contextStr = contextStr
	return fmt.Sprintf("Contexto carregado: %s - %s (%d linhas)", workbook, sheet, len(data.Rows)), nil
}

func (s *Service) GetContextString() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.contextStr
}

func (s *Service) GetPreviewData(workbookName, sheetName string) (*dto.PreviewData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil, fmt.Errorf("não conectado ao Excel")
	}

	data, err := s.client.GetSheetData(workbookName, sheetName, 100)
	if err != nil {
		return nil, err
	}

	s.previewData = data
	s.currentWorkbook = workbookName
	s.currentSheet = sheetName

	// Converter para formato simples
	var rows [][]string
	for _, row := range data.Rows {
		var rowStrings []string
		for _, cell := range row {
			rowStrings = append(rowStrings, cell.Text)
		}
		rows = append(rows, rowStrings)
	}

	return &dto.PreviewData{
		Headers:   data.Headers,
		Rows:      rows,
		TotalRows: len(data.Rows), // Aproximado
		TotalCols: len(data.Headers),
		Workbook:  workbookName,
		Sheet:     sheetName,
	}, nil
}

func (s *Service) GetActiveSelection() (*excel.SheetData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil, fmt.Errorf("não conectado ao Excel")
	}

	return s.client.GetActiveSelection()
}

func (s *Service) UpdateCell(workbook, sheet, cell, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}

	if workbook == "" {
		workbook = s.currentWorkbook
	}
	if sheet == "" {
		sheet = s.currentSheet
	}

	if workbook == "" || sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	// Salvar valor antigo para desfazer
	oldValue, err := s.client.GetCellValue(workbook, sheet, cell)
	if err == nil {
		s.undoStack = append(s.undoStack, dto.UndoAction{
			Workbook: workbook,
			Sheet:    sheet,
			Cell:     cell,
			OldValue: oldValue,
			BatchID:  s.currentBatchID,
		})
	}

	return s.client.SetCellValue(workbook, sheet, cell, value)
}

func (s *Service) StartUndoBatch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentBatchID = time.Now().UnixNano()
}

func (s *Service) EndUndoBatch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentBatchID = 0
}

func (s *Service) UndoLastChange() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.undoStack) == 0 {
		return fmt.Errorf("nada para desfazer")
	}

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}

	undoAction := func(action dto.UndoAction) error {
		return s.client.SetCellValue(action.Workbook, action.Sheet, action.Cell, action.OldValue)
	}

	lastIdx := len(s.undoStack) - 1
	lastAction := s.undoStack[lastIdx]
	s.undoStack = s.undoStack[:lastIdx]

	if err := undoAction(lastAction); err != nil {
		return err
	}

	if lastAction.BatchID != 0 {
		for len(s.undoStack) > 0 {
			idx := len(s.undoStack) - 1
			prevAction := s.undoStack[idx]

			if prevAction.BatchID == lastAction.BatchID {
				s.undoStack = s.undoStack[:idx]
				if err := undoAction(prevAction); err != nil {
					return fmt.Errorf("erro ao desfazer lote: %w", err)
				}
			} else {
				break
			}
		}
	}

	return nil
}

func (s *Service) CreateNewWorkbook() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return "", fmt.Errorf("excel não conectado")
	}
	return s.client.CreateNewWorkbook()
}

func (s *Service) CreateNewSheet(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.InsertNewSheet(s.currentWorkbook, name)
}

func (s *Service) CreateChart(sheet, rangeAddr, chartType, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	if sheet == "" {
		sheet = s.currentSheet
	}
	return s.client.CreateChart(s.currentWorkbook, sheet, rangeAddr, chartType, title)
}

func (s *Service) CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	if sourceSheet == "" {
		sourceSheet = s.currentSheet
	}
	return s.client.CreatePivotTable(s.currentWorkbook, sourceSheet, sourceRange, destSheet, destCell, tableName)
}

func (s *Service) WriteToExcel(row, col int, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("não conectado ao Excel")
	}

	if s.currentWorkbook == "" || s.currentSheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	return s.client.WriteCell(s.currentWorkbook, s.currentSheet, row, col, value)
}

func (s *Service) ApplyFormula(row, col int, formula string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" || s.currentSheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}
	return s.client.ApplyFormula(s.currentWorkbook, s.currentSheet, row, col, formula)
}

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		s.client.Close()
	}
}
