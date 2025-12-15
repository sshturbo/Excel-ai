# Arquitetura do Excel-ai

## Visão Geral

Excel-ai é uma aplicação desktop construída com Wails v2, que combina um backend em Go com um frontend em React/TypeScript. A aplicação se comunica com o Microsoft Excel através de COM (Component Object Model) no Windows.

## Diagrama de Alto Nível

```
┌─────────────────────────────────────────────────────────────┐
│                     Excel-ai Application                     │
│                                                              │
│  ┌─────────────────────┐       ┌─────────────────────────┐ │
│  │   Frontend (React)   │◄─────►│   Backend (Go/Wails)   │ │
│  │  - UI Components     │       │  - Business Logic      │ │
│  │  - State Management  │       │  - API Integration     │ │
│  │  - TypeScript        │       │  - Excel COM Client    │ │
│  └─────────────────────┘       └─────────────────────────┘ │
│                                           │                  │
└───────────────────────────────────────────┼──────────────────┘
                                            │
                    ┌───────────────────────┴────────────────────┐
                    │                                             │
                    ▼                                             ▼
        ┌──────────────────────┐                   ┌──────────────────────┐
        │  Microsoft Excel     │                   │   AI Provider API    │
        │  (COM Interface)     │                   │  - OpenRouter        │
        │  - Workbooks         │                   │  - Groq              │
        │  - Sheets            │                   │  - Google Gemini     │
        │  - Data Manipulation │                   │  - Custom APIs       │
        └──────────────────────┘                   └──────────────────────┘
```

## Estrutura de Diretórios

