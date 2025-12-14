package excel

import "fmt"

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
