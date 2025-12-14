package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StartWorkbookWatcher inicia o monitoramento de mudanças nas planilhas
func (a *App) StartWorkbookWatcher() {
	// Para qualquer watcher existente
	a.StopWorkbookWatcher()

	ctx, cancel := context.WithCancel(a.ctx)
	a.watcherCancel = cancel

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Verificar se ainda está conectado
				status, err := a.excelService.RefreshWorkbooks()
				if err != nil {
					continue
				}
				if status == nil {
					continue
				}

				// Se houver erro de conexão, emitir desconexão
				if status.Error != "" && !status.Connected {
					if a.lastWorkbooksState != "" {
						a.lastWorkbooksState = ""
						runtime.EventsEmit(a.ctx, "excel:workbooks-changed", status)
					}
					continue
				}

				// Serializar estado atual para comparar
				currentState, _ := json.Marshal(status.Workbooks)
				currentStateStr := string(currentState)

				// Se mudou (inclui ficar vazio), emitir evento
				if currentStateStr != a.lastWorkbooksState {
					a.lastWorkbooksState = currentStateStr
					runtime.EventsEmit(a.ctx, "excel:workbooks-changed", status)
				}
			}
		}
	}()
}

// StopWorkbookWatcher para o monitoramento
func (a *App) StopWorkbookWatcher() {
	if a.watcherCancel != nil {
		a.watcherCancel()
		a.watcherCancel = nil
	}
}
