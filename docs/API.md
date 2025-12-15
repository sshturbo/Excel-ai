# Documentação da API - Excel-ai

Esta documentação descreve todos os métodos e APIs disponíveis no backend do Excel-ai.

## Índice

- [Visão Geral](#visão-geral)
- [API de Chat](#api-de-chat)
- [API de Excel](#api-de-excel)
- [API de Configuração](#api-de-configuração)
- [API de Conversas](#api-de-conversas)
- [API de Licença](#api-de-licença)
- [Eventos](#eventos)
- [Tipos e Estruturas](#tipos-e-estruturas)

## Visão Geral

A comunicação entre frontend e backend ocorre através dos bindings do Wails. Métodos Go exportados na struct `App` são automaticamente disponibilizados no frontend.

### Convenções

- Todos os métodos estão no pacote `app.App`
- Erros são retornados como strings
- Operações assíncronas podem usar eventos
- JSON é usado para estruturas complexas

## API de Chat

### SendMessage

Envia uma mensagem para a IA e recebe resposta em stream.

**Assinatura**:
```go
func (a *App) SendMessage(message string) error
```

**Parâmetros**:
- `message` (string): Mensagem do usuário

**Retorno**:
- `error`: Erro se houver falha

**Eventos Emitidos**:
- `message:stream`: Pedaços da resposta da IA
- `message:complete`: Quando a resposta está completa
- `message:error`: Se houver erro durante streaming

**Exemplo (TypeScript)**:
```typescript
import { SendMessage } from '@/wailsjs/go/app/App'
import { EventsOn } from '@/wailsjs/runtime/runtime'

// Escutar stream
EventsOn('message:stream', (chunk: string) => {
  console.log('Received:', chunk)
})

// Enviar mensagem
await SendMessage("Qual é a soma da coluna A?")
```

---

### CancelMessage

Cancela o streaming de mensagem em andamento.

**Assinatura**:
```go
func (a *App) CancelMessage() error
```

**Retorno**:
- `error`: Erro se houver falha

**Exemplo**:
```typescript
import { CancelMessage } from '@/wailsjs/go/app/App'

await CancelMessage()
```

---

### ClearHistory

Limpa o histórico de mensagens da conversa atual.

**Assinatura**:
```go
func (a *App) ClearHistory() error
```

**Retorno**:
- `error`: Erro se houver falha

---

## API de Excel

### GetWorkbooks

Retorna lista de workbooks abertos no Excel.

**Assinatura**:
```go
func (a *App) GetWorkbooks() ([]string, error)
```

**Retorno**:
- `[]string`: Array com nomes dos workbooks
- `error`: Erro se houver falha (ex: Excel não aberto)

**Exemplo**:
```typescript
import { GetWorkbooks } from '@/wailsjs/go/app/App'

const workbooks = await GetWorkbooks()
// ["Vendas.xlsx", "Relatório.xlsx"]
```

---

### GetSheets

Retorna lista de sheets (abas) de um workbook.

**Assinatura**:
```go
func (a *App) GetSheets(workbookName string) ([]string, error)
```

**Parâmetros**:
- `workbookName` (string): Nome do workbook

**Retorno**:
- `[]string`: Array com nomes das sheets
- `error`: Erro se houver falha

**Exemplo**:
```typescript
const sheets = await GetSheets("Vendas.xlsx")
// ["Dados", "Resumo", "Gráficos"]
```

---

### GetPreviewData

Obtém dados de um range para preview.

**Assinatura**:
```go
func (a *App) GetPreviewData(workbook, sheet, rangeAddr string) ([][]interface{}, error)
```

**Parâmetros**:
- `workbook` (string): Nome do workbook
- `sheet` (string): Nome da sheet
- `rangeAddr` (string): Endereço do range (ex: "A1:C10")

**Retorno**:
- `[][]interface{}`: Matriz com os dados
- `error`: Erro se houver falha

**Exemplo**:
```typescript
const data = await GetPreviewData("Vendas.xlsx", "Dados", "A1:B10")
// [
//   ["Nome", "Valor"],
//   ["Item 1", 100],
//   ["Item 2", 200]
// ]
```

---

### RefreshWorkbooks

Atualiza a lista de workbooks detectados.

**Assinatura**:
```go
func (a *App) RefreshWorkbooks() ([]string, error)
```

**Retorno**:
- `[]string`: Array atualizado com nomes dos workbooks
- `error`: Erro se houver falha

---

### UndoLastChange

Desfaz a última alteração feita pelo Excel-ai.

**Assinatura**:
```go
func (a *App) UndoLastChange() error
```

**Retorno**:
- `error`: Erro se houver falha

**Exemplo**:
```typescript
await UndoLastChange()
```

---

### StartUndoBatch

Inicia um batch de operações para undo agrupado.

**Assinatura**:
```go
func (a *App) StartUndoBatch() error
```

**Retorno**:
- `error`: Erro se houver falha

---

### EndUndoBatch

Finaliza um batch de operações.

**Assinatura**:
```go
func (a *App) EndUndoBatch() error
```

**Retorno**:
- `error`: Erro se houver falha

**Exemplo de uso conjunto**:
```typescript
await StartUndoBatch()
try {
  // Múltiplas operações
  await someOperation1()
  await someOperation2()
} finally {
  await EndUndoBatch()
}
```

---

### StartWorkbookWatcher

Inicia observador de mudanças nos workbooks.

**Assinatura**:
```go
func (a *App) StartWorkbookWatcher() error
```

**Retorno**:
- `error`: Erro se houver falha

**Eventos Emitidos**:
- `workbooks:changed`: Quando workbooks mudam

---

### StopWorkbookWatcher

Para o observador de mudanças.

**Assinatura**:
```go
func (a *App) StopWorkbookWatcher() error
```

**Retorno**:
- `error`: Erro se houver falha

---

## API de Configuração

### SaveConfig

Salva configurações da aplicação.

**Assinatura**:
```go
func (a *App) SaveConfig(config dto.ConfigDTO) error
```

**Parâmetros**:
- `config` (ConfigDTO): Objeto com configurações

**ConfigDTO**:
```go
type ConfigDTO struct {
    Provider string `json:"provider"` // "openrouter", "groq", "google", "custom"
    APIKey   string `json:"apiKey"`
    Model    string `json:"model"`
    BaseURL  string `json:"baseUrl,omitempty"`
}
```

**Retorno**:
- `error`: Erro se houver falha

**Exemplo**:
```typescript
await SaveConfig({
  provider: "openrouter",
  apiKey: "sk-...",
  model: "openai/gpt-4-turbo",
  baseUrl: "https://openrouter.ai/api/v1"
})
```

---

### GetSavedConfig

Obtém configurações salvas.

**Assinatura**:
```go
func (a *App) GetSavedConfig() (*dto.ConfigDTO, error)
```

**Retorno**:
- `*ConfigDTO`: Configurações salvas (ou nil)
- `error`: Erro se houver falha

**Exemplo**:
```typescript
const config = await GetSavedConfig()
if (config) {
  console.log('Provider:', config.provider)
}
```

---

## API de Conversas

### SaveConversation

Salva a conversa atual.

**Assinatura**:
```go
func (a *App) SaveConversation(title string) error
```

**Parâmetros**:
- `title` (string): Título da conversa

**Retorno**:
- `error`: Erro se houver falha

**Exemplo**:
```typescript
await SaveConversation("Análise de Vendas Q1")
```

---

### LoadConversation

Carrega uma conversa salva.

**Assinatura**:
```go
func (a *App) LoadConversation(conversationID string) ([]domain.Message, error)
```

**Parâmetros**:
- `conversationID` (string): ID da conversa

**Retorno**:
- `[]Message`: Array de mensagens
- `error`: Erro se houver falha

**Message**:
```go
type Message struct {
    Role      string    `json:"role"`      // "user" ou "assistant"
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

**Exemplo**:
```typescript
const messages = await LoadConversation("conv-123")
```

---

### ListConversations

Lista todas as conversas salvas.

**Assinatura**:
```go
func (a *App) ListConversations() ([]domain.ConversationSummary, error)
```

**Retorno**:
- `[]ConversationSummary`: Array de resumos
- `error`: Erro se houver falha

**ConversationSummary**:
```go
type ConversationSummary struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Preview   string    `json:"preview"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

**Exemplo**:
```typescript
const conversations = await ListConversations()
conversations.forEach(conv => {
  console.log(conv.title, conv.updatedAt)
})
```

---

### DeleteConversation

Deleta uma conversa.

**Assinatura**:
```go
func (a *App) DeleteConversation(conversationID string) error
```

**Parâmetros**:
- `conversationID` (string): ID da conversa

**Retorno**:
- `error`: Erro se houver falha

---

## API de Licença

### CheckLicense

Verifica se a licença é válida.

**Assinatura**:
```go
func (a *App) CheckLicense() (bool, string)
```

**Retorno**:
- `bool`: true se válida
- `string`: Mensagem de status

**Exemplo**:
```typescript
const [isValid, message] = await CheckLicense()
if (isValid) {
  console.log('Licença válida:', message)
}
```

---

### GetLicenseMessage

Obtém mensagem de status da licença.

**Assinatura**:
```go
func (a *App) GetLicenseMessage() string
```

**Retorno**:
- `string`: Mensagem de status

---

### IsLicenseValid

Verifica se a licença é válida (boolean apenas).

**Assinatura**:
```go
func (a *App) IsLicenseValid() bool
```

**Retorno**:
- `bool`: true se válida

---

## Eventos

### message:stream

Emitido durante streaming de resposta da IA.

**Payload**: `string` - Pedaço da resposta

**Exemplo**:
```typescript
EventsOn('message:stream', (chunk: string) => {
  appendToMessage(chunk)
})
```

---

### message:complete

Emitido quando resposta está completa.

**Payload**: `string` - Mensagem completa

---

### message:error

Emitido quando há erro durante streaming.

**Payload**: `string` - Mensagem de erro

---

### workbooks:changed

Emitido quando a lista de workbooks muda.

**Payload**: `string[]` - Nova lista de workbooks

**Exemplo**:
```typescript
EventsOn('workbooks:changed', (workbooks: string[]) => {
  updateWorkbookList(workbooks)
})
```

---

### action:executed

Emitido quando uma ação é executada no Excel.

**Payload**: Objeto com detalhes da ação

---

## Tipos e Estruturas

### ConfigDTO

```typescript
interface ConfigDTO {
  provider: 'openrouter' | 'groq' | 'google' | 'custom'
  apiKey: string
  model: string
  baseUrl?: string
}
```

---

### Message

```typescript
interface Message {
  role: 'user' | 'assistant'
  content: string
  timestamp: string // ISO 8601
}
```

---

### ConversationSummary

```typescript
interface ConversationSummary {
  id: string
  title: string
  preview: string
  updatedAt: string // ISO 8601
}
```

---

### ExcelAction

```typescript
interface ExcelAction {
  type: 'read_data' | 'write_data' | 'create_chart' | 'format_range' | 'create_pivot' | 'apply_formula'
  workbook: string
  sheet: string
  range?: string
  data?: any
  chartType?: string
  formatting?: object
  // ... outros campos específicos por tipo
}
```

---

## Tratamento de Erros

### Códigos de Erro Comuns

| Erro | Descrição | Solução |
|------|-----------|---------|
| "Excel not found" | Excel não está aberto | Abrir Excel |
| "Workbook not found" | Workbook especificado não existe | Verificar nome |
| "Sheet not found" | Sheet não existe | Verificar nome da aba |
| "API request failed" | Erro na API de IA | Verificar chave e conexão |
| "Invalid range" | Range inválido | Corrigir formato (ex: A1:B10) |

### Exemplo de Tratamento

```typescript
try {
  await SendMessage(message)
} catch (error) {
  if (error.includes('Excel not found')) {
    showNotification('Por favor, abra o Excel primeiro')
  } else if (error.includes('API request failed')) {
    showNotification('Erro na comunicação com IA. Verifique suas configurações.')
  } else {
    showNotification('Erro: ' + error)
  }
}
```

---

## Rate Limiting

As APIs de IA têm limites de taxa. O Excel-ai não implementa rate limiting interno, mas respeitará os limites dos provedores:

- **OpenRouter**: Varia por modelo
- **Groq**: ~30 req/min (free tier)
- **Google Gemini**: 60 req/min

---

## Autenticação

### API Keys

Chaves de API são armazenadas localmente e não são transmitidas para servidores do Excel-ai. Elas são enviadas apenas para o provedor de IA selecionado.

---

## Exemplos de Fluxos Completos

### Fluxo 1: Enviar Mensagem e Processar Resposta

```typescript
import { SendMessage } from '@/wailsjs/go/app/App'
import { EventsOn } from '@/wailsjs/runtime/runtime'

let currentMessage = ''

// Setup listeners
EventsOn('message:stream', (chunk: string) => {
  currentMessage += chunk
  updateUI(currentMessage)
})

EventsOn('message:complete', (fullMessage: string) => {
  console.log('Complete:', fullMessage)
  currentMessage = ''
})

EventsOn('message:error', (error: string) => {
  showError(error)
  currentMessage = ''
})

// Send message
try {
  await SendMessage("Crie um gráfico de pizza com a coluna B")
} catch (error) {
  showError('Failed to send message: ' + error)
}
```

---

### Fluxo 2: Carregar e Exibir Dados do Excel

```typescript
import { GetWorkbooks, GetSheets, GetPreviewData } from '@/wailsjs/go/app/App'

async function loadExcelData() {
  try {
    // 1. Get workbooks
    const workbooks = await GetWorkbooks()
    if (workbooks.length === 0) {
      throw new Error('No workbooks open')
    }
    
    // 2. Get sheets from first workbook
    const sheets = await GetSheets(workbooks[0])
    
    // 3. Get data from first sheet
    const data = await GetPreviewData(workbooks[0], sheets[0], 'A1:Z100')
    
    // 4. Display
    displayTable(data)
    
  } catch (error) {
    console.error('Error loading Excel data:', error)
  }
}
```

---

### Fluxo 3: Salvar e Carregar Conversas

```typescript
import { 
  SaveConversation, 
  ListConversations, 
  LoadConversation 
} from '@/wailsjs/go/app/App'

// Save current conversation
async function saveCurrentConversation() {
  const title = prompt('Título da conversa:')
  if (title) {
    await SaveConversation(title)
    showNotification('Conversa salva!')
  }
}

// List all conversations
async function showConversationsList() {
  const conversations = await ListConversations()
  
  conversations.forEach(conv => {
    console.log(`${conv.title} - ${conv.updatedAt}`)
    console.log(`Preview: ${conv.preview}`)
  })
}

// Load a specific conversation
async function loadPreviousConversation(id: string) {
  const messages = await LoadConversation(id)
  
  // Restore messages in UI
  messages.forEach(msg => {
    addMessageToUI(msg.role, msg.content)
  })
}
```

---

## Performance

### Otimizações

1. **Batch Operations**: Use `StartUndoBatch` / `EndUndoBatch` para múltiplas operações
2. **Caching**: Workbooks e sheets podem ser cached no frontend
3. **Debouncing**: Debounce chamadas frequentes (ex: auto-refresh)

### Exemplo de Debouncing

```typescript
import { debounce } from 'lodash'

const debouncedRefresh = debounce(async () => {
  await RefreshWorkbooks()
}, 500)

// Use debouncedRefresh() em vez de RefreshWorkbooks()
```

---

## Migração de Versões

### v1.x → v2.x

Mudanças na API:
- `SendMessage` agora retorna `error` em vez de resultado direto
- Streaming é obrigatório via eventos
- `ConfigDTO` adicionou campo `baseUrl`

---

## Referências Adicionais

- [ARCHITECTURE.md](ARCHITECTURE.md) - Arquitetura detalhada
- [DEVELOPMENT.md](DEVELOPMENT.md) - Guia de desenvolvimento
- [Wails Bindings](https://wails.io/docs/howdoesitwork#go-bindings)

---

## Suporte

Para problemas relacionados à API:

1. Verifique os logs no console
2. Teste com `wails dev` para debug
3. Abra uma issue no GitHub com detalhes

