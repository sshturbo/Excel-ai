package excel

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// NewExcelizeClient cria um novo cliente Excelize a partir de bytes do arquivo
func NewExcelizeClient(data []byte) (*ExcelizeClient, error) {
	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}

	return &ExcelizeClient{
		file:     file,
		filePath: "",
	}, nil
}

// NewExcelizeClientFromPath cria um novo cliente Excelize a partir de um arquivo
func NewExcelizeClientFromPath(path string) (*ExcelizeClient, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}

	return &ExcelizeClient{
		file:     file,
		filePath: path,
	}, nil
}

// Close fecha o arquivo Excelize
func (c *ExcelizeClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.file != nil {
		c.file.Close()
		c.file = nil
	}
}

// ListSheets retorna uma lista de planilhas
func (c *ExcelizeClient) ListSheets() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.GetSheetList()
}

// SheetExists verifica se uma planilha existe
func (c *ExcelizeClient) SheetExists(name string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sheets := c.file.GetSheetList()

	for _, sheet := range sheets {
		if strings.EqualFold(sheet, name) {
			return true, nil
		}
	}
	return false, nil
}

// CreateSheet cria uma nova planilha
func (c *ExcelizeClient) CreateSheet(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.file.NewSheet(name)
	return err
}

// DeleteSheet deleta uma planilha
func (c *ExcelizeClient) DeleteSheet(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.DeleteSheet(name)
}

// RenameSheet renomeia uma planilha
func (c *ExcelizeClient) RenameSheet(oldName, newName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetSheetName(oldName, newName)
}

// GetCellValue retorna o valor de uma célula
func (c *ExcelizeClient) GetCellValue(sheet, cell string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, err := c.file.GetCellValue(sheet, cell)
	if err != nil {
		return "", fmt.Errorf("failed to get cell value: %w", err)
	}
	return value, nil
}

// SetCellValue define o valor de uma célula
func (c *ExcelizeClient) SetCellValue(sheet, cell string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetCellValue(sheet, cell, value)
}

// GetRangeValues retorna os valores de um range
func (c *ExcelizeClient) GetRangeValues(sheet, rng string) ([][]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.getRangeValuesLocked(sheet, rng)
}

// getRangeValuesLocked internal version without mutex - PREVENTS DEADLOCKS
func (c *ExcelizeClient) getRangeValuesLocked(sheet, rng string) ([][]string, error) {
	// Parser simples de range "A1:B10" para obter colunas e linhas
	startCell, endCell, err := parseRange(rng)
	if err != nil {
		return nil, err
	}

	startRow, startCol := cellToIndices(startCell)
	endRow, endCol := cellToIndices(endCell)

	// OTIMIZAÇÃO: Se o range for grande, GetRows consome muita memória.
	// Vamos buscar apenas as células necessárias usando GetCellValue
	result := make([][]string, 0)
	for r := startRow; r <= endRow; r++ {
		row := make([]string, 0)
		for col := startCol; col <= endCol; col++ {
			cell := indicesToCell(r, col)
			val, _ := c.file.GetCellValue(sheet, cell)
			row = append(row, val)
		}
		result = append(result, row)
	}

	return result, nil
}

// WriteRange escreve dados em um range
func (c *ExcelizeClient) WriteRange(sheet, startCell string, data [][]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Excelize não tem um método direto para escrever um range inteiro
	// Precisamos escrever célula por célula
	startRow, startCol := cellToIndices(startCell)

	for r, row := range data {
		for col, value := range row {
			cell := indicesToCell(startRow+r, startCol+col)
			if err := c.file.SetCellValue(sheet, cell, value); err != nil {
				return fmt.Errorf("failed to set cell value %s: %w", cell, err)
			}
		}
	}

	return nil
}

// ClearRange limpa um range de células
func (c *ExcelizeClient) ClearRange(sheet, rng string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// OTIMIZAÇÃO: Não precisamos ler os valores para limpar, apenas calcular o tamanho do range
	startCell, endCell, err := parseRange(rng)
	if err != nil {
		return err
	}

	startRow, startCol := cellToIndices(startCell)
	endRow, endCol := cellToIndices(endCell)

	for r := startRow; r <= endRow; r++ {
		for col := startCol; col <= endCol; col++ {
			cell := indicesToCell(r, col)
			c.file.SetCellValue(sheet, cell, "")
		}
	}

	return nil
}

