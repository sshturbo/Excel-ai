package excel

import (
	"excel-ai/internal/dto"
	"excel-ai/pkg/excel"
	"excel-ai/pkg/logger"
	"excel-ai/pkg/storage"
	"fmt"
	"strings"
	"sync"
)

// Service gerencia operações de Excel usando apenas Excelize (sem COM)
// Suporta Linux, macOS e Windows sem necessidade de Excel instalado
type Service struct {
	fileManager         *excel.FileManager // Gerenciador de arquivos Excelize
	currentSessionID    string             // SessionID do arquivo atual
	currentFileName     string             // Nome do arquivo carregado
	mu                  sync.Mutex
	currentSheet        string
	previewData         *excel.SheetData
	undoStack           []dto.UndoAction
	currentBatchID      int64
	lastExecutedBatchID int64
	contextStr          string
	storage             *storage.Storage
	currentConvID       string
}

func NewService() *Service {
	return &Service{
		fileManager: excel.NewFileManager(),
		undoStack:   []dto.UndoAction{},
	}
}

// getClient retorna o cliente Excelize atual (helper interno)
func (s *Service) getClient() (*excel.ExcelizeClient, error) {
	if s.fileManager == nil || s.currentSessionID == "" {
		return nil, fmt.Errorf("nenhum arquivo carregado")
	}
	return s.fileManager.GetClient(s.currentSessionID)
}

// getClientLocked retorna o cliente sem lock (para uso interno quando já está locked)
func (s *Service) getClientLocked() (*excel.ExcelizeClient, error) {
	if s.fileManager == nil || s.currentSessionID == "" {
		return nil, fmt.Errorf("nenhum arquivo carregado")
	}
	return s.fileManager.GetClient(s.currentSessionID)
}

// SetStorage configura o storage para persistência de undo
func (s *Service) SetStorage(store *storage.Storage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage = store
	logger.ExcelInfo("Storage configurado")
}

// SetConversationID define a conversa atual para vincular ações de undo
func (s *Service) SetConversationID(convID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentConvID = convID
	logger.ExcelDebug("ID da conversa definido: " + convID)
}

// GetConversationID retorna a conversa atual
func (s *Service) GetConversationID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentConvID
}

// getFirstSheet retorna a primeira aba quando currentSheet contém múltiplas abas
func (s *Service) getFirstSheet() string {
	if strings.Contains(s.currentSheet, ",") {
		return strings.TrimSpace(strings.Split(s.currentSheet, ",")[0])
	}
	return s.currentSheet
}

// Connect retorna status da conexão (para compatibilidade)
// No modo Excelize, retorna status baseado em arquivo carregado
func (s *Service) Connect() (*dto.ExcelStatus, error) {
	logger.ExcelInfo("Verificando status do serviço Excel")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentSessionID != "" {
		// Já tem arquivo carregado
		client, err := s.fileManager.GetClient(s.currentSessionID)
		if err == nil {
			sheets := client.ListSheets()
			workbook := excel.Workbook{
				Name:   s.currentFileName,
				Sheets: sheets,
			}
			return &dto.ExcelStatus{
				Connected: true,
				Workbooks: []excel.Workbook{workbook},
			}, nil
		}
	}

	// Sem arquivo carregado - retorna pronto para receber upload
	return &dto.ExcelStatus{
		Connected: false,
		Error:     "Faça upload de um arquivo .xlsx para começar",
	}, nil
}

// RefreshWorkbooks atualiza lista de workbooks (retorna arquivo atual)
func (s *Service) RefreshWorkbooks() (*dto.ExcelStatus, error) {
	return s.Connect()
}

// GetActiveWorkbookName retorna o nome do arquivo atual
func (s *Service) GetActiveWorkbookName() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentFileName == "" {
		return "", fmt.Errorf("nenhum arquivo carregado")
	}
	return s.currentFileName, nil
}

// Close fecha o serviço e libera recursos
func (s *Service) Close() {
	logger.ExcelInfo("Fechando serviço Excel")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fileManager != nil {
		s.fileManager.CloseAll()
	}
	s.currentSessionID = ""
	s.currentFileName = ""
	s.currentSheet = ""
}

// SaveUndoAction salva uma ação de undo no banco de dados
func (s *Service) SaveUndoAction(opType, workbook, sheet, cell, oldValue, undoData string) error {
	logger.ExcelDebug(fmt.Sprintf("Salvando ação de undo: %s", opType))
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.storage != nil && s.currentConvID != "" && s.currentBatchID != 0 {
		err := s.storage.SaveUndoActionFull(s.currentConvID, s.currentBatchID, opType, workbook, sheet, cell, oldValue, undoData)
		if err != nil {
			logger.ExcelError("Erro ao salvar ação de undo: " + err.Error())
		} else {
			logger.ExcelDebug("Ação de undo salva com sucesso")
		}
		return err
	}
	return nil
}

// ===== Métodos para gerenciamento de arquivos =====

