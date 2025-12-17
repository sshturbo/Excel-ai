package ai

import "encoding/json"

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

OPERAÇÕES: write_cell, write_range, create_sheet, delete_sheet, rename_sheet, format_range, autofit_columns, clear_range, insert_rows, delete_rows, merge_cells, set_borders, sort_range, apply_filter, create_chart, create_table

Exemplo: {actions: [{tool:"write_range", args:{sheet:"Plan1", cell:"A1", data:[["Nome","Idade"],["Ana",30]]}}]}`,
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
