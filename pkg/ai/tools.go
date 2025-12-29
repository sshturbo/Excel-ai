package ai

import (
	"encoding/json"
	"strings"
)

// =============================================================================
// Types for Function Calling (Compatible with OpenAI/Gemini APIs)
// =============================================================================

// Tool represents a tool that can be called by the AI
type Tool struct {
	Type     string              `json:"type"` // Always "function"
	Function FunctionDeclaration `json:"function"`
}

// FunctionDeclaration defines a function the AI can call
type FunctionDeclaration struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Parameters  FunctionParameters `json:"parameters"`
}

// FunctionParameters defines the parameters schema for a function
type FunctionParameters struct {
	Type       string                      `json:"type"` // Always "object"
	Properties map[string]FunctionProperty `json:"properties"`
	Required   []string                    `json:"required,omitempty"`
}

// FunctionProperty defines a single parameter
type FunctionProperty struct {
	Type        string            `json:"type"`
	Description string            `json:"description,omitempty"`
	Enum        []string          `json:"enum,omitempty"`
	Items       *FunctionProperty `json:"items,omitempty"` // For array types
}

// ToolCall represents a function call requested by the AI
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall contains the function name and arguments
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string of arguments
}

// ParsedFunctionCall contains parsed arguments as a map
type ParsedFunctionCall struct {
	Name      string
	Arguments map[string]interface{}
}

// ParseArguments parses the JSON arguments string into a map
func (tc *ToolCall) ParseArguments() (map[string]interface{}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return nil, err
	}
	return args, nil
}

// =============================================================================
// Excel Tool Definitions - CONSOLIDADO (apenas 6 ferramentas)
// =============================================================================

// GetExcelTools returns all available Excel tools for function calling
func GetExcelTools() []Tool {
	return []Tool{
		// =========================================================================
		// QUERY TOOLS
		// =========================================================================
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "list_sheets",
				Description: "Lista todas as planilhas da pasta de trabalho atual. Use primeiro para verificar conexão com Excel.",
				Parameters: FunctionParameters{
					Type:       "object",
					Properties: map[string]FunctionProperty{},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "query_batch",
				Description: "FERRAMENTA PRINCIPAL DE CONSULTA. Executa múltiplas consultas de uma vez. SEMPRE use esta ferramenta em vez de consultas individuais.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"queries": {
							Type:        "array",
							Description: "Lista de consultas: 'headers', 'row_count', 'used_range', 'sample_data', 'column_count', 'has_filter', 'charts', 'tables'",
							Items: &FunctionProperty{
								Type: "string",
								Enum: []string{"headers", "row_count", "used_range", "sample_data", "column_count", "has_filter", "charts", "tables"},
							},
						},
						"sample_rows": {
							Type:        "integer",
							Description: "Número de linhas de amostra (default: 5)",
						},
					},
					Required: []string{"sheet", "queries"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "get_range_values",
				Description: "Obtém valores de um intervalo específico. Use max_rows para limitar dados.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de células (ex: 'A1:C10')",
						},
						"max_rows": {
							Type:        "integer",
							Description: "Limite máximo de linhas",
						},
						"filter_column": {
							Type:        "string",
							Description: "Coluna para filtrar",
						},
						"filter_value": {
							Type:        "string",
							Description: "Valor do filtro",
						},
					},
					Required: []string{"sheet", "range"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "get_cell_formula",
				Description: "Obtém a fórmula de uma célula específica.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"cell": {
							Type:        "string",
							Description: "Endereço da célula (ex: 'A1')",
						},
					},
					Required: []string{"sheet", "cell"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "get_active_cell",
				Description: "Obtém endereço da célula selecionada.",
				Parameters: FunctionParameters{
					Type:       "object",
					Properties: map[string]FunctionProperty{},
				},
			},
		},

		// =========================================================================
		// ACTION TOOLS - Consolidado em execute_macro
		// =========================================================================
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name: "execute_macro",
				Description: `FERRAMENTA ÚNICA DE AÇÕES. Executa uma ou mais operações no Excel.

OPERAÇÕES DISPONÍVEIS:
• create_sheet: {name} - Cria aba já com o nome final (NÃO precisa rename depois!)
• write_cell: {sheet, cell, value}
• write_range: {sheet, cell, data} - data é array 2D: [["col1","col2"],["v1","v2"]]
• delete_sheet: {name}
• rename_sheet: {old_name, new_name} - APENAS para abas existentes, NÃO após create_sheet!
• format_range: {sheet, range, bold, italic, font_size, font_color, bg_color}
• autofit_columns: {sheet, range}
• clear_range: {sheet, range}
• insert_rows/delete_rows: {sheet, row, count}
• merge_cells: {sheet, range}
• set_borders: {sheet, range, style}
• sort_range: {sheet, range, column, ascending}
• apply_filter: {sheet, range}
• create_chart: {sheet, range, chart_type, title}

Exemplo criar aba e escrever: {actions: [{tool:"create_sheet", args:{name:"MinhaAba"}}, {tool:"write_range", args:{sheet:"MinhaAba", cell:"A1", data:[["Nome","Idade"]]}}]}`,
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"actions": {
							Type:        "array",
							Description: "Lista de ações. Cada ação: {tool, args}",
							Items: &FunctionProperty{
								Type:        "object",
								Description: "Uma ação com 'tool' e 'args'",
							},
						},
					},
					Required: []string{"actions"},
				},
			},
		},
	}
}

