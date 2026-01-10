// pivot.go - Excel pivot table operations
package excel

import (
	"fmt"
	"strings"

	"github.com/go-ole/go-ole/oleutil"

	"excel-ai/pkg/logger"
)

// CreatePivotTable cria uma tabela dinâmica usando PivotTableWizard
func (c *Client) CreatePivotTable(workbookName, sourceSheet, sourceRange, destSheet, destCell, tableName string) error {
	return c.runOnCOMThread(func() error {
		// Obter workbook
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return fmt.Errorf("falha ao acessar Workbooks: %w", err)
		}
		workbooksDisp := workbooks.ToIDispatch()
		defer workbooksDisp.Release()

		wb, err := oleutil.GetProperty(workbooksDisp, "Item", workbookName)
		if err != nil {
			return fmt.Errorf("pasta de trabalho '%s' não encontrada: %w", workbookName, err)
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		// Ativar o workbook
		oleutil.CallMethod(wbDisp, "Activate")

		// Obter Worksheets (não Sheets) - Worksheets é específico para planilhas
		worksheets, err := oleutil.GetProperty(wbDisp, "Worksheets")
		if err != nil {
			return fmt.Errorf("falha ao acessar Worksheets: %w", err)
		}
		worksheetsDisp := worksheets.ToIDispatch()
		defer worksheetsDisp.Release()

		// Obter aba de origem
		srcSheet, err := oleutil.GetProperty(worksheetsDisp, "Item", sourceSheet)
		if err != nil {
			return fmt.Errorf("a aba de origem '%s' não existe: %w", sourceSheet, err)
		}
		srcSheetDisp := srcSheet.ToIDispatch()
		defer srcSheetDisp.Release()

		// Expandir range se necessário
		expandedSourceRange := strings.TrimSpace(sourceRange)

		// Remover referência à aba se a IA a incluiu (ex: 'Custo de 2024'!A1:F200 ou Custo de 2024!A1:F200)
		if idx := strings.Index(expandedSourceRange, "!"); idx != -1 {
			expandedSourceRange = expandedSourceRange[idx+1:]
		}

		// Se apenas colunas (ex: A:F), expande para A1:F<lastRow>
		if strings.Contains(expandedSourceRange, ":") {
			parts := strings.Split(expandedSourceRange, ":")
			if len(parts) == 2 {
				hasDigit0 := strings.IndexFunc(parts[0], func(r rune) bool { return r >= '0' && r <= '9' }) != -1
				hasDigit1 := strings.IndexFunc(parts[1], func(r rune) bool { return r >= '0' && r <= '9' }) != -1
				if !hasDigit0 && !hasDigit1 {
					startCol := strings.ToUpper(strings.TrimSpace(parts[0]))
					endCol := strings.ToUpper(strings.TrimSpace(parts[1]))

					usedRange, uErr := oleutil.GetProperty(srcSheetDisp, "UsedRange")
					if uErr == nil {
						usedRangeDisp := usedRange.ToIDispatch()
						urRows, _ := oleutil.GetProperty(usedRangeDisp, "Rows")
						urRowsDisp := urRows.ToIDispatch()
						countVar, _ := oleutil.GetProperty(urRowsDisp, "Count")
						startRowVar, _ := oleutil.GetProperty(usedRangeDisp, "Row")
						lastRow := int(startRowVar.Val) + int(countVar.Val) - 1
						if lastRow < 2 {
							lastRow = 2
						}
						expandedSourceRange = fmt.Sprintf("%s1:%s%d", startCol, endCol, lastRow)
						urRowsDisp.Release()
						usedRangeDisp.Release()
					}
				}
			}
		}

		logger.ExcelDebug(fmt.Sprintf("[DEBUG] Range expandido: %s", expandedSourceRange))

		// Obter range de origem
		srcRange, err := oleutil.GetProperty(srcSheetDisp, "Range", expandedSourceRange)
		if err != nil {
			return fmt.Errorf("range de origem inválido '%s': %w", expandedSourceRange, err)
		}
		srcRangeDisp := srcRange.ToIDispatch()
		defer srcRangeDisp.Release()

		// Obter endereço externo (inclui nome do workbook e aba automaticamente)
		// Address(RowAbsolute, ColumnAbsolute, ReferenceStyle, External, RelativeTo)
		srcAddressExternal, err := oleutil.GetProperty(srcRangeDisp, "Address", true, true, 1, true) // 1 = xlA1, true = External
		var fullSourceAddress string
		if err == nil && srcAddressExternal.ToString() != "" {
			fullSourceAddress = srcAddressExternal.ToString()
		} else {
			// Fallback: construir manualmente
			srcAddressProp, _ := oleutil.GetProperty(srcRangeDisp, "Address", true, true)
			fullSourceAddress = fmt.Sprintf("'%s'!%s", sourceSheet, srcAddressProp.ToString())
		}
		logger.ExcelDebug(fmt.Sprintf("[DEBUG] Endereço completo da fonte: %s", fullSourceAddress))

		// Obter aba de destino
		destSheetObj, err := oleutil.GetProperty(worksheetsDisp, "Item", destSheet)
		if err != nil {
			return fmt.Errorf("a aba de destino '%s' não existe. Crie a aba primeiro usando create-sheet", destSheet)
		}
		destSheetDisp := destSheetObj.ToIDispatch()
		defer destSheetDisp.Release()

		// Ativar aba de destino
		oleutil.CallMethod(destSheetDisp, "Activate")

		// Obter célula de destino
		destRange, err := oleutil.GetProperty(destSheetDisp, "Range", destCell)
		if err != nil {
			return fmt.Errorf("célula de destino inválida '%s': %w", destCell, err)
		}
		destRangeDisp := destRange.ToIDispatch()
		defer destRangeDisp.Release()

		// Construir endereço de destino como string
		destAddressProp, _ := oleutil.GetProperty(destRangeDisp, "Address", true, true)
		fullDestAddress := fmt.Sprintf("'%s'!%s", destSheet, destAddressProp.ToString())

		// Usar PivotTableWizard
		// Parâmetros: SourceType, SourceData, TableDestination, TableName
		// xlDatabase = 1
		logger.ExcelDebug(fmt.Sprintf("[DEBUG] Chamando PivotTableWizard com:\n  Source: %s\n  Dest: %s\n  Name: %s", fullSourceAddress, fullDestAddress, tableName))

		// Tentar primeiro usando objeto Range para fonte (mais preciso)
		_, err = oleutil.CallMethod(destSheetDisp, "PivotTableWizard",
			1,             // SourceType = xlDatabase
			srcRangeDisp,  // SourceData como Range object
			destRangeDisp, // TableDestination como Range object
			tableName,     // TableName
		)
		if err != nil {
			logger.ExcelDebug(fmt.Sprintf("[DEBUG] PivotTableWizard com Range objects falhou: %v", err))

			// Tentar usando a aba de ORIGEM para chamar PivotTableWizard
			_, err = oleutil.CallMethod(srcSheetDisp, "PivotTableWizard",
				1,             // SourceType = xlDatabase
				srcRangeDisp,  // SourceData como Range object
				destRangeDisp, // TableDestination como Range object
				tableName,     // TableName
			)
			if err != nil {
				logger.ExcelDebug(fmt.Sprintf("[DEBUG] PivotTableWizard via srcSheet falhou: %v", err))

				// Tentar com endereço como string para fonte
				_, err = oleutil.CallMethod(destSheetDisp, "PivotTableWizard",
					1,                 // SourceType = xlDatabase
					fullSourceAddress, // SourceData como string
					destRangeDisp,     // TableDestination como Range
					tableName,         // TableName
				)
				if err != nil {
					errStr := err.Error()
					logger.ExcelDebug(fmt.Sprintf("[DEBUG] PivotTableWizard com string source falhou: %v", err))
					// Verificar se é erro de campos inválidos
					if strings.Contains(errStr, "campo") || strings.Contains(errStr, "field") || strings.Contains(errStr, "colunas rotuladas") {
						return fmt.Errorf("os dados de origem têm colunas sem cabeçalho. Verifique se todas as colunas na primeira linha têm um título")
					}
					return fmt.Errorf("falha ao criar Tabela Dinâmica: %w", err)
				}
			}
		}

		logger.ExcelDebug("[DEBUG] Tabela Dinâmica criada com sucesso!")
		return nil
	})
}

