# Migração para Arquitetura Web com Upload de Arquivos

## Objetivo

Migrar o sistema Excel-AI de **COM (Windows-only)** para **Excelize (cross-platform)** com arquitetura baseada em upload de arquivos e visualização via WebView.

---

## User Review Required

> [!IMPORTANT]
> **Breaking Change**: O sistema deixará de funcionar em tempo real com Excel aberto. O novo fluxo será:
> 1. Upload do arquivo `.xlsx`
> 2. IA modifica arquivo em memória
> 3. WebView mostra preview das modificações
> 4. Download do arquivo final

> [!WARNING]
> **Limitações do Excelize**:
> - Pivot Tables: Criação básica (sem atualização dinâmica)
> - Macros VBA: Não suportado
> - Fórmulas: Salvas mas não calculadas automaticamente

---

## Proposed Changes

### Fase 1: Backend - Excelize Client

#### [NEW] [excelize_client.go](file:///c:/Projetos/Excel-ai/pkg/excel/excelize_client.go)

Novo cliente usando biblioteca Excelize com mesma interface do COM.

```go
type ExcelizeClient struct {
    file       *excelize.File
    filePath   string
    mu         sync.Mutex
}

// Métodos implementados:
// - GetSheetList() []string
// - GetCellValue(sheet, cell) string
// - SetCellValue(sheet, cell, value) error
// - GetRows(sheet) [][]string
// - SetSheetRow(sheet, cell, row) error
// - AddChart(sheet, format) error
// - SetCellStyle(sheet, hCell, vCell, styleID) error
// - NewStyle(style) (int, error)
// - MergeCell(sheet, hCell, vCell) error
// - UnmergeCell(sheet, hCell, vCell) error
// - InsertRows(sheet, row, n) error
// - RemoveRow(sheet, row) error
// - SetColWidth(sheet, startCol, endCol, width) error
// - SetRowHeight(sheet, row, height) error
// - AutoFilter(sheet, hCell, vCell) error
// - SaveAs(path) error
// - Write(w io.Writer) error
```

---

#### [NEW] [file_manager.go](file:///c:/Projetos/Excel-ai/pkg/excel/file_manager.go)

Gerenciador de arquivos em memória para sessões de usuário.

```go
type FileManager struct {
    sessions map[string]*ExcelizeClient  // sessionID -> client
    mu       sync.RWMutex
}

func (fm *FileManager) LoadFile(sessionID string, data []byte) error
func (fm *FileManager) GetClient(sessionID string) (*ExcelizeClient, error)
func (fm *FileManager) Export(sessionID string) ([]byte, error)
func (fm *FileManager) Close(sessionID string)
```

---

#### [MODIFY] [types.go](file:///c:/Projetos/Excel-ai/pkg/excel/types.go)

Remover dependência de `go-ole`, criar tipos neutros.

```diff
-import (
-    "github.com/go-ole/go-ole"
-)
-
-type Client struct {
-    excelApp *ole.IDispatch
-    ...
-}
+import "sync"
+
+// ClientType define o tipo de cliente (COM ou Excelize)
+type ClientType int
+
+const (
+    ClientTypeCOM ClientType = iota
+    ClientTypeExcelize
+)
```

---

#### [NEW] [interface.go](file:///c:/Projetos/Excel-ai/pkg/excel/interface.go)

Interface abstrata para ambos os clientes.

```go
type ExcelClient interface {
    // Sheets
    ListSheets() ([]string, error)
    SheetExists(name string) (bool, error)
    CreateSheet(name string) error
    DeleteSheet(name string) error
    RenameSheet(oldName, newName string) error
    
    // Data
    GetCellValue(sheet, cell string) (string, error)
    SetCellValue(sheet, cell string, value interface{}) error
    GetRangeValues(sheet, rng string) ([][]string, error)
    WriteRange(sheet, startCell string, data [][]interface{}) error
    ClearRange(sheet, rng string) error
    
    // Formatting
    FormatRange(sheet, rng string, format Format) error
    SetColumnWidth(sheet, col string, width float64) error
    SetRowHeight(sheet, row string, height float64) error
    MergeCells(sheet, rng string) error
    UnmergeCells(sheet, rng string) error
    SetBorders(sheet, rng, style string) error
    AutoFitColumns(sheet, rng string) error
    
    // Structure
    InsertRows(sheet string, row, count int) error
    DeleteRows(sheet string, row, count int) error
    
    // Objects
    CreateChart(sheet, rng, chartType, title string) error
    DeleteChart(sheet, name string) error
    ListCharts(sheet string) ([]string, error)
    CreateTable(sheet, rng, name, style string) error
    DeleteTable(sheet, name string) error
    ListTables(sheet string) ([]string, error)
    CreatePivotTable(srcSheet, srcRange, destSheet, destCell, name string) error
    ListPivotTables(sheet string) ([]string, error)
    
    // Filters
    ApplyFilter(sheet, rng string) error
    ClearFilters(sheet string) error
    HasFilter(sheet string) (bool, error)
    SortRange(sheet, rng string, col int, ascending bool) error
    
    // Query
    GetUsedRange(sheet string) (string, error)
    GetRowCount(sheet string) (int, error)
    GetColumnCount(sheet string) (int, error)
    GetHeaders(sheet, rng string) ([]string, error)
    GetCellFormula(sheet, cell string) (string, error)
    
    // Lifecycle
    Close()
}
```

---

### Fase 2: Serviço Excel

#### [MODIFY] [service.go](file:///c:/Projetos/Excel-ai/internal/services/excel/service.go)