// GetGeminiTools converts Tools to Gemini-specific format
func GetGeminiTools() []map[string]interface{} {
	tools := GetExcelTools()

	// Gemini uses a single object with functionDeclarations array
	functionDeclarations := make([]map[string]interface{}, len(tools))

	for i, tool := range tools {
		functionDeclarations[i] = map[string]interface{}{
			"name":        tool.Function.Name,
			"description": tool.Function.Description,
			"parameters":  tool.Function.Parameters,
		}
	}

	return []map[string]interface{}{
		{
			"functionDeclarations": functionDeclarations,
		},
	}
}

// IsQueryTool returns true if the tool is a read-only query
func IsQueryTool(name string) bool {
	queryTools := map[string]bool{
		"list_sheets":      true,
		"get_range_values": true,
		"get_cell_formula": true,
		"get_active_cell":  true,
		"query_batch":      true,
	}
	return queryTools[name]
}

// IsActionTool returns true if the tool modifies Excel
func IsActionTool(name string) bool {
	return !IsQueryTool(name)
}

// FilterValidToolCalls filtra tool calls inválidos (nome vazio, null, etc)
func FilterValidToolCalls(toolCalls []ToolCall) []ToolCall {
	validTools := map[string]bool{
		"list_sheets":      true,
		"query_batch":      true,
		"get_range_values": true,
		"get_cell_formula": true,
		"get_active_cell":  true,
		"execute_macro":    true,
		"write_cell":       true,
		"write_range":      true,
		"create_sheet":     true,
		"delete_sheet":     true,
		"rename_sheet":     true,
		"format_range":     true,
		"clear_range":      true,
		"insert_rows":      true,
		"delete_rows":      true,
		"merge_cells":      true,
		"set_borders":      true,
		"sort_range":       true,
		"apply_filter":     true,
		"create_chart":     true,
		"autofit_columns":  true,
	}

	var valid []ToolCall
	for _, tc := range toolCalls {
		// Verificar se o nome da função é válido e não vazio
		if tc.Function.Name != "" && validTools[tc.Function.Name] {
			valid = append(valid, tc)
		}
	}
	return valid
}

// IsPartialToolCallJSON detecta se o texto contém um JSON parcial de tool call
// Retorna true se parece que estamos no meio de um JSON que pode ser uma tool call
func IsPartialToolCallJSON(text string) bool {
	// Verificar se há indicadores de tool call no texto
	toolIndicators := []string{
		`"name":`,
		`"name": null`,
		`"arguments":`,
		`"arguments": {}`,
		`"tool":`,
		`"args":`,
		`"function":`,
		`"function":{}`,
		`"type":"function"`,
		`"type": "function"`,
		`"parameters":`,
		`"parameters":{}`,
		"```json",
		"list_sheets",
		"query_batch",
		"get_range_values",
		"execute_macro",
		"write_cell",
		"write_range",
	}

	hasIndicator := false
	for _, indicator := range toolIndicators {
		if containsString(text, indicator) {
			hasIndicator = true
			break
		}
	}

	if !hasIndicator {
		return false
	}

	// Contar chaves abertas e fechadas
	openBraces := 0
	closeBraces := 0
	inString := false
	escaped := false

	for _, c := range text {
		if escaped {
			escaped = false
			continue
		}
		switch {
		case c == '\\' && inString:
			escaped = true
		case c == '"':
			inString = !inString
		case inString:
			// skip characters inside strings
		case c == '{':
			openBraces++
		case c == '}':
			closeBraces++
		}
	}

	// Se temos chaves abertas, verificar se é um JSON de tool call completo ou parcial
	if openBraces > 0 {
		// Se mais abertas que fechadas, estamos no meio
		if openBraces > closeBraces {
			return true
		}
		// Se igual, o JSON está completo - ainda retornar true para limpar depois
		if openBraces == closeBraces {
			return true
		}
	}

	return false
}

