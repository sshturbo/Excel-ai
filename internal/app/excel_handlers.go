package app

import (
	"fmt"

	"excel-ai/internal/dto"
	"excel-ai/pkg/excel"
)

// ConnectExcel tenta conectar a uma instância do Excel
func (a *App) ConnectExcel() (*dto.ExcelStatus, error) {
	result, err := a.excelService.Connect()
	if err == nil && result.Connected {
		// Iniciar watcher automaticamente ao conectar
		a.StartWorkbookWatcher()
	}
	return result, err
}

// RefreshWorkbooks atualiza a lista de pastas de trabalho
func (a *App) RefreshWorkbooks() (*dto.ExcelStatus, error) {
	return a.excelService.RefreshWorkbooks()
}

// GetPreviewData obtém preview dos dados antes de enviar para IA
func (a *App) GetPreviewData(workbookName, sheetName string) (*dto.PreviewData, error) {
	return a.excelService.GetPreviewData(workbookName, sheetName)
}

// SetExcelContext define o contexto do Excel para uso no chat
func (a *App) SetExcelContext(workbookName, sheetName string) (string, error) {
	// Carregar configurações do storage
	cfg, err := a.storage.LoadConfig()
	if err != nil {
		// Usar valores padrão se não conseguir carregar
		return a.excelService.SetContextWithConfig(workbookName, sheetName, 50, true)
	}

	// Usar configurações do usuário
	maxRows := cfg.MaxRowsContext
	if maxRows <= 0 {
		maxRows = 50 // Padrão
	}

	return a.excelService.SetContextWithConfig(workbookName, sheetName, maxRows, cfg.IncludeHeaders)
}

// GetActiveSelection obtém a seleção atual do Excel
func (a *App) GetActiveSelection() (*excel.SheetData, error) {
	return a.excelService.GetActiveSelection()
}

// UpdateExcelCell atualiza uma célula no Excel
func (a *App) UpdateExcelCell(workbook, sheet, cell, value string) error {
	fmt.Printf("[DEBUG] UpdateExcelCell request: WB='%s' Sheet='%s' Cell='%s' Value='%s'\n", workbook, sheet, cell, value)
	return a.excelService.UpdateCell(workbook, sheet, cell, value)
}

// CreateNewWorkbook cria uma nova pasta de trabalho
func (a *App) CreateNewWorkbook() (string, error) {
	return a.excelService.CreateNewWorkbook()
}

// CreateChart cria um gráfico
func (a *App) CreateChart(sheet, rangeAddr, chartType, title string) error {
	return a.excelService.CreateChart(sheet, rangeAddr, chartType, title)
}

// CreatePivotTable cria uma tabela dinâmica
func (a *App) CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName string) error {
	return a.excelService.CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName)
}

// ConfigurePivotFields configura os campos de uma tabela dinâmica
func (a *App) ConfigurePivotFields(sheetName, tableName string, rowFields []string, dataFields []map[string]string) error {
	return a.excelService.ConfigurePivotFields(sheetName, tableName, rowFields, dataFields)
}

// UndoLastChange desfaz a última alteração
func (a *App) UndoLastChange() error {
	return a.excelService.UndoLastChange()
}

// StartUndoBatch inicia um lote de alterações
func (a *App) StartUndoBatch() {
	a.excelService.StartUndoBatch()
}

// EndUndoBatch finaliza o lote de alterações
func (a *App) EndUndoBatch() {
	a.excelService.EndUndoBatch()
}

// GetLastBatchID retorna o ID do último lote executado
func (a *App) GetLastBatchID() int64 {
	return a.excelService.GetLastBatchID()
}

// ClearLastBatchID limpa o ID do último lote (quando usuário confirma alterações)
func (a *App) ClearLastBatchID() {
	a.excelService.ClearLastBatchID()
}

