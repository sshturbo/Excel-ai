package excel

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"excel-ai/internal/dto"
)

func (s *Service) UpdateCell(workbook, sheet, cell, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}

	// Check active workbook/sheet first if not specified
	if workbook == "" || sheet == "" {
		activeWb, activeSheet, err := s.client.GetActiveWorkbookAndSheet()
		if err == nil {
			if workbook == "" {
				workbook = activeWb
			}
			if sheet == "" {
				sheet = activeSheet
			}
		}
	}

	// Fallback to context if still empty
	if workbook == "" {
		workbook = s.currentWorkbook
	}
	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	if workbook == "" || sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	// Salvar valor antigo para desfazer
	oldValue, err := s.client.GetCellValue(workbook, sheet, cell)
	if err == nil && s.currentBatchID != 0 {
		// Usar banco de dados se disponível
		if s.storage != nil && s.currentConvID != "" {
			s.storage.SaveUndoAction(s.currentConvID, s.currentBatchID, workbook, sheet, cell, oldValue)
		} else {
			// Fallback para pilha em memória
			s.undoStack = append(s.undoStack, dto.UndoAction{
				Workbook: workbook,
				Sheet:    sheet,
				Cell:     cell,
				OldValue: oldValue,
				BatchID:  s.currentBatchID,
			})
		}
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
	s.lastExecutedBatchID = s.currentBatchID
	s.currentBatchID = 0
}

// GetLastBatchID returns the ID of the last executed batch (for safe undo)
func (s *Service) GetLastBatchID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastExecutedBatchID
}

// ClearLastBatchID clears the last batch ID (called when user confirms changes)
func (s *Service) ClearLastBatchID() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastExecutedBatchID = 0
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

// UndoByConversation desfaz ações pendentes de uma conversa específica
func (s *Service) UndoByConversation(convID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.storage == nil {
		return 0, fmt.Errorf("storage não configurado")
	}

	if s.client == nil {
		return 0, fmt.Errorf("excel não conectado")
	}

	// Buscar ações pendentes desta conversa
	actions, err := s.storage.GetPendingUndoActions(convID)
	if err != nil {
		return 0, err
	}

	if len(actions) == 0 {
		return 0, fmt.Errorf("nada para desfazer nesta conversa")
	}

	// Agrupar por batch_id para encontrar o último batch
	lastBatchID := actions[0].BatchID // Já está ordenado DESC

	undoneCount := 0
	for _, action := range actions {
		if action.BatchID != lastBatchID {
			break // Só desfaz o último batch
		}

		var err error
		switch action.OperationType {
		case "write":
			err = s.client.SetCellValue(action.Workbook, action.Sheet, action.Cell, action.OldValue)
		case "create-sheet":
			var data map[string]string
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if sheetName, ok := data["sheetName"]; ok {
					err = s.client.DeleteSheet(action.Workbook, sheetName) // Added workbook
				}
			} else {
				err = s.client.DeleteSheet(action.Workbook, action.Sheet) // Added workbook
			}
		case "rename-sheet":
			var data map[string]string
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if oldName, ok := data["oldName"]; ok {
					err = s.client.RenameSheet(action.Workbook, data["newName"], oldName) // Added workbook
				}
			}
		case "merge-cells":
			// Undo merge = Unmerge
			// action.Cell contains the range
			err = s.client.UnmergeCells(action.Workbook, action.Sheet, action.Cell)
		case "unmerge-cells":
			// Undo unmerge = Merge
			err = s.client.MergeCells(action.Workbook, action.Sheet, action.Cell)
		case "insert-rows":
			// Undo insert = Delete
			var data map[string]int
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				// Use workbook from action
				err = s.client.DeleteRows(action.Workbook, action.Sheet, data["row"], data["count"])
			}
		case "delete-rows":
			// Undo delete = Insert + Restore Data
			var data struct {
				Row   int        `json:"row"`
				Count int        `json:"count"`
				Data  [][]string `json:"data"`
			}
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				// 1. Re-insert empty rows
				if insErr := s.client.InsertRows(action.Workbook, action.Sheet, data.Row, data.Count); insErr != nil {
					err = insErr
				} else {
					// 2. Restore data if available
					if len(data.Data) > 0 {
						// Need to restore row by row or range?
						// WriteRange expects [][]interface{} but we stored [][]string. Convert.
						batchData := make([][]interface{}, len(data.Data))
						for i, rowVals := range data.Data {
							rowInterface := make([]interface{}, len(rowVals))
							for j, val := range rowVals {
								rowInterface[j] = val
							}
							batchData[i] = rowInterface
						}
						// Write back starting at the inserted row, col 1 (A)
						err = s.client.WriteRange(action.Workbook, action.Sheet, data.Row, 1, batchData)
					}
				}
			}
		case "clear-range":
			// Undo clear = Restore Data
			var data map[string][][]string
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if oldData, ok := data["data"]; ok && len(oldData) > 0 {
					// Need to calculate start row/col from action.Cell (which is Range)
					// This is tricky without parsing "A1:B2".
					// Workaround: Assume WriteRange accepts range address like "A1" or "A1:B2"...
					// But our wrapper takes (row, col).
					// We need a helper to ParseRangeAddress to (row, col).
					// Since we don't have it easily here, let's try to pass the range string to new WriteRange method overload?
					// No overload.
					// Let's implement ParseRangeAddress in service or client.
					// Or just use "A1" if range starts there? We don't know start.

					// Let's assume action.Cell contains "A1:B2".
					// We need top-left cell row/col.
					// Let's ask client for coordinates?
					// s.client.GetCellCoordinates(cell) would be useful.

					// Temporary: Try to parse crudely or add helper.
					// Better: Add WriteRangeString(sheet, address, data) to client?
					// Or just restore blindly to A1 if we can't parse? No.

					// Let's rely on a new service helper: WriteRangeToAddress
					err = s.WriteRangeToAddress(action.Workbook, action.Sheet, action.Cell, oldData)
				}
			}
		case "sort":
			// Undo sort = Restore original data order
			var data map[string][][]string
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if oldData, ok := data["data"]; ok && len(oldData) > 0 {
					err = s.WriteRangeToAddress(action.Workbook, action.Sheet, action.Cell, oldData)
				}
			}
		case "set-column-width":
			// Undo width = restore width
			var data map[string]float64
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if w, ok := data["width"]; ok {
					err = s.client.SetColumnWidth(action.Workbook, action.Sheet, action.Cell, w)
				}
			}
		case "set-row-height":
			// Undo height = restore height
			var data map[string]float64
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if h, ok := data["height"]; ok {
					err = s.client.SetRowHeight(action.Workbook, action.Sheet, action.Cell, h)
				}
			}
		default:
			err = s.client.SetCellValue(action.Workbook, action.Sheet, action.Cell, action.OldValue)
		}

		if err != nil {
			return undoneCount, fmt.Errorf("erro ao desfazer ação %s: %w", action.OperationType, err)
		}
		undoneCount++
	}

	// Remover ações desfeitas do banco
	if err := s.storage.DeleteUndoActions(convID, lastBatchID); err != nil {
		return undoneCount, fmt.Errorf("erro ao limpar ações desfeitas: %w", err)
	}

	return undoneCount, nil
}

