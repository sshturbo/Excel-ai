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
