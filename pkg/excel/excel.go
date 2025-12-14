package excel

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

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
		var err error
		// Tentar até 3 vezes com backoff se o Excel estiver ocupado
		for i := 0; i < 3; i++ {
			err = fn()
			if err == nil {
				break
			}
			// Verificar se é erro de "Call was rejected by callee" (Excel ocupado)
			if strings.Contains(err.Error(), "Call was rejected by callee") ||
				strings.Contains(err.Error(), "A chamada foi rejeitada pelo chamado") {
				time.Sleep(time.Millisecond * 500 * time.Duration(i+1))
				continue
			}
			break
		}
		errChan <- err
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
		var v T
		var err error
		// Tentar até 3 vezes com backoff se o Excel estiver ocupado
		for i := 0; i < 3; i++ {
			v, err = fn()
			if err == nil {
				break
			}
			// Verificar se é erro de "Call was rejected by callee" (Excel ocupado)
			if strings.Contains(err.Error(), "Call was rejected by callee") ||
				strings.Contains(err.Error(), "A chamada foi rejeitada pelo chamado") {
				time.Sleep(time.Millisecond * 500 * time.Duration(i+1))
				continue
			}
			break
		}
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

// CreatePivotTable cria uma tabela dinâmica usando PivotTableWizard
func (c *Client) CreatePivotTable(workbookName, sourceSheet, sourceRange, destSheet, destCell, tableName string) error {
	return c.runOnCOMThread(func() error {
		// Obter workbook
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return fmt.Errorf("falha ao acessar Workbooks: %w", err)
		}
		workbooksDisp := workbooks.ToIDispatch()
		defer workbooksDisp.Release()

		wb, err := oleutil.GetProperty(workbooksDisp, "Item", workbookName)
		if err != nil {
			return fmt.Errorf("pasta de trabalho '%s' não encontrada: %w", workbookName, err)
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		// Ativar o workbook
		oleutil.CallMethod(wbDisp, "Activate")

		// Obter Worksheets (não Sheets) - Worksheets é específico para planilhas
		worksheets, err := oleutil.GetProperty(wbDisp, "Worksheets")
		if err != nil {
			return fmt.Errorf("falha ao acessar Worksheets: %w", err)
		}
		worksheetsDisp := worksheets.ToIDispatch()
		defer worksheetsDisp.Release()

		// Obter aba de origem
		srcSheet, err := oleutil.GetProperty(worksheetsDisp, "Item", sourceSheet)
		if err != nil {
			return fmt.Errorf("a aba de origem '%s' não existe: %w", sourceSheet, err)
		}
		srcSheetDisp := srcSheet.ToIDispatch()
		defer srcSheetDisp.Release()

		// Expandir range se necessário
		expandedSourceRange := strings.TrimSpace(sourceRange)

		// Remover referência à aba se a IA a incluiu (ex: 'Custo de 2024'!A1:F200 ou Custo de 2024!A1:F200)
		if idx := strings.Index(expandedSourceRange, "!"); idx != -1 {
			expandedSourceRange = expandedSourceRange[idx+1:]
		}

		// Se apenas colunas (ex: A:F), expande para A1:F<lastRow>
		if strings.Contains(expandedSourceRange, ":") {
			parts := strings.Split(expandedSourceRange, ":")
			if len(parts) == 2 {
				hasDigit0 := strings.IndexFunc(parts[0], func(r rune) bool { return r >= '0' && r <= '9' }) != -1
				hasDigit1 := strings.IndexFunc(parts[1], func(r rune) bool { return r >= '0' && r <= '9' }) != -1
				if !hasDigit0 && !hasDigit1 {
					startCol := strings.ToUpper(strings.TrimSpace(parts[0]))
					endCol := strings.ToUpper(strings.TrimSpace(parts[1]))

					usedRange, uErr := oleutil.GetProperty(srcSheetDisp, "UsedRange")
					if uErr == nil {
						usedRangeDisp := usedRange.ToIDispatch()
						urRows, _ := oleutil.GetProperty(usedRangeDisp, "Rows")
						urRowsDisp := urRows.ToIDispatch()
						countVar, _ := oleutil.GetProperty(urRowsDisp, "Count")
						startRowVar, _ := oleutil.GetProperty(usedRangeDisp, "Row")
						lastRow := int(startRowVar.Val) + int(countVar.Val) - 1
						if lastRow < 2 {
							lastRow = 2
						}
						expandedSourceRange = fmt.Sprintf("%s1:%s%d", startCol, endCol, lastRow)
						urRowsDisp.Release()
						usedRangeDisp.Release()
					}
				}
			}
		}

		fmt.Printf("[DEBUG] Range expandido: %s\n", expandedSourceRange)

		// Obter range de origem
		srcRange, err := oleutil.GetProperty(srcSheetDisp, "Range", expandedSourceRange)
		if err != nil {
			return fmt.Errorf("range de origem inválido '%s': %w", expandedSourceRange, err)
		}
		srcRangeDisp := srcRange.ToIDispatch()
		defer srcRangeDisp.Release()

		// Obter endereço externo (inclui nome do workbook e aba automaticamente)
		// Address(RowAbsolute, ColumnAbsolute, ReferenceStyle, External, RelativeTo)
		srcAddressExternal, err := oleutil.GetProperty(srcRangeDisp, "Address", true, true, 1, true) // 1 = xlA1, true = External
		var fullSourceAddress string
		if err == nil && srcAddressExternal.ToString() != "" {
			fullSourceAddress = srcAddressExternal.ToString()
		} else {
			// Fallback: construir manualmente
			srcAddressProp, _ := oleutil.GetProperty(srcRangeDisp, "Address", true, true)
			fullSourceAddress = fmt.Sprintf("'%s'!%s", sourceSheet, srcAddressProp.ToString())
		}
		fmt.Printf("[DEBUG] Endereço completo da fonte: %s\n", fullSourceAddress)

		// Obter aba de destino
		destSheetObj, err := oleutil.GetProperty(worksheetsDisp, "Item", destSheet)
		if err != nil {
			return fmt.Errorf("a aba de destino '%s' não existe. Crie a aba primeiro usando create-sheet", destSheet)
		}
		destSheetDisp := destSheetObj.ToIDispatch()
		defer destSheetDisp.Release()

		// Ativar aba de destino
		oleutil.CallMethod(destSheetDisp, "Activate")

		// Obter célula de destino
		destRange, err := oleutil.GetProperty(destSheetDisp, "Range", destCell)
		if err != nil {
			return fmt.Errorf("célula de destino inválida '%s': %w", destCell, err)
		}
		destRangeDisp := destRange.ToIDispatch()
		defer destRangeDisp.Release()

		// Construir endereço de destino como string
		destAddressProp, _ := oleutil.GetProperty(destRangeDisp, "Address", true, true)
		fullDestAddress := fmt.Sprintf("'%s'!%s", destSheet, destAddressProp.ToString())

		// Usar PivotTableWizard
		// Parâmetros: SourceType, SourceData, TableDestination, TableName
		// xlDatabase = 1
		fmt.Printf("[DEBUG] Chamando PivotTableWizard com:\n  Source: %s\n  Dest: %s\n  Name: %s\n", fullSourceAddress, fullDestAddress, tableName)

		// Tentar primeiro usando objeto Range para fonte (mais preciso)
		_, err = oleutil.CallMethod(destSheetDisp, "PivotTableWizard",
			1,             // SourceType = xlDatabase
			srcRangeDisp,  // SourceData como Range object
			destRangeDisp, // TableDestination como Range object
			tableName,     // TableName
		)
		if err != nil {
			fmt.Printf("[DEBUG] PivotTableWizard com Range objects falhou: %v\n", err)

			// Tentar usando a aba de ORIGEM para chamar PivotTableWizard
			_, err = oleutil.CallMethod(srcSheetDisp, "PivotTableWizard",
				1,             // SourceType = xlDatabase
				srcRangeDisp,  // SourceData como Range object
				destRangeDisp, // TableDestination como Range object
				tableName,     // TableName
			)
			if err != nil {
				fmt.Printf("[DEBUG] PivotTableWizard via srcSheet falhou: %v\n", err)

				// Tentar com endereço como string para fonte
				_, err = oleutil.CallMethod(destSheetDisp, "PivotTableWizard",
					1,                 // SourceType = xlDatabase
					fullSourceAddress, // SourceData como string
					destRangeDisp,     // TableDestination como Range
					tableName,         // TableName
				)
				if err != nil {
					errStr := err.Error()
					fmt.Printf("[DEBUG] PivotTableWizard com string source falhou: %v\n", err)
					// Verificar se é erro de campos inválidos
					if strings.Contains(errStr, "campo") || strings.Contains(errStr, "field") || strings.Contains(errStr, "colunas rotuladas") {
						return fmt.Errorf("os dados de origem têm colunas sem cabeçalho. Verifique se todas as colunas na primeira linha têm um título")
					}
					return fmt.Errorf("falha ao criar Tabela Dinâmica: %w", err)
				}
			}
		}

		fmt.Println("[DEBUG] Tabela Dinâmica criada com sucesso!")
		return nil
	})
}