// ApproveActions marca ações pendentes de uma conversa como aprovadas
func (s *Service) ApproveActions(convID string) error {
	if s.storage == nil {
		return nil // Sem storage, não há o que aprovar
	}
	return s.storage.ApproveUndoActions(convID)
}

// WriteRangeToAddress escreve dados em um range definido por string (ex: "A1:B2")
func (s *Service) WriteRangeToAddress(workbook, sheet, address string, data [][]string) error {
	// Pega apenas a célula inicial (ex: "A1" de "A1:B2")
	startCell := strings.Split(address, ":")[0]
	row, col, err := parseCellAddress(startCell)
	if err != nil {
		return fmt.Errorf("endereço inválido '%s': %w", startCell, err)
	}

	// Converter [][]string para [][]interface{}
	interfaceData := make([][]interface{}, len(data))
	for i, r := range data {
		rowSlice := make([]interface{}, len(r))
		for j, c := range r {
			rowSlice[j] = c
		}
		interfaceData[i] = rowSlice
	}

	return s.client.WriteRange(workbook, sheet, row, col, interfaceData)
}

func parseCellAddress(address string) (int, int, error) {
	address = strings.ToUpper(address)
	colStr := ""
	rowStr := ""

	for _, char := range address {
		if char >= 'A' && char <= 'Z' {
			colStr += string(char)
		} else if char >= '0' && char <= '9' {
			rowStr += string(char)
		}
	}

	if colStr == "" || rowStr == "" {
		return 0, 0, fmt.Errorf("formato inválido")
	}

	// Col string to int (A=1, AA=27)
	col := 0
	for i := 0; i < len(colStr); i++ {
		col *= 26
		col += int(colStr[i] - 'A' + 1)
	}

	// Row string to int
	row := 0
	fmt.Sscanf(rowStr, "%d", &row)

	return row, col, nil
}

// HasPendingUndoActions verifica se há ações pendentes para uma conversa
func (s *Service) HasPendingUndoActions(convID string) (bool, error) {
	if s.storage == nil {
		return false, nil
	}
	return s.storage.HasPendingUndoActions(convID)
}

func (s *Service) WriteToExcel(row, col int, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("não conectado ao Excel")
	}

	workbook := s.currentWorkbook
	sheet := s.getFirstSheet()

	// Logic update: prefer active if context is empty
	if workbook == "" || sheet == "" {
		activeWb, activeSheet, err := s.client.GetActiveWorkbookAndSheet()
		if err == nil {
			if workbook == "" {
				workbook = activeWb
			}
			if sheet == "" {
				sheet = activeSheet
			}
		}
	}

	if workbook == "" || sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	return s.client.WriteCell(workbook, sheet, row, col, value)
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
