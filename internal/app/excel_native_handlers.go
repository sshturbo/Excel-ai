package app

import (
	"excel-ai/pkg/logger"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// OpenFileNative abre um seletor de arquivos nativo e carrega o arquivo
func (a *App) OpenFileNative() (string, error) {
	logger.AppInfo("Abrindo seletor de arquivos nativo")

	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar Arquivo Excel",
		Filters: []runtime.FileFilter{
			{DisplayName: "Arquivos Excel (*.xlsx)", Pattern: "*.xlsx"},
		},
	})

	if err != nil {
		logger.AppError("Erro ao abrir seletor: " + err.Error())
		return "", fmt.Errorf("erro ao selecionar arquivo: %w", err)
	}

	if path == "" {
		logger.AppInfo("Seleção cancelada pelo usuário")
		return "", nil // Retorna vazio se cancelou
	}

	// Gerar sessionID único
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())

	// Carregar via path
	if err := a.excelService.ConnectFilePath(sessionID, path); err != nil {
		logger.AppError("Erro ao carregar arquivo do path: " + err.Error())
		return "", fmt.Errorf("erro ao carregar arquivo: %w", err)
	}

	// Vincular o path à conversa atual no banco de dados, se houver uma conversa
	convID := a.chatService.GetCurrentConversationID()
	if convID != "" && a.storage != nil {
		if err := a.storage.SetConversationExcelPath(convID, path); err != nil {
			logger.AppWarn("Falha ao salvar path do Excel no banco: " + err.Error())
		} else {
			logger.AppInfo("Caminho do Excel vinculado à conversa " + convID)
		}
	}

	logger.AppInfo("Arquivo carregado via path com sucesso: " + path)
	return sessionID, nil
}

// SaveFileNative salva o arquivo atual no disco (sobrescreve o original)
func (a *App) SaveFileNative() error {
	logger.AppInfo("Solicitando salvamento nativo no disco")

	if err := a.excelService.SaveToDisk(); err != nil {
		logger.AppError("Erro ao salvar no disco: " + err.Error())
		return fmt.Errorf("erro ao salvar arquivo: %w", err)
	}

	logger.AppInfo("Salvamento nativo concluído")
	return nil
}
