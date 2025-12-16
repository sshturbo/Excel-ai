// formatting.go - Excel cell/range formatting operations
package excel

import (
	"fmt"
	"strings"

	"github.com/go-ole/go-ole/oleutil"
)

// FormatRange aplica formatação a um range de células
// options: bold, italic, fontSize, fontColor (hex), bgColor (hex)
func (c *Client) FormatRange(workbookName, sheetName, rangeAddr string, bold, italic bool, fontSize int, fontColor, bgColor string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		// Font
		font, _ := oleutil.GetProperty(rangeDisp, "Font")
		fontDisp := font.ToIDispatch()
		defer fontDisp.Release()

		if bold {
			oleutil.PutProperty(fontDisp, "Bold", true)
		}
		if italic {
			oleutil.PutProperty(fontDisp, "Italic", true)
		}
		if fontSize > 0 {
			oleutil.PutProperty(fontDisp, "Size", fontSize)
		}
		if fontColor != "" {
			// Converter hex para RGB
			color := hexToRGB(fontColor)
			oleutil.PutProperty(fontDisp, "Color", color)
		}

		// Background color
		if bgColor != "" {
			interior, _ := oleutil.GetProperty(rangeDisp, "Interior")
			interiorDisp := interior.ToIDispatch()
			defer interiorDisp.Release()
			color := hexToRGB(bgColor)
			oleutil.PutProperty(interiorDisp, "Color", color)
		}

		return nil
	})
}

// hexToRGB converte cor hex (#RRGGBB) para valor RGB do Excel (BGR)
func hexToRGB(hex string) int {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	// Excel usa BGR
	return b<<16 | g<<8 | r
}

// MergeCells mescla células em um range
func (c *Client) MergeCells(workbookName, sheetName, rangeAddr string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		_, err = oleutil.CallMethod(rangeDisp, "Merge")
		return err
	})
}

// UnmergeCells desfaz mesclagem de células
func (c *Client) UnmergeCells(workbookName, sheetName, rangeAddr string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		_, err = oleutil.CallMethod(rangeDisp, "UnMerge")
		return err
	})
}

// SetBorders adiciona bordas a um range
// style: "thin", "medium", "thick"
func (c *Client) SetBorders(workbookName, sheetName, rangeAddr, style string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		// Determinar peso da borda
		weight := 2 // xlThin
		switch style {
		case "medium":
			weight = -4138 // xlMedium
		case "thick":
			weight = 4 // xlThick
		}

		// Aplicar bordas em todos os lados (xlEdgeLeft=7, xlEdgeTop=8, xlEdgeBottom=9, xlEdgeRight=10, xlInsideVertical=11, xlInsideHorizontal=12)
		borders, _ := oleutil.GetProperty(rangeDisp, "Borders")
		bordersDisp := borders.ToIDispatch()
		defer bordersDisp.Release()

		for _, edge := range []int{7, 8, 9, 10, 11, 12} {
			border, err := oleutil.GetProperty(bordersDisp, "Item", edge)
			if err != nil {
				continue
			}
			borderDisp := border.ToIDispatch()
			oleutil.PutProperty(borderDisp, "LineStyle", 1) // xlContinuous
			oleutil.PutProperty(borderDisp, "Weight", weight)
			borderDisp.Release()
		}

		return nil
	})
}

// SetColumnWidth define largura de colunas
func (c *Client) SetColumnWidth(workbookName, sheetName, rangeAddr string, width float64) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		cols, _ := oleutil.GetProperty(rangeDisp, "Columns")
		colsDisp := cols.ToIDispatch()
		defer colsDisp.Release()

		_, err = oleutil.PutProperty(colsDisp, "ColumnWidth", width)
		return err
	})
}

// SetRowHeight define altura de linhas
func (c *Client) SetRowHeight(workbookName, sheetName, rangeAddr string, height float64) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		rows, _ := oleutil.GetProperty(rangeDisp, "Rows")
		rowsDisp := rows.ToIDispatch()
		defer rowsDisp.Release()

		_, err = oleutil.PutProperty(rowsDisp, "RowHeight", height)
		return err
	})
}

// AutoFitColumns ajusta automaticamente a largura das colunas de um range
func (c *Client) AutoFitColumns(workbookName, sheetName, rangeAddr string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		// Obter colunas do range
		cols, _ := oleutil.GetProperty(rangeDisp, "Columns")
		colsDisp := cols.ToIDispatch()
		defer colsDisp.Release()

		_, err = oleutil.CallMethod(colsDisp, "AutoFit")
		return err
	})
}

// ClearRange limpa o conteúdo de um range (mantém formatação)
func (c *Client) ClearRange(workbookName, sheetName, rangeAddr string) error {
	return c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		_, err = oleutil.CallMethod(rangeDisp, "ClearContents")
		return err
	})
}