// ConfigurePivotFields configura os campos de uma tabela dinâmica
// rowFields: campos para as linhas
// dataFields: campos para os valores (com função de agregação)
func (c *Client) ConfigurePivotFields(workbookName, sheetName, tableName string, rowFields []string, dataFields []map[string]string) error {
	return c.runOnCOMThread(func() error {
		// Obter workbook
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return err
		}
		workbooksDisp := workbooks.ToIDispatch()
		defer workbooksDisp.Release()

		wb, err := oleutil.GetProperty(workbooksDisp, "Item", workbookName)
		if err != nil {
			return err
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		// Obter worksheet
		worksheets, err := oleutil.GetProperty(wbDisp, "Worksheets")
		if err != nil {
			return err
		}
		worksheetsDisp := worksheets.ToIDispatch()
		defer worksheetsDisp.Release()

		sheet, err := oleutil.GetProperty(worksheetsDisp, "Item", sheetName)
		if err != nil {
			return err
		}
		sheetDisp := sheet.ToIDispatch()
		defer sheetDisp.Release()

		// Obter PivotTables
		pivotTables, err := oleutil.GetProperty(sheetDisp, "PivotTables")
		if err != nil {
			return fmt.Errorf("falha ao acessar PivotTables: %w", err)
		}
		pivotTablesDisp := pivotTables.ToIDispatch()
		defer pivotTablesDisp.Release()

		// Obter a tabela dinâmica pelo nome
		pivotTable, err := oleutil.GetProperty(pivotTablesDisp, "Item", tableName)
		if err != nil {
			// Tentar pelo índice 1 (primeira tabela)
			pivotTable, err = oleutil.GetProperty(pivotTablesDisp, "Item", 1)
			if err != nil {
				return fmt.Errorf("tabela dinâmica '%s' não encontrada: %w", tableName, err)
			}
		}
		pivotTableDisp := pivotTable.ToIDispatch()
		defer pivotTableDisp.Release()

		// Configurar campos de linha
		for _, fieldName := range rowFields {
			pivotField, err := oleutil.CallMethod(pivotTableDisp, "PivotFields", fieldName)
			if err != nil {
				fmt.Printf("[DEBUG] Campo '%s' não encontrado: %v\n", fieldName, err)
				continue
			}
			pivotFieldDisp := pivotField.ToIDispatch()

			// xlRowField = 1
			_, err = oleutil.PutProperty(pivotFieldDisp, "Orientation", 1)
			if err != nil {
				fmt.Printf("[DEBUG] Erro ao definir campo linha '%s': %v\n", fieldName, err)
			} else {
				fmt.Printf("[DEBUG] Campo '%s' adicionado às linhas\n", fieldName)
			}
			pivotFieldDisp.Release()
		}

		// Configurar campos de dados (valores)
		for _, dataField := range dataFields {
			fieldName := dataField["field"]
			function := dataField["function"]
			if fieldName == "" {
				continue
			}

			pivotField, err := oleutil.CallMethod(pivotTableDisp, "PivotFields", fieldName)
			if err != nil {
				fmt.Printf("[DEBUG] Campo '%s' não encontrado: %v\n", fieldName, err)
				continue
			}
			pivotFieldDisp := pivotField.ToIDispatch()

			// xlDataField = 4
			_, err = oleutil.PutProperty(pivotFieldDisp, "Orientation", 4)
			if err != nil {
				fmt.Printf("[DEBUG] Erro ao definir campo dados '%s': %v\n", fieldName, err)
				pivotFieldDisp.Release()
				continue
			}

			// Definir função de agregação
			// xlSum = -4157, xlCount = -4112, xlAverage = -4106
			var funcVal int
			switch strings.ToLower(function) {
			case "sum", "soma":
				funcVal = -4157
			case "count", "contar":
				funcVal = -4112
			case "average", "média", "media":
				funcVal = -4106
			case "max", "máximo", "maximo":
				funcVal = -4136
			case "min", "mínimo", "minimo":
				funcVal = -4139
			default:
				funcVal = -4157 // Padrão: soma
			}

			_, err = oleutil.PutProperty(pivotFieldDisp, "Function", funcVal)
			if err != nil {
				fmt.Printf("[DEBUG] Erro ao definir função '%s' para '%s': %v\n", function, fieldName, err)
			} else {
				fmt.Printf("[DEBUG] Campo '%s' adicionado aos valores com função '%s'\n", fieldName, function)
			}
			pivotFieldDisp.Release()
		}

		fmt.Println("[DEBUG] Campos da tabela dinâmica configurados!")
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

// ========== OPERAÇÕES DE CONSULTA (QUERY) ==========

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

		var sheetNames []string
		for i := 1; i <= count; i++ {
			sheet, err := oleutil.GetProperty(sheetsDisp, "Item", i)
			if err != nil {
				continue
			}
			sheetDisp := sheet.ToIDispatch()
			nameVar, _ := oleutil.GetProperty(sheetDisp, "Name")
			sheetNames = append(sheetNames, nameVar.ToString())
			sheetDisp.Release()
		}

		return sheetNames, nil
	})
}

