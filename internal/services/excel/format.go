package excel

import (
	"excel-ai/pkg/excel"
	"fmt"
)

// FormatRange formata um range de células
func (s *Service) FormatRange(sheet, rangeAddr string, bold, italic bool, fontSize int, fontColor, bgColor string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	format := excel.Format{
		Bold:      bold,
		Italic:    italic,
		FontSize:  fontSize,
		FontColor: fontColor,
		BgColor:   bgColor,
	}

	return client.FormatRange(sheet, rangeAddr, format)
}

// SetBorders define bordas em um range
func (s *Service) SetBorders(sheet, rangeAddr, style string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.SetBorders(sheet, rangeAddr, style)
}

// SetColumnWidth define largura de coluna
func (s *Service) SetColumnWidth(sheet, col string, width float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.SetColumnWidth(sheet, col, width)
}

// SetRowHeight define altura de linha
func (s *Service) SetRowHeight(sheet, row string, height float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.SetRowHeight(sheet, row, height)
}

// ApplyFilter aplica filtro em um range
func (s *Service) ApplyFilter(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.ApplyFilter(sheet, rangeAddr)
}

// ClearFilters limpa filtros de uma planilha
func (s *Service) ClearFilters(sheet string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.ClearFilters(sheet)
}

// SortRange ordena um range
func (s *Service) SortRange(sheet, rangeAddr string, column int, ascending bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.SortRange(sheet, rangeAddr, column, ascending)
}

// CopyRange copia um range para outro destino
func (s *Service) CopyRange(sheet, sourceRange, destRange string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	// Ler dados do range de origem
	data, err := client.GetRangeValues(sheet, sourceRange)
	if err != nil {
		return fmt.Errorf("erro ao ler range de origem: %w", err)
	}

	// Converter para [][]interface{}
	interfaceData := make([][]interface{}, len(data))
	for i, row := range data {
		interfaceData[i] = make([]interface{}, len(row))
		for j, val := range row {
			interfaceData[i][j] = val
		}
	}

	// Escrever no destino (primeira célula do range de destino)
	destCell := destRange
	if idx := len(destRange); idx > 0 {
		// Se for range "A1:B2", pegar só primeira célula
		for i, c := range destRange {
			if c == ':' {
				destCell = destRange[:i]
				break
			}
		}
	}

	return client.WriteRange(sheet, destCell, interfaceData)
}

// GetFormat retorna formato de um range (não suportado totalmente no Excelize)
func (s *Service) GetFormat(sheet, rangeAddr string) (*excel.Format, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentSessionID == "" {
		return nil, fmt.Errorf("nenhum arquivo carregado")
	}

	// Excelize não tem método direto para ler formato
	// Retornar formato vazio
	return &excel.Format{}, nil
}

// GetColumnWidth retorna largura de coluna (não suportado totalmente)
func (s *Service) GetColumnWidth(sheet, col string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentSessionID == "" {
		return 0, fmt.Errorf("nenhum arquivo carregado")
	}

	// Retornar valor padrão - Excelize não expõe isso facilmente
	return 8.43, nil // Largura padrão do Excel
}

// GetRowHeight retorna altura de linha (não suportado totalmente)
func (s *Service) GetRowHeight(sheet, row string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentSessionID == "" {
		return 0, fmt.Errorf("nenhum arquivo carregado")
	}

	// Retornar valor padrão
	return 15, nil // Altura padrão do Excel
}
