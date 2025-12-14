package excel

import (
	"excel-ai/internal/dto"
	"excel-ai/pkg/excel"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Service struct {
	client          *excel.Client
	mu              sync.Mutex
	currentWorkbook string
	currentSheet    string
	previewData     *excel.SheetData
	undoStack       []dto.UndoAction
	currentBatchID  int64
	contextStr      string
}

func NewService() *Service {
	return &Service{
		undoStack: []dto.UndoAction{},
	}
}

// getFirstSheet retorna a primeira aba quando currentSheet contém múltiplas abas
func (s *Service) getFirstSheet() string {
	if strings.Contains(s.currentSheet, ",") {
		return strings.TrimSpace(strings.Split(s.currentSheet, ",")[0])
	}
	return s.currentSheet
}

func (s *Service) Connect() (*dto.ExcelStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		s.client.Close()
	}

	client, err := excel.NewClient()
	if err != nil {
		return &dto.ExcelStatus{Connected: false, Error: err.Error()}, nil
	}

	s.client = client
	workbooks, err := s.client.GetOpenWorkbooks()
	if err != nil {
		// Se falhar ao listar, consideramos que a conexão não foi totalmente bem sucedida
		// para permitir que o usuário tente novamente
		s.client.Close()
		s.client = nil
		return &dto.ExcelStatus{Connected: false, Error: "Conectado ao Excel, mas falha ao listar planilhas: " + err.Error()}, nil
	}

	return &dto.ExcelStatus{Connected: true, Workbooks: workbooks}, nil
}

func (s *Service) RefreshWorkbooks() (*dto.ExcelStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return &dto.ExcelStatus{Connected: false, Error: "Não conectado"}, nil
	}

	workbooks, err := s.client.GetOpenWorkbooks()
	if err != nil {
		return &dto.ExcelStatus{Connected: true, Error: err.Error()}, nil
	}

	return &dto.ExcelStatus{Connected: true, Workbooks: workbooks}, nil
}

func (s *Service) SetContext(workbook, sheet string) (string, error) {
	// Usar valores padrão
	return s.SetContextWithConfig(workbook, sheet, 50, true)
}

// SetContextWithConfig define o contexto com configurações personalizadas
func (s *Service) SetContextWithConfig(workbook, sheet string, maxRowsContext int, includeHeaders bool) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return "", fmt.Errorf("não conectado ao Excel")
	}

	// Suporte a múltiplas abas separadas por vírgula
	sheets := strings.Split(sheet, ",")
	var contextStr string
	totalRows := 0

	// Usar configuração do usuário, com limite por aba quando há múltiplas
	maxRowsPerSheet := maxRowsContext
	if len(sheets) > 1 && maxRowsContext > 20 {
		maxRowsPerSheet = maxRowsContext / len(sheets) // Dividir entre abas
		if maxRowsPerSheet < 10 {
			maxRowsPerSheet = 10 // Mínimo de 10 linhas por aba
		}
	}

	for i, sheetName := range sheets {
		sheetName = strings.TrimSpace(sheetName)
		if sheetName == "" {
			continue
		}

		data, err := s.client.GetSheetData(workbook, sheetName, maxRowsPerSheet)
		if err != nil {
			return "", fmt.Errorf("erro ao ler aba %s: %w", sheetName, err)
		}

		if i == 0 {
			s.currentWorkbook = workbook
			s.currentSheet = sheetName
		}

		// Construir representação em string para a IA
		if i > 0 {
			contextStr += "\n---\n\n"
		}
		contextStr += fmt.Sprintf("=== ABA: %s ===\n\n", sheetName)

		// Cabeçalhos (opcional)
		if includeHeaders && len(data.Headers) > 0 {
			for j, h := range data.Headers {
				if j > 0 {
					contextStr += " | "
				}
				// Truncar cabeçalhos muito longos
				if len(h) > 50 {
					h = h[:47] + "..."
				}
				contextStr += h
			}
			contextStr += "\n"
		}

		// Linhas
		for _, row := range data.Rows {
			for j, cell := range row {
				if j > 0 {
					contextStr += " | "
				}
				// Truncar células muito longas
				cellText := cell.Text
				if len(cellText) > 100 {
					cellText = cellText[:97] + "..."
				}
				contextStr += cellText
			}
			contextStr += "\n"
		}

		totalRows += len(data.Rows)
	}

	// Limitar tamanho total do contexto (aproximadamente 4000 tokens = ~16000 caracteres)
	maxContextLen := 12000
	if len(contextStr) > maxContextLen {
		contextStr = contextStr[:maxContextLen] + "\n\n[... contexto truncado para economizar tokens ...]"
	}

	s.contextStr = fmt.Sprintf("Planilha: %s\nAbas selecionadas: %s\n\nDados:\n%s", workbook, sheet, contextStr)

	if len(sheets) > 1 {
		return fmt.Sprintf("Contexto carregado: %s - %d abas (%d linhas total)", workbook, len(sheets), totalRows), nil
	}
	return fmt.Sprintf("Contexto carregado: %s - %s (%d linhas)", workbook, sheets[0], totalRows), nil
}