// containsString verifica se uma string contém outra (case-sensitive)
func containsString(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ParseToolCallsFromText extrai tool calls de texto quando a IA retorna JSON no conteúdo
// Isso é necessário para modelos como Ollama que podem não usar o formato tool_calls estruturado
func ParseToolCallsFromText(text string) ([]ToolCall, string) {
	var toolCalls []ToolCall
	cleanedText := text

	// Padrões comuns de tool calls no texto
	// Padrão 1: {"name": "tool_name", "arguments": {...}}
	// Padrão 2: ```json\n{"name": "tool_name", ...}\n```
	// Padrão 3: {"tool": "tool_name", "args": {...}} (formato execute_macro)

	// Lista de tools válidas para detectar
	validTools := map[string]bool{
		"list_sheets":      true,
		"query_batch":      true,
		"get_range_values": true,
		"get_cell_formula": true,
		"get_active_cell":  true,
		"execute_macro":    true,
		"write_cell":       true,
		"write_range":      true,
		"create_sheet":     true,
		"delete_sheet":     true,
		"rename_sheet":     true,
		"format_range":     true,
		"clear_range":      true,
		"insert_rows":      true,
		"delete_rows":      true,
		"merge_cells":      true,
		"set_borders":      true,
		"sort_range":       true,
		"apply_filter":     true,
		"create_chart":     true,
		"autofit_columns":  true,
	}

	// Tentar extrair JSON do texto
	jsonPatterns := []struct {
		start string
		end   string
	}{
		{"```json\n", "\n```"},
		{"```\n", "\n```"},
		{"{", "}"},
	}

	for _, pattern := range jsonPatterns {
		startIdx := 0
		for {
			idx := findJSONStart(cleanedText[startIdx:], pattern.start)
			if idx == -1 {
				break
			}
			idx += startIdx

			// Encontrar o JSON completo (balanceando chaves)
			jsonStr, endIdx := extractBalancedJSON(cleanedText[idx:])
			if jsonStr == "" {
				startIdx = idx + 1
				continue
			}

			// Verificar se é um JSON de tool call (válido ou inválido)
			isToolCallJSON, isValid := checkToolCallJSON(jsonStr, validTools)

			if isToolCallJSON {
				if isValid {
					// Tool call válido - parsear e executar
					tc, ok := parseAsToolCall(jsonStr, validTools)
					if ok {
						toolCalls = append(toolCalls, tc)
					}
				}
				// Remover o JSON do texto (válido ou inválido)
				cleanedText = cleanedText[:idx] + cleanedText[idx+endIdx:]
				// Não avançar startIdx pois removemos texto
				continue
			}

			startIdx = idx + 1
		}
	}

	// Limpar texto de artefatos
	cleanedText = cleanTextArtifacts(cleanedText)

	return toolCalls, cleanedText
}

// checkToolCallJSON verifica se um JSON parece ser uma tentativa de tool call
// Retorna (isToolCallJSON, isValid)
func checkToolCallJSON(jsonStr string, validTools map[string]bool) (bool, bool) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return false, false
	}

	// Verificar se tem estrutura de tool call
	hasName := false
	hasArguments := false
	hasTool := false
	hasArgs := false
	hasType := false
	hasFunction := false
	hasParameters := false

	if _, ok := data["name"]; ok {
		hasName = true
	}
	if _, ok := data["arguments"]; ok {
		hasArguments = true
	}
	if _, ok := data["tool"]; ok {
		hasTool = true
	}
	if _, ok := data["args"]; ok {
		hasArgs = true
	}
	if typeVal, ok := data["type"]; ok {
		hasType = true
		// Se é type:function, é uma definição de tool malformada
		if typeVal == "function" {
			return true, false
		}
	}
	if _, ok := data["function"]; ok {
		hasFunction = true
	}
	if _, ok := data["parameters"]; ok {
		hasParameters = true
	}

	// Se não tem nenhuma das chaves de tool call, não é um tool call JSON
	if !hasName && !hasArguments && !hasTool && !hasArgs && !hasType && !hasFunction && !hasParameters {
		return false, false
	}

	// Se tem type:function ou function:{} ou parameters:{}, é definição malformada
	if hasType || hasFunction || hasParameters {
		return true, false
	}

	// É um JSON de tool call, verificar se é válido
	// Formato 1: {"name": "tool_name", "arguments": {...}}
	if name, ok := data["name"].(string); ok && name != "" {
		if validTools[name] {
			return true, true
		}
	}

	// Formato 2: {"tool": "tool_name", "args": {...}}
	if tool, ok := data["tool"].(string); ok && tool != "" {
		if validTools[tool] {
			return true, true
		}
	}

	// É um JSON de tool call mas inválido (name: null, etc)
	return true, false
}

