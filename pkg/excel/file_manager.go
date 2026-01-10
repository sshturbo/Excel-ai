package excel

import (
	"fmt"
	"sync"
)

// FileManager gerencia múltiplas sessões de arquivos Excelize
type FileManager struct {
	sessions map[string]*ExcelizeClient // sessionID -> client
	mu       sync.RWMutex
}

// NewFileManager cria um novo FileManager
func NewFileManager() *FileManager {
	return &FileManager{
		sessions: make(map[string]*ExcelizeClient),
	}
}

// LoadFile carrega um arquivo a partir de bytes e associa a uma sessão
func (fm *FileManager) LoadFile(sessionID string, data []byte) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	// Fechar sessão anterior se existir
	if existingClient, exists := fm.sessions[sessionID]; exists && existingClient != nil {
		existingClient.Close()
	}
	
	// Criar novo cliente
	client, err := NewExcelizeClient(data)
	if err != nil {
		return fmt.Errorf("failed to load Excel file: %w", err)
	}
	
	fm.sessions[sessionID] = client
	return nil
}

// GetClient retorna o cliente de uma sessão
func (fm *FileManager) GetClient(sessionID string) (*ExcelizeClient, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	client, exists := fm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	if client == nil {
		return nil, fmt.Errorf("client is nil for session: %s", sessionID)
	}
	
	return client, nil
}

// Export exporta o arquivo de uma sessão para bytes
func (fm *FileManager) Export(sessionID string) ([]byte, error) {
	client, err := fm.GetClient(sessionID)
	if err != nil {
		return nil, err
	}
	
	return client.Write()
}

// Close fecha uma sessão e libera recursos
func (fm *FileManager) Close(sessionID string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	if client, exists := fm.sessions[sessionID]; exists && client != nil {
		client.Close()
		delete(fm.sessions, sessionID)
	}
}

// ListSessions retorna uma lista de todos os IDs de sessão ativos
func (fm *FileManager) ListSessions() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	sessions := make([]string, 0, len(fm.sessions))
	for id := range fm.sessions {
		sessions = append(sessions, id)
	}
	return sessions
}

// SessionExists verifica se uma sessão existe
func (fm *FileManager) SessionExists(sessionID string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	_, exists := fm.sessions[sessionID]
	return exists
}

// CloseAll fecha todas as sessões
func (fm *FileManager) CloseAll() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	for id, client := range fm.sessions {
		if client != nil {
			client.Close()
		}
		delete(fm.sessions, id)
	}
}