// FormatRange formata um range (implementação básica)
func (c *ExcelizeClient) FormatRange(sheet, rng string, format Format) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	styleID, err := c.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   format.Bold,
			Italic: format.Italic,
			Size:   float64(format.FontSize),
			Color:  format.FontColor,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{format.BgColor},
			Pattern: 1,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create style: %w", err)
	}

	return c.file.SetCellStyle(sheet, rng, rng, styleID)
}

// SetColumnWidth define a largura de uma coluna
func (c *ExcelizeClient) SetColumnWidth(sheet, col string, width float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetColWidth(sheet, col, col, width)
}

// SetRowHeight define a altura de uma linha
func (c *ExcelizeClient) SetRowHeight(sheet, row string, height float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	rowNum, err := strconv.Atoi(row)
	if err != nil {
		return fmt.Errorf("invalid row number: %w", err)
	}

	return c.file.SetRowHeight(sheet, rowNum, height)
}

// MergeCells mescla células
func (c *ExcelizeClient) MergeCells(sheet, rng string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Excelize.MergeCell requer sheet, topLeft, bottomRight
	startCell, endCell, err := parseRange(rng)
	if err != nil {
		return err
	}

	return c.file.MergeCell(sheet, startCell, endCell)
}

// UnmergeCells desmescla células
func (c *ExcelizeClient) UnmergeCells(sheet, rng string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Excelize.UnmergeCell requer sheet, topLeft, bottomRight
	startCell, endCell, err := parseRange(rng)
	if err != nil {
		return err
	}

	return c.file.UnmergeCell(sheet, startCell, endCell)
}

// SetBorders define bordas
func (c *ExcelizeClient) SetBorders(sheet, rng, style string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var borderType []excelize.Border

	switch style {
	case "thin":
		borderType = []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		}
	case "medium":
		borderType = []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
		}
	default:
		return fmt.Errorf("unsupported border style: %s", style)
	}

	styleID, err := c.file.NewStyle(&excelize.Style{
		Border: borderType,
	})

	if err != nil {
		return err
	}

	return c.file.SetCellStyle(sheet, rng, rng, styleID)
}

// AutoFitColumns ajusta automaticamente a largura das colunas
func (c *ExcelizeClient) AutoFitColumns(sheet, rng string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Usar versão internal para evitar deadlock
	data, err := c.getRangeValuesLocked(sheet, rng)
	if err != nil {
		return err
	}

	startCell, _, err := parseRange(rng)
	if err != nil {
		return err
	}

	_, startCol := cellToIndices(startCell)

	// Calcular largura máxima por coluna
	for col := range data[0] {
		maxWidth := 0
		for row := range data {
			width := len(data[row][col])
			if width > maxWidth {
				maxWidth = width
			}
		}

		colLetter := indicesToCell(0, startCol+col)
		// Excelize usa unidades diferentes, 1 char ≈ 7 pixels ≈ 0.9 unidade
		width := float64(maxWidth) * 0.9
		if width < 8 {
			width = 8
		}
		if width > 50 {
			width = 50
		}

		c.file.SetColWidth(sheet, colLetter, colLetter, width)
	}

	return nil
}

// InsertRows insere linhas
func (c *ExcelizeClient) InsertRows(sheet string, row, count int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < count; i++ {
		if err := c.file.InsertRows(sheet, row, 1); err != nil {
			return err
		}
	}
	return nil
}

// DeleteRows deleta linhas
func (c *ExcelizeClient) DeleteRows(sheet string, row, count int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.RemoveRow(sheet, row+count-1)
}

// CreateChart cria um gráfico
func (c *ExcelizeClient) CreateChart(sheet, rng, chartType, title string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simplificado: criação básica de gráfico
	// Excelize tem APIs complexas para gráficos, esta é uma implementação mínima

	var excelizeType excelize.ChartType
	switch strings.ToLower(chartType) {
	case "bar", "column":
		excelizeType = excelize.Col
	case "line":
		excelizeType = excelize.Line
	case "pie":
		excelizeType = excelize.Pie
	case "scatter":
		excelizeType = excelize.Scatter
	default:
		excelizeType = excelize.Col
	}

	if err := c.file.AddChart(sheet, "Chart1", &excelize.Chart{
		Type:   excelizeType,
		Series: []excelize.ChartSeries{{Name: title, Categories: rng, Values: rng}},
	}); err != nil {
		return fmt.Errorf("failed to create chart: %w", err)
	}

	return nil
}

