package app

import (
	"fmt"

	"excel-ai/internal/dto"
	apperrors "excel-ai/pkg/errors"
	"excel-ai/pkg/excel"
	"excel-ai/pkg/logger"
)

// ConnectExcel tenta conectar a uma instância do Excel
func (a *App) ConnectExcel() (*dto.ExcelStatus, error) {
	logger.ExcelInfo("Tentando conectar ao Excel...")
	result, err := a.excelService.Connect()
	if err != nil {
		logger.ExcelError("Falha ao conectar ao Excel: " + err.Error())
		return nil, apperrors.ExcelNotConnected("não foi possível conectar ao Excel: " + err.Error())
	}

	if result.Connected {
		logger.ExcelInfo("Conectado ao Excel com sucesso")
		// Iniciar watcher automaticamente ao conectar
		a.StartWorkbookWatcher()
	}
	return result, nil
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
	logger.ExcelDebug(fmt.Sprintf("UpdateExcelCell: WB='%s' Sheet='%s' Cell='%s' Value='%s'", workbook, sheet, cell, value))

	err := a.excelService.UpdateCell(workbook, sheet, cell, value)
	if err != nil {
		logger.ExcelError("Erro ao atualizar célula: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao atualizar célula")
	}

	return nil
}

// CreateNewWorkbook cria uma nova pasta de trabalho
func (a *App) CreateNewWorkbook() (string, error) {
	return a.excelService.CreateNewWorkbook()
}

// CreateChart cria um gráfico
func (a *App) CreateChart(sheet, rangeAddr, chartType, title string) error {
	logger.ExcelInfo(fmt.Sprintf("Criando gráfico: tipo=%s, range=%s", chartType, rangeAddr))

	err := a.excelService.CreateChart(sheet, rangeAddr, chartType, title)
	if err != nil {
		logger.ExcelError("Erro ao criar gráfico: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao criar gráfico")
	}

	return nil
}

// CreatePivotTable cria uma tabela dinâmica
func (a *App) CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName string) error {
	logger.ExcelInfo(fmt.Sprintf("Criando tabela dinâmica: %s", tableName))

	err := a.excelService.CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName)
	if err != nil {
		logger.ExcelError("Erro ao criar tabela dinâmica: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao criar tabela dinâmica")
	}

	return nil
}

// ConfigurePivotFields configura os campos de uma tabela dinâmica
func (a *App) ConfigurePivotFields(sheetName, tableName string, rowFields []string, dataFields []map[string]string) error {
	return a.excelService.ConfigurePivotFields(sheetName, tableName, rowFields, dataFields)
}

// UndoLastChange desfaz a última alteração
func (a *App) UndoLastChange() error {
	logger.AppInfo("Desfazendo última alteração")

	err := a.excelService.UndoLastChange()
	if err != nil {
		logger.AppError("Erro ao desfazer alteração: " + err.Error())
		return err
	}

	logger.AppInfo("Alteração desfeita com sucesso")
	return nil
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
	logger.ExcelInfo(fmt.Sprintf("Aplicando formatação: range=%s", rangeAddr))

	err := a.excelService.FormatRange(sheet, rangeAddr, bold, italic, fontSize, fontColor, bgColor)
	if err != nil {
		logger.ExcelError("Erro ao aplicar formatação: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao aplicar formatação")
	}

	return nil
}