// UndoByConversation desfaz ações pendentes de uma conversa específica
func (a *App) UndoByConversation(convID string) (int, error) {
	return a.excelService.UndoByConversation(convID)
}

// ApproveUndoActions marca ações de uma conversa como aprovadas (não podem mais ser desfeitas)
func (a *App) ApproveUndoActions(convID string) error {
	return a.excelService.ApproveActions(convID)
}

// HasPendingUndoActionsForConversation verifica se há ações pendentes para uma conversa
func (a *App) HasPendingUndoActionsForConversation(convID string) (bool, error) {
	return a.excelService.HasPendingUndoActions(convID)
}

// SetConversationIDForUndo define a conversa atual para vincular ações de undo
func (a *App) SetConversationIDForUndo(convID string) {
	a.excelService.SetConversationID(convID)
}

// FormatRange aplica formatação a um range
func (a *App) FormatRange(sheet, rangeAddr string, bold, italic bool, fontSize int, fontColor, bgColor string) error {
	return a.excelService.FormatRange(sheet, rangeAddr, bold, italic, fontSize, fontColor, bgColor)
}

// DeleteSheet exclui uma aba
func (a *App) DeleteSheet(sheetName string) error {
	return a.excelService.DeleteSheet(sheetName)
}

// RenameSheet renomeia uma aba
func (a *App) RenameSheet(oldName, newName string) error {
	return a.excelService.RenameSheet(oldName, newName)
}

// ClearRange limpa o conteúdo de um range
func (a *App) ClearRange(sheet, rangeAddr string) error {
	return a.excelService.ClearRange(sheet, rangeAddr)
}

// AutoFitColumns ajusta largura das colunas
func (a *App) AutoFitColumns(sheet, rangeAddr string) error {
	return a.excelService.AutoFitColumns(sheet, rangeAddr)
}

// InsertRows insere linhas
func (a *App) InsertRows(sheet string, rowNumber, count int) error {
	return a.excelService.InsertRows(sheet, rowNumber, count)
}

// DeleteRows exclui linhas
func (a *App) DeleteRows(sheet string, rowNumber, count int) error {
	return a.excelService.DeleteRows(sheet, rowNumber, count)
}

// MergeCells mescla células
func (a *App) MergeCells(sheet, rangeAddr string) error {
	return a.excelService.MergeCells(sheet, rangeAddr)
}

// UnmergeCells desfaz mesclagem
func (a *App) UnmergeCells(sheet, rangeAddr string) error {
	return a.excelService.UnmergeCells(sheet, rangeAddr)
}

// SetBorders adiciona bordas
func (a *App) SetBorders(sheet, rangeAddr, style string) error {
	return a.excelService.SetBorders(sheet, rangeAddr, style)
}

// SetColumnWidth define largura
func (a *App) SetColumnWidth(sheet, rangeAddr string, width float64) error {
	return a.excelService.SetColumnWidth(sheet, rangeAddr, width)
}

// SetRowHeight define altura
func (a *App) SetRowHeight(sheet, rangeAddr string, height float64) error {
	return a.excelService.SetRowHeight(sheet, rangeAddr, height)
}

// ApplyFilter aplica filtro
func (a *App) ApplyFilter(sheet, rangeAddr string) error {
	return a.excelService.ApplyFilter(sheet, rangeAddr)
}

// ClearFilters limpa filtros
func (a *App) ClearFilters(sheet string) error {
	return a.excelService.ClearFilters(sheet)
}

// SortRange ordena dados
func (a *App) SortRange(sheet, rangeAddr string, column int, ascending bool) error {
	return a.excelService.SortRange(sheet, rangeAddr, column, ascending)
}

// CopyRange copia range
func (a *App) CopyRange(sheet, sourceRange, destRange string) error {
	return a.excelService.CopyRange(sheet, sourceRange, destRange)
}

// ListCharts lista gráficos
func (a *App) ListCharts(sheet string) ([]string, error) {
	return a.excelService.ListCharts(sheet)
}