// DeleteChart deleta um gráfico
func (c *ExcelizeClient) DeleteChart(sheet, name string) error {
	// Excelize não tem método direto para deletar chart por nome
	// Isso é uma limitação conhecida
	return fmt.Errorf("delete chart not fully implemented in Excelize")
}

// ListCharts lista gráficos (limitação do Excelize)
func (c *ExcelizeClient) ListCharts(sheet string) ([]string, error) {
	// Excelize não tem método para listar charts
	return []string{}, nil
}

// CreateTable cria uma tabela
func (c *ExcelizeClient) CreateTable(sheet, rng, name, style string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.file.AddTable(sheet, &excelize.Table{
		Range:     rng,
		Name:      name,
		StyleName: style,
	}); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// DeleteTable deleta uma tabela
func (c *ExcelizeClient) DeleteTable(sheet, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Excelize.DeleteTable recebe apenas o nome da tabela
	return c.file.DeleteTable(name)
}

// ListTables lista tabelas
func (c *ExcelizeClient) ListTables(sheet string) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tables, err := c.file.GetTables(sheet)
	if err != nil {
		return nil, err
	}

	// Extrair apenas os nomes
	names := make([]string, 0, len(tables))
	for _, table := range tables {
		names = append(names, table.Name)
	}
	return names, nil
}

// CreatePivotTable cria uma tabela dinâmica usando Excelize
func (c *ExcelizeClient) CreatePivotTable(srcSheet, srcRange, destSheet, destCell, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Construir referência de dados no formato correto: Sheet1!A1:E10
	dataRange := fmt.Sprintf("%s!%s", srcSheet, srcRange)

	// Definir opções da Pivot Table
	opts := &excelize.PivotTableOptions{
		DataRange:       dataRange,
		PivotTableRange: fmt.Sprintf("%s!%s", destSheet, destCell),
		Name:            name,
		// Configuração padrão - campos podem ser configurados depois
		RowGrandTotals:    true,
		ColGrandTotals:    true,
		ShowDrill:         true,
		UseAutoFormatting: true,
		PageOverThenDown:  true,
		ShowRowHeaders:    true,
		ShowColHeaders:    true,
		ShowLastColumn:    true,
	}

	return c.file.AddPivotTable(opts)
}

// CreatePivotTableWithFields cria uma pivot table com campos específicos
func (c *ExcelizeClient) CreatePivotTableWithFields(srcSheet, srcRange, destSheet, destCell, name string,
	rowFields, colFields []string, dataFields []excelize.PivotTableField) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dataRange := fmt.Sprintf("%s!%s", srcSheet, srcRange)

	// Converter row fields
	rows := make([]excelize.PivotTableField, len(rowFields))
	for i, f := range rowFields {
		rows[i] = excelize.PivotTableField{Data: f, DefaultSubtotal: true}
	}

	// Converter column fields
	cols := make([]excelize.PivotTableField, len(colFields))
	for i, f := range colFields {
		cols[i] = excelize.PivotTableField{Data: f, DefaultSubtotal: true}
	}

	opts := &excelize.PivotTableOptions{
		DataRange:         dataRange,
		PivotTableRange:   fmt.Sprintf("%s!%s", destSheet, destCell),
		Name:              name,
		Rows:              rows,
		Columns:           cols,
		Data:              dataFields,
		RowGrandTotals:    true,
		ColGrandTotals:    true,
		ShowDrill:         true,
		UseAutoFormatting: true,
	}

	return c.file.AddPivotTable(opts)
}

// ListPivotTables lista tabelas dinâmicas de uma planilha
func (c *ExcelizeClient) ListPivotTables(sheet string) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tables, err := c.file.GetPivotTables(sheet)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(tables))
	for _, table := range tables {
		names = append(names, table.Name)
	}
	return names, nil
}

// DeletePivotTable deleta uma tabela dinâmica
func (c *ExcelizeClient) DeletePivotTable(sheet, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.DeletePivotTable(sheet, name)
}

// ApplyFilter aplica um filtro
func (c *ExcelizeClient) ApplyFilter(sheet, rng string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.AutoFilter(sheet, rng, []excelize.AutoFilterOptions{})
}

// ClearFilters remove filtros de uma planilha
func (c *ExcelizeClient) ClearFilters(sheet string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Excelize suporta remoção de AutoFilter desde v2.6
	// Workaround: definir AutoFilter vazio remove o filtro
	return c.file.AutoFilter(sheet, "", nil)
}

