package excel

import "fmt"

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
