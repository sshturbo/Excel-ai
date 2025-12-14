package chat

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"excel-ai/internal/services/excel"
)

// TestExcelIntegration executa um teste completo das ferramentas do agente contra o Excel real
// Requer Excel aberto.
// Execute com: go test -v ./internal/services/chat/integration_test.go
func TestExcelIntegration(t *testing.T) {
	fmt.Println("--- INICIANDO TESTE INTEGRADO DE FERRAMENTAS ---")

	// 1. Setup Services
	excelSvc := excel.NewService()
	chatSvc := NewService(nil)
	chatSvc.SetExcelService(excelSvc)

	// 2. Connect
	fmt.Println("Conectando ao Excel...")
	status, err := excelSvc.Connect()
	if err != nil {
		t.Fatalf("Falha na conexão inicial: %v", err)
	}
	if !status.Connected {
		t.Fatalf("Excel não conectado: %s", status.Error)
	}
	fmt.Printf("Conectado! Workbooks abertos: %d\n", len(status.Workbooks))

	// Helper para executar tool
	exec := func(name string, toolType ToolType, payload map[string]interface{}) string {
		t.Helper()
		fmt.Printf(">> Executando: %s\n", name)

		cmd := ToolCommand{
			Type:    toolType,
			Payload: payload,
		}

		res, err := chatSvc.ExecuteTool(cmd)
		if err != nil {
			t.Fatalf("ERRO em %s: %v", name, err)
		}
		fmt.Printf("   Resultado: %s\n", res)
		return res
	}

	// 3. Test Cycle using Tool Commands

	// A. Create Workbook
	wbName := fmt.Sprintf("AutoTest_%d", time.Now().Unix())
	exec("Criar Workbook", ToolTypeAction, map[string]interface{}{
		"op":   "create-workbook",
		"name": wbName,
	})

	// Wait for Excel to settle
	time.Sleep(2 * time.Second)

	// B. Write Data
	exec("Escrever em A1", ToolTypeAction, map[string]interface{}{
		"op":    "write",
		"cell":  "A1",
		"value": "TesteIntegrado",
	})

	// C. Read Data (Query)
	// Como criamos um workbook novo, ele deve ser o ativo.
	// O backend usa 'active' se sheet/workbook não forem passados.
	res := exec("Ler Fórmula A1", ToolTypeQuery, map[string]interface{}{
		"type": "get-cell-formula",
		"cell": "A1",
	})

	if !contains(res, "TesteIntegrado") {
		t.Errorf("Falha leitura A1. Esperado conter 'TesteIntegrado', veio: %s", res)
	}

	// D. Create Sheet
	sheetName := "AbaTeste"
	exec("Criar Aba", ToolTypeAction, map[string]interface{}{
		"op":   "create-sheet",
		"name": sheetName,
	})

	// E. Check Sheet Exists
	res = exec("Verificar Aba", ToolTypeQuery, map[string]interface{}{
		"type": "sheet-exists",
		"name": sheetName,
	})
	if !contains(res, "true") {
		t.Errorf("Aba %s deveria existir", sheetName)
	}

	// F. Write in New Sheet
	exec("Escrever na Aba Nova", ToolTypeAction, map[string]interface{}{
		"op":    "write",
		"cell":  "B2",
		"value": "123",
	})

	// G. Format Range
	exec("Formatar Célula", ToolTypeAction, map[string]interface{}{
		"op":        "format-range",
		"range":     "B2",
		"bold":      true,
		"fontColor": "#FF0000",
	})

	// H. Insert Rows
	exec("Inserir Linhas", ToolTypeAction, map[string]interface{}{
		"op":    "insert-rows",
		"sheet": sheetName,
		"row":   1,
		"count": 2,
	})

	// I. Merge Cells
	exec("Mesclar Células", ToolTypeAction, map[string]interface{}{
		"op":    "merge-cells",
		"sheet": sheetName,
		"range": "D2:E3",
	})

	// J. Set Borders
	exec("Bordas", ToolTypeAction, map[string]interface{}{
		"op":    "set-borders",
		"sheet": sheetName,
		"range": "B2",
		"style": "thick",
	})

	// K. Create Table
	exec("Criar Tabela", ToolTypeAction, map[string]interface{}{
		"op":    "create-table",
		"sheet": sheetName,
		"range": "G1:H5",
		"name":  "TabelaVendas",
		"style": "TableStyleMedium9",
	})

	// L. Delete Sheet (Cleanup)
	time.Sleep(1 * time.Second)
	// exec("Deletar Aba", ToolTypeAction, map[string]interface{}{
	// 	"op": "delete-sheet",
	// 	"name": sheetName,
	// })

	// I. Final Report
	fmt.Println("--- TESTE CONCLUÍDO COM SUCESSO ---")
	fmt.Println("Verifique o arquivo criado no Excel.")
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