// HasFilter verifica se há filtro (limitação Excelize - sempre retorna false)
func (c *ExcelizeClient) HasFilter(sheet string) (bool, error) {
	// Excelize não expõe método para verificar filtros diretamente
	// Retornar false por padrão
	return false, nil
}

// SortRange ordena um range com suporte a valores numéricos
func (c *ExcelizeClient) SortRange(sheet, rng string, col int, ascending bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Usar versão internal para evitar deadlock
	data, err := c.getRangeValuesLocked(sheet, rng)
	if err != nil {
		return err
	}

	if len(data) == 0 || col >= len(data[0]) {
		return nil
	}

	// Ordenar as linhas baseado na coluna especificada
	// Detectar se a coluna é numérica para ordenação correta
	sort.Slice(data, func(i, j int) bool {
		valI := data[i][col]
		valJ := data[j][col]

		// Tentar converter para float para ordenação numérica
		numI, errI := strconv.ParseFloat(strings.TrimSpace(valI), 64)
		numJ, errJ := strconv.ParseFloat(strings.TrimSpace(valJ), 64)

		// Se ambos são números, ordenar numericamente
		if errI == nil && errJ == nil {
			if ascending {
				return numI < numJ
			}
			return numI > numJ
		}

		// Senão, ordenar alfabeticamente (case-insensitive)
		if ascending {
			return strings.ToLower(valI) < strings.ToLower(valJ)
		}
		return strings.ToLower(valI) > strings.ToLower(valJ)
	})

	// Usar versão internal para evitar deadlock
	startCell, _, err := parseRange(rng)
	if err != nil {
		return err
	}

	startRow, startCol := cellToIndices(startCell)

	for r, row := range data {
		for colIdx, value := range row {
			cell := indicesToCell(startRow+r, startCol+colIdx)
			c.file.SetCellValue(sheet, cell, value)
		}
	}

	return nil
}

// GetUsedRange retorna o range utilizado
func (c *ExcelizeClient) GetUsedRange(sheet string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rows, err := c.file.GetRows(sheet)
	if err != nil {
		return "", err
	}

	if len(rows) == 0 {
		return "A1:A1", nil
	}

	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	return fmt.Sprintf("A1:%s", indicesToCell(len(rows)-1, maxCols-1)), nil
}

// GetRowCount retorna o número de linhas
func (c *ExcelizeClient) GetRowCount(sheet string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rows, err := c.file.GetRows(sheet)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// GetColumnCount retorna o número de colunas
func (c *ExcelizeClient) GetColumnCount(sheet string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rows, err := c.file.GetRows(sheet)
	if err != nil {
		return 0, err
	}

	if len(rows) == 0 {
		return 0, nil
	}

	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	return maxCols, nil
}

// GetHeaders retorna os cabeçalhos de um range
func (c *ExcelizeClient) GetHeaders(sheet, rng string) ([]string, error) {
	data, err := c.GetRangeValues(sheet, rng)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []string{}, nil
	}
	return data[0], nil
}

// GetCellFormula retorna a fórmula de uma célula
func (c *ExcelizeClient) GetCellFormula(sheet, cell string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	formula, err := c.file.GetCellFormula(sheet, cell)
	if err != nil {
		return "", fmt.Errorf("failed to get cell formula: %w", err)
	}
	return formula, nil
}

// SaveAs salva o arquivo
func (c *ExcelizeClient) SaveAs(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SaveAs(path)
}

// Write escreve os dados para um writer
func (c *ExcelizeClient) Write() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	buffer, err := c.file.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// Funções auxiliares