// GetColumnWidth retorna a largura da primeira coluna do range
func (c *Client) GetColumnWidth(workbookName, sheetName, rangeAddr string) (float64, error) {
	var width float64
	err := c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		// Columns
		cols, _ := oleutil.GetProperty(rangeDisp, "Columns")
		colsDisp := cols.ToIDispatch()
		defer colsDisp.Release()

		w, err := oleutil.GetProperty(colsDisp, "ColumnWidth")
		if err != nil {
			return err
		}
		if val, ok := w.Value().(float64); ok {
			width = val
		} else {
			// Pode retornar null se larguras variadas. Retorna a padrão ou erro?
			// Vamos assumir primeira coluna.
			col1, _ := oleutil.GetProperty(colsDisp, "Item", 1)
			col1Disp := col1.ToIDispatch()
			defer col1Disp.Release()
			w1, _ := oleutil.GetProperty(col1Disp, "ColumnWidth")
			if val1, ok1 := w1.Value().(float64); ok1 {
				width = val1
			}
		}
		return nil
	})
	return width, err
}

// GetRowHeight retorna a altura da primeira linha do range
func (c *Client) GetRowHeight(workbookName, sheetName, rangeAddr string) (float64, error) {
	var height float64
	err := c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		// Rows
		rows, _ := oleutil.GetProperty(rangeDisp, "Rows")
		rowsDisp := rows.ToIDispatch()
		defer rowsDisp.Release()

		h, err := oleutil.GetProperty(rowsDisp, "RowHeight")
		if err != nil {
			return err
		}
		if val, ok := h.Value().(float64); ok {
			height = val
		} else {
			// Variadas
			row1, _ := oleutil.GetProperty(rowsDisp, "Item", 1)
			row1Disp := row1.ToIDispatch()
			defer row1Disp.Release()
			h1, _ := oleutil.GetProperty(row1Disp, "RowHeight")
			if val1, ok1 := h1.Value().(float64); ok1 {
				height = val1
			}
		}
		return nil
	})
	return height, err
}

// GetFormat lê a formatação atual de um range
// Lê da primeira célula do range (comum no Excel)
func (c *Client) GetFormat(workbookName, sheetName, rangeAddr string) (*Format, error) {
	var format *Format
	err := c.runOnCOMThread(func() error {
		sheet, err := c.getSheetInternal(workbookName, sheetName)
		if err != nil {
			return err
		}
		defer sheet.Release()

		rangeObj, err := oleutil.GetProperty(sheet, "Range", rangeAddr)
		if err != nil {
			return fmt.Errorf("range '%s' inválido: %w", rangeAddr, err)
		}
		rangeDisp := rangeObj.ToIDispatch()
		defer rangeDisp.Release()

		format = &Format{}

		// Font
		font, _ := oleutil.GetProperty(rangeDisp, "Font")
		if font.ToIDispatch() != nil {
			fontDisp := font.ToIDispatch()
			defer fontDisp.Release()

			boldVar, _ := oleutil.GetProperty(fontDisp, "Bold")
			if boldVar.Value() != nil {
				format.Bold, _ = boldVar.Value().(bool)
			}

			italicVar, _ := oleutil.GetProperty(fontDisp, "Italic")
			if italicVar.Value() != nil {
				format.Italic, _ = italicVar.Value().(bool)
			}

			sizeVar, _ := oleutil.GetProperty(fontDisp, "Size")
			if sizeVar.Value() != nil {
				if val, ok := sizeVar.Value().(float64); ok {
					format.FontSize = int(val)
				}
			}

			colorVar, _ := oleutil.GetProperty(fontDisp, "Color")
			if colorVar.Value() != nil {
				if val, ok := colorVar.Value().(float64); ok { // OLE returns currency/double for colors sometimes? int usually.
					format.FontColor = rgbToHex(int(val))
				} else if val, ok := colorVar.Value().(int64); ok {
					format.FontColor = rgbToHex(int(val))
				}
			}
		}

		// Background color
		interior, _ := oleutil.GetProperty(rangeDisp, "Interior")
		if interior.ToIDispatch() != nil {
			interiorDisp := interior.ToIDispatch()
			defer interiorDisp.Release()

			colorVar, _ := oleutil.GetProperty(interiorDisp, "Color")
			if colorVar.Value() != nil {
				if val, ok := colorVar.Value().(float64); ok {
					format.BgColor = rgbToHex(int(val))
				} else if val, ok := colorVar.Value().(int64); ok {
					format.BgColor = rgbToHex(int(val))
				}
			}
		}

		return nil
	})
	return format, err
}

// rgbToHex converte int RGB (do Excel, que é BGR) para hex #RRGGBB
func rgbToHex(color int) string {
	r := color & 0xFF
	g := (color >> 8) & 0xFF
	b := (color >> 16) & 0xFF
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