// SheetExists verifica se uma aba existe
func (c *Client) SheetExists(workbookName, sheetName string) (bool, error) {
	sheets, err := c.ListSheets(workbookName)
	if err != nil {
		return false, err
	}
	for _, s := range sheets {
		if strings.EqualFold(s, sheetName) {
			return true, nil
		}
	}
	return false, nil
}

// ListPivotTables retorna lista de tabelas dinâmicas em uma aba
func (c *Client) ListPivotTables(workbookName, sheetName string) ([]string, error) {
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

		pivotTables, err := oleutil.GetProperty(sheetDisp, "PivotTables")
		if err != nil {
			return []string{}, nil // Sem pivot tables
		}
		pivotTablesDisp := pivotTables.ToIDispatch()
		defer pivotTablesDisp.Release()

		countVar, _ := oleutil.GetProperty(pivotTablesDisp, "Count")
		count := int(countVar.Val)

		var names []string
		for i := 1; i <= count; i++ {
			pt, err := oleutil.GetProperty(pivotTablesDisp, "Item", i)
			if err != nil {
				continue
			}
			ptDisp := pt.ToIDispatch()
			nameVar, _ := oleutil.GetProperty(ptDisp, "Name")
			names = append(names, nameVar.ToString())
			ptDisp.Release()
		}

		return names, nil
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

// FormatRange aplica formatação a um range de células
// options: bold, italic, fontSize, fontColor (hex), bgColor (hex)
func (c *Client) FormatRange(workbookName, sheetName, rangeAddr string, bold, italic bool, fontSize int, fontColor, bgColor string) error {
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

		// Font
		font, _ := oleutil.GetProperty(rangeDisp, "Font")
		fontDisp := font.ToIDispatch()
		defer fontDisp.Release()

		if bold {
			oleutil.PutProperty(fontDisp, "Bold", true)
		}
		if italic {
			oleutil.PutProperty(fontDisp, "Italic", true)
		}
		if fontSize > 0 {
			oleutil.PutProperty(fontDisp, "Size", fontSize)
		}
		if fontColor != "" {
			// Converter hex para RGB
			color := hexToRGB(fontColor)
			oleutil.PutProperty(fontDisp, "Color", color)
		}

		// Background color
		if bgColor != "" {
			interior, _ := oleutil.GetProperty(rangeDisp, "Interior")
			interiorDisp := interior.ToIDispatch()
			defer interiorDisp.Release()
			color := hexToRGB(bgColor)
			oleutil.PutProperty(interiorDisp, "Color", color)
		}

		return nil
	})
}

