package excel

// advanced.go - Métodos avançados do Excel Service
// Wrappers para features avançadas do Excelize

// AddDropdownList adiciona uma lista dropdown a um range
func (s *Service) AddDropdownList(sheet, rng string, options []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.AddDropdownList(sheet, rng, options)
}

// AddCellComment adiciona um comentário a uma célula
func (s *Service) AddCellComment(sheet, cell, author, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.AddCellComment(sheet, cell, author, text)
}

// DeleteCellComment remove um comentário de uma célula
func (s *Service) DeleteCellComment(sheet, cell string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.DeleteCellComment(sheet, cell)
}

// AddHyperlink adiciona um hyperlink a uma célula
func (s *Service) AddHyperlink(sheet, cell, url, display string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.AddHyperlink(sheet, cell, url, display)
}

// FreezePane congela linhas/colunas
func (s *Service) FreezePane(sheet, cell string, freezeRows, freezeCols int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.FreezePane(sheet, cell, freezeRows, freezeCols)
}

// UnfreezePane remove o congelamento de painéis
func (s *Service) UnfreezePane(sheet string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.UnfreezePane(sheet)
}

// HideSheet oculta uma planilha
func (s *Service) HideSheet(sheet string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.HideSheet(sheet)
}

// ShowSheet exibe uma planilha oculta
func (s *Service) ShowSheet(sheet string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.ShowSheet(sheet)
}

// ProtectSheet protege uma planilha com senha
func (s *Service) ProtectSheet(sheet, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.ProtectSheet(sheet, password, nil)
}

// UnprotectSheet remove a proteção de uma planilha
func (s *Service) UnprotectSheet(sheet, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.UnprotectSheet(sheet, password)
}

// SetCellLocked define se uma célula está bloqueada
func (s *Service) SetCellLocked(sheet, cell string, locked bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.SetCellLocked(sheet, cell, locked)
}

// SetCellFormula define uma fórmula em uma célula
func (s *Service) SetCellFormula(sheet, cell, formula string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.SetCellFormula(sheet, cell, formula)
}

// AddSimpleConditionalFormat adiciona formatação condicional simplificada
func (s *Service) AddSimpleConditionalFormat(sheet, rng, criteria, value, bgColor string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.AddSimpleConditionalFormat(sheet, rng, criteria, value, bgColor)
}

// DeletePivotTable remove uma tabela dinâmica
func (s *Service) DeletePivotTable(sheet, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := s.getClientLocked()
	if err != nil {
		return err
	}
	return client.DeletePivotTable(sheet, name)
}
