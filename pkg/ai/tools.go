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
// Excel Tool Definitions
// =============================================================================

// GetExcelTools returns all available Excel tools for function calling
func GetExcelTools() []Tool {
	return []Tool{
		// =========================================================================
		// QUERY TOOLS (Read-only operations)
		// =========================================================================
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "list_sheets",
				Description: "Lista todas as planilhas da pasta de trabalho atual do Excel. Use esta função primeiro para verificar se o Excel está conectado e quais planilhas existem.",
				Parameters: FunctionParameters{
					Type:       "object",
					Properties: map[string]FunctionProperty{},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "sheet_exists",
				Description: "Verifica se uma planilha específica existe na pasta de trabalho.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"name": {
							Type:        "string",
							Description: "Nome da planilha a verificar",
						},
					},
					Required: []string{"name"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "get_used_range",
				Description: "Obtém o intervalo de células utilizadas em uma planilha (ex: 'A1:D100').",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
					},
					Required: []string{"sheet"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "get_headers",
				Description: "Obtém os cabeçalhos (primeira linha) de um intervalo de colunas.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de colunas (ex: 'A:F' ou 'A1:F1')",
						},
					},
					Required: []string{"sheet", "range"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "get_range_values",
				Description: "Obtém os valores de um intervalo de células. Use max_rows para limitar a quantidade de dados retornados e evitar excesso de tokens.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de células (ex: 'A1:C10' ou 'A:C' para todas as linhas)",
						},
						"max_rows": {
							Type:        "integer",
							Description: "Limite máximo de linhas a retornar (ex: 10, 50, 100). Use para economizar tokens.",
						},
						"filter_column": {
							Type:        "string",
							Description: "Coluna para filtrar (ex: 'A', 'Status'). Opcional.",
						},
						"filter_value": {
							Type:        "string",
							Description: "Valor para filtrar na coluna especificada. Opcional.",
						},
					},
					Required: []string{"sheet", "range"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "get_row_count",
				Description: "Obtém o número de linhas com dados em uma planilha.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
					},
					Required: []string{"sheet"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "get_column_count",
				Description: "Obtém o número de colunas com dados em uma planilha.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
					},
					Required: []string{"sheet"},
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
				Description: "Obtém o endereço da célula atualmente selecionada.",
				Parameters: FunctionParameters{
					Type:       "object",
					Properties: map[string]FunctionProperty{},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "has_filter",
				Description: "Verifica se a planilha tem filtro aplicado.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
					},
					Required: []string{"sheet"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "list_charts",
				Description: "Lista todos os gráficos de uma planilha.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
					},
					Required: []string{"sheet"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "list_tables",
				Description: "Lista todas as tabelas formatadas de uma planilha.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
					},
					Required: []string{"sheet"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "list_pivot_tables",
				Description: "Lista todas as tabelas dinâmicas de uma planilha.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
					},
					Required: []string{"sheet"},
				},
			},
		},

		// =========================================================================
		// ACTION TOOLS (Write operations)
		// =========================================================================
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "write_cell",
				Description: "Escreve um valor ou fórmula em uma célula específica.",
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
						"value": {
							Type:        "string",
							Description: "Valor a escrever. Para fórmulas, inicie com '=' (ex: '=SOMA(A1:A10)')",
						},
					},
					Required: []string{"sheet", "cell", "value"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "write_range",
				Description: "Escreve dados em lote a partir de uma célula inicial. Use para inserir tabelas ou múltiplos valores de uma vez.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"cell": {
							Type:        "string",
							Description: "Célula inicial (ex: 'A1')",
						},
						"data": {
							Type:        "array",
							Description: "Array 2D de valores. Cada sub-array é uma linha. Ex: [['Nome','Idade'],['João',25],['Maria',30]]",
							Items: &FunctionProperty{
								Type:        "array",
								Description: "Uma linha de dados",
								Items: &FunctionProperty{
									Type: "string",
								},
							},
						},
					},
					Required: []string{"sheet", "cell", "data"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "create_sheet",
				Description: "Cria uma nova planilha na pasta de trabalho.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"name": {
							Type:        "string",
							Description: "Nome da nova planilha",
						},
					},
					Required: []string{"name"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "delete_sheet",
				Description: "Exclui uma planilha da pasta de trabalho. CUIDADO: Esta ação é irreversível.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"name": {
							Type:        "string",
							Description: "Nome da planilha a excluir",
						},
					},
					Required: []string{"name"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "rename_sheet",
				Description: "Renomeia uma planilha.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"old_name": {
							Type:        "string",
							Description: "Nome atual da planilha",
						},
						"new_name": {
							Type:        "string",
							Description: "Novo nome",
						},
					},
					Required: []string{"old_name", "new_name"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "format_range",
				Description: "Aplica formatação a um intervalo de células.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de células (ex: 'A1:D1')",
						},
						"bold": {
							Type:        "boolean",
							Description: "Aplicar negrito",
						},
						"italic": {
							Type:        "boolean",
							Description: "Aplicar itálico",
						},
						"font_size": {
							Type:        "integer",
							Description: "Tamanho da fonte",
						},
						"font_color": {
							Type:        "string",
							Description: "Cor do texto em hex (ex: '#FF0000' para vermelho)",
						},
						"bg_color": {
							Type:        "string",
							Description: "Cor de fundo em hex (ex: '#FFFF00' para amarelo)",
						},
					},
					Required: []string{"sheet", "range"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "autofit_columns",
				Description: "Ajusta automaticamente a largura das colunas para caber o conteúdo.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de colunas (ex: 'A:D')",
						},
					},
					Required: []string{"sheet", "range"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "clear_range",
				Description: "Limpa o conteúdo de um intervalo de células.",
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
					},
					Required: []string{"sheet", "range"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "insert_rows",
				Description: "Insere novas linhas em uma posição específica.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"row": {
							Type:        "integer",
							Description: "Número da linha onde inserir (1-indexed)",
						},
						"count": {
							Type:        "integer",
							Description: "Quantidade de linhas a inserir",
						},
					},
					Required: []string{"sheet", "row", "count"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "delete_rows",
				Description: "Exclui linhas de uma planilha.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"row": {
							Type:        "integer",
							Description: "Número da primeira linha a excluir (1-indexed)",
						},
						"count": {
							Type:        "integer",
							Description: "Quantidade de linhas a excluir",
						},
					},
					Required: []string{"sheet", "row", "count"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "merge_cells",
				Description: "Mescla um intervalo de células.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de células a mesclar (ex: 'A1:C1')",
						},
					},
					Required: []string{"sheet", "range"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "unmerge_cells",
				Description: "Desfaz a mesclagem de células.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de células mescladas",
						},
					},
					Required: []string{"sheet", "range"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "set_borders",
				Description: "Aplica bordas a um intervalo de células.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de células",
						},
						"style": {
							Type:        "string",
							Description: "Estilo da borda",
							Enum:        []string{"thin", "medium", "thick", "double"},
						},
					},
					Required: []string{"sheet", "range", "style"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "set_column_width",
				Description: "Define a largura de colunas.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de colunas (ex: 'A:B')",
						},
						"width": {
							Type:        "number",
							Description: "Largura em pontos",
						},
					},
					Required: []string{"sheet", "range", "width"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "set_row_height",
				Description: "Define a altura de linhas.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de linhas (ex: '1:5')",
						},
						"height": {
							Type:        "number",
							Description: "Altura em pontos",
						},
					},
					Required: []string{"sheet", "range", "height"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "apply_filter",
				Description: "Aplica filtro automático a um intervalo.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de dados (ex: 'A1:D100')",
						},
					},
					Required: []string{"sheet", "range"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "clear_filters",
				Description: "Remove todos os filtros de uma planilha.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
					},
					Required: []string{"sheet"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "sort_range",
				Description: "Ordena um intervalo de dados por uma coluna.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de dados a ordenar",
						},
						"column": {
							Type:        "integer",
							Description: "Índice da coluna para ordenar (1-indexed)",
						},
						"ascending": {
							Type:        "boolean",
							Description: "Ordenar em ordem crescente (true) ou decrescente (false)",
						},
					},
					Required: []string{"sheet", "range", "column", "ascending"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "copy_range",
				Description: "Copia um intervalo de células para outro local.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"source": {
							Type:        "string",
							Description: "Intervalo de origem (ex: 'A1:B10')",
						},
						"dest": {
							Type:        "string",
							Description: "Célula de destino (ex: 'D1')",
						},
					},
					Required: []string{"sheet", "source", "dest"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "create_chart",
				Description: "Cria um gráfico a partir de dados.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha com os dados",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de dados para o gráfico (ex: 'A1:B10')",
						},
						"chart_type": {
							Type:        "string",
							Description: "Tipo de gráfico",
							Enum:        []string{"column", "bar", "line", "pie", "area", "scatter"},
						},
						"title": {
							Type:        "string",
							Description: "Título do gráfico",
						},
					},
					Required: []string{"sheet", "range", "chart_type", "title"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "delete_chart",
				Description: "Exclui um gráfico da planilha.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"name": {
							Type:        "string",
							Description: "Nome do gráfico a excluir",
						},
					},
					Required: []string{"sheet", "name"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "create_pivot_table",
				Description: "Cria uma tabela dinâmica a partir de dados.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"source_sheet": {
							Type:        "string",
							Description: "Planilha com os dados de origem",
						},
						"source_range": {
							Type:        "string",
							Description: "Intervalo de dados de origem (ex: 'A:F')",
						},
						"dest_sheet": {
							Type:        "string",
							Description: "Planilha de destino para a tabela dinâmica",
						},
						"dest_cell": {
							Type:        "string",
							Description: "Célula inicial da tabela dinâmica (ex: 'A1')",
						},
						"name": {
							Type:        "string",
							Description: "Nome da tabela dinâmica",
						},
					},
					Required: []string{"source_sheet", "source_range", "dest_sheet", "dest_cell", "name"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "create_table",
				Description: "Cria uma tabela formatada a partir de um intervalo de dados.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"range": {
							Type:        "string",
							Description: "Intervalo de dados (ex: 'A1:D10')",
						},
						"name": {
							Type:        "string",
							Description: "Nome da tabela",
						},
						"style": {
							Type:        "string",
							Description: "Estilo da tabela (ex: 'TableStyleMedium2')",
						},
					},
					Required: []string{"sheet", "range", "name"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "delete_table",
				Description: "Remove uma tabela formatada (mantém os dados).",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha",
						},
						"name": {
							Type:        "string",
							Description: "Nome da tabela a remover",
						},
					},
					Required: []string{"sheet", "name"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "execute_macro",
				Description: "Executa múltiplas ações/consultas em sequência de forma atômica. Use quando precisar fazer várias operações relacionadas (ex: criar planilha + escrever dados + formatar). Também suporta consultas como list_sheets, get_headers, get_range_values, etc.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"actions": {
							Type:        "array",
							Description: "Lista de ações/consultas a executar em sequência. Cada ação deve ter 'tool' (nome da ferramenta) e 'args' (argumentos). Ex: [{\"tool\":\"get_headers\",\"args\":{\"sheet\":\"Plan1\",\"range\":\"A:E\"}}, {\"tool\":\"get_row_count\",\"args\":{\"sheet\":\"Plan1\"}}]",
							Items: &FunctionProperty{
								Type:        "object",
								Description: "Uma ação individual com 'tool' e 'args'",
							},
						},
					},
					Required: []string{"actions"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDeclaration{
				Name:        "query_batch",
				Description: "Executa múltiplas consultas de uma vez para obter informações da planilha. Use para economizar chamadas de API quando precisar de várias informações simultaneamente.",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]FunctionProperty{
						"sheet": {
							Type:        "string",
							Description: "Nome da planilha a consultar",
						},
						"queries": {
							Type:        "array",
							Description: "Lista de consultas a executar. Opções: 'headers', 'row_count', 'used_range', 'sample_data', 'column_count'",
							Items: &FunctionProperty{
								Type: "string",
								Enum: []string{"headers", "row_count", "used_range", "sample_data", "column_count"},
							},
						},
						"sample_rows": {
							Type:        "integer",
							Description: "Número de linhas de amostra a retornar quando 'sample_data' está nas queries (default: 5)",
						},
					},
					Required: []string{"sheet", "queries"},
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
		"list_sheets":       true,
		"sheet_exists":      true,
		"get_used_range":    true,
		"get_headers":       true,
		"get_range_values":  true,
		"get_row_count":     true,
		"get_column_count":  true,
		"get_cell_formula":  true,
		"get_active_cell":   true,
		"has_filter":        true,
		"list_charts":       true,
		"list_tables":       true,
		"list_pivot_tables": true,
		"query_batch":       true,
	}
	return queryTools[name]
}

// IsActionTool returns true if the tool modifies Excel
func IsActionTool(name string) bool {
	return !IsQueryTool(name)
}
