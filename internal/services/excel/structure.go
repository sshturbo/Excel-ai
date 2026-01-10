package excel

import "fmt"

// CreateNewWorkbook cria um novo arquivo Excel em memória
func (s *Service) CreateNewWorkbook() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// No modo Excelize, criar um arquivo novo é feito via upload
	// Retornar erro informativo
	return "", fmt.Errorf("use upload de arquivo para criar uma nova planilha")
}

// CreateNewSheet cria uma nova planilha no arquivo atual
func (s *Service) CreateNewSheet(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	return client.CreateSheet(name)
}

// DeleteSheet deleta uma planilha
func (s *Service) DeleteSheet(sheetName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	return client.DeleteSheet(sheetName)
}

// RenameSheet renomeia uma planilha
func (s *Service) RenameSheet(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	return client.RenameSheet(oldName, newName)
}

// ClearRange limpa um range de células
func (s *Service) ClearRange(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.ClearRange(sheet, rangeAddr)
}

// AutoFitColumns ajusta automaticamente a largura das colunas
func (s *Service) AutoFitColumns(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.AutoFitColumns(sheet, rangeAddr)
}

// InsertRows insere linhas
func (s *Service) InsertRows(sheet string, rowNumber, count int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.InsertRows(sheet, rowNumber, count)
}

// DeleteRows deleta linhas
func (s *Service) DeleteRows(sheet string, rowNumber, count int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.DeleteRows(sheet, rowNumber, count)
}

// MergeCells mescla células
func (s *Service) MergeCells(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.MergeCells(sheet, rangeAddr)
}

// UnmergeCells desmescla células
func (s *Service) UnmergeCells(sheet, rangeAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}

	if sheet == "" {
		sheet = s.getFirstSheet()
	}

	return client.UnmergeCells(sheet, rangeAddr)
}