// ConfigurePivotFields configura os campos de uma tabela dinâmica
// rowFields: campos para as linhas
// dataFields: campos para os valores (com função de agregação)
func (c *Client) ConfigurePivotFields(workbookName, sheetName, tableName string, rowFields []string, dataFields []map[string]string) error {
	return c.runOnCOMThread(func() error {
		// Obter workbook
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return err
		}
		workbooksDisp := workbooks.ToIDispatch()
		defer workbooksDisp.Release()

		wb, err := oleutil.GetProperty(workbooksDisp, "Item", workbookName)
		if err != nil {
			return err
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		// Obter worksheet
		worksheets, err := oleutil.GetProperty(wbDisp, "Worksheets")
		if err != nil {
			return err
		}
		worksheetsDisp := worksheets.ToIDispatch()
		defer worksheetsDisp.Release()

		sheet, err := oleutil.GetProperty(worksheetsDisp, "Item", sheetName)
		if err != nil {
			return err
		}
		sheetDisp := sheet.ToIDispatch()
		defer sheetDisp.Release()

		// Obter PivotTables
		pivotTables, err := oleutil.GetProperty(sheetDisp, "PivotTables")
		if err != nil {
			return fmt.Errorf("falha ao acessar PivotTables: %w", err)
		}
		pivotTablesDisp := pivotTables.ToIDispatch()
		defer pivotTablesDisp.Release()

		// Obter a tabela dinâmica pelo nome
		pivotTable, err := oleutil.GetProperty(pivotTablesDisp, "Item", tableName)
		if err != nil {
			// Tentar pelo índice 1 (primeira tabela)
			pivotTable, err = oleutil.GetProperty(pivotTablesDisp, "Item", 1)
			if err != nil {
				return fmt.Errorf("tabela dinâmica '%s' não encontrada: %w", tableName, err)
			}
		}
		pivotTableDisp := pivotTable.ToIDispatch()
		defer pivotTableDisp.Release()

		// Configurar campos de linha
		for _, fieldName := range rowFields {
			pivotField, err := oleutil.CallMethod(pivotTableDisp, "PivotFields", fieldName)
			if err != nil {
				logger.ExcelDebug(fmt.Sprintf("[DEBUG] Campo '%s' não encontrado: %v", fieldName, err))
				continue
			}
			pivotFieldDisp := pivotField.ToIDispatch()

			// xlRowField = 1
			_, err = oleutil.PutProperty(pivotFieldDisp, "Orientation", 1)
			if err != nil {
				logger.ExcelDebug(fmt.Sprintf("[DEBUG] Erro ao definir campo linha '%s': %v", fieldName, err))
			} else {
				logger.ExcelDebug(fmt.Sprintf("[DEBUG] Campo '%s' adicionado às linhas", fieldName))
			}
			pivotFieldDisp.Release()
		}

		// Configurar campos de dados (valores)
		for _, dataField := range dataFields {
			fieldName := dataField["field"]
			function := dataField["function"]
			if fieldName == "" {
				continue
			}

			pivotField, err := oleutil.CallMethod(pivotTableDisp, "PivotFields", fieldName)
			if err != nil {
				logger.ExcelDebug(fmt.Sprintf("[DEBUG] Campo '%s' não encontrado: %v", fieldName, err))
				continue
			}
			pivotFieldDisp := pivotField.ToIDispatch()

			// xlDataField = 4
			_, err = oleutil.PutProperty(pivotFieldDisp, "Orientation", 4)
			if err != nil {
				logger.ExcelDebug(fmt.Sprintf("[DEBUG] Erro ao definir campo dados '%s': %v", fieldName, err))
				pivotFieldDisp.Release()
				continue
			}

			// Definir função de agregação
			// xlSum = -4157, xlCount = -4112, xlAverage = -4106
			var funcVal int
			switch strings.ToLower(function) {
			case "sum", "soma":
				funcVal = -4157
			case "count", "contar":
				funcVal = -4112
			case "average", "média", "media":
				funcVal = -4106
			case "max", "máximo", "maximo":
				funcVal = -4136
			case "min", "mínimo", "minimo":
				funcVal = -4139
			default:
				funcVal = -4157 // Padrão: soma
			}

			_, err = oleutil.PutProperty(pivotFieldDisp, "Function", funcVal)
			if err != nil {
				logger.ExcelDebug(fmt.Sprintf("[DEBUG] Erro ao definir função '%s' para '%s': %v", function, fieldName, err))
			} else {
				logger.ExcelDebug(fmt.Sprintf("[DEBUG] Campo '%s' adicionado aos valores com função '%s'", fieldName, function))
			}
			pivotFieldDisp.Release()
		}

		logger.ExcelDebug("[DEBUG] Campos da tabela dinâmica configurados!")
		return nil
	})
}

