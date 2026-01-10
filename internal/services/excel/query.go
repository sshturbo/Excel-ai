package excel

import "fmt"

// ListSheets retorna lista de planilhas do arquivo atual
func (s *Service) ListSheets() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return nil, err
	}
	return client.ListSheets(), nil
}

// SheetExists verifica se uma planilha existe
func (s *Service) SheetExists(sheetName string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return false, err
	}
	return client.SheetExists(sheetName)
}

// ListPivotTables lista tabelas dinâmicas (limitação do Excelize)
func (s *Service) ListPivotTables(sheetName string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return nil, err
	}
	return client.ListPivotTables(sheetName)
}

// GetHeaders retorna os cabeçalhos de um range
func (s *Service) GetHeaders(sheetName, rangeAddr string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return nil, err
	}

	if sheetName == "" {
		sheetName = s.getFirstSheet()
	}

	return client.GetHeaders(sheetName, rangeAddr)
}

// GetUsedRange retorna o range utilizado na planilha
func (s *Service) GetUsedRange(sheetName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return "", err
	}

	if sheetName == "" {
		sheetName = s.getFirstSheet()
	}

	return client.GetUsedRange(sheetName)
}

// GetRowCount retorna número de linhas usadas
func (s *Service) GetRowCount(sheetName string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return 0, err
	}

	if sheetName == "" {
		sheetName = s.getFirstSheet()
	}

	return client.GetRowCount(sheetName)
}

// GetColumnCount retorna número de colunas usadas
func (s *Service) GetColumnCount(sheetName string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return 0, err
	}

	if sheetName == "" {
		sheetName = s.getFirstSheet()
	}

	return client.GetColumnCount(sheetName)
}

// GetCellFormula retorna a fórmula de uma célula
func (s *Service) GetCellFormula(sheetName, cellAddress string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return "", err
	}

	if sheetName == "" {
		sheetName = s.getFirstSheet()
	}

	return client.GetCellFormula(sheetName, cellAddress)
}

// HasFilter verifica se a planilha tem filtro aplicado
func (s *Service) HasFilter(sheetName string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return false, err
	}

	if sheetName == "" {
		sheetName = s.getFirstSheet()
	}

	return client.HasFilter(sheetName)
}

// GetActiveCell retorna célula ativa (sempre A1 no modo Excelize)
func (s *Service) GetActiveCell() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentSessionID == "" {
		return "", fmt.Errorf("nenhum arquivo carregado")
	}

	// Excelize não tem conceito de célula ativa - retornar A1
	return "A1", nil
}

// GetRangeValues retorna valores de um range
func (s *Service) GetRangeValues(sheetName, rangeAddr string) ([][]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return nil, err
	}

	if sheetName == "" {
		sheetName = s.getFirstSheet()
	}

	return client.GetRangeValues(sheetName, rangeAddr)
}
