// table.go - Excel table (ListObject) operations
package excel

import (
	"fmt"

	"github.com/go-ole/go-ole/oleutil"
)

// CreateTable cria uma tabela formatada (ListObject) a partir de um range
// style: nome do estilo como "TableStyleLight1", "TableStyleMedium2", etc
func (c *Client) CreateTable(workbookName, sheetName, rangeAddr, tableName, style string) error {
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

		// Obter coleção ListObjects
		listObjects, err := oleutil.GetProperty(sheet, "ListObjects")
		if err != nil {
			return fmt.Errorf("erro ao acessar ListObjects: %w", err)
		}
		listObjectsDisp := listObjects.ToIDispatch()
		defer listObjectsDisp.Release()

		// Criar tabela: ListObjects.Add(xlSrcRange=1, Source, , XlListObjectHasHeaders.xlYes=1)
		table, err := oleutil.CallMethod(listObjectsDisp, "Add", 1, rangeDisp, nil, 1)
		if err != nil {
			return fmt.Errorf("erro ao criar tabela: %w", err)
		}
		tableDisp := table.ToIDispatch()
		defer tableDisp.Release()

		// Definir nome da tabela
		if tableName != "" {
			oleutil.PutProperty(tableDisp, "Name", tableName)
		}

		// Aplicar estilo
		if style != "" {
			oleutil.PutProperty(tableDisp, "TableStyle", style)
		} else {
			oleutil.PutProperty(tableDisp, "TableStyle", "TableStyleMedium2")
		}

		return nil
	})
}

// ListTables lista as tabelas (ListObjects) em uma aba
func (c *Client) ListTables(workbookName, sheetName string) ([]string, error) {
	return runOnCOMThreadWithResult(c, func() ([]string, error) {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return nil, err
		}
		defer sheet.Release()

		listObjects, err := oleutil.GetProperty(sheet, "ListObjects")
		if err != nil {
			return []string{}, nil
		}
		listObjectsDisp := listObjects.ToIDispatch()
		defer listObjectsDisp.Release()

		countVar, _ := oleutil.GetProperty(listObjectsDisp, "Count")
		count := int(countVar.Val)

		tables := make([]string, 0, count)
		for i := 1; i <= count; i++ {
			item, err := oleutil.GetProperty(listObjectsDisp, "Item", i)
			if err != nil {
				continue
			}
			itemDisp := item.ToIDispatch()
			nameVar, _ := oleutil.GetProperty(itemDisp, "Name")
			tables = append(tables, nameVar.ToString())
			itemDisp.Release()
		}

		return tables, nil
	})
}

// DeleteTable exclui uma tabela (converte para range)
func (c *Client) DeleteTable(workbookName, sheetName, tableName string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		listObjects, err := oleutil.GetProperty(sheet, "ListObjects")
		if err != nil {
			return fmt.Errorf("erro ao acessar ListObjects: %w", err)
		}
		listObjectsDisp := listObjects.ToIDispatch()
		defer listObjectsDisp.Release()

		table, err := oleutil.GetProperty(listObjectsDisp, "Item", tableName)
		if err != nil {
			return fmt.Errorf("tabela '%s' não encontrada: %w", tableName, err)
		}
		tableDisp := table.ToIDispatch()
		defer tableDisp.Release()

		// Unlist converte tabela para range normal
		_, err = oleutil.CallMethod(tableDisp, "Unlist")
		return err
	})
}
