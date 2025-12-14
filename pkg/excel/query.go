// query.go - Excel query and information operations
package excel

import (
	"fmt"

	"github.com/go-ole/go-ole/oleutil"
)

// GetRowCount retorna o número de linhas com dados
func (c *Client) GetRowCount(workbookName, sheetName string) (int, error) {
	return runOnCOMThreadWithResult(c, func() (int, error) {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return 0, err
		}
		defer sheet.Release()

		usedRange, err := oleutil.GetProperty(sheet, "UsedRange")
		if err != nil {
			return 0, err
		}
		usedRangeDisp := usedRange.ToIDispatch()
		defer usedRangeDisp.Release()

		rows, _ := oleutil.GetProperty(usedRangeDisp, "Rows")
		rowsDisp := rows.ToIDispatch()
		defer rowsDisp.Release()

		countVar, _ := oleutil.GetProperty(rowsDisp, "Count")
		return int(countVar.Val), nil
	})
}

// GetColumnCount retorna o número de colunas com dados
func (c *Client) GetColumnCount(workbookName, sheetName string) (int, error) {
	return runOnCOMThreadWithResult(c, func() (int, error) {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return 0, err
		}
		defer sheet.Release()

		usedRange, err := oleutil.GetProperty(sheet, "UsedRange")
		if err != nil {
			return 0, err
		}
		usedRangeDisp := usedRange.ToIDispatch()
		defer usedRangeDisp.Release()

		cols, _ := oleutil.GetProperty(usedRangeDisp, "Columns")
		colsDisp := cols.ToIDispatch()
		defer colsDisp.Release()

		countVar, _ := oleutil.GetProperty(colsDisp, "Count")
		return int(countVar.Val), nil
	})
}

