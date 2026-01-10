package app

import (
	"excel-ai/pkg/logger"
	"fmt"
	"time"
)

// UploadExcel recebe um arquivo .xlsx e retorna o sessionID
func (a *App) UploadExcel(filename string, data []byte) (string, error) {
	logger.AppInfo("Recebendo upload de arquivo: " + filename)

	// Gerar sessionID único usando timestamp
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())

	// Conectar ao arquivo via Excelize
	if err := a.excelService.ConnectFile(sessionID, data); err != nil {
		logger.AppError("Erro ao conectar arquivo: " + err.Error())
		return "", fmt.Errorf("erro ao conectar arquivo: %w", err)
	}

	logger.AppInfo("Arquivo carregado com sucesso. SessionID: " + sessionID)
	return sessionID, nil
}

// DownloadExcel retorna o arquivo .xlsx modificado
func (a *App) DownloadExcel(sessionID string) ([]byte, error) {
	logger.AppInfo("Requisitando download do arquivo. SessionID: " + sessionID)

	data, err := a.excelService.ExportFile()
	if err != nil {
		logger.AppError("Erro ao exportar arquivo: " + err.Error())
		return nil, fmt.Errorf("erro ao exportar arquivo: %w", err)
	}

	logger.AppInfo("Arquivo exportado com sucesso")
	return data, nil
}

// GetExcelPreview retorna os dados para o viewer
type PreviewData struct {
	SessionID  string         `json:"sessionId"`
	FileName   string         `json:"fileName"`
	Sheets     []SheetPreview `json:"sheets"`
	ActiveSheet string         `json:"activeSheet"`
}

type SheetPreview struct {
	Name string `json:"name"`
	Rows int    `json:"rows"`
	Cols int    `json:"cols"`
}

// GetExcelPreview retorna um preview dos dados do arquivo
func (a *App) GetExcelPreview(sessionID string) (*PreviewData, error) {
	logger.AppInfo("Gerando preview. SessionID: " + sessionID)

	// Obter cliente Excelize
	excelClient, err := a.excelService.GetExcelClient()
	if err != nil {
		logger.AppError("Erro ao obter cliente Excelize: " + err.Error())
		return nil, fmt.Errorf("erro ao obter cliente: %w", err)
	}

	// Obter lista de planilhas
	sheets := excelClient.ListSheets()
	if len(sheets) == 0 {
		logger.AppError("Nenhuma planilha encontrada no arquivo")
		return nil, fmt.Errorf("arquivo não contém planilhas")
	}

	// Criar preview de cada planilha
	sheetPreviews := make([]SheetPreview, 0, len(sheets))
	for _, sheetName := range sheets {
		rows, err := excelClient.GetRowCount(sheetName)
		if err != nil {
			logger.AppWarn("Erro ao obter contagem de linhas: " + err.Error())
			rows = 0
		}

		cols, err := excelClient.GetColumnCount(sheetName)
		if err != nil {
			logger.AppWarn("Erro ao obter contagem de colunas: " + err.Error())
			cols = 0
		}

		sheetPreviews = append(sheetPreviews, SheetPreview{
			Name: sheetName,
			Rows: rows,
			Cols: cols,
		})
	}

	preview := &PreviewData{
		SessionID: sessionID,
		FileName:  "arquivo.xlsx", // TODO: Armazenar nome original
		Sheets:    sheetPreviews,
		// ActiveSheet será obtido via getter quando implementado
	}

	logger.AppInfo("Preview gerado com sucesso")
	return preview, nil
}

// GetSheetData retorna os dados de uma planilha específica
func (a *App) GetSheetData(sessionID, sheetName string) ([][]string, error) {
	logger.AppInfo("Obtendo dados da planilha: " + sheetName)

	// Obter cliente Excelize
	excelClient, err := a.excelService.GetExcelClient()
	if err != nil {
		logger.AppError("Erro ao obter cliente Excelize: " + err.Error())
		return nil, fmt.Errorf("erro ao obter cliente: %w", err)
	}

	// Obter range usado
	usedRange, err := excelClient.GetUsedRange(sheetName)
	if err != nil {
		logger.AppError("Erro ao obter range usado: " + err.Error())
		return nil, fmt.Errorf("erro ao obter range: %w", err)
	}

	// Obter dados do range
	data, err := excelClient.GetRangeValues(sheetName, usedRange)
	if err != nil {
		logger.AppError("Erro ao obter valores: " + err.Error())
		return nil, fmt.Errorf("erro ao obter valores: %w", err)
	}

	logger.AppInfo("Dados da planilha obtidos com sucesso")
	return data, nil
}

// CloseSession fecha uma sessão de arquivo
func (a *App) CloseSession(sessionID string) error {
	logger.AppInfo("Fechando sessão: " + sessionID)

	// Implementação futura: adicionar método no Service para fechar sessão específica
	// Por enquanto, fechamos todo o serviço
	a.excelService.Close()

	logger.AppInfo("Sessão fechada com sucesso")
	return nil
}
