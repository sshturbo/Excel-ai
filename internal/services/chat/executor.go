package chat

import (
	"encoding/json"
	"fmt"
)

// executeToolCall executa uma tool call da Z.ai diretamente
func (s *Service) executeToolCall(toolName string, args map[string]interface{}) (string, error) {
	// Mapear ferramentas Z.ai para o sistema interno
	// Consultas (queries) não precisam de confirmação
	queryTools := map[string]bool{
		"list_sheets":      true,
		"query_batch":      true,
		"get_range_values": true,
		"get_cell_formula": true,
		"get_active_cell":  true,
	}

	var cmd ToolCommand
	if queryTools[toolName] {
		cmd.Type = ToolTypeQuery
	} else {
		cmd.Type = ToolTypeAction
	}

	// Converter args para ToolCommand format
	// Para query_batch: extrair queries array
	if toolName == "query_batch" {
		queries, ok := args["queries"].([]interface{})
		if ok && len(queries) > 0 {
			// Criar comandos individuais para cada query
			results := make([]string, 0, len(queries))
			sheet, _ := args["sheet"].(string)
			
			for _, q := range queries {
				if queryStr, ok := q.(string); ok {
					queryCmd := ToolCommand{
						Type: ToolTypeQuery,
						Payload: map[string]interface{}{
							"type":  convertQueryType(queryStr),
							"sheet": sheet,
						},
					}
					result, err := s.ExecuteTool(queryCmd)
					if err != nil {
						results = append(results, fmt.Sprintf("ERROR: %v", err))
					} else {
						results = append(results, result)
					}
				}
			}
			return fmt.Sprintf("QUERY_BATCH (%d queries):\n%s", len(results), joinResults(results)), nil
		}
	}

	// Para outras ferramentas, converter args para formato interno
	cmd.Payload = convertToolArguments(toolName, args)

	return s.ExecuteTool(cmd)
}

// convertQueryType converte query string para tipo interno
func convertQueryType(query string) string {
	switch query {
	case "headers":
		return "get-headers"
	case "row_count":
		return "get-row-count"
	case "used_range":
		return "get-used-range"
	case "sample_data":
		return "get-range-values"
	case "column_count":
		return "get-column-count"
	case "has_filter":
		return "has-filter"
	case "charts":
		return "list-charts"
	case "tables":
		return "list-tables"
	default:
		return "get-range-values"
	}
}

// convertToolArguments converte args da Z.ai para formato interno
func convertToolArguments(toolName string, args map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Para get_range_values
	if toolName == "get_range_values" {
		result["type"] = "get-range-values"
		if sheet, ok := args["sheet"].(string); ok {
			result["sheet"] = sheet
		}
		if rng, ok := args["range"].(string); ok {
			result["range"] = rng
		}
		return result
	}

	// Para get_cell_formula
	if toolName == "get_cell_formula" {
		result["type"] = "get-cell-formula"
		if sheet, ok := args["sheet"].(string); ok {
			result["sheet"] = sheet
		}
		if cell, ok := args["cell"].(string); ok {
			result["cell"] = cell
		}
		return result
	}

	// Para list_sheets
	if toolName == "list_sheets" {
		result["type"] = "list-sheets"
		return result
	}

	// Para get_active_cell
	if toolName == "get_active_cell" {
		result["type"] = "get-active-cell"
		return result
	}

	// Para execute_macro - converter para macro
	if toolName == "execute_macro" {
		result["op"] = "macro"
		if actions, ok := args["actions"].([]interface{}); ok {
			result["actions"] = actions
		}
		return result
	}

	return result
}

// ExecuteTool executa o comando extraído da IA
func (s *Service) ExecuteTool(cmd ToolCommand) (string, error) {
	if s.excelService == nil {
		return "", fmt.Errorf("excel service not connected")
	}

	payloadMap, ok := cmd.Payload.(map[string]interface{})
	if !ok {
		// Tentar converter de struct se falhou antes?
		// Assumindo que o parser entregou map[string]interface{} vindo do json.Decode
		// Se não, marshal/unmarshal hack
		b, _ := json.Marshal(cmd.Payload)
		if err := json.Unmarshal(b, &payloadMap); err != nil {
			return "", fmt.Errorf("invalid payload format: %v", err)
		}
	}

	switch cmd.Type {
	case ToolTypeQuery:
		return s.executeQuery(payloadMap)
	case ToolTypeAction:
		return s.executeAction(payloadMap)
	default:
		return "", fmt.Errorf("unknown tool type: %s", cmd.Type)
	}
}