// ConnectFile carrega um arquivo Excel via Excelize
func (s *Service) ConnectFile(sessionID string, data []byte) error {
	logger.ExcelInfo("Carregando arquivo Excel (sessionID: " + sessionID + ")")
	s.mu.Lock()
	defer s.mu.Unlock()

	// Carregar arquivo
	if err := s.fileManager.LoadFile(sessionID, data); err != nil {
		logger.ExcelError("Erro ao carregar arquivo: " + err.Error())
		return fmt.Errorf("erro ao carregar arquivo: %w", err)
	}

	s.currentSessionID = sessionID

	// Obter primeira planilha como padrão
	client, err := s.fileManager.GetClient(sessionID)
	if err == nil {
		sheets := client.ListSheets()
		if len(sheets) > 0 {
			s.currentSheet = sheets[0]
		}
	}

	logger.ExcelInfo("Arquivo carregado com sucesso")
	return nil
}

// ConnectFileWithName carrega um arquivo e guarda o nome
func (s *Service) ConnectFileWithName(sessionID, fileName string, data []byte) error {
	err := s.ConnectFile(sessionID, data)
	if err == nil {
		s.mu.Lock()
		s.currentFileName = fileName
		s.mu.Unlock()
	}
	return err
}

// ConnectFilePath carrega um arquivo a partir de um caminho no disco
func (s *Service) ConnectFilePath(sessionID, path string) error {
	logger.ExcelInfo(fmt.Sprintf("Carregando arquivo do disco: %s (sessionID: %s)", path, sessionID))
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.fileManager.LoadFileFromPath(sessionID, path); err != nil {
		logger.ExcelError("Erro ao carregar arquivo do disco: " + err.Error())
		return fmt.Errorf("erro ao carregar arquivo do disco: %w", err)
	}

	s.currentSessionID = sessionID

	// Extrair nome do arquivo do path
	parts := strings.Split(path, "\\")
	if len(parts) == 1 {
		parts = strings.Split(path, "/")
	}
	s.currentFileName = parts[len(parts)-1]

	// Obter primeira planilha como padrão
	client, err := s.fileManager.GetClient(sessionID)
	if err == nil {
		sheets := client.ListSheets()
		if len(sheets) > 0 {
			s.currentSheet = sheets[0]
		}
	}

	logger.ExcelInfo("Arquivo carregado via path com sucesso")
	return nil
}

// SaveToDisk salva as alterações de volta ao disco
func (s *Service) SaveToDisk() error {
	logger.ExcelInfo("Salvando alterações no disco")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fileManager == nil || s.currentSessionID == "" {
		return fmt.Errorf("nenhum arquivo carregado")
	}

	client, err := s.fileManager.GetClient(s.currentSessionID)
	if err != nil {
		return err
	}

	path := client.GetFilePath()
	if path == "" {
		return fmt.Errorf("o arquivo não tem um caminho de disco associado. Use ExportFile em vez disso.")
	}

	if err := client.SaveAs(path); err != nil {
		logger.ExcelError("Erro ao salvar no disco: " + err.Error())
		// Fornecer mensagem mais amigável se o arquivo estiver bloqueado
		errMsg := err.Error()
		if strings.Contains(errMsg, "process cannot access") || strings.Contains(errMsg, "used by another process") || strings.Contains(errMsg, "permission denied") {
			return fmt.Errorf("o arquivo está aberto em outro programa (como o Excel). Por favor, feche o arquivo no outro programa e tente salvar novamente")
		}
		return fmt.Errorf("erro ao salvar no disco: %w", err)
	}

	logger.ExcelInfo("Arquivo salvo no disco com sucesso em: " + path)
	return nil
}

// ExportFile exporta o arquivo atual como bytes
func (s *Service) ExportFile() ([]byte, error) {
	logger.ExcelInfo("Exportando arquivo")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fileManager == nil || s.currentSessionID == "" {
		return nil, fmt.Errorf("nenhum arquivo carregado")
	}

	data, err := s.fileManager.Export(s.currentSessionID)
	if err != nil {
		logger.ExcelError("Erro ao exportar arquivo: " + err.Error())
		return nil, fmt.Errorf("erro ao exportar arquivo: %w", err)
	}

	logger.ExcelInfo("Arquivo exportado com sucesso")
	return data, nil
}

// IsFileMode sempre retorna true (para compatibilidade)
func (s *Service) IsFileMode() bool {
	return true
}

// IsConnected retorna true se há um arquivo carregado
func (s *Service) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentSessionID != ""
}

// GetExcelClient retorna o cliente Excelize atual
func (s *Service) GetExcelClient() (*excel.ExcelizeClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getClientLocked()
}

// GetCurrentFileName retorna o nome do arquivo atual
func (s *Service) GetCurrentFileName() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentFileName
}

// GetCurrentSheet retorna a planilha atual
func (s *Service) GetCurrentSheet() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentSheet
}

// SetCurrentSheet define a planilha atual
func (s *Service) SetCurrentSheet(sheet string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentSheet = sheet
}

// CloseFile fecha o arquivo atual
func (s *Service) CloseFile() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fileManager != nil && s.currentSessionID != "" {
		s.fileManager.Close(s.currentSessionID)
	}
	s.currentSessionID = ""
	s.currentFileName = ""
	s.currentSheet = ""
}
