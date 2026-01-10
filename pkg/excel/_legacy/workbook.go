// workbook.go - Excel workbook and sheet management operations
package excel

import (
	"fmt"

	"github.com/go-ole/go-ole/oleutil"
)

// GetOpenWorkbooks retorna lista de pastas de trabalho abertas
func (c *Client) GetOpenWorkbooks() ([]Workbook, error) {
	return runOnCOMThreadWithResult(c, func() ([]Workbook, error) {
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return nil, err
		}
		defer workbooks.ToIDispatch().Release()

		count, err := oleutil.GetProperty(workbooks.ToIDispatch(), "Count")
		if err != nil {
			return nil, err
		}

		var result []Workbook
		for i := 1; i <= int(count.Val); i++ {
			wb, err := oleutil.GetProperty(workbooks.ToIDispatch(), "Item", i)
			if err != nil {
				continue
			}
			wbDisp := wb.ToIDispatch()

			name, _ := oleutil.GetProperty(wbDisp, "Name")
			path, _ := oleutil.GetProperty(wbDisp, "Path")

			sheets, err := c.getSheetsInternal(wbDisp)
			if err != nil {
				sheets = []string{}
			}

			result = append(result, Workbook{
				Name:   name.ToString(),
				Path:   path.ToString(),
				Sheets: sheets,
			})
			wbDisp.Release()
		}

		return result, nil
	})
}

// InsertNewSheet cria uma nova aba na pasta de trabalho
func (c *Client) InsertNewSheet(workbookName, sheetName string) error {
	return c.runOnCOMThread(func() error {
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return err
		}
		defer workbooks.ToIDispatch().Release()

		wb, err := oleutil.GetProperty(workbooks.ToIDispatch(), "Item", workbookName)
		if err != nil {
			return fmt.Errorf("pasta de trabalho não encontrada: %w", err)
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		sheets, err := oleutil.GetProperty(wbDisp, "Sheets")
		if err != nil {
			return err
		}
		sheetsDisp := sheets.ToIDispatch()
		defer sheetsDisp.Release()

		newSheet, err := oleutil.CallMethod(sheetsDisp, "Add")
		if err != nil {
			return fmt.Errorf("falha ao criar aba: %w", err)
		}
		newSheetDisp := newSheet.ToIDispatch()
		defer newSheetDisp.Release()

		_, err = oleutil.PutProperty(newSheetDisp, "Name", sheetName)
		if err != nil {
			return fmt.Errorf("falha ao renomear aba: %w", err)
		}

		return nil
	})
}

// CreateNewWorkbook cria uma nova pasta de trabalho
func (c *Client) CreateNewWorkbook() (string, error) {
	return runOnCOMThreadWithResult(c, func() (string, error) {
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return "", err
		}
		defer workbooks.ToIDispatch().Release()

		wb, err := oleutil.CallMethod(workbooks.ToIDispatch(), "Add")
		if err != nil {
			return "", err
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		name, _ := oleutil.GetProperty(wbDisp, "Name")
		return name.ToString(), nil
	})
}

// GetActiveWorkbookAndSheet retorna o nome da pasta e aba ativa
func (c *Client) GetActiveWorkbookAndSheet() (workbook, sheet string, err error) {
	type result struct {
		workbook string
		sheet    string
	}
	res, err := runOnCOMThreadWithResult(c, func() (result, error) {
		activeWb, err := oleutil.GetProperty(c.excelApp, "ActiveWorkbook")
		if err != nil {
			return result{}, err
		}
		wbDisp := activeWb.ToIDispatch()
		if wbDisp == nil {
			return result{}, fmt.Errorf("nenhuma pasta de trabalho ativa")
		}
		defer wbDisp.Release()

		wbName, _ := oleutil.GetProperty(wbDisp, "Name")

		activeSheet, err := oleutil.GetProperty(c.excelApp, "ActiveSheet")
		if err != nil {
			return result{workbook: wbName.ToString()}, err
		}
		sheetDisp := activeSheet.ToIDispatch()
		if sheetDisp == nil {
			return result{workbook: wbName.ToString()}, fmt.Errorf("nenhuma aba ativa")
		}
		defer sheetDisp.Release()

		sheetNameProp, _ := oleutil.GetProperty(sheetDisp, "Name")

		return result{
			workbook: wbName.ToString(),
			sheet:    sheetNameProp.ToString(),
		}, nil
	})
	return res.workbook, res.sheet, err
}

// ListSheets retorna a lista de abas de um workbook
func (c *Client) ListSheets(workbookName string) ([]string, error) {
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

		countVar, _ := oleutil.GetProperty(sheetsDisp, "Count")
		count := int(countVar.Val)

		var names []string
		for i := 1; i <= count; i++ {
			sheet, err := oleutil.GetProperty(sheetsDisp, "Item", i)
			if err != nil {
				continue
			}
			sheetDisp := sheet.ToIDispatch()
			nameVar, _ := oleutil.GetProperty(sheetDisp, "Name")
			names = append(names, nameVar.ToString())
			sheetDisp.Release()
		}

		return names, nil
	})
}

// SheetExists verifica se uma aba existe
func (c *Client) SheetExists(workbookName, sheetName string) (bool, error) {
	sheets, err := c.ListSheets(workbookName)
	if err != nil {
		return false, err
	}
	for _, s := range sheets {
		if s == sheetName {
			return true, nil
		}
	}
	return false, nil
}