func (s *Service) GetContextString() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.contextStr
}

func (s *Service) GetPreviewData(workbookName, sheetName string) (*dto.PreviewData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil, fmt.Errorf("não conectado ao Excel")
	}

	data, err := s.client.GetSheetData(workbookName, sheetName, 100)
	if err != nil {
		return nil, err
	}

	s.previewData = data
	s.currentWorkbook = workbookName
	s.currentSheet = sheetName

	// Converter para formato simples
	var rows [][]string
	for _, row := range data.Rows {
		var rowStrings []string
		for _, cell := range row {
			rowStrings = append(rowStrings, cell.Text)
		}
		rows = append(rows, rowStrings)
	}

	return &dto.PreviewData{
		Headers:   data.Headers,
		Rows:      rows,
		TotalRows: len(data.Rows), // Aproximado
		TotalCols: len(data.Headers),
		Workbook:  workbookName,
		Sheet:     sheetName,
	}, nil
}

func (s *Service) GetActiveSelection() (*excel.SheetData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil, fmt.Errorf("não conectado ao Excel")
	}

	return s.client.GetActiveSelection()
}

func (s *Service) UpdateCell(workbook, sheet, cell, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}

	if workbook == "" {
		workbook = s.currentWorkbook
	}
	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	// Se ainda não tiver workbook, tentar obter o ativo
	if workbook == "" {
		activeWb, activeSheet, err := s.client.GetActiveWorkbookAndSheet()
		if err == nil {
			workbook = activeWb
			if sheet == "" {
				sheet = activeSheet
			}
		}
	}

	if workbook == "" || sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	// Salvar valor antigo para desfazer
	oldValue, err := s.client.GetCellValue(workbook, sheet, cell)
	if err == nil {
		s.undoStack = append(s.undoStack, dto.UndoAction{
			Workbook: workbook,
			Sheet:    sheet,
			Cell:     cell,
			OldValue: oldValue,
			BatchID:  s.currentBatchID,
		})
	}

	return s.client.SetCellValue(workbook, sheet, cell, value)
}

func (s *Service) StartUndoBatch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentBatchID = time.Now().UnixNano()
}

func (s *Service) EndUndoBatch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentBatchID = 0
}

func (s *Service) UndoLastChange() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.undoStack) == 0 {
		return fmt.Errorf("nada para desfazer")
	}

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}

	undoAction := func(action dto.UndoAction) error {
		return s.client.SetCellValue(action.Workbook, action.Sheet, action.Cell, action.OldValue)
	}

	lastIdx := len(s.undoStack) - 1
	lastAction := s.undoStack[lastIdx]
	s.undoStack = s.undoStack[:lastIdx]

	if err := undoAction(lastAction); err != nil {
		return err
	}

	if lastAction.BatchID != 0 {
		for len(s.undoStack) > 0 {
			idx := len(s.undoStack) - 1
			prevAction := s.undoStack[idx]

			if prevAction.BatchID == lastAction.BatchID {
				s.undoStack = s.undoStack[:idx]
				if err := undoAction(prevAction); err != nil {
					return fmt.Errorf("erro ao desfazer lote: %w", err)
				}
			} else {
				break
			}
		}
	}

	return nil
}

func (s *Service) CreateNewWorkbook() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return "", fmt.Errorf("excel não conectado")
	}
	return s.client.CreateNewWorkbook()
}

func (s *Service) CreateNewSheet(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.InsertNewSheet(s.currentWorkbook, name)
}