func (s *Service) executeQuery(params map[string]interface{}) (string, error) {
	queryType, _ := params["type"].(string)

	switch queryType {
	case "list-sheets":
		sheets, err := s.excelService.ListSheets()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("SHEETS: %v", sheets), nil

	case "sheet-exists":
		name, _ := params["name"].(string)
		exists, err := s.excelService.SheetExists(name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("EXISTS (%s): %v", name, exists), nil

	case "get-used-range":
		sheet, _ := params["sheet"].(string)
		rng, err := s.excelService.GetUsedRange(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("USED RANGE (%s): %s", sheet, rng), nil

	case "get-headers":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		headers, err := s.excelService.GetHeaders(sheet, rng)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("HEADERS (%s!%s): %v", sheet, rng, headers), nil

	case "get-cell-formula":
		sheet, _ := params["sheet"].(string)
		cell, _ := params["cell"].(string)
		formula, err := s.excelService.GetCellFormula(sheet, cell)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("FORMULA (%s!%s): %s", sheet, cell, formula), nil

	case "get-active-cell":
		cell, err := s.excelService.GetActiveCell()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("ACTIVE CELL: %s", cell), nil

	case "list-tables":
		sheet, _ := params["sheet"].(string)
		tables, err := s.excelService.ListTables(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("TABLES: %v", tables), nil

	case "list-charts":
		sheet, _ := params["sheet"].(string)
		charts, err := s.excelService.ListCharts(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("CHARTS: %v", charts), nil

	case "list-pivot-tables":
		sheet, _ := params["sheet"].(string)
		pivots, err := s.excelService.ListPivotTables(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("PIVOTS: %v", pivots), nil

	case "get-range-values":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		values, err := s.excelService.GetRangeValues(sheet, rng)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("RANGE VALUES (%s!%s): %v", sheet, rng, values), nil

	case "has-filter":
		sheet, _ := params["sheet"].(string)
		hasFilter, err := s.excelService.HasFilter(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("HAS FILTER (%s): %v", sheet, hasFilter), nil

	case "get-row-count":
		sheet, _ := params["sheet"].(string)
		count, err := s.excelService.GetRowCount(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("ROW COUNT (%s): %d", sheet, count), nil

	case "get-column-count":
		sheet, _ := params["sheet"].(string)
		count, err := s.excelService.GetColumnCount(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("COLUMN COUNT (%s): %d", sheet, count), nil
	}

	return "", fmt.Errorf("unknown query type: %s", queryType)
}

// Helper para extrair int de interface{} (suporta float64 e int)
func getInt(v interface{}) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	if i, ok := v.(int); ok {
		return i
	}
	return 0
}

// Helper para extrair float64 de interface{}
func getFloat(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	if i, ok := v.(int); ok {
		return float64(i)
	}
	return 0.0
}

// Helper para juntar resultados de macro
func joinResults(results []string) string {
	result := ""
	for _, r := range results {
		result += "  - " + r + "\n"
	}
	return result
}

func (s *Service) executeAction(params map[string]interface{}) (string, error) {
	op, _ := params["op"].(string)

	switch op {
	case "macro":
		// MACRO: Executa múltiplas ações em sequência
		// {"op": "macro", "actions": [{...}, {...}, {...}]}
		actions, ok := params["actions"].([]interface{})
		if !ok {
			return "", fmt.Errorf("macro requires 'actions' array")
		}

		// DEBUG: Log macro execution
		fmt.Printf("[DEBUG] 🎯 MACRO detected with %d actions:\n", len(actions))
		for i, action := range actions {
			if actionMap, ok := action.(map[string]interface{}); ok {
				fmt.Printf("[DEBUG]   Action %d: op=%s\n", i+1, actionMap["op"])
			}
		}

		// Start undo batch so all actions can be undone together
		s.excelService.StartUndoBatch()

		var results []string
		for i, action := range actions {
			actionMap, ok := action.(map[string]interface{})
			if !ok {
				results = append(results, fmt.Sprintf("Action %d: SKIP (invalid format)", i+1))
				continue
			}

			result, err := s.executeAction(actionMap)
			if err != nil {
				results = append(results, fmt.Sprintf("Action %d (%s): ERROR - %v", i+1, actionMap["op"], err))
				// Stop execution on error to prevent cascading failures
				break
			} else {
				results = append(results, fmt.Sprintf("Action %d (%s): %s", i+1, actionMap["op"], result))
			}
		}

		// End undo batch
		s.excelService.EndUndoBatch()

		fmt.Printf("[DEBUG] ✅ MACRO completed: %d actions executed\n", len(actions))
		return fmt.Sprintf("MACRO OK (%d actions):\n%s", len(actions), joinResults(results)), nil

	case "write":
		// Suporta dois formatos:
		// 1. Célula única: {"op": "write", "cell": "A1", "value": "xyz"}
		// 2. Lote (batch): {"op": "write", "cell": "A1", "data": [["a","b"],["c","d"]]}

		cell, _ := params["cell"].(string)
		sheet, _ := params["sheet"].(string)

		// Verificar se é batch (data array) ou single (value)
		if data, ok := params["data"].([]interface{}); ok {
			// Converter para [][]interface{}
			batchData := make([][]interface{}, len(data))
			for i, row := range data {
				if rowArr, ok := row.([]interface{}); ok {
					batchData[i] = rowArr
				} else {
					// Se não for array, criar array de 1 elemento
					batchData[i] = []interface{}{row}
				}
			}

			// DEBUG: Ver formato dos dados
			fmt.Printf("[DEBUG] Batch write: cell=%s, rows=%d\n", cell, len(batchData))
			for i, row := range batchData {
				fmt.Printf("[DEBUG]   Row %d: %d cols, values: %v\n", i, len(row), row)
			}

			err := s.excelService.WriteRange(sheet, cell, batchData)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("WRITE BATCH OK: %d linhas a partir de %s", len(batchData), cell), nil
		}

		// Modo single cell

		// Verificar se existe parâmetro "formula" explicito
		if formula, ok := params["formula"].(string); ok && formula != "" {
			params["value"] = formula
		}

		var valStr string
		if params["value"] == nil {
			valStr = ""
		} else {
			valStr = fmt.Sprintf("%v", params["value"])
		}

		err := s.excelService.UpdateCell("", sheet, cell, valStr)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("WRITE OK: %s = %s", cell, valStr), nil

	case "create-workbook":
		createdName, err := s.excelService.CreateNewWorkbook()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("CREATE WORKBOOK OK: %s", createdName), nil

	case "create-sheet":
		name, _ := params["name"].(string)
		err := s.excelService.CreateNewSheet(name)
		if err != nil {
			return "", err
		}
		// Undo: create-sheet -> só precisa saber o nome para deletar
		wbName, _ := s.excelService.GetActiveWorkbookName() // Tenta obter nome do workbook atual (best effort)
		undoData, _ := json.Marshal(map[string]string{"sheetName": name})
		s.excelService.SaveUndoAction("create-sheet", wbName, name, "", "", string(undoData))
		return fmt.Sprintf("CREATE SHEET OK: %s", name), nil

	case "create-chart":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		cType, _ := params["chartType"].(string)
		title, _ := params["title"].(string)

		err := s.excelService.CreateChart(sheet, rng, cType, title)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("CREATE CHART OK: %s", title), nil

	case "create-pivot":
		srcSheet, _ := params["sourceSheet"].(string)
		srcRange, _ := params["sourceRange"].(string)
		destSheet, _ := params["destSheet"].(string)
		destCell, _ := params["destCell"].(string)
		tableName, _ := params["tableName"].(string)

		err := s.excelService.CreatePivotTable(srcSheet, srcRange, destSheet, destCell, tableName)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("CREATE PIVOT OK: %s", tableName), nil

	case "format-range":
		rng, _ := params["range"].(string)
		bold, _ := params["bold"].(bool)
		italic, _ := params["italic"].(bool)
		fontSize := getFloat(params["fontSize"])
		fontColor, _ := params["fontColor"].(string)
		bgColor, _ := params["bgColor"].(string)
		sheet, _ := params["sheet"].(string)

		err := s.excelService.FormatRange(sheet, rng, bold, italic, int(fontSize), fontColor, bgColor)
		if err != nil {
			return "", err
		}
		// TODO: Implement undo for format-range (complex)
		return "FORMAT RANGE OK", nil

	case "delete-sheet":
		name, _ := params["name"].(string)
		err := s.excelService.DeleteSheet(name)
		if err != nil {
			return "", err
		}
		// Undo: delete-sheet é irreversível sem snapshot completo
		return fmt.Sprintf("DELETE SHEET OK: %s", name), nil

	case "rename-sheet":
		oldName, _ := params["oldName"].(string)
		newName, _ := params["newName"].(string)
		err := s.excelService.RenameSheet(oldName, newName)
		if err != nil {
			return "", err
		}
		// Undo: rename-sheet -> renomear de volta
		undoData, _ := json.Marshal(map[string]string{"oldName": oldName, "newName": newName})
		s.excelService.SaveUndoAction("rename-sheet", "", newName, "", "", string(undoData))
		return fmt.Sprintf("RENAME SHEET OK: %s -> %s", oldName, newName), nil

	case "clear-range":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)

		// Save data before clearing
		oldData, _ := s.excelService.GetRangeValues(sheet, rng)

		err := s.excelService.ClearRange(sheet, rng)
		if err != nil {
			return "", err
		}

		// Undo: clear -> restore data
		if len(oldData) > 0 {
			undoMap := map[string]interface{}{
				"data": oldData,
			}
			undoData, _ := json.Marshal(undoMap)
			s.excelService.SaveUndoAction("clear-range", "", sheet, rng, "", string(undoData))
		}
		return "CLEAR RANGE OK", nil

	case "autofit":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		err := s.excelService.AutoFitColumns(sheet, rng)
		if err != nil {
			return "", err
		}
		return "AUTOFIT OK", nil

	case "insert-rows":
		sheet, _ := params["sheet"].(string)
		row := getInt(params["row"])
		count := getInt(params["count"])
		err := s.excelService.InsertRows(sheet, row, count)
		if err != nil {
			return "", err
		}
		// Undo: insert-rows -> delete-rows
		undoData, _ := json.Marshal(map[string]int{"row": row, "count": count})
		s.excelService.SaveUndoAction("insert-rows", "", sheet, "", "", string(undoData))
		return fmt.Sprintf("INSERT ROWS OK: %d at %d", count, row), nil

	case "delete-rows":
		sheet, _ := params["sheet"].(string)
		row := getInt(params["row"])
		count := getInt(params["count"])

		// Antes de deletar, salvar os dados!
		// Ler range das linhas afetadas (ex: "5:7")
		rangeAddr := fmt.Sprintf("%d:%d", row, row+count-1)

		// Tentar ler dados. GetRangeValues pode suportar "1:1" dependendo da impl do client
		// Se não suportar, fallback: ler UsedRange intersectado?
		// Assumindo suporte básico ou que o usuário vai recuperar via un-delete vazio + re-escrita

		// Por segurança vamos ler apenas UsedRows se possível, mas GetRangeValues é genérico.
		// Vamos tentar ler. Se falhar, salvamos sem dados (pelo menos restaura as linhas vazias).
		rowsData, _ := s.excelService.GetRangeValues(sheet, rangeAddr)

		err := s.excelService.DeleteRows(sheet, row, count)
		if err != nil {
			return "", err
		}

		// Undo: delete-rows -> insert-rows + write data
		undoMap := map[string]interface{}{
			"row":   row,
			"count": count,
			"data":  rowsData, // [][]string
		}
		undoData, _ := json.Marshal(undoMap)
		s.excelService.SaveUndoAction("delete-rows", "", sheet, "", "", string(undoData))

		return fmt.Sprintf("DELETE ROWS OK: %d at %d", count, row), nil

	case "merge-cells":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		err := s.excelService.MergeCells(sheet, rng)
		if err != nil {
			return "", err
		}
		// Undo: merge -> unmerge
		s.excelService.SaveUndoAction("merge-cells", "", sheet, rng, "", "")
		return "MERGE OK", nil

	case "unmerge-cells":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		err := s.excelService.UnmergeCells(sheet, rng)
		if err != nil {
			return "", err
		}
		// Undo: unmerge -> merge
		s.excelService.SaveUndoAction("unmerge-cells", "", sheet, rng, "", "")
		return "UNMERGE OK", nil

	case "set-borders":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		style, _ := params["style"].(string)
		err := s.excelService.SetBorders(sheet, rng, style)
		if err != nil {
			return "", err
		}
		return "BORDERS OK", nil

	case "set-column-width":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		width := getFloat(params["width"])

		oldWidth, _ := s.excelService.GetColumnWidth(sheet, rng)

		err := s.excelService.SetColumnWidth(sheet, rng, width)
		if err != nil {
			return "", err
		}

		// Undo: restore width
		if oldWidth > 0 {
			undoMap := map[string]float64{
				"width": oldWidth,
			}
			undoData, _ := json.Marshal(undoMap)
			s.excelService.SaveUndoAction("set-column-width", "", sheet, rng, "", string(undoData))
		}
		return "COL WIDTH OK", nil

	case "set-row-height":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		height := getFloat(params["height"])

		oldHeight, _ := s.excelService.GetRowHeight(sheet, rng)

		err := s.excelService.SetRowHeight(sheet, rng, height)
		if err != nil {
			return "", err
		}

		// Undo: restore height
		if oldHeight > 0 {
			undoMap := map[string]float64{
				"height": oldHeight,
			}
			undoData, _ := json.Marshal(undoMap)
			s.excelService.SaveUndoAction("set-row-height", "", sheet, rng, "", string(undoData))
		}
		return "ROW HEIGHT OK", nil

	case "apply-filter":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		err := s.excelService.ApplyFilter(sheet, rng)
		if err != nil {
			return "", err
		}
		return "FILTER APPLIED OK", nil

	case "clear-filters":
		sheet, _ := params["sheet"].(string)
		err := s.excelService.ClearFilters(sheet)
		if err != nil {
			return "", err
		}
		return "FILTERS CLEARED OK", nil

	case "sort":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		col := getInt(params["column"])
		asc, _ := params["ascending"].(bool)

		// Sort é complexo. Undo data = save range data before sort.
		oldData, _ := s.excelService.GetRangeValues(sheet, rng)

		err := s.excelService.SortRange(sheet, rng, col, asc)
		if err != nil {
			return "", err
		}

		// Undo: restore original data order
		if len(oldData) > 0 {
			undoMap := map[string]interface{}{
				"data": oldData,
			}
			undoData, _ := json.Marshal(undoMap)
			s.excelService.SaveUndoAction("sort", "", sheet, rng, "", string(undoData))
		}
		return "SORT OK", nil

	case "copy-range":
		sheet, _ := params["sheet"].(string)
		src, _ := params["source"].(string)
		dest, _ := params["dest"].(string)
		err := s.excelService.CopyRange(sheet, src, dest)
		if err != nil {
			return "", err
		}
		return "COPY OK", nil

	case "list-charts":
		sheet, _ := params["sheet"].(string)
		charts, err := s.excelService.ListCharts(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("CHARTS: %v", charts), nil

	case "delete-chart":
		sheet, _ := params["sheet"].(string)
		name, _ := params["name"].(string)
		err := s.excelService.DeleteChart(sheet, name)
		if err != nil {
			return "", err
		}
		return "DELETE CHART OK", nil

	case "create-table":
		sheet, _ := params["sheet"].(string)
		rng, _ := params["range"].(string)
		name, _ := params["name"].(string)
		style, _ := params["style"].(string)
		err := s.excelService.CreateTable(sheet, rng, name, style)
		if err != nil {
			return "", err
		}
		return "CREATE TABLE OK", nil

	case "list-tables":
		sheet, _ := params["sheet"].(string)
		tables, err := s.excelService.ListTables(sheet)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("TABLES: %v", tables), nil

	case "delete-table":
		sheet, _ := params["sheet"].(string)
		name, _ := params["name"].(string)
		err := s.excelService.DeleteTable(sheet, name)
		if err != nil {
			return "", err
		}
		return "DELETE TABLE OK", nil
	}

	return "", fmt.Errorf("unknown action op: %s", op)
}