// GetCellFormula retorna a fórmula de uma célula
func (c *Client) GetCellFormula(workbookName, sheetName, cellAddress string) (string, error) {
	return runOnCOMThreadWithResult(c, func() (string, error) {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return "", err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", cellAddress)
		if err != nil {
			return "", fmt.Errorf("célula '%s' inválida: %w", cellAddress, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		formula, _ := oleutil.GetProperty(rangeDisp, "Formula")
		return fmt.Sprintf("%v", formula.Value()), nil
	})
}

// HasFilter verifica se a aba tem filtro ativo
func (c *Client) HasFilter(workbookName, sheetName string) (bool, error) {
	return runOnCOMThreadWithResult(c, func() (bool, error) {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return false, err
		}
		defer sheet.Release()

		autoFilterMode, _ := oleutil.GetProperty(sheet, "AutoFilterMode")
		return autoFilterMode.Val != 0, nil
	})
}

// GetActiveCell retorna o endereço da célula ativa
func (c *Client) GetActiveCell() (string, error) {
	return runOnCOMThreadWithResult(c, func() (string, error) {
		activeCell, err := oleutil.GetProperty(c.excelApp, "ActiveCell")
		if err != nil {
			return "", err
		}
		activeCellDisp := activeCell.ToIDispatch()
		defer activeCellDisp.Release()

		addrVar, _ := oleutil.GetProperty(activeCellDisp, "Address")
		return addrVar.ToString(), nil
	})
}

// GetRangeValues retorna os valores de um range como array 2D
func (c *Client) GetRangeValues(workbookName, sheetName, rangeAddr string) ([][]string, error) {
	return runOnCOMThreadWithResult(c, func() ([][]string, error) {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return nil, err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return nil, fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		// Obter número de linhas e colunas
		rowsObj, _ := oleutil.GetProperty(rangeDisp, "Rows")
		rowsDisp := rowsObj.ToIDispatch()
		rowCountVar, _ := oleutil.GetProperty(rowsDisp, "Count")
		rowCount := int(rowCountVar.Val)
		rowsDisp.Release()

		colsObj, _ := oleutil.GetProperty(rangeDisp, "Columns")
		colsDisp := colsObj.ToIDispatch()
		colCountVar, _ := oleutil.GetProperty(colsDisp, "Count")
		colCount := int(colCountVar.Val)
		colsDisp.Release()

		// Limitar para não sobrecarregar
		if rowCount > 100 {
			rowCount = 100
		}
		if colCount > 26 {
			colCount = 26
		}

		result := make([][]string, rowCount)
		for i := 0; i < rowCount; i++ {
			result[i] = make([]string, colCount)
			for j := 0; j < colCount; j++ {
				cell, _ := oleutil.GetProperty(rangeDisp, "Cells", i+1, j+1)
				cellDisp := cell.ToIDispatch()
				valVar, _ := oleutil.GetProperty(cellDisp, "Value")
				if valVar.Val != 0 {
					result[i][j] = fmt.Sprintf("%v", valVar.Value())
				}
				cellDisp.Release()
			}
		}

		return result, nil
	})
}

// GetHeaders retorna os cabeçalhos (primeira linha) de um range
func (c *Client) GetHeaders(workbookName, sheetName, rangeAddr string) ([]string, error) {
	return runOnCOMThreadWithResult(c, func() ([]string, error) {
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return nil, err
		}
		workbooksDisp := workbooks.ToIDispatch()
		defer workbooksDisp.Release()

		wb, err := oleutil.GetProperty(workbooksDisp, "Item", workbookName)
		if err != nil {
			return nil, fmt.Errorf("workbook '%s' não encontrado", workbookName)
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		sheets, err := oleutil.GetProperty(wbDisp, "Worksheets")
		if err != nil {
			return nil, err
		}
		sheetsDisp := sheets.ToIDispatch()
		defer sheetsDisp.Release()

		sheet, err := oleutil.GetProperty(sheetsDisp, "Item", sheetName)
		if err != nil {
			return nil, fmt.Errorf("aba '%s' não encontrada", sheetName)
		}
		sheetDisp := sheet.ToIDispatch()
		defer sheetDisp.Release()

		// Obter range
		rng, err := oleutil.GetProperty(sheetDisp, "Range", rangeAddr)
		if err != nil {
			return nil, fmt.Errorf("range '%s' inválido", rangeAddr)
		}
		rngDisp := rng.ToIDispatch()
		defer rngDisp.Release()

		// Pegar primeira linha do range
		firstRow, err := oleutil.GetProperty(rngDisp, "Rows", 1)
		if err != nil {
			return nil, err
		}
		firstRowDisp := firstRow.ToIDispatch()
		defer firstRowDisp.Release()

		// Contar colunas
		colsVar, _ := oleutil.GetProperty(firstRowDisp, "Columns")
		colsDisp := colsVar.ToIDispatch()
		defer colsDisp.Release()

		countVar, _ := oleutil.GetProperty(colsDisp, "Count")
		count := int(countVar.Val)

		var headers []string
		for i := 1; i <= count; i++ {
			cell, err := oleutil.GetProperty(firstRowDisp, "Cells", 1, i)
			if err != nil {
				continue
			}
			cellDisp := cell.ToIDispatch()
			valueVar, _ := oleutil.GetProperty(cellDisp, "Value")

			var header string
			if valueVar.Val != 0 {
				header = fmt.Sprintf("%v", valueVar.Value())
			}
			headers = append(headers, header)
			cellDisp.Release()
		}

		return headers, nil
	})
}

// GetUsedRange retorna o endereço do range utilizado em uma aba
func (c *Client) GetUsedRange(workbookName, sheetName string) (string, error) {
	return runOnCOMThreadWithResult(c, func() (string, error) {
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return "", err
		}
		workbooksDisp := workbooks.ToIDispatch()
		defer workbooksDisp.Release()

		wb, err := oleutil.GetProperty(workbooksDisp, "Item", workbookName)
		if err != nil {
			return "", fmt.Errorf("workbook '%s' não encontrado", workbookName)
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		sheets, err := oleutil.GetProperty(wbDisp, "Worksheets")
		if err != nil {
			return "", err
		}
		sheetsDisp := sheets.ToIDispatch()
		defer sheetsDisp.Release()

		sheet, err := oleutil.GetProperty(sheetsDisp, "Item", sheetName)
		if err != nil {
			return "", fmt.Errorf("aba '%s' não encontrada", sheetName)
		}
		sheetDisp := sheet.ToIDispatch()
		defer sheetDisp.Release()

		usedRange, err := oleutil.GetProperty(sheetDisp, "UsedRange")
		if err != nil {
			return "", err
		}
		usedRangeDisp := usedRange.ToIDispatch()
		defer usedRangeDisp.Release()

		addrVar, _ := oleutil.GetProperty(usedRangeDisp, "Address")
		return addrVar.ToString(), nil
	})
}