Alterar para usar interface abstrata e suportar ambos os modos.

```diff
 type Service struct {
-    client              *excel.Client
+    client              excel.ExcelClient
+    clientType          excel.ClientType
+    fileManager         *excel.FileManager
+    currentSessionID    string
     mu                  sync.Mutex
     ...
 }

+// ConnectFile conecta a um arquivo via Excelize
+func (s *Service) ConnectFile(sessionID string, data []byte) error
+
+// ExportFile exporta o arquivo atual
+func (s *Service) ExportFile() ([]byte, error)
```

---

### Fase 3: API de Upload/Download

#### [NEW] [upload_handlers.go](file:///c:/Projetos/Excel-ai/internal/app/upload_handlers.go)

Endpoints HTTP para upload e download.

```go
// POST /api/excel/upload
// Recebe arquivo .xlsx e retorna sessionID
func (a *App) UploadExcel(filename string, data []byte) (string, error)

// GET /api/excel/download
// Retorna arquivo .xlsx modificado
func (a *App) DownloadExcel(sessionID string) ([]byte, error)

// GET /api/excel/preview
// Retorna dados para o viewer (JSON com sheets e dados)
func (a *App) GetExcelPreview(sessionID string) (*PreviewData, error)
```

---

### Fase 4: Frontend

#### [NEW] [UploadExcel.tsx](file:///c:/Projetos/Excel-ai/frontend/src/components/UploadExcel.tsx)

Componente de upload de arquivos.

```tsx
// Drag & drop + file picker
// Envia para backend via Wails binding
// Mostra progresso e status
```

---

#### [NEW] [ExcelViewer.tsx](file:///c:/Projetos/Excel-ai/frontend/src/components/ExcelViewer.tsx)

Visualizador de planilhas usando biblioteca web (ex: Handsontable).

```tsx
// Renderiza dados da planilha
// Suporta múltiplas abas
// Atualiza via WebSocket quando IA modifica
// Read-only por padrão
```

---

### Fase 5: Migração de Tools

#### [MODIFY] [executor.go](file:///c:/Projetos/Excel-ai/internal/services/chat/executor.go)

Sem mudanças no executor - ele usa `excelService` que agora é interface.
Os métodos do `excelService` passam a chamar Excelize internamente.

---

#### Mapeamento de Operações

| Tool | COM Method | Excelize Method |
|------|------------|-----------------|
| `list-sheets` | [getSheetsInternal()](file:///c:/Projetos/Excel-ai/pkg/excel/client.go#224-252) | `f.GetSheetList()` |
| `get-range-values` | `ReadRangeData()` | `f.GetRows()` |
| `write` | `WriteCell()` via COM | `f.SetCellValue()` |
| `write` (batch) | [WriteRange()](file:///c:/Projetos/Excel-ai/internal/services/excel/edit.go#415-451) | `f.SetSheetRow()` loop |
| `format-range` | `ApplyFormatting()` | `f.SetCellStyle()` |
| `create-chart` | COM `AddChart` | `f.AddChart()` |
| `create-pivot` | COM `PivotTables.Add` | ⚠️ Limitado |
| `merge-cells` | COM `Merge` | `f.MergeCell()` |
| `insert-rows` | COM `Rows.Insert` | `f.InsertRows()` |
| `sort` | COM `Sort.Apply` | ⚠️ Manual sort |
| `apply-filter` | COM `AutoFilter` | `f.AutoFilter()` |

---

## Verification Plan

### Automated Tests

1. **Testes Unitários do ExcelizeClient**
   ```bash
   go test ./pkg/excel/... -v -run TestExcelize
   ```
   - Criar arquivo `pkg/excel/excelize_client_test.go`
   - Testar cada método do ExcelizeClient isoladamente

2. **Teste de Integração Existente (Adaptado)**
   ```bash
   go test ./internal/services/chat/... -v -run TestAllTools
   ```
   - Modificar [integration_test.go](file:///c:/Projetos/Excel-ai/internal/services/chat/integration_test.go) para usar flag `--mode=excelize`
   - Reutilizar os 30+ testes existentes

3. **Teste de Upload/Download**
   ```bash
   go test ./internal/app/... -v -run TestUploadDownload
   ```
   - Criar arquivo de teste para endpoints

### Manual Verification

1. **Fluxo Completo E2E**
   - Subir aplicação: `wails dev`
   - Fazer upload de arquivo `.xlsx` de teste
   - Enviar comando de chat: "Adicione uma coluna C com valores 1,2,3"
   - Verificar preview no WebView
   - Baixar arquivo e abrir no Excel real

2. **Teste de Compatibilidade**
   - Abrir arquivo modificado no Microsoft Excel
   - Verificar formatação preservada
   - Verificar fórmulas preservadas

---

## Estimativa de Esforço

| Fase | Arquivos | Linhas | Tempo |
|------|----------|--------|-------|
| 1. Excelize Client | 3 | ~500 | 2-3 dias |
| 2. Interface Abstrata | 2 | ~200 | 1 dia |
| 3. API Upload | 1 | ~150 | 1 dia |
| 4. Frontend | 2 | ~400 | 2 dias |
| 5. Migração Tools | 0 | ~0 | 0 (sem mudanças) |
| 6. Testes | 2 | ~300 | 2 dias |
| **Total** | **10** | **~1550** | **8-10 dias** |

---

## Dependências a Adicionar

```bash
go get github.com/xuri/excelize/v2
```

Frontend (escolher uma):
```bash
cd frontend && npm install handsontable @handsontable/react
# ou
cd frontend && npm install xlsx sheetjs-style
```