// hexToRGB converte cor hex (#RRGGBB) para valor RGB do Excel (BGR)
func hexToRGB(hex string) int {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	// Excel usa BGR
	return b<<16 | g<<8 | r
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
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, oldName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		_, err = oleutil.PutProperty(sheet, "Name", newName)
		return err
	})
}

// ClearRange limpa o conteúdo de um range (mantém formatação)
func (c *Client) ClearRange(workbookName, sheetName, rangeAddr string) error {
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

		_, err = oleutil.CallMethod(rangeDisp, "ClearContents")
		return err
	})
}

// AutoFitColumns ajusta automaticamente a largura das colunas de um range
func (c *Client) AutoFitColumns(workbookName, sheetName, rangeAddr string) error {
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

		// Obter colunas do range
		cols, _ := oleutil.GetProperty(rangeDisp, "Columns")
		colsDisp := cols.ToIDispatch()
		defer colsDisp.Release()

		_, err = oleutil.CallMethod(colsDisp, "AutoFit")
		return err
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
			rowObj, err := oleutil.GetProperty(sheet, "Rows", rowAddr)
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
		rowObj, err := oleutil.GetProperty(sheet, "Rows", rowAddr)
		if err != nil {
			return err
		}
		rowDisp := rowObj.ToIDispatch()
		defer rowDisp.Release()

		_, err = oleutil.CallMethod(rowDisp, "Delete")
		return err
	})
}