// ListPivotTables retorna lista de tabelas dinâmicas em uma aba
func (c *Client) ListPivotTables(workbookName, sheetName string) ([]string, error) {
	return runOnCOMThreadWithResult(c, func() ([]string, error) {
		workbooks, err := oleutil.GetProperty(c.excelApp, "Workbooks")
		if err != nil {
			return nil, err
		}
		workbooksDisp := workbooks.ToIDispatch()
		defer workbooksDisp.Release()

		wb, err := oleutil.GetProperty(workbooksDisp, "Item", workbookName)
		if err != nil {
			return nil, fmt.Errorf("workbook '%s' não encontrado", workbookName)
		}
		wbDisp := wb.ToIDispatch()
		defer wbDisp.Release()

		sheets, err := oleutil.GetProperty(wbDisp, "Worksheets")
		if err != nil {
			return nil, err
		}
		sheetsDisp := sheets.ToIDispatch()
		defer sheetsDisp.Release()

		sheet, err := oleutil.GetProperty(sheetsDisp, "Item", sheetName)
		if err != nil {
			return nil, fmt.Errorf("aba '%s' não encontrada", sheetName)
		}
		sheetDisp := sheet.ToIDispatch()
		defer sheetDisp.Release()

		pivotTables, err := oleutil.GetProperty(sheetDisp, "PivotTables")
		if err != nil {
			return []string{}, nil // Sem pivot tables
		}
		pivotTablesDisp := pivotTables.ToIDispatch()
		defer pivotTablesDisp.Release()

		countVar, _ := oleutil.GetProperty(pivotTablesDisp, "Count")
		count := int(countVar.Val)

		var names []string
		for i := 1; i <= count; i++ {
			pt, err := oleutil.GetProperty(pivotTablesDisp, "Item", i)
			if err != nil {
				continue
			}
			ptDisp := pt.ToIDispatch()
			nameVar, _ := oleutil.GetProperty(ptDisp, "Name")
			names = append(names, nameVar.ToString())
			ptDisp.Release()
		}

		return names, nil
	})
}
