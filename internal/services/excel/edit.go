package excel

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"excel-ai/internal/dto"
)

// UpdateCell atualiza o valor de uma célula
func (s *Service) UpdateCell(workbook, sheet, cell, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	// workbook é ignorado no modo Excelize
	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	if sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	// Salvar valor antigo para desfazer
	oldValue, err := client.GetCellValue(sheet, cell)
	if err == nil && s.currentBatchID != 0 {
		if s.storage != nil && s.currentConvID != "" {
			s.storage.SaveUndoAction(s.currentConvID, s.currentBatchID, s.currentFileName, sheet, cell, oldValue)
		} else {
			s.undoStack = append(s.undoStack, dto.UndoAction{
				Workbook: s.currentFileName,
				Sheet:    sheet,
				Cell:     cell,
				OldValue: oldValue,
				BatchID:  s.currentBatchID,
			})
		}
	}

	return client.SetCellValue(sheet, cell, value)
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

// GetLastBatchID returns the ID of the last executed batch
func (s *Service) GetLastBatchID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastExecutedBatchID
}

// ClearLastBatchID clears the last batch ID
func (s *Service) ClearLastBatchID() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastExecutedBatchID = 0
}

// UndoLastChange desfaz última alteração da pilha em memória
func (s *Service) UndoLastChange() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.undoStack) == 0 {
		return fmt.Errorf("nada para desfazer")
	}

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	undoAction := func(action dto.UndoAction) error {
		return client.SetCellValue(action.Sheet, action.Cell, action.OldValue)
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

	client, err := s.getClientLocked()
	if err != nil {
		return 0, err
	}

	// Buscar ações pendentes desta conversa
	actions, err := s.storage.GetPendingUndoActions(convID)
	if err != nil {
		return 0, err
	}

	if len(actions) == 0 {
		return 0, fmt.Errorf("nada para desfazer nesta conversa")
	}

	lastBatchID := actions[0].BatchID

	undoneCount := 0
	for _, action := range actions {
		if action.BatchID != lastBatchID {
			break
		}

		var err error
		switch action.OperationType {
		case "write":
			err = client.SetCellValue(action.Sheet, action.Cell, action.OldValue)
		case "create-sheet":
			var data map[string]string
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if sheetName, ok := data["sheetName"]; ok {
					err = client.DeleteSheet(sheetName)
				}
			} else {
				err = client.DeleteSheet(action.Sheet)
			}
		case "rename-sheet":
			var data map[string]string
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if oldName, ok := data["oldName"]; ok {
					err = client.RenameSheet(data["newName"], oldName)
				}
			}
		case "merge-cells":
			err = client.UnmergeCells(action.Sheet, action.Cell)
		case "unmerge-cells":
			err = client.MergeCells(action.Sheet, action.Cell)
		case "insert-rows":
			var data map[string]int
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				err = client.DeleteRows(action.Sheet, data["row"], data["count"])
			}
		case "delete-rows":
			var data struct {
				Row   int        `json:"row"`
				Count int        `json:"count"`
				Data  [][]string `json:"data"`
			}
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if insErr := client.InsertRows(action.Sheet, data.Row, data.Count); insErr != nil {
					err = insErr
				} else if len(data.Data) > 0 {
					batchData := make([][]interface{}, len(data.Data))
					for i, rowVals := range data.Data {
						rowInterface := make([]interface{}, len(rowVals))
						for j, val := range rowVals {
							rowInterface[j] = val
						}
						batchData[i] = rowInterface
					}
					startCell := fmt.Sprintf("A%d", data.Row)
					err = client.WriteRange(action.Sheet, startCell, batchData)
				}
			}
		case "clear-range":
			var data map[string][][]string
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if oldData, ok := data["data"]; ok && len(oldData) > 0 {
					startCell := strings.Split(action.Cell, ":")[0]
					interfaceData := make([][]interface{}, len(oldData))
					for i, r := range oldData {
						rowSlice := make([]interface{}, len(r))
						for j, c := range r {
							rowSlice[j] = c
						}
						interfaceData[i] = rowSlice
					}
					err = client.WriteRange(action.Sheet, startCell, interfaceData)
				}
			}
		case "sort":
			var data map[string][][]string
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if oldData, ok := data["data"]; ok && len(oldData) > 0 {
					startCell := strings.Split(action.Cell, ":")[0]
					interfaceData := make([][]interface{}, len(oldData))
					for i, r := range oldData {
						rowSlice := make([]interface{}, len(r))
						for j, c := range r {
							rowSlice[j] = c
						}
						interfaceData[i] = rowSlice
					}
					err = client.WriteRange(action.Sheet, startCell, interfaceData)
				}
			}
		case "set-column-width":
			var data map[string]float64
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if w, ok := data["width"]; ok {
					err = client.SetColumnWidth(action.Sheet, action.Cell, w)
				}
			}
		case "set-row-height":
			var data map[string]float64
			if jsonErr := json.Unmarshal([]byte(action.UndoData), &data); jsonErr == nil {
				if h, ok := data["height"]; ok {
					err = client.SetRowHeight(action.Sheet, action.Cell, h)
				}
			}
		default:
			err = client.SetCellValue(action.Sheet, action.Cell, action.OldValue)
		}

		if err != nil {
			return undoneCount, fmt.Errorf("erro ao desfazer ação %s: %w", action.OperationType, err)
		}
		undoneCount++
	}

	if err := s.storage.DeleteUndoActions(convID, lastBatchID); err != nil {
		return undoneCount, fmt.Errorf("erro ao limpar ações desfeitas: %w", err)
	}

	return undoneCount, nil
}

// ApproveActions marca ações pendentes de uma conversa como aprovadas
func (s *Service) ApproveActions(convID string) error {
	if s.storage == nil {
		return nil
	}
	return s.storage.ApproveUndoActions(convID)
}

// HasPendingUndoActions verifica se há ações pendentes para uma conversa
func (s *Service) HasPendingUndoActions(convID string) (bool, error) {
	if s.storage == nil {
		return false, nil
	}
	return s.storage.HasPendingUndoActions(convID)
}

// WriteToExcel escreve um valor em posição específica
func (s *Service) WriteToExcel(row, col int, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	sheet := s.getFirstSheet()
	if sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	// Converter row/col para endereço de célula
	cell := colToLetter(col) + fmt.Sprintf("%d", row)
	return client.SetCellValue(sheet, cell, value)
}

// ApplyFormula aplica uma fórmula em uma célula
func (s *Service) ApplyFormula(row, col int, formula string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	sheet := s.getFirstSheet()
	if sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	cell := colToLetter(col) + fmt.Sprintf("%d", row)
	// Fórmulas no Excelize são valores que começam com =
	return client.SetCellValue(sheet, cell, formula)
}

// WriteRange escreve múltiplos valores em um range
func (s *Service) WriteRange(sheet, startCell string, data [][]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	if sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	return client.WriteRange(sheet, startCell, data)
}

// colToLetter converte número de coluna para letra (1=A, 27=AA)
func colToLetter(col int) string {
	result := ""
	for col > 0 {
		col--
		result = string(rune('A'+(col%26))) + result
		col = col / 26
	}
	return result
}