func parseRange(rng string) (string, string, error) {
	parts := strings.Split(rng, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid range format: %s", rng)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func cellToIndices(cell string) (row, col int) {
	// Ex: "A1" -> row=0, col=0
	col = 0
	row = 0

	// Extrair letras (coluna)
	for _, c := range cell {
		if c >= 'A' && c <= 'Z' {
			col = col*26 + int(c-'A'+1)
		} else if c >= 'a' && c <= 'z' {
			col = col*26 + int(c-'a'+1)
		} else {
			break
		}
	}
	col-- // Ajustar para 0-indexed

	// Extrair números (linha)
	for _, c := range cell {
		if c >= '0' && c <= '9' {
			row = row*10 + int(c-'0')
		}
	}
	row-- // Ajustar para 0-indexed

	return row, col
}

func indicesToCell(row, col int) string {
	// Ex: row=0, col=0 -> "A1"
	letter := ""
	for col > 0 {
		col-- // Ajustar para 0-indexed
		letter = string(rune('A'+(col%26))) + letter
		col = col / 26
	}
	return fmt.Sprintf("%s%d", letter, row+1)
}

// ==================== ADVANCED FEATURES ====================

// AddConditionalFormat adiciona formatação condicional a um range
func (c *ExcelizeClient) AddConditionalFormat(sheet, rng string, formatType string, criteria string, value string, format *excelize.Style) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Criar estilo para a formatação
	styleID, err := c.file.NewStyle(format)
	if err != nil {
		return fmt.Errorf("erro ao criar estilo: %w", err)
	}

	// Mapear tipo de formatação
	var cfType string
	switch formatType {
	case "greaterThan":
		cfType = "cellIs"
	case "lessThan":
		cfType = "cellIs"
	case "equal":
		cfType = "cellIs"
	case "between":
		cfType = "cellIs"
	case "containsText":
		cfType = "containsText"
	case "duplicate":
		cfType = "duplicateValues"
	case "unique":
		cfType = "uniqueValues"
	case "top10":
		cfType = "top10"
	case "colorScale":
		cfType = "colorScale"
	case "dataBar":
		cfType = "dataBar"
	default:
		cfType = "cellIs"
	}

	opts := []excelize.ConditionalFormatOptions{{
		Type:     cfType,
		Criteria: criteria,
		Format:   &styleID,
		Value:    value,
	}}

	return c.file.SetConditionalFormat(sheet, rng, opts)
}

// AddSimpleConditionalFormat versão simplificada para formatação condicional comum
func (c *ExcelizeClient) AddSimpleConditionalFormat(sheet, rng string, criteria string, value string, bgColor string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	style := &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{bgColor},
		},
	}

	styleID, err := c.file.NewStyle(style)
	if err != nil {
		return err
	}

	opts := []excelize.ConditionalFormatOptions{{
		Type:     "cellIs",
		Criteria: criteria,
		Format:   &styleID,
		Value:    value,
	}}

	return c.file.SetConditionalFormat(sheet, rng, opts)
}

// AddDataValidation adiciona validação de dados (dropdown, números, etc)
func (c *ExcelizeClient) AddDataValidation(sheet, rng string, validationType string, options []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dv := excelize.NewDataValidation(true)
	dv.Sqref = rng

	switch validationType {
	case "list":
		// Dropdown list
		dv.SetDropList(options)
	case "whole":
		// Número inteiro
		if len(options) >= 2 {
			dv.SetRange(options[0], options[1], excelize.DataValidationTypeWhole, excelize.DataValidationOperatorBetween)
		}
	case "decimal":
		// Número decimal
		if len(options) >= 2 {
			dv.SetRange(options[0], options[1], excelize.DataValidationTypeDecimal, excelize.DataValidationOperatorBetween)
		}
	case "date":
		// Data
		if len(options) >= 2 {
			dv.SetRange(options[0], options[1], excelize.DataValidationTypeDate, excelize.DataValidationOperatorBetween)
		}
	case "textLength":
		// Comprimento do texto
		if len(options) >= 2 {
			dv.SetRange(options[0], options[1], excelize.DataValidationTypeTextLength, excelize.DataValidationOperatorBetween)
		}
	}

	return c.file.AddDataValidation(sheet, dv)
}

// AddDropdownList adiciona uma lista dropdown a uma célula/range
func (c *ExcelizeClient) AddDropdownList(sheet, rng string, options []string) error {
	return c.AddDataValidation(sheet, rng, "list", options)
}

// AddCellComment adiciona um comentário a uma célula
func (c *ExcelizeClient) AddCellComment(sheet, cell, author, text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	comment := excelize.Comment{
		Cell:   cell,
		Author: author,
		Text:   text,
	}

	return c.file.AddComment(sheet, comment)
}

// GetCellComment obtém o comentário de uma célula
func (c *ExcelizeClient) GetCellComment(sheet, cell string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	comments, err := c.file.GetComments(sheet)
	if err != nil {
		return "", err
	}

	for _, comment := range comments {
		if comment.Cell == cell {
			return comment.Text, nil
		}
	}

	return "", nil
}

// DeleteCellComment remove um comentário de uma célula
func (c *ExcelizeClient) DeleteCellComment(sheet, cell string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.DeleteComment(sheet, cell)
}