// DeleteSheet exclui uma aba
func (a *App) DeleteSheet(sheetName string) error {
	logger.ExcelInfo("Excluindo planilha: " + sheetName)

	err := a.excelService.DeleteSheet(sheetName)
	if err != nil {
		logger.ExcelError("Erro ao excluir planilha: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeInvalidSheet, "falha ao excluir planilha")
	}

	return nil
}

// RenameSheet renomeia uma aba
func (a *App) RenameSheet(oldName, newName string) error {
	logger.ExcelInfo(fmt.Sprintf("Renomeando planilha: %s -> %s", oldName, newName))

	err := a.excelService.RenameSheet(oldName, newName)
	if err != nil {
		logger.ExcelError("Erro ao renomear planilha: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeInvalidSheet, "falha ao renomear planilha")
	}

	return nil
}

// ClearRange limpa o conteúdo de um range
func (a *App) ClearRange(sheet, rangeAddr string) error {
	logger.ExcelInfo(fmt.Sprintf("Limpando range: %s", rangeAddr))

	err := a.excelService.ClearRange(sheet, rangeAddr)
	if err != nil {
		logger.ExcelError("Erro ao limpar range: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao limpar range")
	}

	return nil
}

// AutoFitColumns ajusta largura das colunas
func (a *App) AutoFitColumns(sheet, rangeAddr string) error {
	logger.ExcelInfo(fmt.Sprintf("Auto-fit colunas: %s", rangeAddr))

	err := a.excelService.AutoFitColumns(sheet, rangeAddr)
	if err != nil {
		logger.ExcelError("Erro ao auto-fit colunas: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao auto-fit colunas")
	}

	return nil
}

// InsertRows insere linhas
func (a *App) InsertRows(sheet string, rowNumber, count int) error {
	logger.ExcelInfo(fmt.Sprintf("Inserindo %d linhas na linha %d", count, rowNumber))

	err := a.excelService.InsertRows(sheet, rowNumber, count)
	if err != nil {
		logger.ExcelError("Erro ao inserir linhas: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao inserir linhas")
	}

	return nil
}

// DeleteRows exclui linhas
func (a *App) DeleteRows(sheet string, rowNumber, count int) error {
	logger.ExcelInfo(fmt.Sprintf("Excluindo %d linhas a partir da linha %d", count, rowNumber))

	err := a.excelService.DeleteRows(sheet, rowNumber, count)
	if err != nil {
		logger.ExcelError("Erro ao excluir linhas: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao excluir linhas")
	}

	return nil
}

// MergeCells mescla células
func (a *App) MergeCells(sheet, rangeAddr string) error {
	logger.ExcelInfo(fmt.Sprintf("Mesclando células: %s", rangeAddr))

	err := a.excelService.MergeCells(sheet, rangeAddr)
	if err != nil {
		logger.ExcelError("Erro ao mesclar células: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao mesclar células")
	}

	return nil
}

// UnmergeCells desfaz mesclagem
func (a *App) UnmergeCells(sheet, rangeAddr string) error {
	logger.ExcelInfo(fmt.Sprintf("Desfazendo mesclagem: %s", rangeAddr))
	
	err := a.excelService.UnmergeCells(sheet, rangeAddr)
	if err != nil {
		logger.ExcelError("Erro ao desfazer mesclagem: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao desfazer mesclagem")
	}
	
	return nil
}

// SetBorders adiciona bordas
func (a *App) SetBorders(sheet, rangeAddr, style string) error {
	logger.ExcelInfo(fmt.Sprintf("Adicionando bordas: %s, estilo=%s", rangeAddr, style))

	err := a.excelService.SetBorders(sheet, rangeAddr, style)
	if err != nil {
		logger.ExcelError("Erro ao adicionar bordas: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao adicionar bordas")
	}

	return nil
}

// SetColumnWidth define largura
func (a *App) SetColumnWidth(sheet, rangeAddr string, width float64) error {
	logger.ExcelInfo(fmt.Sprintf("Definindo largura de colunas: %s, width=%.2f", rangeAddr, width))
	
	err := a.excelService.SetColumnWidth(sheet, rangeAddr, width)
	if err != nil {
		logger.ExcelError("Erro ao definir largura de colunas: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao definir largura de colunas")
	}
	
	return nil
}

// SetRowHeight define altura
func (a *App) SetRowHeight(sheet, rangeAddr string, height float64) error {
	logger.ExcelInfo(fmt.Sprintf("Definindo altura de linhas: %s, height=%.2f", rangeAddr, height))
	
	err := a.excelService.SetRowHeight(sheet, rangeAddr, height)
	if err != nil {
		logger.ExcelError("Erro ao definir altura de linhas: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao definir altura de linhas")
	}
	
	return nil
}

// ApplyFilter aplica filtro
func (a *App) ApplyFilter(sheet, rangeAddr string) error {
	logger.ExcelInfo(fmt.Sprintf("Aplicando filtro: %s", rangeAddr))
	
	err := a.excelService.ApplyFilter(sheet, rangeAddr)
	if err != nil {
		logger.ExcelError("Erro ao aplicar filtro: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao aplicar filtro")
	}
	
	return nil
}

// ClearFilters limpa filtros
func (a *App) ClearFilters(sheet string) error {
	logger.ExcelInfo("Limpando filtros da planilha: " + sheet)
	
	err := a.excelService.ClearFilters(sheet)
	if err != nil {
		logger.ExcelError("Erro ao limpar filtros: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao limpar filtros")
	}
	
	return nil
}

// SortRange ordena dados
func (a *App) SortRange(sheet, rangeAddr string, column int, ascending bool) error {
	order := "crescente"
	if !ascending {
		order = "decrescente"
	}
	logger.ExcelInfo(fmt.Sprintf("Ordenando range %s: col=%d, ordem=%s", rangeAddr, column, order))
	
	err := a.excelService.SortRange(sheet, rangeAddr, column, ascending)
	if err != nil {
		logger.ExcelError("Erro ao ordenar dados: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao ordenar dados")
	}
	
	return nil
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
	logger.ExcelDebug(fmt.Sprintf("Escrevendo na célula: row=%d, col=%d, value=%s", row, col, value))

	err := a.excelService.WriteToExcel(row, col, value)
	if err != nil {
		logger.ExcelError("Erro ao escrever valor: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao escrever valor")
	}

	return nil
}

// ApplyFormula aplica fórmula
func (a *App) ApplyFormula(row, col int, formula string) error {
	return a.excelService.ApplyFormula(row, col, formula)
}

// CreateNewSheet cria nova aba
func (a *App) CreateNewSheet(name string) error {
	logger.ExcelInfo("Criando nova planilha: " + name)

	err := a.excelService.CreateNewSheet(name)
	if err != nil {
		logger.ExcelError("Erro ao criar planilha: " + err.Error())
		return apperrors.Wrap(err, apperrors.ErrCodeExcelNotFound, "falha ao criar planilha")
	}

	return nil
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