```
excel-ai/
│
├── main.go                      # Ponto de entrada da aplicação Wails
│
├── internal/                    # Código interno da aplicação (não exportável)
│   │
│   ├── app/                    # Camada de aplicação
│   │   ├── app.go             # Estrutura principal e inicialização
│   │   ├── chat_handlers.go   # Handlers para operações de chat
│   │   ├── config_handlers.go # Handlers para configurações
│   │   ├── excel_handlers.go  # Handlers para Excel
│   │   ├── license_handlers.go # Handlers de licença
│   │   └── watcher.go         # Observador de mudanças no Excel
│   │
│   ├── domain/                 # Modelos de domínio
│   │   ├── conversation.go    # Modelo de conversação
│   │   ├── message.go         # Modelo de mensagem
│   │   └── model.go           # Outros modelos
│   │
│   ├── dto/                    # Data Transfer Objects
│   │   └── types.go           # Tipos compartilhados
│   │
│   └── services/              # Serviços de negócio
│       │
│       ├── chat/              # Serviço de Chat/IA
│       │   ├── service.go    # Lógica principal do chat
│       │   ├── parser.go     # Parse de comandos
│       │   ├── executor.go   # Execução de ações
│       │   ├── streaming.go  # Streaming de respostas
│       │   ├── history.go    # Gerenciamento de histórico
│       │   ├── conversation.go # Gerenciamento de conversas
│       │   └── models.go     # Modelos internos
│       │
│       └── excel/            # Serviço de Excel
│           ├── service.go    # Lógica principal
│           ├── context.go    # Gerenciamento de contexto
│           ├── query.go      # Consultas de dados
│           ├── edit.go       # Edição de células
│           ├── format.go     # Formatação
│           ├── objects.go    # Objetos do Excel
│           └── structure.go  # Estrutura de workbooks
│
├── pkg/                       # Pacotes reutilizáveis (exportáveis)
│   │
│   ├── ai/                   # Clientes de IA
│   │   ├── openrouter.go    # Cliente OpenRouter/OpenAI compatible
│   │   └── gemini.go        # Cliente Google Gemini
│   │
│   ├── excel/                # Cliente COM do Excel
│   │   ├── client.go        # Cliente COM principal
│   │   ├── workbook.go      # Operações de workbook
│   │   ├── data.go          # Manipulação de dados
│   │   ├── formatting.go    # Formatação
│   │   ├── charts.go        # Criação de gráficos
│   │   ├── pivot.go         # Tabelas dinâmicas
│   │   ├── table.go         # Tabelas
│   │   ├── query.go         # Consultas
│   │   ├── operations.go    # Operações diversas
│   │   └── types.go         # Tipos e constantes
│   │
│   ├── license/              # Sistema de licenciamento
│   │   ├── license.go       # Lógica de licença
│   │   └── adapter.go       # Adaptador para validação
│   │
│   └── storage/              # Persistência de dados
│       └── storage.go       # SQLite storage para configs e conversas
│
└── frontend/                 # Interface React
    │
    ├── src/
    │   │
    │   ├── components/      # Componentes React
    │   │   │
    │   │   ├── ui/         # Componentes UI base (shadcn)
    │   │   │   ├── button.tsx
    │   │   │   ├── input.tsx
    │   │   │   ├── card.tsx
    │   │   │   └── ...
    │   │   │
    │   │   ├── layout/     # Componentes de layout
    │   │   │   ├── Header.tsx
    │   │   │   └── Sidebar.tsx
    │   │   │
    │   │   ├── chat/       # Componentes de chat
    │   │   │   ├── ChatInput.tsx
    │   │   │   ├── MessageBubble.tsx
    │   │   │   └── EmptyState.tsx
    │   │   │
    │   │   ├── excel/      # Componentes relacionados ao Excel
    │   │   │   ├── DataPreview.tsx
    │   │   │   ├── ChartViewer.tsx
    │   │   │   ├── Toolbar.tsx
    │   │   │   └── PendingActions.tsx
    │   │   │
    │   │   ├── markdown/   # Renderização de Markdown
    │   │   │   └── MarkdownRenderer.tsx
    │   │   │
    │   │   └── settings/   # Componentes de configuração
    │   │       ├── ApiTab.tsx
    │   │       ├── DataTab.tsx
    │   │       └── SettingsHeader.tsx
    │   │
    │   ├── hooks/          # Custom React Hooks
    │   │   ├── useChat.ts
    │   │   ├── useConversations.ts
    │   │   ├── useExcelConnection.ts
    │   │   ├── useStreamingMessage.ts
    │   │   └── useTheme.ts
    │   │
    │   ├── services/       # Serviços frontend
    │   │   ├── excelActions.ts
    │   │   ├── aiProcessor.ts
    │   │   └── contentCleaner.ts
    │   │
    │   ├── types/          # Tipos TypeScript
    │   │   └── index.ts
    │   │
    │   ├── lib/            # Utilitários
    │   │   └── utils.ts
    │   │
    │   ├── App.tsx         # Componente principal
    │   ├── Settings.tsx    # Tela de configurações
    │   └── main.tsx        # Entry point
    │
    ├── wailsjs/           # Bindings gerados pelo Wails
    │   └── go/            # Funções Go acessíveis do frontend
    │
    └── package.json       # Dependências npm

```

## Camadas da Aplicação

### 1. Camada de Apresentação (Frontend)

**Tecnologias**: React 18, TypeScript, Tailwind CSS, shadcn/ui

**Responsabilidades**:
- Interface do usuário
- Gerenciamento de estado local
- Comunicação com backend via Wails bindings
- Renderização de Markdown e syntax highlighting
- Visualizações de dados com Chart.js

**Principais Componentes**:
- `App.tsx`: Componente raiz da aplicação
- `Settings.tsx`: Interface de configurações
- `components/`: Componentes reutilizáveis organizados por função

### 2. Camada de Aplicação (App Layer)

**Localização**: `internal/app/`

**Responsabilidades**:
- Coordenação entre serviços
- Handlers para operações expostas ao frontend
- Gerenciamento do ciclo de vida da aplicação
- Observação de mudanças no Excel

**Principais Arquivos**:
- `app.go`: Estrutura principal e inicialização
- `*_handlers.go`: Handlers específicos por domínio

### 3. Camada de Serviços (Business Logic)

**Localização**: `internal/services/`

#### Serviço de Chat (`chat/`)
- Comunicação com APIs de IA
- Parse de respostas e extração de ações
- Streaming de mensagens
- Gerenciamento de histórico e conversas
- Integração com serviço de Excel

#### Serviço de Excel (`excel/`)
- Abstração de alto nível para operações no Excel
- Gerenciamento de contexto (workbooks, sheets)
- Operações de leitura e escrita
- Formatação e manipulação de dados