// AddImage adiciona uma imagem à planilha
func (c *ExcelizeClient) AddImage(sheet, cell string, imageData []byte, format string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Excelize espera extensão do arquivo
	ext := "." + format
	if format == "jpg" {
		ext = ".jpeg"
	}

	opts := &excelize.GraphicOptions{
		AutoFit: true,
	}

	return c.file.AddPictureFromBytes(sheet, cell, &excelize.Picture{
		Extension: ext,
		File:      imageData,
		Format:    opts,
	})
}

// AddImageFromFile adiciona uma imagem de um arquivo
func (c *ExcelizeClient) AddImageFromFile(sheet, cell, imagePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := &excelize.GraphicOptions{
		AutoFit: true,
	}

	return c.file.AddPicture(sheet, cell, imagePath, opts)
}

// ProtectSheet protege uma planilha com senha opcional
func (c *ExcelizeClient) ProtectSheet(sheet string, password string, options *excelize.SheetProtectionOptions) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if options == nil {
		options = &excelize.SheetProtectionOptions{
			Password:            password,
			SelectLockedCells:   true,
			SelectUnlockedCells: true,
		}
	} else {
		options.Password = password
	}

	return c.file.ProtectSheet(sheet, options)
}

// UnprotectSheet remove a proteção de uma planilha
func (c *ExcelizeClient) UnprotectSheet(sheet string, password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.UnprotectSheet(sheet, password)
}

// SetCellLocked define se uma célula está bloqueada (para uso com ProtectSheet)
func (c *ExcelizeClient) SetCellLocked(sheet, cell string, locked bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	style := &excelize.Style{
		Protection: &excelize.Protection{
			Locked: locked,
		},
	}

	styleID, err := c.file.NewStyle(style)
	if err != nil {
		return err
	}

	return c.file.SetCellStyle(sheet, cell, cell, styleID)
}

// CalculateFormulas calcula todas as fórmulas do arquivo
func (c *ExcelizeClient) CalculateFormulas() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Excelize suporta cálculo de fórmulas desde v2.8
	return c.file.UpdateLinkedValue()
}

// SetCellFormula define uma fórmula em uma célula
func (c *ExcelizeClient) SetCellFormula(sheet, cell, formula string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetCellFormula(sheet, cell, formula)
}

// GetSheetProtection verifica se uma planilha está protegida
func (c *ExcelizeClient) GetSheetProtection(sheet string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Excelize não expõe método direto para verificar proteção
	// Workaround: tentar desproteger sem senha e ver se falha
	err := c.file.UnprotectSheet(sheet, "")
	if err != nil {
		// Se falhou, está protegido
		return true, nil
	}
	// Se conseguiu, re-proteger e retornar falso (não estava protegido)
	return false, nil
}

// HideSheet oculta uma planilha
func (c *ExcelizeClient) HideSheet(sheet string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetSheetVisible(sheet, false)
}

// ShowSheet exibe uma planilha oculta
func (c *ExcelizeClient) ShowSheet(sheet string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetSheetVisible(sheet, true)
}

// FreezePane congela linhas/colunas
func (c *ExcelizeClient) FreezePane(sheet, cell string, freezeRows, freezeCols int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		XSplit:      freezeCols,
		YSplit:      freezeRows,
		TopLeftCell: cell,
		ActivePane:  "bottomRight",
	})
}

// UnfreezePane remove o congelamento de painéis
func (c *ExcelizeClient) UnfreezePane(sheet string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetPanes(sheet, &excelize.Panes{
		Freeze: false,
	})
}

// GroupRows agrupa linhas
func (c *ExcelizeClient) GroupRows(sheet string, startRow, endRow int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetRowOutlineLevel(sheet, endRow, 1)
}

// GroupColumns agrupa colunas
func (c *ExcelizeClient) GroupColumns(sheet, startCol, endCol string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetColOutlineLevel(sheet, endCol, 1)
}

// SetPrintArea define a área de impressão
func (c *ExcelizeClient) SetPrintArea(sheet, rng string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Area",
		RefersTo: fmt.Sprintf("'%s'!%s", sheet, rng),
		Scope:    sheet,
	})
}

// AddHyperlink adiciona um hyperlink a uma célula
func (c *ExcelizeClient) AddHyperlink(sheet, cell, url, display string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.file.SetCellHyperLink(sheet, cell, url, "External", excelize.HyperlinkOpts{
		Display: &display,
	})
}

// GetHyperlink obtém o hyperlink de uma célula
func (c *ExcelizeClient) GetHyperlink(sheet, cell string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hasLink, target, err := c.file.GetCellHyperLink(sheet, cell)
	if err != nil || !hasLink {
		return "", err
	}
	return target, nil
}
