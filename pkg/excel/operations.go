// operations.go - Excel general data operations (copy, sort, filter, rows)
package excel

import (
	"fmt"

	"github.com/go-ole/go-ole/oleutil"
)

// CopyRange copia um range para outro local
func (c *Client) CopyRange(workbookName, sheetName, sourceRange, destRange string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		srcObj, err := oleutil.GetProperty(sheet, "Range", sourceRange)
		if err != nil {
			return fmt.Errorf("range origem '%s' inválido: %w", sourceRange, err)
		}
		srcDisp := srcObj.ToIDispatch()
		defer srcDisp.Release()

		destObj, err := oleutil.GetProperty(sheet, "Range", destRange)
		if err != nil {
			return fmt.Errorf("range destino '%s' inválido: %w", destRange, err)
		}
		destDisp := destObj.ToIDispatch()
		defer destDisp.Release()

		_, err = oleutil.CallMethod(srcDisp, "Copy", destDisp)
		return err
	})
}

// SortRange ordena dados em um range
// ascending: true para A-Z, false para Z-A
func (c *Client) SortRange(workbookName, sheetName, rangeAddr string, column int, ascending bool) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		// Obter célula de ordenação (Key1)
		// Estratégia: Range.Columns(col).Cells(1)
		columnsObj, err := oleutil.GetProperty(rangeDisp, "Columns", column)
		if err != nil {
			return fmt.Errorf("falha ao selecionar coluna %d do range: %w", column, err)
		}
		columnsDisp := columnsObj.ToIDispatch()
		defer columnsDisp.Release()

		// Pegar primeira célula dessa coluna para usar como chave
		keyCell, err := oleutil.GetProperty(columnsDisp, "Cells", 1)
		if err != nil {
			return fmt.Errorf("falha ao obter KeyCell (Cells(1)) da coluna: %w", err)
		}
		keyCellDisp := keyCell.ToIDispatch()
		defer keyCellDisp.Release()

		order := 1 // xlAscending
		if !ascending {
			order = 2 // xlDescending
		}

		_, err = oleutil.CallMethod(rangeDisp, "Sort", keyCellDisp, order)
		return err
	})
}

// ApplyFilter aplica filtro automático a um range
func (c *Client) ApplyFilter(workbookName, sheetName, rangeAddr string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		_, err = oleutil.CallMethod(rangeDisp, "AutoFilter")
		return err
	})
}

// ClearFilters remove todos os filtros de uma aba
func (c *Client) ClearFilters(workbookName, sheetName string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		// Tentar desativar filtro (AutoFilterMode = false) blindamente.
		// Se não houver suporte ou falhar, ignoramos.
		oleutil.PutProperty(sheet, "AutoFilterMode", false)

		// Tenta ShowAllData para garantir que linhas ocultas sejam mostradas.
		// CallMethod retornará erro se não houver dados filtrados, ignoramos.
		oleutil.CallMethod(sheet, "ShowAllData")

		return nil
	})
}

// InsertRows insere linhas em uma posição específica
func (c *Client) InsertRows(workbookName, sheetName string, rowNumber, count int) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		for i := 0; i < count; i++ {
			// Selecionar linha
			rowAddr := fmt.Sprintf("%d:%d", rowNumber, rowNumber)
			rowObj, err := oleutil.GetProperty(sheet, "Range", rowAddr)
			if err != nil {
				return err
			}
			rowDisp := rowObj.ToIDispatch()
			_, err = oleutil.CallMethod(rowDisp, "Insert")
			rowDisp.Release()
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteRows exclui linhas a partir de uma posição
func (c *Client) DeleteRows(workbookName, sheetName string, rowNumber, count int) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		// Desabilitar alertas
		oleutil.PutProperty(c.excelApp, "DisplayAlerts", false)
		defer oleutil.PutProperty(c.excelApp, "DisplayAlerts", true)

		// Selecionar range de linhas
		rowAddr := fmt.Sprintf("%d:%d", rowNumber, rowNumber+count-1)
		// Em COM, sheet.Range("1:1") funciona melhor que sheet.Rows("1:1") as vezes
		rowObj, err := oleutil.GetProperty(sheet, "Range", rowAddr)
		if err != nil {
			return fmt.Errorf("falha ao selecionar linhas %s: %w", rowAddr, err)
		}
		rowDisp := rowObj.ToIDispatch()
		defer rowDisp.Release()

		_, err = oleutil.CallMethod(rowDisp, "Delete")
		return err
	})
}

// DeleteSheet exclui uma aba da pasta de trabalho
func (c *Client) DeleteSheet(workbookName, sheetName string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		// Desabilitar alertas para não pedir confirmação
		oleutil.PutProperty(c.excelApp, "DisplayAlerts", false)
		defer oleutil.PutProperty(c.excelApp, "DisplayAlerts", true)

		_, err = oleutil.CallMethod(sheet, "Delete")
		return err
	})
}

// RenameSheet renomeia uma aba
func (c *Client) RenameSheet(workbookName, oldName, newName string) error {
	if oldName == "" {
		return fmt.Errorf("nome antigo da aba está vazio")
	}
	if newName == "" {
		return fmt.Errorf("nome novo da aba está vazio")
	}

	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, oldName)
		if err != nil {
			return fmt.Errorf("aba '%s' não encontrada no workbook '%s': %w", oldName, workbookName, err)
		}
		defer sheet.Release()

		_, err = oleutil.PutProperty(sheet, "Name", newName)
		if err != nil {
			return fmt.Errorf("falha ao renomear aba '%s' para '%s': %w", oldName, newName, err)
		}
		return nil
	})
}
