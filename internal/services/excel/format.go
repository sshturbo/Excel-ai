package excel

import (
	"excel-ai/pkg/excel"
	"fmt"
)

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

func (s *Service) GetFormat(sheet, rangeAddr string) (*excel.Format, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return nil, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return nil, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.GetFormat(s.currentWorkbook, sheet, rangeAddr)
}

func (s *Service) GetColumnWidth(sheet, rangeAddr string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return 0, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return 0, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.GetColumnWidth(s.currentWorkbook, sheet, rangeAddr)
}

func (s *Service) GetRowHeight(sheet, rangeAddr string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return 0, fmt.Errorf("excel não conectado")
	}
	if s.currentWorkbook == "" {
		return 0, fmt.Errorf("nenhuma pasta de trabalho selecionada")
	}
	return s.client.GetRowHeight(s.currentWorkbook, sheet, rangeAddr)
}