### 4. Camada de Domínio

**Localização**: `internal/domain/`

**Responsabilidades**:
- Definição de entidades de negócio
- Regras de negócio fundamentais
- Modelos de dados core

**Entidades Principais**:
- `Conversation`: Representa uma conversa completa
- `Message`: Representa uma mensagem individual
- `Model`: Configuração de modelo de IA

### 5. Camada de Infraestrutura

**Localização**: `pkg/`

#### Cliente de IA (`pkg/ai/`)
- Integração com OpenRouter (compatível com OpenAI)
- Integração com Google Gemini
- Streaming de respostas
- Gestão de requisições HTTP

#### Cliente Excel COM (`pkg/excel/`)
- Comunicação COM com Microsoft Excel
- Thread-safe operations
- Gerenciamento de recursos COM
- Operações de baixo nível (células, ranges, gráficos, etc.)

#### Storage (`pkg/storage/`)
- Persistência em SQLite
- Armazenamento de configurações
- Histórico de conversas
- Migrations de schema

#### Licenciamento (`pkg/license/`)
- Validação de licenças
- Controle de acesso

## Fluxo de Dados

### 1. Fluxo de Comando do Usuário

```
Usuario → Frontend (React)
    ↓
    Chat Input Component
    ↓
    useChat Hook
    ↓
    Wails Binding (SendMessage)
    ↓
    Backend: app.SendMessage (handler)
    ↓
    chatService.ProcessMessage
    ↓
    AI Provider API
    ↓
    Stream Response → Parse Actions
    ↓
    Execute Actions → Excel Service
    ↓
    Excel COM Client → Microsoft Excel
    ↓
    Response Stream → Frontend
    ↓
    UI Update
```

### 2. Fluxo de Ações no Excel

```
AI Response → Parser
    ↓
    Extract Actions (JSON)
    ↓
    Executor Service
    ↓
    Switch by Action Type:
        - read_data → excel.GetData
        - write_data → excel.SetData
        - create_chart → excel.CreateChart
        - format_range → excel.FormatRange
        - etc.
    ↓
    Excel COM Operations
    ↓
    Excel File Updated
    ↓
    Watcher Detects Change
    ↓
    Notify Frontend
```

### 3. Fluxo de Configuração

```
Settings UI → SaveConfig
    ↓
    Backend: app.SaveConfig
    ↓
    storage.SaveConfig (SQLite)
    ↓
    chatService.SetAPIKey/SetModel
    ↓
    Update AI Client Config
```

## Comunicação Frontend-Backend

### Wails Bindings

Wails gera automaticamente bindings TypeScript para funções Go exportadas. Exemplo:

**Backend (Go)**:
```go
// internal/app/chat_handlers.go
func (a *App) SendMessage(message string) error {
    return a.chatService.ProcessMessage(message)
}
```

**Frontend (TypeScript)**:
```typescript
import { SendMessage } from '@/wailsjs/go/app/App'

// Uso
await SendMessage("Qual é a soma da coluna A?")
```

### Event System

Para comunicação assíncrona (Backend → Frontend):

**Backend**:
```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

runtime.EventsEmit(a.ctx, "excel:changed", data)
```

**Frontend**:
```typescript
import { EventsOn } from '@/wailsjs/runtime/runtime'

EventsOn('excel:changed', (data) => {
    console.log('Excel changed:', data)
})
```

## Integração COM com Excel

### Thread Safety

O Excel COM requer operações em uma thread específica. O design do cliente garante isso:

```go
type Client struct {
    excelApp *ole.IDispatch
    cmdChan  chan func()
    doneChan chan struct{}
    mu       sync.Mutex
}

func (c *Client) runOnCOMThread(fn func() error) error {
    errChan := make(chan error, 1)
    c.cmdChan <- func() {
        errChan <- fn()
    }
    return <-errChan
}
```

Todas as operações COM são serializadas através de `cmdChan`, garantindo que ocorram na thread COM dedicada.

### Gerenciamento de Recursos

Objetos COM devem ser explicitamente liberados:

```go
workbooks := oleutil.GetProperty(c.excelApp, "Workbooks")
defer workbooks.ToIDispatch().Release()
```

