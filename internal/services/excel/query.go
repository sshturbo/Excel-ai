package excel

import "fmt"

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

func (s *Service) GetActiveCell() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return "", fmt.Errorf("excel não conectado")
	}
	return s.client.GetActiveCell()
}

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
