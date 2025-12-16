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

	if s.client == nil {
		return "", fmt.Errorf("não conectado ao Excel")
	}

	// Suporte a múltiplas abas separadas por vírgula
	sheets := strings.Split(sheet, ",")
	var contextStr string
	totalRows := 0

	// Usar configuração do usuário, com limite por aba quando há múltiplas
	maxRowsPerSheet := maxRowsContext
	if len(sheets) > 1 && maxRowsContext > 10 { // Adjusted logic
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

	// Limitar tamanho total do contexto de forma agressiva para Groq 8k TPM
	// Valor configurável pelo usuário
	maxContextLen := 6000 // Default safe
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
