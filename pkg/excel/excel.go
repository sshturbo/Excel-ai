package excel

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// Client gerencia a conexão COM com o Excel usando uma thread dedicada
type Client struct {
	excelApp *ole.IDispatch
	cmdChan  chan func()
	doneChan chan struct{}
	mu       sync.Mutex
}

// Workbook representa uma pasta de trabalho aberta
type Workbook struct {
	Name   string   `json:"name"`
	Path   string   `json:"path"`
	Sheets []string `json:"sheets"`
}

// CellData representa dados de uma célula
type CellData struct {
	Row   int         `json:"row"`
	Col   int         `json:"col"`
	Value interface{} `json:"value"`
	Text  string      `json:"text"`
}

// SheetData representa dados de uma planilha
type SheetData struct {
	Name    string       `json:"name"`
	Headers []string     `json:"headers"`
	Rows    [][]CellData `json:"rows"`
}

// NewClient cria um novo cliente Excel conectando-se a uma instância existente
func NewClient() (*Client, error) {
	c := &Client{
		cmdChan:  make(chan func()),
		doneChan: make(chan struct{}),
	}

	errChan := make(chan error, 1)

	// Iniciar goroutine dedicada para operações COM
	go func() {
		// Bloquear esta goroutine na mesma thread do SO
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// Inicializar COM nesta thread
		ole.CoInitialize(0)
		defer ole.CoUninitialize()

		// Conectar ao Excel
		unknown, err := oleutil.GetActiveObject("Excel.Application")
		if err != nil {
			errChan <- fmt.Errorf("nenhuma instância do Excel encontrada: %w", err)
			return
		}

		excelApp, err := unknown.QueryInterface(ole.IID_IDispatch)
		if err != nil {
			errChan <- fmt.Errorf("falha ao obter interface Excel: %w", err)
			return
		}

		c.excelApp = excelApp
		errChan <- nil

		// Loop de processamento de comandos
		for {
			select {
			case cmd := <-c.cmdChan:
				cmd()
			case <-c.doneChan:
				if c.excelApp != nil {
					c.excelApp.Release()
				}
				return
			}
		}
	}()

	// Aguardar inicialização
	if err := <-errChan; err != nil {
		return nil, err
	}

	return c, nil
}

// runOnCOMThread executa uma função na thread COM e retorna o resultado
func (c *Client) runOnCOMThread(fn func() error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	errChan := make(chan error, 1)
	c.cmdChan <- func() {
		errChan <- fn()
	}
	return <-errChan
}

// runOnCOMThreadWithResult executa uma função e retorna resultado genérico
func runOnCOMThreadWithResult[T any](c *Client, fn func() (T, error)) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	type result struct {
		value T
		err   error
	}
	resChan := make(chan result, 1)
	c.cmdChan <- func() {
		v, err := fn()
		resChan <- result{v, err}
	}
	res := <-resChan
	return res.value, res.err
}

// Close libera os recursos COM
func (c *Client) Close() {
	close(c.doneChan)
}

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

// getSheetsInternal retorna as abas de uma pasta de trabalho (chamado internamente na thread COM)
func (c *Client) getSheetsInternal(wb *ole.IDispatch) ([]string, error) {
	sheets, err := oleutil.GetProperty(wb, "Sheets")
	if err != nil {
		return nil, err
	}
	defer sheets.ToIDispatch().Release()

	count, err := oleutil.GetProperty(sheets.ToIDispatch(), "Count")
	if err != nil {
		return nil, err
	}

	var result []string
	for i := 1; i <= int(count.Val); i++ {
		sheet, err := oleutil.GetProperty(sheets.ToIDispatch(), "Item", i)
		if err != nil {
			continue
		}
		sheetDisp := sheet.ToIDispatch()

		name, _ := oleutil.GetProperty(sheetDisp, "Name")
		result = append(result, name.ToString())
		sheetDisp.Release()
	}

	return result, nil
}

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

