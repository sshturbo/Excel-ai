// charts.go - Excel chart operations
package excel

import (
	"fmt"

	"github.com/go-ole/go-ole/oleutil"
)

// CreateChart cria um gráfico a partir de um range
func (c *Client) CreateChart(workbookName, sheetName, rangeAddress, chartType, title string) error {
	return c.runOnCOMThread(func() error {
		sheetDisp, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheetDisp.Release()

		// Selecionar dados
		sourceRange, err := oleutil.GetProperty(sheetDisp, "Range", rangeAddress)
		if err != nil {
			return fmt.Errorf("range inválido: %w", err)
		}
		sourceRangeDisp := sourceRange.ToIDispatch()
		defer sourceRangeDisp.Release()

		// Criar gráfico (Shapes.AddChart2)
		shapes, err := oleutil.GetProperty(sheetDisp, "Shapes")
		if err != nil {
			return err
		}
		shapesDisp := shapes.ToIDispatch()
		defer shapesDisp.Release()

		// 201 = xlColumnClustered (padrão)
		// -1 = Style default
		chartShape, err := oleutil.CallMethod(shapesDisp, "AddChart2", -1, 201)
		if err != nil {
			// Fallback para AddChart (Excel antigo)
			chartObjects, err := oleutil.GetProperty(sheetDisp, "ChartObjects")
			if err != nil {
				return fmt.Errorf("falha ao criar gráfico: %w", err)
			}
			chartObjectsDisp := chartObjects.ToIDispatch()
			defer chartObjectsDisp.Release()

			chartShape, err = oleutil.CallMethod(chartObjectsDisp, "Add", 10, 10, 300, 200)
			if err != nil {
				return fmt.Errorf("falha ao criar gráfico: %w", err)
			}
		}
		chartShapeDisp := chartShape.ToIDispatch()
		defer chartShapeDisp.Release()

		chart, err := oleutil.GetProperty(chartShapeDisp, "Chart")
		if err != nil {
			return err
		}
		chartDisp := chart.ToIDispatch()
		defer chartDisp.Release()

		// Definir dados
		oleutil.CallMethod(chartDisp, "SetSourceData", sourceRangeDisp)

		// Definir título
		if title != "" {
			oleutil.PutProperty(chartDisp, "HasTitle", true)
			chartTitle, err := oleutil.GetProperty(chartDisp, "ChartTitle")
			if err == nil {
				oleutil.PutProperty(chartTitle.ToIDispatch(), "Text", title)
				chartTitle.ToIDispatch().Release()
			}
		}

		return nil
	})
}

// ListCharts lista os gráficos em uma aba
func (c *Client) ListCharts(workbookName, sheetName string) ([]string, error) {
	return runOnCOMThreadWithResult(c, func() ([]string, error) {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return nil, err
		}
		defer sheet.Release()

		chartObjs, err := oleutil.GetProperty(sheet, "ChartObjects")
		if err != nil {
			return []string{}, nil
		}
		chartObjsDisp := chartObjs.ToIDispatch()
		defer chartObjsDisp.Release()

		countVar, _ := oleutil.GetProperty(chartObjsDisp, "Count")
		count := int(countVar.Val)

		charts := make([]string, 0, count)
		for i := 1; i <= count; i++ {
			item, err := oleutil.GetProperty(chartObjsDisp, "Item", i)
			if err != nil {
				continue
			}
			itemDisp := item.ToIDispatch()
			nameVar, _ := oleutil.GetProperty(itemDisp, "Name")
			charts = append(charts, nameVar.ToString())
			itemDisp.Release()
		}

		return charts, nil
	})
}

// DeleteChart exclui um gráfico pelo nome
func (c *Client) DeleteChart(workbookName, sheetName, chartName string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		chartObjs, err := oleutil.GetProperty(sheet, "ChartObjects")
		if err != nil {
			return fmt.Errorf("erro ao acessar gráficos: %w", err)
		}
		chartObjsDisp := chartObjs.ToIDispatch()
		defer chartObjsDisp.Release()

		chart, err := oleutil.GetProperty(chartObjsDisp, "Item", chartName)
		if err != nil {
			return fmt.Errorf("gráfico '%s' não encontrado: %w", chartName, err)
		}
		chartDisp := chart.ToIDispatch()
		defer chartDisp.Release()

		_, err = oleutil.CallMethod(chartDisp, "Delete")
		return err
	})
}