func (s *Service) CreateChart(sheet, rangeAddr, chartType, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	if sheet == "" {
		sheet = s.getFirstSheet()
	}
	return s.client.CreateChart(s.currentWorkbook, sheet, rangeAddr, chartType, title)
}

func (s *Service) CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Printf("[DEBUG] CreatePivotTable chamado:\n")
	fmt.Printf("  sourceSheet: %s\n", sourceSheet)
	fmt.Printf("  sourceRange: %s\n", sourceRange)
	fmt.Printf("  destSheet: %s\n", destSheet)
	fmt.Printf("  destCell: %s\n", destCell)
	fmt.Printf("  tableName: %s\n", tableName)
	fmt.Printf("  currentWorkbook: %s\n", s.currentWorkbook)

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}

	// Se não tiver workbook selecionado, tentar obter o ativo
	workbook := s.currentWorkbook
	if workbook == "" {
		activeWb, _, err := s.client.GetActiveWorkbookAndSheet()
		if err != nil {
			return fmt.Errorf("nenhuma pasta de trabalho selecionada e não foi possível obter a ativa: %w", err)
		}
		workbook = activeWb
		fmt.Printf("  workbook (ativo): %s\n", workbook)
	}

	if sourceSheet == "" {
		sourceSheet = s.getFirstSheet()
		fmt.Printf("  sourceSheet (fallback): %s\n", sourceSheet)
	}

	err := s.client.CreatePivotTable(workbook, sourceSheet, sourceRange, destSheet, destCell, tableName)
	if err != nil {
		fmt.Printf("[DEBUG] Erro ao criar PivotTable: %v\n", err)
	}
	return err
}

func (s *Service) ConfigurePivotFields(sheetName, tableName string, rowFields []string, dataFields []map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}

	workbook := s.currentWorkbook
	if workbook == "" {
		activeWb, _, err := s.client.GetActiveWorkbookAndSheet()
		if err != nil {
			return fmt.Errorf("nenhuma pasta de trabalho selecionada: %w", err)
		}
		workbook = activeWb
	}

	return s.client.ConfigurePivotFields(workbook, sheetName, tableName, rowFields, dataFields)
}

func (s *Service) WriteToExcel(row, col int, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("não conectado ao Excel")
	}

	sheet := s.getFirstSheet()
	if s.currentWorkbook == "" || sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}

	return s.client.WriteCell(s.currentWorkbook, sheet, row, col, value)
}

func (s *Service) ApplyFormula(row, col int, formula string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	sheet := s.getFirstSheet()
	if s.currentWorkbook == "" || sheet == "" {
		return fmt.Errorf("nenhuma planilha selecionada")
	}
	return s.client.ApplyFormula(s.currentWorkbook, sheet, row, col, formula)
}

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		s.client.Close()
	}
}

// ========== OPERAÇÕES DE CONSULTA (QUERY) ==========

// ListSheets lista todas as abas do workbook atual
func (s *Service) ListSheets() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return nil, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.ListSheets(s.currentWorkbook)
}

// SheetExists verifica se uma aba existe
func (s *Service) SheetExists(sheetName string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return false, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return false, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.SheetExists(s.currentWorkbook, sheetName)
}

// ListPivotTables lista tabelas dinâmicas em uma aba
func (s *Service) ListPivotTables(sheetName string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return nil, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.ListPivotTables(s.currentWorkbook, sheetName)
}

// GetHeaders retorna os cabeçalhos de um range
func (s *Service) GetHeaders(sheetName, rangeAddr string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return nil, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.GetHeaders(s.currentWorkbook, sheetName, rangeAddr)
}

// GetUsedRange retorna o range utilizado de uma aba
func (s *Service) GetUsedRange(sheetName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return "", fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return "", fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.GetUsedRange(s.currentWorkbook, sheetName)
}

// GetRowCount retorna número de linhas
func (s *Service) GetRowCount(sheetName string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return 0, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return 0, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.GetRowCount(s.currentWorkbook, sheetName)
}

// GetColumnCount retorna número de colunas
func (s *Service) GetColumnCount(sheetName string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return 0, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return 0, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.GetColumnCount(s.currentWorkbook, sheetName)
}

// GetCellFormula retorna fórmula de célula
func (s *Service) GetCellFormula(sheetName, cellAddress string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return "", fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return "", fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.GetCellFormula(s.currentWorkbook, sheetName, cellAddress)
}

