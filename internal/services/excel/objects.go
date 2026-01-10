package excel

import "fmt"

// CreateChart cria um gráfico
func (s *Service) CreateChart(sheet, rangeAddr, chartType, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.CreateChart(sheet, rangeAddr, chartType, title)
}

// CreatePivotTable cria uma tabela dinâmica
func (s *Service) CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sourceSheet == "" {
		sourceSheet = s.getFirstSheet()
	}

	return client.CreatePivotTable(sourceSheet, sourceRange, destSheet, destCell, tableName)
}

// ConfigurePivotFields configura campos da tabela dinâmica (limitado no Excelize)
func (s *Service) ConfigurePivotFields(sheetName, tableName string, rowFields []string, dataFields []map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentSessionID == "" {
		return fmt.Errorf("nenhum arquivo carregado")
	}

	// Excelize não suporta configuração avançada de pivot tables
	return fmt.Errorf("configuração de campos de pivot table não suportada no modo Excelize")
}

// ListCharts lista gráficos (limitação do Excelize)
func (s *Service) ListCharts(sheet string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return nil, err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.ListCharts(sheet)
}

// DeleteChart deleta um gráfico
func (s *Service) DeleteChart(sheet, chartName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.DeleteChart(sheet, chartName)
}

// CreateTable cria uma tabela
func (s *Service) CreateTable(sheet, rangeAddr, tableName, style string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.CreateTable(sheet, rangeAddr, tableName, style)
}

// ListTables lista tabelas
func (s *Service) ListTables(sheet string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return nil, err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.ListTables(sheet)
}

// DeleteTable deleta uma tabela
func (s *Service) DeleteTable(sheet, tableName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.DeleteTable(sheet, tableName)
}
