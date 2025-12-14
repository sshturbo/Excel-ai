// data.go - Excel data reading and writing operations
package excel

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// GetActiveSelection retorna os dados da seleção atual
func (c *Client) GetActiveSelection() (*SheetData, error) {
	return runOnCOMThreadWithResult(c, func() (*SheetData, error) {
		selection, err := oleutil.GetProperty(c.excelApp, "Selection")
		if err != nil {
			return nil, fmt.Errorf("falha ao obter seleção: %w", err)
		}
		selDisp := selection.ToIDispatch()
		defer selDisp.Release()

		return c.readRangeDataInternal(selDisp)
	})
}

// GetSheetData lê dados de uma planilha específica
func (c *Client) GetSheetData(workbookName, sheetName string, maxRows int) (*SheetData, error) {
	return runOnCOMThreadWithResult(c, func() (*SheetData, error) {
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return nil, err
		}
		defer workbooks.ToIDispatch().Release()

		wb, err := oleutil.GetProperty(workbooks.ToIDispatch(), "Item", workbookName)
		if err != nil {
			return nil, fmt.Errorf("pasta de trabalho '%s' não encontrada: %w", workbookName, err)
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		sheets, err := oleutil.GetProperty(wbDisp, "Sheets")
		if err != nil {
			return nil, err
		}
		defer sheets.ToIDispatch().Release()

		sheet, err := oleutil.GetProperty(sheets.ToIDispatch(), "Item", sheetName)
		if err != nil {
			return nil, fmt.Errorf("aba '%s' não encontrada: %w", sheetName, err)
		}
		sheetDisp := sheet.ToIDispatch()
		defer sheetDisp.Release()

		usedRange, err := oleutil.GetProperty(sheetDisp, "UsedRange")
		if err != nil {
			return nil, err
		}
		rangeDisp := usedRange.ToIDispatch()
		defer rangeDisp.Release()

		return c.readRangeDataInternal(rangeDisp)
	})
}

// readRangeDataInternal lê os dados de um range do Excel (chamado internamente na thread COM)
func (c *Client) readRangeDataInternal(rangeDisp *ole.IDispatch) (*SheetData, error) {
	rows, err := oleutil.GetProperty(rangeDisp, "Rows")
	if err != nil {
		return nil, err
	}
	defer rows.ToIDispatch().Release()

	cols, err := oleutil.GetProperty(rangeDisp, "Columns")
	if err != nil {
		return nil, err
	}
	defer cols.ToIDispatch().Release()

	rowCount, _ := oleutil.GetProperty(rows.ToIDispatch(), "Count")
	colCount, _ := oleutil.GetProperty(cols.ToIDispatch(), "Count")

	numRows := int(rowCount.Val)
	numCols := int(colCount.Val)

	// Limitar linhas para performance
	if numRows > 100 {
		numRows = 100
	}

	data := &SheetData{
		Headers: []string{},
		Rows:    [][]CellData{},
	}

	// Ler cabeçalhos (primeira linha)
	for col := 1; col <= numCols; col++ {
		cell, err := oleutil.GetProperty(rangeDisp, "Cells", 1, col)
		if err != nil {
			data.Headers = append(data.Headers, "")
			continue
		}
		cellDisp := cell.ToIDispatch()
		text, _ := oleutil.GetProperty(cellDisp, "Text")
		data.Headers = append(data.Headers, text.ToString())
		cellDisp.Release()
	}

	// Ler dados (a partir da segunda linha)
	for row := 2; row <= numRows; row++ {
		rowData := []CellData{}
		for col := 1; col <= numCols; col++ {
			cell, err := oleutil.GetProperty(rangeDisp, "Cells", row, col)
			if err != nil {
				rowData = append(rowData, CellData{Row: row, Col: col})
				continue
			}
			cellDisp := cell.ToIDispatch()
			value, _ := oleutil.GetProperty(cellDisp, "Value")
			text, _ := oleutil.GetProperty(cellDisp, "Text")

			rowData = append(rowData, CellData{
				Row:   row,
				Col:   col,
				Value: value.Value(),
				Text:  text.ToString(),
			})
			cellDisp.Release()
		}
		data.Rows = append(data.Rows, rowData)
	}

	return data, nil
}

// WriteCell escreve um valor em uma célula específica
func (c *Client) WriteCell(workbookName, sheetName string, row, col int, value interface{}) error {
	return c.runOnCOMThread(func() error {
		sheetDisp, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheetDisp.Release()

		cell, err := oleutil.GetProperty(sheetDisp, "Cells", row, col)
		if err != nil {
			return fmt.Errorf("falha ao acessar célula: %w", err)
		}
		cellDisp := cell.ToIDispatch()
		defer cellDisp.Release()

		_, err = oleutil.PutProperty(cellDisp, "Value", value)
		if err != nil {
			return fmt.Errorf("falha ao escrever valor: %w", err)
		}

		return nil
	})
}