// HasFilter verifica se tem filtro ativo
func (s *Service) HasFilter(sheetName string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return false, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return false, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.HasFilter(s.currentWorkbook, sheetName)
}

// GetActiveCell retorna célula ativa
func (s *Service) GetActiveCell() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return "", fmt.Errorf("excel não conectado")
	}
	return s.client.GetActiveCell()
}

// GetRangeValues retorna valores de um range
func (s *Service) GetRangeValues(sheetName, rangeAddr string) ([][]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return nil, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return nil, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.GetRangeValues(s.currentWorkbook, sheetName, rangeAddr)
}

// FormatRange aplica formatação a um range
func (s *Service) FormatRange(sheet, rangeAddr string, bold, italic bool, fontSize int, fontColor, bgColor string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.FormatRange(s.currentWorkbook, sheet, rangeAddr, bold, italic, fontSize, fontColor, bgColor)
}

// DeleteSheet exclui uma aba
func (s *Service) DeleteSheet(sheetName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.DeleteSheet(s.currentWorkbook, sheetName)
}

// RenameSheet renomeia uma aba
func (s *Service) RenameSheet(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.RenameSheet(s.currentWorkbook, oldName, newName)
}

// ClearRange limpa o conteúdo de um range
func (s *Service) ClearRange(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.ClearRange(s.currentWorkbook, sheet, rangeAddr)
}

// AutoFitColumns ajusta largura das colunas
func (s *Service) AutoFitColumns(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.AutoFitColumns(s.currentWorkbook, sheet, rangeAddr)
}

// InsertRows insere linhas
func (s *Service) InsertRows(sheet string, rowNumber, count int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.InsertRows(s.currentWorkbook, sheet, rowNumber, count)
}

// DeleteRows exclui linhas
func (s *Service) DeleteRows(sheet string, rowNumber, count int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.DeleteRows(s.currentWorkbook, sheet, rowNumber, count)
}

// MergeCells mescla células
func (s *Service) MergeCells(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.MergeCells(s.currentWorkbook, sheet, rangeAddr)
}

// UnmergeCells desfaz mesclagem
func (s *Service) UnmergeCells(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.UnmergeCells(s.currentWorkbook, sheet, rangeAddr)
}

// SetBorders adiciona bordas
func (s *Service) SetBorders(sheet, rangeAddr, style string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.SetBorders(s.currentWorkbook, sheet, rangeAddr, style)
}

// SetColumnWidth define largura de colunas
func (s *Service) SetColumnWidth(sheet, rangeAddr string, width float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.SetColumnWidth(s.currentWorkbook, sheet, rangeAddr, width)
}

// SetRowHeight define altura de linhas
func (s *Service) SetRowHeight(sheet, rangeAddr string, height float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.SetRowHeight(s.currentWorkbook, sheet, rangeAddr, height)
}

// ApplyFilter aplica filtro automático
func (s *Service) ApplyFilter(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.ApplyFilter(s.currentWorkbook, sheet, rangeAddr)
}

// ClearFilters limpa filtros
func (s *Service) ClearFilters(sheet string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.ClearFilters(s.currentWorkbook, sheet)
}

// SortRange ordena dados
func (s *Service) SortRange(sheet, rangeAddr string, column int, ascending bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.SortRange(s.currentWorkbook, sheet, rangeAddr, column, ascending)
}

// CopyRange copia range
func (s *Service) CopyRange(sheet, sourceRange, destRange string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.CopyRange(s.currentWorkbook, sheet, sourceRange, destRange)
}

// ListCharts lista gráficos
func (s *Service) ListCharts(sheet string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return nil, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return nil, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.ListCharts(s.currentWorkbook, sheet)
}

// DeleteChart exclui gráfico
func (s *Service) DeleteChart(sheet, chartName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.DeleteChart(s.currentWorkbook, sheet, chartName)
}

// CreateTable cria tabela formatada
func (s *Service) CreateTable(sheet, rangeAddr, tableName, style string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.CreateTable(s.currentWorkbook, sheet, rangeAddr, tableName, style)
}

// ListTables lista tabelas
func (s *Service) ListTables(sheet string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return nil, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return nil, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.ListTables(s.currentWorkbook, sheet)
}

// DeleteTable remove tabela
func (s *Service) DeleteTable(sheet, tableName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.DeleteTable(s.currentWorkbook, sheet, tableName)
}