### Retry Logic

Para lidar com Excel ocupado:

```go
for i := 0; i < 10; i++ {
    err = fn()
    if err == nil {
        break
    }
    if strings.Contains(err.Error(), "Call was rejected") {
        time.Sleep(1 * time.Second)
        continue
    }
    break
}
```

## Sistema de IA

### Provedores Suportados

1. **OpenRouter** (padrão)
   - Gateway para múltiplos modelos (GPT-4, Claude, etc.)
   - API compatível com OpenAI

2. **Groq**
   - Inferência ultra-rápida
   - Modelos otimizados

3. **Google Gemini**
   - Cliente específico para API do Google
   - Suporte a modelos Gemini Pro/Ultra

4. **Custom APIs**
   - Qualquer API compatível com OpenAI

### Streaming de Respostas

Utiliza Server-Sent Events (SSE) para streaming em tempo real:

```go
func (c *Client) ChatStream(ctx context.Context, messages []Message, 
    callback func(string)) error {
    
    reader := bufio.NewReader(resp.Body)
    for {
        line, err := reader.ReadBytes('\n')
        if err != nil {
            break
        }
        
        if bytes.HasPrefix(line, []byte("data: ")) {
            data := bytes.TrimPrefix(line, []byte("data: "))
            callback(string(data))
        }
    }
}
```

### Parse de Ações

Respostas da IA podem incluir ações em JSON:

```json
{
  "type": "read_data",
  "workbook": "Vendas.xlsx",
  "sheet": "Dados",
  "range": "A1:B10"
}
```

O parser extrai essas ações e o executor as processa.

## Persistência de Dados

### SQLite Schema

```sql
-- Configurações
CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Conversas
CREATE TABLE IF NOT EXISTS conversations (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    context TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Mensagens
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
```

## Segurança

### API Keys

- Armazenadas localmente em SQLite
- Não transmitidas para servidores externos (exceto provedores de IA)
- Configuráveis pelo usuário

### Validação de Licença

Sistema de licenciamento integrado para controle de uso.

## Performance

### Otimizações

1. **Thread Dedicada COM**: Evita overhead de inicialização COM
2. **Caching**: Referências a workbooks e sheets mantidas quando possível
3. **Batching**: Operações múltiplas agrupadas com `StartUndoBatch`/`EndUndoBatch`
4. **Streaming**: Respostas da IA streamadas para feedback imediato
5. **Lazy Loading**: Dados carregados sob demanda

### Limitações

- Uma instância do Excel por vez
- Excel deve estar aberto antes de iniciar operações
- Windows específico para COM (suporte parcial em outras plataformas)

## Escalabilidade

### Horizontal

Não aplicável (aplicação desktop single-user)

### Vertical

- Múltiplas conversas simultâneas suportadas
- Histórico ilimitado (limitado por SQLite e disco)
- Suporte a workbooks grandes (limitado pelo Excel)

## Manutenibilidade

### Princípios Seguidos

- **Separation of Concerns**: Camadas bem definidas
- **Dependency Injection**: Serviços injetados onde necessário
- **Interface Segregation**: Interfaces pequenas e focadas
- **Single Responsibility**: Cada módulo tem uma responsabilidade clara

### Testes

```
internal/services/chat/integration_test.go    # Testes de integração
```

### Logs

Logs estruturados para debug:
```go
fmt.Printf("[ExcelClient] Erro: %s\n", err)
fmt.Println("[LICENSE] ✅ Licença válida")
```

## Roadmap Técnico

### Futuras Melhorias

1. **Cross-platform**: Suporte para Excel Online, Google Sheets
2. **Plugin System**: Extensões customizadas
3. **Testes Automatizados**: Maior cobertura de testes
4. **Telemetria**: Analytics de uso (opt-in)
5. **CI/CD**: Pipeline automatizado de build e release
6. **Multi-idioma**: Internacionalização completa

## Referências

- [Wails Documentation](https://wails.io/docs/introduction)
- [go-ole Documentation](https://github.com/go-ole/go-ole)
- [React Documentation](https://react.dev)
- [Microsoft Excel Object Model](https://docs.microsoft.com/en-us/office/vba/api/overview/excel)
