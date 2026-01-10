package excel

import (
	"fmt"
	"strings"

	"excel-ai/internal/dto"
	"excel-ai/pkg/excel"
)

func (s *Service) SetContext(workbook, sheet string) (string, error) {
	// Usar valores padrão mais conservadores para evitar estouro de TPM (Limit 8000)
	return s.SetContextWithConfig(workbook, sheet, 20, true)
}

// SetContextWithConfig define o contexto com configurações personalizadas
func (s *Service) SetContextWithConfig(workbook, sheet string, maxRowsContext int, includeHeaders bool) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return "", err
	}

	// workbook é ignorado no modo Excelize (sempre é o arquivo atual)
	if workbook == "" {
		workbook = s.currentFileName
	}

	// Suporte a múltiplas abas separadas por vírgula
	sheets := strings.Split(sheet, ",")
	var contextStr string
	totalRows := 0

	// Usar configuração do usuário, com limite por aba quando há múltiplas
	maxRowsPerSheet := maxRowsContext
	if len(sheets) > 1 && maxRowsContext > 10 {
		maxRowsPerSheet = maxRowsContext / len(sheets)
		if maxRowsPerSheet < 5 {
			maxRowsPerSheet = 5
		}
	}

	for i, sheetName := range sheets {
		sheetName = strings.TrimSpace(sheetName)
		if sheetName == "" {
			continue
		}

		// Obter range usado
		usedRange, err := client.GetUsedRange(sheetName)
		if err != nil {
			return "", fmt.Errorf("erro ao ler aba %s: %w", sheetName, err)
		}

		// Obter dados
		data, err := client.GetRangeValues(sheetName, usedRange)
		if err != nil {
			return "", fmt.Errorf("erro ao ler dados da aba %s: %w", sheetName, err)
		}

		if i == 0 {
			s.currentSheet = sheetName
		}

		// Construir representação em string para a IA
		if i > 0 {
			contextStr += "\n---\n\n"
		}
		contextStr += fmt.Sprintf("=== ABA: %s ===\n\n", sheetName)

		// Limitar linhas
		rowsToShow := len(data)
		if rowsToShow > maxRowsPerSheet {
			rowsToShow = maxRowsPerSheet
		}

		// Cabeçalhos (primeira linha)
		if includeHeaders && len(data) > 0 {
			for j, h := range data[0] {
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

		// Linhas (a partir da segunda se includeHeaders, senão todas)
		startRow := 0
		if includeHeaders && len(data) > 0 {
			startRow = 1
		}

		for r := startRow; r < rowsToShow && r < len(data); r++ {
			for j, cell := range data[r] {
				if j > 0 {
					contextStr += " | "
				}
				// Truncar células muito longas
				if len(cell) > 100 {
					cell = cell[:97] + "..."
				}
				contextStr += cell
			}
			contextStr += "\n"
		}

		totalRows += rowsToShow
	}

	// Limitar tamanho total do contexto
	maxContextLen := 6000
	if s.storage != nil {
		if cfg, err := s.storage.LoadConfig(); err == nil {
			if cfg.MaxContextChars > 0 {
				maxContextLen = cfg.MaxContextChars
			}
		}
	}

	if len(contextStr) > maxContextLen {
		contextStr = contextStr[:maxContextLen] + fmt.Sprintf("\n\n[... contexto truncado em %d chars para economizar tokens ...]", maxContextLen)
	}

	s.contextStr = fmt.Sprintf("Arquivo: %s\nAbas selecionadas: %s\n\nDados:\n%s", workbook, sheet, contextStr)

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

// GetActiveContext retorna apenas o workbook e sheet ativos (sem dados)
func (s *Service) GetActiveContext() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentFileName == "" {
		return ""
	}

	if s.currentSheet == "" {
		return fmt.Sprintf("Arquivo ativo: %s", s.currentFileName)
	}

	return fmt.Sprintf("Arquivo: %s | Aba selecionada: %s", s.currentFileName, s.currentSheet)
}

func (s *Service) GetPreviewData(workbookName, sheetName string) (*dto.PreviewData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return nil, err
	}

	if sheetName == "" {
		sheets := client.ListSheets()
		if len(sheets) > 0 {
			sheetName = sheets[0]
		}
	}

	// Obter range usado
	usedRange, err := client.GetUsedRange(sheetName)
	if err != nil {
		return nil, err
	}

	// Obter dados
	data, err := client.GetRangeValues(sheetName, usedRange)
	if err != nil {
		return nil, err
	}

	// Limitar a 100 linhas
	maxRows := 100
	if len(data) > maxRows {
		data = data[:maxRows]
	}

	s.currentSheet = sheetName

	// Separar headers e rows
	var headers []string
	var rows [][]string

	if len(data) > 0 {
		headers = data[0]
		if len(data) > 1 {
			rows = data[1:]
		}
	}

	return &dto.PreviewData{
		Headers:   headers,
		Rows:      rows,
		TotalRows: len(data),
		TotalCols: len(headers),
		Workbook:  s.currentFileName,
		Sheet:     sheetName,
	}, nil
}

func (s *Service) GetActiveSelection() (*excel.SheetData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return nil, err
	}

	sheet := s.currentSheet
	if sheet == "" {
		sheets := client.ListSheets()
		if len(sheets) > 0 {
			sheet = sheets[0]
		}
	}

	// Obter range usado
	usedRange, err := client.GetUsedRange(sheet)
	if err != nil {
		return nil, err
	}

	// Obter dados
	data, err := client.GetRangeValues(sheet, usedRange)
	if err != nil {
		return nil, err
	}

	// Converter para SheetData
	var headers []string
	var rows [][]excel.CellData

	if len(data) > 0 {
		headers = data[0]
		for i := 1; i < len(data) && i <= 100; i++ {
			row := make([]excel.CellData, len(data[i]))
			for j, cell := range data[i] {
				row[j] = excel.CellData{Text: cell}
			}
			rows = append(rows, row)
		}
	}

	return &excel.SheetData{
		Headers: headers,
		Rows:    rows,
	}, nil
}
