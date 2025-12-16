// types.go - Excel types and data structures
package excel

import (
	"sync"

	"github.com/go-ole/go-ole"
)

// Client gerencia a conexão COM com o Excel usando uma thread dedicada
type Client struct {
	excelApp *ole.IDispatch
	cmdChan  chan func()
	doneChan chan struct{}
	mu       sync.Mutex
}

// Workbook representa uma pasta de trabalho aberta
type Workbook struct {
	Name   string   `json:"name"`
	Path   string   `json:"path"`
	Sheets []string `json:"sheets"`
}

// CellData representa dados de uma célula
type CellData struct {
	Row   int         `json:"row"`
	Col   int         `json:"col"`
	Value interface{} `json:"value"`
	Text  string      `json:"text"`
}

// SheetData representa dados de uma planilha
type SheetData struct {
	Name    string       `json:"name"`
	Headers []string     `json:"headers"`
	Rows    [][]CellData `json:"rows"`
}

// Format representa formatação de célula/range
type Format struct {
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	FontSize  int    `json:"fontSize"`
	FontColor string `json:"fontColor"` // Hex
	BgColor   string `json:"bgColor"`   // Hex
}