// MergeCells mescla células em um range
func (c *Client) MergeCells(workbookName, sheetName, rangeAddr string) error {
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

		_, err = oleutil.CallMethod(rangeDisp, "Merge")
		return err
	})
}

// UnmergeCells desfaz mesclagem de células
func (c *Client) UnmergeCells(workbookName, sheetName, rangeAddr string) error {
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

		_, err = oleutil.CallMethod(rangeDisp, "UnMerge")
		return err
	})
}

// SetBorders adiciona bordas a um range
// style: "thin", "medium", "thick"
func (c *Client) SetBorders(workbookName, sheetName, rangeAddr, style string) error {
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

		// Determinar peso da borda
		weight := 2 // xlThin
		if style == "medium" {
			weight = -4138 // xlMedium
		} else if style == "thick" {
			weight = 4 // xlThick
		}

		// Aplicar bordas em todos os lados (xlEdgeLeft=7, xlEdgeTop=8, xlEdgeBottom=9, xlEdgeRight=10, xlInsideVertical=11, xlInsideHorizontal=12)
		borders, _ := oleutil.GetProperty(rangeDisp, "Borders")
		bordersDisp := borders.ToIDispatch()
		defer bordersDisp.Release()

		for _, edge := range []int{7, 8, 9, 10, 11, 12} {
			border, err := oleutil.GetProperty(bordersDisp, "Item", edge)
			if err != nil {
				continue
			}
			borderDisp := border.ToIDispatch()
			oleutil.PutProperty(borderDisp, "LineStyle", 1) // xlContinuous
			oleutil.PutProperty(borderDisp, "Weight", weight)
			borderDisp.Release()
		}

		return nil
	})
}

// SetColumnWidth define largura de colunas
func (c *Client) SetColumnWidth(workbookName, sheetName, rangeAddr string, width float64) error {
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

		cols, _ := oleutil.GetProperty(rangeDisp, "Columns")
		colsDisp := cols.ToIDispatch()
		defer colsDisp.Release()

		_, err = oleutil.PutProperty(colsDisp, "ColumnWidth", width)
		return err
	})
}

// SetRowHeight define altura de linhas
func (c *Client) SetRowHeight(workbookName, sheetName, rangeAddr string, height float64) error {
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

		rows, _ := oleutil.GetProperty(rangeDisp, "Rows")
		rowsDisp := rows.ToIDispatch()
		defer rowsDisp.Release()

		_, err = oleutil.PutProperty(rowsDisp, "RowHeight", height)
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

		// Verificar se há AutoFilter ativo
		autoFilter, err := oleutil.GetProperty(sheet, "AutoFilter")
		if err != nil || autoFilter.Val == 0 {
			return nil // Sem filtro ativo
		}
		autoFilterDisp := autoFilter.ToIDispatch()
		if autoFilterDisp != nil {
			defer autoFilterDisp.Release()
			oleutil.CallMethod(autoFilterDisp, "ShowAllData")
		}
		return nil
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

		// Obter célula de ordenação
		cells, _ := oleutil.GetProperty(rangeDisp, "Cells")
		cellsDisp := cells.ToIDispatch()
		defer cellsDisp.Release()

		keyCell, _ := oleutil.GetProperty(cellsDisp, "Item", 1, column)
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
