// client.go - Excel COM client and thread management
package excel

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"

	"excel-ai/pkg/logger"
)

// NewClient cria um novo cliente Excel conectando-se a uma instância existente
func NewClient() (*Client, error) {
	c := &Client{
		cmdChan:  make(chan func()),
		doneChan: make(chan struct{}),
	}

	errChan := make(chan error, 1)

	// Iniciar goroutine dedicada para operações COM
	go func() {
		// Bloquear esta goroutine na mesma thread do SO
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// Inicializar COM nesta thread
		ole.CoInitialize(0)
		defer ole.CoUninitialize()

		// Conectar ao Excel com Retry Loop
		var excelApp *ole.IDispatch
		for i := 0; i < 10; i++ {
			unknown, err := oleutil.GetActiveObject("Excel.Application")
			if err == nil {
				excelApp, err = unknown.QueryInterface(ole.IID_IDispatch)
				if err == nil {
					break
				}
			}
			time.Sleep(500 * time.Millisecond)
		}

		if excelApp == nil {
			errChan <- fmt.Errorf("falha ao conectar ao Excel (verifique se ele está aberto e não está em modo de edição)")
			return
		}

		c.excelApp = excelApp
		errChan <- nil

		// Loop de processamento de comandos
		for {
			select {
			case cmd := <-c.cmdChan:
				cmd()
			case <-c.doneChan:
				if c.excelApp != nil {
					c.excelApp.Release()
				}
				return
			}
		}
	}()

	// Aguardar inicialização
	if err := <-errChan; err != nil {
		return nil, err
	}

	return c, nil
}

// Close libera os recursos COM
func (c *Client) Close() {
	close(c.doneChan)
}

// runOnCOMThread executa uma função na thread COM e retorna o resultado
func (c *Client) runOnCOMThread(fn func() error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	errChan := make(chan error, 1)
	c.cmdChan <- func() {
		var err error
		// Tentar até 10 vezes com backoff se o Excel estiver ocupado
		for i := 0; i < 10; i++ {
			err = fn()
			if err == nil {
				break
			}
			// Verificar erro COM de "Busy" ou "Call rejected"
			errStr := err.Error()
			logger.ExcelDebug(fmt.Sprintf("[ExcelClient] Erro runOnCOMThread tentativa %d/10: %s", i+1, errStr))

			if strings.Contains(errStr, "Call was rejected by callee") ||
				strings.Contains(errStr, "A chamada foi rejeitada pelo chamado") ||
				strings.Contains(errStr, "80010001") {

				logger.ExcelWarn("[ExcelClient] Excel ocupado. Aguardando 1s...")
				time.Sleep(time.Millisecond * 1000)
				continue
			}
			break
		}
		errChan <- err
	}
	return <-errChan
}

// runOnCOMThreadWithResult executa uma função e retorna resultado genérico
func runOnCOMThreadWithResult[T any](c *Client, fn func() (T, error)) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	type result struct {
		value T
		err   error
	}
	resChan := make(chan result, 1)
	c.cmdChan <- func() {
		var v T
		var err error
		// Tentar até 10 vezes com backoff se o Excel estiver ocupado
		for i := 0; i < 10; i++ {
			v, err = fn()
			if err == nil {
				break
			}
			// Verificar erro COM de "Busy" ou "Call rejected"
			errStr := err.Error()
			logger.ExcelDebug(fmt.Sprintf("[ExcelClient] Erro runWithResult tentativa %d/10: %s", i+1, errStr))

			if strings.Contains(errStr, "Call was rejected by callee") ||
				strings.Contains(errStr, "A chamada foi rejeitada pelo chamado") ||
				strings.Contains(errStr, "80010001") {

				logger.ExcelWarn("[ExcelClient] Excel ocupado. Aguardando 1s...")
				time.Sleep(time.Millisecond * 1000)
				continue
			}
			break
		}
		resChan <- result{v, err}
	}
	res := <-resChan
	return res.value, res.err
}

// getWorkbookInternal obtém uma referência para uma pasta de trabalho (chamado internamente na thread COM)
func (c *Client) getWorkbookInternal(workbookName string) (*ole.IDispatch, error) {
	// Se nome vazio, retornar ActiveWorkbook
	if workbookName == "" {
		wbObj, err := oleutil.GetProperty(c.excelApp, "ActiveWorkbook")
		if err != nil {
			return nil, fmt.Errorf("falha ao obter ActiveWorkbook (nenhuma pasta aberta?): %w", err)
		}
		if wbObj.Val == 0 { // Check for null dispatch
			return nil, fmt.Errorf("nenhuma pasta de trabalho ativa")
		}
		return wbObj.ToIDispatch(), nil
	}

	workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
	if err != nil {
		return nil, err
	}
	defer workbooks.ToIDispatch().Release()

	wb, err := oleutil.GetProperty(workbooks.ToIDispatch(), "Item", workbookName)
	if err != nil {
		return nil, fmt.Errorf("pasta de trabalho '%s' não encontrada: %w", workbookName, err)
	}
	return wb.ToIDispatch(), nil
}

// GetActiveWorkbookName retorna o nome da pasta de trabalho ativa
func (c *Client) GetActiveWorkbookName() (string, error) {
	return runOnCOMThreadWithResult(c, func() (string, error) {
		wbObj, err := oleutil.GetProperty(c.excelApp, "ActiveWorkbook")
		if err != nil {
			return "", err
		}
		if wbObj.Val == 0 {
			return "", fmt.Errorf("nenhuma pasta de trabalho ativa")
		}

		wbDisp := wbObj.ToIDispatch()
		defer wbDisp.Release()

		nameVar, err := oleutil.GetProperty(wbDisp, "Name")
		if err != nil {
			return "", err
		}
		return nameVar.ToString(), nil
	})
}

// getSheetInternal obtém uma referência para uma aba específica (chamado internamente na thread COM)
func (c *Client) getSheetInternal(workbookName, sheetName string) (*ole.IDispatch, error) {
	wbDisp, err := c.getWorkbookInternal(workbookName)
	if err != nil {
		return nil, err
	}
	defer wbDisp.Release()

	sheets, err := oleutil.GetProperty(wbDisp, "Sheets")
	if err != nil {
		return nil, err
	}
	defer sheets.ToIDispatch().Release()

	sheet, err := oleutil.GetProperty(sheets.ToIDispatch(), "Item", sheetName)
	if err != nil {
		return nil, fmt.Errorf("aba '%s' não encontrada: %w", sheetName, err)
	}

	return sheet.ToIDispatch(), nil
}

// getSheetsInternal retorna as abas de uma pasta de trabalho (chamado internamente na thread COM)
func (c *Client) getSheetsInternal(wb *ole.IDispatch) ([]string, error) {
	sheets, err := oleutil.GetProperty(wb, "Sheets")
	if err != nil {
		return nil, err
	}
	defer sheets.ToIDispatch().Release()

	count, err := oleutil.GetProperty(sheets.ToIDispatch(), "Count")
	if err != nil {
		return nil, err
	}

	var result []string
	for i := 1; i <= int(count.Val); i++ {
		sheet, err := oleutil.GetProperty(sheets.ToIDispatch(), "Item", i)
		if err != nil {
			continue
		}
		sheetDisp := sheet.ToIDispatch()

		name, _ := oleutil.GetProperty(sheetDisp, "Name")
		result = append(result, name.ToString())
		sheetDisp.Release()
	}

	return result, nil
}
