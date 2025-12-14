package excel

import (
	"fmt"
	"time"

	"excel-ai/internal/dto"
)

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
		sheet = s.getFirstSheet()
	}

	// Se ainda não tiver workbook, tentar obter o ativo
	if workbook == "" {
		activeWb, activeSheet, err := s.client.GetActiveWorkbookAndSheet()
		if err == nil {
			workbook = activeWb
			if sheet == "" {
				sheet = activeSheet
			}
		}
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

func (s *Service) WriteToExcel(row, col int, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("não conectado ao Excel")
	}

	sheet := s.getFirstSheet()
	if s.currentWorkbook == "" || sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	return s.client.WriteCell(s.currentWorkbook, sheet, row, col, value)
}

func (s *Service) ApplyFormula(row, col int, formula string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	sheet := s.getFirstSheet()
	if s.currentWorkbook == "" || sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}
	return s.client.ApplyFormula(s.currentWorkbook, sheet, row, col, formula)
}

// WriteRange escreve múltiplos valores em um range a partir de uma célula inicial
func (s *Service) WriteRange(sheet, startCell string, data [][]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}

	workbook := s.currentWorkbook

	// Sempre tentar obter workbook e sheet ativos se não especificados
	activeWb, activeSheet, _ := s.client.GetActiveWorkbookAndSheet()

	if workbook == "" {
		workbook = activeWb
	}
	if sheet == "" {
		sheet = activeSheet
	}
	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	if workbook == "" || sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	// Converter endereço de célula para linha/coluna
	startRow, startCol := cellToRowCol(startCell)
	if startRow == 0 || startCol == 0 {
		return fmt.Errorf("endereço de célula inválido: %s", startCell)
	}

	return s.client.WriteRange(workbook, sheet, startRow, startCol, data)
}

// cellToRowCol converte endereço de célula (ex: "A1") para linha e coluna (1-indexed)
func cellToRowCol(cell string) (row, col int) {
	col = 0
	row = 0

	for i, c := range cell {
		if c >= 'A' && c <= 'Z' {
			col = col*26 + int(c-'A'+1)
		} else if c >= 'a' && c <= 'z' {
			col = col*26 + int(c-'a'+1)
		} else if c >= '0' && c <= '9' {
			for j := i; j < len(cell); j++ {
				if cell[j] >= '0' && cell[j] <= '9' {
					row = row*10 + int(cell[j]-'0')
				}
			}
			break
		}
	}
	return row, col
}
