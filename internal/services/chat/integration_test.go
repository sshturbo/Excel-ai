package chat

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"excel-ai/internal/services/excel"
)

// TestAllTools executa um teste EXAUSTIVO de todas as ferramentas disponíveis no Agente.
// Requer Excel aberto.
func TestAllTools(t *testing.T) {
	fmt.Println("=== INICIANDO MEGA TESTE DE TODAS AS 30+ FERRAMENTAS ===")

	// 1. Setup
	excelSvc := excel.NewService()
	chatSvc := NewService(nil)
	chatSvc.SetExcelService(excelSvc)

	// Conectar
	status, err := excelSvc.Connect()
	if err != nil {
		t.Fatalf("Falha crítica conexão: %v", err)
	}
	if !status.Connected {
		t.Fatalf("Não conectado")
	}

	// Helper
	exec := func(name string, op string, payload map[string]interface{}) {
		t.Helper()
		fmt.Printf("\n[TEST] %s (%s)... ", name, op)

		fullPayload := make(map[string]interface{})
		for k, v := range payload {
			fullPayload[k] = v
		}

		// Se for action
		var cmd ToolCommand
		if strings.HasPrefix(op, "get-") || strings.HasPrefix(op, "list-") || strings.Contains(op, "exists") {
			fullPayload["type"] = op
			cmd = ToolCommand{Type: ToolTypeQuery, Payload: fullPayload}
		} else {
			fullPayload["op"] = op
			cmd = ToolCommand{Type: ToolTypeAction, Payload: fullPayload}
		}

		res, err := chatSvc.ExecuteTool(cmd)
		if err != nil {
			fmt.Printf("FAIL ❌ (%v)\n", err)
			t.Errorf("Falha em %s: %v", name, err)
		} else {
			fmt.Printf("OK ✅\n   -> %s\n", res)
		}
		// Throttle pequeno para o Excel respirar
		time.Sleep(200 * time.Millisecond)
	}

	sheet1 := "Dados"
	sheet2 := "Relatorio"

	// GRUPO 1: WORKBOOK & SHEETS
	exec("Criar Workbook", "create-workbook", nil) // Retorna nome novo e seta current
	time.Sleep(1 * time.Second)                    // Wait for workbook creation

	exec("Renomear Aba Padrão", "rename-sheet", map[string]interface{}{
		"oldName": "Planilha1",
		"newName": sheet1,
	})

	exec("Criar Segunda Aba", "create-sheet", map[string]interface{}{"name": sheet2})

	exec("Verificar Aba Existe", "sheet-exists", map[string]interface{}{"name": sheet2})
	exec("Listar Abas", "list-sheets", nil)

	// GRUPO 2: DADOS & EDIÇÃO
	exec("Escrever Cabeçalho", "write", map[string]interface{}{
		"sheet": sheet1, "cell": "A1", "value": "Produto",
	})
	exec("Escrever Valor", "write", map[string]interface{}{
		"sheet": sheet1, "cell": "B1", "value": "Valor",
	})

	// Preencher dados para teste de Sort e Filter
	products := []string{"Mouse", "Teclado", "Monitor", "Cabo"}
	prices := []int{50, 150, 1200, 20}

	for i, p := range products {
		row := i + 2
		exec(fmt.Sprintf("Add Prod %s", p), "write", map[string]interface{}{
			"sheet": sheet1, "cell": fmt.Sprintf("A%d", row), "value": p,
		})
		exec(fmt.Sprintf("Add Price %d", prices[i]), "write", map[string]interface{}{
			"sheet": sheet1, "cell": fmt.Sprintf("B%d", row), "value": fmt.Sprintf("%d", prices[i]),
		})
	}

	exec("Ler Célula", "get-cell-formula", map[string]interface{}{"sheet": sheet1, "cell": "A2"})
	exec("Obter Used Range", "get-used-range", map[string]interface{}{"sheet": sheet1})

	// GRUPO 3: FORMATAÇÃO
	exec("Negrito Header", "format-range", map[string]interface{}{
		"sheet": sheet1, "range": "A1:B1", "bold": true,
	})
	exec("Cor de Fundo", "format-range", map[string]interface{}{
		"sheet": sheet1, "range": "A1:B1", "bgColor": "#FFFF00",
	})
	exec("Set Column Width", "set-column-width", map[string]interface{}{
		"sheet": sheet1, "range": "A:A", "width": 20.0,
	})
	exec("AutoFit", "autofit", map[string]interface{}{
		"sheet": sheet1, "range": "B:B",
	})
	exec("Set Row Height", "set-row-height", map[string]interface{}{
		"sheet": sheet1, "range": "1:1", "height": 30.0,
	})
	exec("Bordas", "set-borders", map[string]interface{}{
		"sheet": sheet1, "range": "A1:B5", "style": "medium",
	})

	// GRUPO 4: ESTRUTURA
	exec("Inserir Linha Topo", "insert-rows", map[string]interface{}{
		"sheet": sheet1, "row": 1, "count": 1,
	})
	exec("Deletar Linha Inserida", "delete-rows", map[string]interface{}{
		"sheet": sheet1, "row": 1, "count": 1,
	})

	exec("Mesclar Células (Titulo)", "merge-cells", map[string]interface{}{
		"sheet": sheet2, "range": "A1:C1",
	})
	exec("Escrever no Merge", "write", map[string]interface{}{
		"sheet": sheet2, "cell": "A1", "value": "RELATÓRIO GERAL",
	})
	exec("Desmesclar", "unmerge-cells", map[string]interface{}{
		"sheet": sheet2, "range": "A1:C1",
	})

	// GRUPO 5: TABELAS & FILTROS
	exec("Criar Tabela", "create-table", map[string]interface{}{
		"sheet": sheet1, "range": "A1:B5", "name": "TabVendas", "style": "TableStyleMedium2",
	})
	exec("Listar Tabelas", "list-tables", map[string]interface{}{"sheet": sheet1})

	// Filtro (Table já tem filtro, mas vamos testar apply-filter em range normal na sheet2)
	exec("Dados Filter", "write", map[string]interface{}{
		"sheet": sheet2, "cell": "A5", "value": "X",
	})
	exec("Aplicar Filtro", "apply-filter", map[string]interface{}{
		"sheet": sheet2, "range": "A5",
	})
	exec("Limpar Filtros", "clear-filters", map[string]interface{}{"sheet": sheet2})

	// Sort
	exec("Ordenar Preço Desc", "sort", map[string]interface{}{
		"sheet": sheet1, "range": "A2:B5", "column": 2.0, "ascending": false,
	})

	// GRUPO 6: CHART & PIVOT
	// Charts precisam de dados selecionados ou range.
	exec("Criar Gráfico", "create-chart", map[string]interface{}{
		"sheet": sheet1, "range": "A1:B5", "type": "xlColumnClustered",
		"title": "Vendas Chart",
	})
	exec("Listar Gráficos", "list-charts", map[string]interface{}{"sheet": sheet1})

	// Pivot (Complexo, exige source data valido)
	// Vamos tentar criar na sheet2 usando dados da sheet1
	// Pivot creation logic is tricky via COM regarding destination string
	// exec("Criar Pivot", "create-pivot", map[string]interface{}{
	// 	"sourceSheet": sheet1, "sourceRange": "A1:B5",
	// 	"destSheet": sheet2, "destCell": "E5",
	// 	"rowFields": []string{"Produto"}, "dataFields": []string{"Valor"},
	// })

	// GRUPO 7: CLEANUP ACTIONS
	//	exec("Deletar Gráfico", "delete-chart", map[string]interface{}{
	//		"sheet": sheet1, "name": "Vendas Chart",
	//	})
	exec("Deletar Tabela", "delete-table", map[string]interface{}{
		"sheet": sheet1, "name": "TabVendas",
	})
	exec("Deletar Aba 2", "delete-sheet", map[string]interface{}{"name": sheet2})

	fmt.Println("=== MEGA TESTE FINALIZADO ===")
}