// DeleteChartByName exclui gráfico
func (a *App) DeleteChartByName(sheet, chartName string) error {
	return a.excelService.DeleteChart(sheet, chartName)
}

// CreateTable cria tabela formatada
func (a *App) CreateTable(sheet, rangeAddr, tableName, style string) error {
	return a.excelService.CreateTable(sheet, rangeAddr, tableName, style)
}

// ListTables lista tabelas
func (a *App) ListTables(sheet string) ([]string, error) {
	return a.excelService.ListTables(sheet)
}

// DeleteTable remove tabela
func (a *App) DeleteTable(sheet, tableName string) error {
	return a.excelService.DeleteTable(sheet, tableName)
}

// WriteToExcel escreve valor
func (a *App) WriteToExcel(row, col int, value string) error {
	return a.excelService.WriteToExcel(row, col, value)
}

// ApplyFormula aplica fórmula
func (a *App) ApplyFormula(row, col int, formula string) error {
	return a.excelService.ApplyFormula(row, col, formula)
}

// CreateNewSheet cria nova aba
func (a *App) CreateNewSheet(name string) error {
	return a.excelService.CreateNewSheet(name)
}

// ========== OPERAÇÕES DE CONSULTA (QUERY) ==========

// ListSheets lista todas as abas do workbook atual
func (a *App) ListSheets() ([]string, error) {
	return a.excelService.ListSheets()
}

// SheetExists verifica se uma aba existe
func (a *App) SheetExists(sheetName string) (bool, error) {
	return a.excelService.SheetExists(sheetName)
}

// ListPivotTables lista tabelas dinâmicas em uma aba
func (a *App) ListPivotTables(sheetName string) ([]string, error) {
	return a.excelService.ListPivotTables(sheetName)
}

// GetHeaders retorna os cabeçalhos de um range
func (a *App) GetHeaders(sheetName, rangeAddr string) ([]string, error) {
	return a.excelService.GetHeaders(sheetName, rangeAddr)
}

// GetUsedRange retorna o range utilizado de uma aba
func (a *App) GetUsedRange(sheetName string) (string, error) {
	return a.excelService.GetUsedRange(sheetName)
}

// QueryResult resultado de uma query
type QueryResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// QueryExcel executa uma query genérica no Excel
func (a *App) QueryExcel(queryType string, params map[string]string) QueryResult {
	switch queryType {
	case "list-sheets":
		sheets, err := a.ListSheets()
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: sheets}

	case "sheet-exists":
		exists, err := a.SheetExists(params["name"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: exists}

	case "list-pivot-tables":
		pivots, err := a.ListPivotTables(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: pivots}

	case "get-headers":
		headers, err := a.GetHeaders(params["sheet"], params["range"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: headers}

	case "get-used-range":
		usedRange, err := a.GetUsedRange(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: usedRange}

	case "get-row-count":
		count, err := a.excelService.GetRowCount(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: count}

	case "get-column-count":
		count, err := a.excelService.GetColumnCount(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: count}

	case "get-cell-formula":
		formula, err := a.excelService.GetCellFormula(params["sheet"], params["cell"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: formula}

	case "has-filter":
		hasFilter, err := a.excelService.HasFilter(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: hasFilter}

	case "get-active-cell":
		cell, err := a.excelService.GetActiveCell()
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: cell}

	case "get-range-values":
		values, err := a.excelService.GetRangeValues(params["sheet"], params["range"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: values}

	case "list-charts":
		charts, err := a.excelService.ListCharts(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: charts}

	case "list-tables":
		tables, err := a.excelService.ListTables(params["sheet"])
		if err != nil {
			return QueryResult{Success: false, Error: err.Error()}
		}
		return QueryResult{Success: true, Data: tables}

	default:
		return QueryResult{Success: false, Error: fmt.Sprintf("query type '%s' não reconhecido", queryType)}
	}
}
