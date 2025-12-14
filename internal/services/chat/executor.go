package chat

import (
	"encoding/json"
	"fmt"
)

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
	}

	return "", fmt.Errorf("unknown query type: %s", queryType)
}

func (s *Service) executeAction(params map[string]interface{}) (string, error) {
	op, _ := params["op"].(string)

	switch op {
	case "write":
		// {"op": "write", "cell": "A1", "value": "xyz"}
		// Precisa de current workbook context ou params
		// s.excelService.UpdateCell usa current context se params vazios

		cell, _ := params["cell"].(string)
		// val, _ := params["value"].(string) // Simplificação: assume string
		// Se valor for numérico no json, vem float64.
		// UpdateCell espera string no backend atual?
		// s.excelService.UpdateCell(wb, sheet, cell, value string)
		// Vou converter qualquer valor para string.

		valStr := fmt.Sprintf("%v", params["value"])

		// UpdateCell recebe (workbook, sheet, cell, value)
		// O backend usa current se vazio.
		err := s.excelService.UpdateCell("", "", cell, valStr)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("WRITE OK: %s = %s", cell, valStr), nil

	case "create-workbook":
		// ExcelService CreateNewWorkbook() returns (string, error)
		createdName, err := s.excelService.CreateNewWorkbook()
		if err != nil {
			return "", err
		}
		// O backend cria com nome padrão (Pasta1, Pasta2...)
		// Se o usuário passou "name", ignoramos por enquanto pois a func não aceita.
		return fmt.Sprintf("CREATE WORKBOOK OK: %s", createdName), nil

	case "create-sheet":
		name, _ := params["name"].(string)
		err := s.excelService.CreateNewSheet(name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("CREATE SHEET OK: %s", name), nil

	case "create-chart":
		sheet, _ := params["sheet"].(string) // Optional in service
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
		fontSize, _ := params["fontSize"].(float64)
		fontColor, _ := params["fontColor"].(string)
		bgColor, _ := params["bgColor"].(string)

		// sheet optional
		sheet, _ := params["sheet"].(string)

		err := s.excelService.FormatRange(sheet, rng, bold, italic, int(fontSize), fontColor, bgColor)
		if err != nil {
			return "", err
		}
		return "FORMAT RANGE OK", nil
	}

	return "", fmt.Errorf("unknown action op: %s", op)
}