// findJSONStart encontra o início de um bloco JSON
func findJSONStart(text, pattern string) int {
	idx := -1
	for i := 0; i <= len(text)-len(pattern); i++ {
		if text[i:i+len(pattern)] == pattern {
			idx = i
			if pattern != "{" {
				idx += len(pattern)
			}
			break
		}
	}
	return idx
}

// extractBalancedJSON extrai um JSON balanceado começando de uma abertura {
func extractBalancedJSON(text string) (string, int) {
	if len(text) == 0 || text[0] != '{' {
		// Procurar primeiro {
		idx := -1
		for i, c := range text {
			if c == '{' {
				idx = i
				break
			}
		}
		if idx == -1 {
			return "", 0
		}
		text = text[idx:]
	}

	depth := 0
	inString := false
	escaped := false

	for i, c := range text {
		if escaped {
			escaped = false
			continue
		}

		if c == '\\' && inString {
			escaped = true
			continue
		}

		switch {
		case c == '"':
			inString = !inString
		case inString:
			// skip characters inside strings
		case c == '{':
			depth++
		case c == '}':
			depth--
			if depth == 0 {
				return text[:i+1], i + 1
			}
		}
	}

	return "", 0
}

// parseAsToolCall tenta parsear JSON como uma tool call
func parseAsToolCall(jsonStr string, validTools map[string]bool) (ToolCall, bool) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ToolCall{}, false
	}

	var toolName string
	var arguments map[string]interface{}

	// Formato 1: {"name": "tool_name", "arguments": {...}}
	if name, ok := data["name"].(string); ok {
		toolName = name
		if args, ok := data["arguments"].(map[string]interface{}); ok {
			arguments = args
		} else if args, ok := data["arguments"]; ok {
			// Arguments pode ser outro tipo
			arguments = map[string]interface{}{"value": args}
		}
	}

	// Formato 2: {"tool": "tool_name", "args": {...}} (para ações individuais)
	if toolName == "" {
		if tool, ok := data["tool"].(string); ok {
			toolName = tool
			if args, ok := data["args"].(map[string]interface{}); ok {
				arguments = args
			}
		}
	}

	// Verificar se é uma tool válida
	if toolName == "" || !validTools[toolName] {
		return ToolCall{}, false
	}

	// Converter arguments de volta para JSON string
	argsJSON, err := json.Marshal(arguments)
	if err != nil {
		argsJSON = []byte("{}")
	}

	return ToolCall{
		ID:   generateToolCallID(),
		Type: "function",
		Function: FunctionCall{
			Name:      toolName,
			Arguments: string(argsJSON),
		},
	}, true
}

// generateToolCallID gera um ID único para tool calls
func generateToolCallID() string {
	return "call_" + randomString(8)
}

// randomString gera uma string aleatória
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// cleanTextArtifacts remove artefatos de texto deixados após extração de tool calls
func cleanTextArtifacts(text string) string {
	// Padrões comuns que modelos Ollama escrevem ao tentar chamar tools incorretamente
	junkPatterns := []string{
		`{"name": null, "arguments": {}}`,
		`{"name":null,"arguments":{}}`,
		`{"name": "", "arguments": {}}`,
		`{"name":"","arguments":{}}`,
		`{"type":"function","function":{}}`,
		`{"type": "function", "function": {}}`,
		`{"type":"function","function":{},"parameters":{}}`,
		`{"parameters": {}}`,
		`{"parameters":{}}`,
		`"parameters": {}`,
		`; {"name":`,
		`}, {"name":`,
		`}; {"name":`,
	}

	for _, pattern := range junkPatterns {
		text = replacePattern(text, pattern, "")
	}

	// Remover blocos de código vazios
	for {
		newText := text
		// Remover ```json\n``` vazios
		newText = replacePattern(newText, "```json\n```", "")
		newText = replacePattern(newText, "```\n```", "")
		newText = replacePattern(newText, "```json```", "")
		newText = replacePattern(newText, "``````", "")
		// Remover múltiplas quebras de linha
		for containsDouble(newText, "\n\n\n") {
			newText = replacePattern(newText, "\n\n\n", "\n\n")
		}
		// Remover pontuação solta
		newText = replacePattern(newText, "; ; ", " ")
		newText = replacePattern(newText, ";;", "")
		newText = replacePattern(newText, "; ;", "")

		if newText == text {
			break
		}
		text = newText
	}

	// Remover espaços extras no início e fim
	return strings.TrimSpace(text)
}

func replacePattern(s, old, new string) string {
	result := ""
	for {
		idx := -1
		for i := 0; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

func containsDouble(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
