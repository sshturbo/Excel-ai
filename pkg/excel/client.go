// client.go - Excel COM client and thread management
package excel

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
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

		// Conectar ao Excel
		unknown, err := oleutil.GetActiveObject("Excel.Application")
		if err != nil {
			errChan <- fmt.Errorf("nenhuma instância do Excel encontrada: %w", err)
			return
		}

		excelApp, err := unknown.QueryInterface(ole.IID_IDispatch)
		if err != nil {
			errChan <- fmt.Errorf("falha ao obter interface Excel: %w", err)
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
		// Tentar até 3 vezes com backoff se o Excel estiver ocupado
		for i := 0; i < 3; i++ {
			err = fn()
			if err == nil {
				break
			}
			// Verificar se é erro de "Call was rejected by callee" (Excel ocupado)
			if strings.Contains(err.Error(), "Call was rejected by callee") ||
				strings.Contains(err.Error(), "A chamada foi rejeitada pelo chamado") {
				time.Sleep(time.Millisecond * 500 * time.Duration(i+1))
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
		// Tentar até 3 vezes com backoff se o Excel estiver ocupado
		for i := 0; i < 3; i++ {
			v, err = fn()
			if err == nil {
				break
			}
			// Verificar se é erro de "Call was rejected by callee" (Excel ocupado)
			if strings.Contains(err.Error(), "Call was rejected by callee") ||
				strings.Contains(err.Error(), "A chamada foi rejeitada pelo chamado") {
				time.Sleep(time.Millisecond * 500 * time.Duration(i+1))
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