// WriteRange escreve múltiplos valores em um range
func (c *Client) WriteRange(workbookName, sheetName string, startRow, startCol int, data [][]interface{}) error {
	return c.runOnCOMThread(func() error {
		sheetDisp, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheetDisp.Release()

		for rowIdx, rowData := range data {
			for colIdx, value := range rowData {
				cell, err := oleutil.GetProperty(sheetDisp, "Cells", startRow+rowIdx, startCol+colIdx)
				if err != nil {
					continue
				}
				cellDisp := cell.ToIDispatch()
				oleutil.PutProperty(cellDisp, "Value", value)
				cellDisp.Release()
			}
		}

		return nil
	})
}

// ApplyFormula aplica uma fórmula em uma célula
func (c *Client) ApplyFormula(workbookName, sheetName string, row, col int, formula string) error {
	return c.runOnCOMThread(func() error {
		sheetDisp, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheetDisp.Release()

		cell, err := oleutil.GetProperty(sheetDisp, "Cells", row, col)
		if err != nil {
			return fmt.Errorf("falha ao acessar célula: %w", err)
		}
		cellDisp := cell.ToIDispatch()
		defer cellDisp.Release()

		_, err = oleutil.PutProperty(cellDisp, "Formula", formula)
		if err != nil {
			return fmt.Errorf("falha ao aplicar fórmula: %w", err)
		}

		return nil
	})
}

// GetCellValue lê o valor de uma célula específica
func (c *Client) GetCellValue(workbookName, sheetName, cellAddress string) (string, error) {
	return runOnCOMThreadWithResult(c, func() (string, error) {
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return "", err
		}
		defer workbooks.ToIDispatch().Release()

		wb, err := oleutil.GetProperty(workbooks.ToIDispatch(), "Item", workbookName)
		if err != nil {
			return "", fmt.Errorf("pasta de trabalho '%s' não encontrada: %w", workbookName, err)
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		sheets, err := oleutil.GetProperty(wbDisp, "Sheets")
		if err != nil {
			return "", err
		}
		defer sheets.ToIDispatch().Release()

		sheet, err := oleutil.GetProperty(sheets.ToIDispatch(), "Item", sheetName)
		if err != nil {
			return "", fmt.Errorf("aba '%s' não encontrada: %w", sheetName, err)
		}
		sheetDisp := sheet.ToIDispatch()
		defer sheetDisp.Release()

		cell, err := oleutil.GetProperty(sheetDisp, "Range", cellAddress)
		if err != nil {
			return "", fmt.Errorf("célula '%s' inválida: %w", cellAddress, err)
		}
		cellDisp := cell.ToIDispatch()
		defer cellDisp.Release()

		val, err := oleutil.GetProperty(cellDisp, "Value")
		if err != nil {
			return "", nil // Célula vazia ou erro ao ler
		}

		if val.Value() == nil {
			return "", nil
		}

		return val.ToString(), nil
	})
}

// SetCellValue escreve um valor em uma célula específica
func (c *Client) SetCellValue(workbookName, sheetName, cellAddress, value string) error {
	return c.runOnCOMThread(func() error {
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return err
		}
		defer workbooks.ToIDispatch().Release()

		wb, err := oleutil.GetProperty(workbooks.ToIDispatch(), "Item", workbookName)
		if err != nil {
			return fmt.Errorf("pasta de trabalho '%s' não encontrada: %w", workbookName, err)
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		sheets, err := oleutil.GetProperty(wbDisp, "Sheets")
		if err != nil {
			return err
		}
		defer sheets.ToIDispatch().Release()

		sheet, err := oleutil.GetProperty(sheets.ToIDispatch(), "Item", sheetName)
		if err != nil {
			return fmt.Errorf("aba '%s' não encontrada: %w", sheetName, err)
		}
		sheetDisp := sheet.ToIDispatch()
		defer sheetDisp.Release()

		cell, err := oleutil.GetProperty(sheetDisp, "Range", cellAddress)
		if err != nil {
			return fmt.Errorf("célula '%s' inválida: %w", cellAddress, err)
		}
		cellDisp := cell.ToIDispatch()
		defer cellDisp.Release()

		_, err = oleutil.PutProperty(cellDisp, "Value", value)
		return err
	})
}