// getWorkbookInternal obtém uma referência para uma pasta de trabalho (chamado internamente na thread COM)
func (c *Client) getWorkbookInternal(workbookName string) (*ole.IDispatch, error) {
	workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
	if err != nil {
		return nil, err
	}
	defer workbooks.ToIDispatch().Release()

	wb, err := oleutil.GetProperty(workbooks.ToIDispatch(), "Item", workbookName)
	if err != nil {
		return nil, fmt.Errorf("pasta de trabalho '%s' não encontrada: %w", workbookName, err)
	}
	return wb.ToIDispatch(), nil
}

// getSheetInternal obtém uma referência para uma aba específica (chamado internamente na thread COM)
func (c *Client) getSheetInternal(workbookName, sheetName string) (*ole.IDispatch, error) {
	wbDisp, err := c.getWorkbookInternal(workbookName)
	if err != nil {
		return nil, err
	}
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

	return sheet.ToIDispatch(), nil
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

// CreatePivotTable cria uma tabela dinâmica
func (c *Client) CreatePivotTable(workbookName, sourceSheet, sourceRange, destSheet, destCell, tableName string) error {
	return c.runOnCOMThread(func() error {
		wbDisp, err := c.getWorkbookInternal(workbookName)
		if err != nil {
			return err
		}
		defer wbDisp.Release()

		// Source Data Range
		srcSheetDisp, err := c.getSheetInternal(workbookName, sourceSheet)
		if err != nil {
			return err
		}
		defer srcSheetDisp.Release()

		srcRange, err := oleutil.GetProperty(srcSheetDisp, "Range", sourceRange)
		if err != nil {
			return fmt.Errorf("range de origem inválido: %w", err)
		}
		srcRangeDisp := srcRange.ToIDispatch()
		defer srcRangeDisp.Release()

		// Destination Range
		destSheetDisp, err := c.getSheetInternal(workbookName, destSheet)
		if err != nil {
			return err
		}
		defer destSheetDisp.Release()

		destRange, err := oleutil.GetProperty(destSheetDisp, "Range", destCell)
		if err != nil {
			return fmt.Errorf("célula de destino inválida: %w", err)
		}
		destRangeDisp := destRange.ToIDispatch()
		defer destRangeDisp.Release()

		// PivotCaches.Create
		pivotCaches, err := oleutil.GetProperty(wbDisp, "PivotCaches")
		if err != nil {
			return err
		}
		pivotCachesDisp := pivotCaches.ToIDispatch()
		defer pivotCachesDisp.Release()

		// xlDatabase = 1
		pivotCache, err := oleutil.CallMethod(pivotCachesDisp, "Create", 1, srcRangeDisp, 6) // 6 = xlPivotTableVersion14
		if err != nil {
			return fmt.Errorf("falha ao criar PivotCache: %w", err)
		}
		pivotCacheDisp := pivotCache.ToIDispatch()
		defer pivotCacheDisp.Release()

		// CreatePivotTable
		_, err = oleutil.CallMethod(pivotCacheDisp, "CreatePivotTable", destRangeDisp, tableName)
		if err != nil {
			return fmt.Errorf("falha ao criar Tabela Dinâmica: %w", err)
		}

		return nil
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
		defer wbDisp.Release()

		wbName, _ := oleutil.GetProperty(wbDisp, "Name")

		activeSheet, err := oleutil.GetProperty(c.excelApp, "ActiveSheet")
		if err != nil {
			return result{workbook: wbName.ToString()}, err
		}
		sheetDisp := activeSheet.ToIDispatch()
		defer sheetDisp.Release()

		sheetNameProp, _ := oleutil.GetProperty(sheetDisp, "Name")

		return result{
			workbook: wbName.ToString(),
			sheet:    sheetNameProp.ToString(),
		}, nil
	})
	return res.workbook, res.sheet, err
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
