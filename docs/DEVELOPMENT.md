# Guia de Desenvolvimento - Excel-ai

Este guia fornece informações para desenvolvedores que desejam contribuir ou estender o Excel-ai.

## Índice

- [Configuração do Ambiente](#configuração-do-ambiente)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Desenvolvimento Frontend](#desenvolvimento-frontend)
- [Desenvolvimento Backend](#desenvolvimento-backend)
- [Testes](#testes)
- [Build e Deploy](#build-e-deploy)
- [Convenções de Código](#convenções-de-código)
- [Debugging](#debugging)
- [Contribuindo](#contribuindo)

## Configuração do Ambiente

### Pré-requisitos

Certifique-se de ter instalado:
- Go 1.23+
- Node.js 18+
- Wails CLI
- Git
- IDE recomendada: VSCode, GoLand, ou WebStorm

### Clone e Setup

```bash
# Clone o repositório
git clone https://github.com/sshturbo/Excel-ai.git
cd Excel-ai

# Instale dependências
go mod download
cd frontend && npm install && cd ..

# Verifique o ambiente
wails doctor
```

### VSCode Setup

Instale as extensões recomendadas:

```json
{
  "recommendations": [
    "golang.go",
    "esbenp.prettier-vscode",
    "dbaeumer.vscode-eslint",
    "bradlc.vscode-tailwindcss",
    "ms-vscode.vscode-typescript-next"
  ]
}
```

Crie `.vscode/settings.json`:

```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

## Estrutura do Projeto

```
excel-ai/
├── main.go                 # Entry point
├── internal/               # Código privado
│   ├── app/               # Handlers e lógica de aplicação
│   ├── domain/            # Entidades de negócio
│   ├── dto/               # Data Transfer Objects
│   └── services/          # Serviços de negócio
├── pkg/                   # Código público/reutilizável
│   ├── ai/               # Clientes de IA
│   ├── excel/            # Cliente COM Excel
│   ├── license/          # Sistema de licenças
│   └── storage/          # Persistência
└── frontend/             # React app
    └── src/
        ├── components/   # Componentes React
        ├── hooks/        # Custom hooks
        ├── services/     # Lógica de negócio frontend
        └── types/        # Tipos TypeScript
```

## Desenvolvimento Frontend

### Iniciar Dev Server

```bash
# Terminal 1: Backend
wails dev

# Ou apenas frontend (para desenvolvimento UI)
cd frontend
npm run dev
```

### Componentes

Usamos **shadcn/ui** como base. Para adicionar novos componentes:

```bash
cd frontend
npx shadcn-ui@latest add button
npx shadcn-ui@latest add dialog
```

### Custom Hooks

Exemplo de hook personalizado:

```typescript
// hooks/useExcelData.ts
import { useState, useEffect } from 'react'
import { GetPreviewData } from '@/wailsjs/go/app/App'

export function useExcelData(workbook: string, sheet: string) {
  const [data, setData] = useState<string[][]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    async function fetchData() {
      setLoading(true)
      try {
        const result = await GetPreviewData(workbook, sheet, 'A1:Z100')
        setData(result)
      } catch (error) {
        console.error('Error fetching data:', error)
      } finally {
        setLoading(false)
      }
    }
    
    if (workbook && sheet) {
      fetchData()
    }
  }, [workbook, sheet])

  return { data, loading }
}
```

### Tailwind CSS

Configuração está em `frontend/tailwind.config.js`. Para adicionar cores customizadas:

```javascript
module.exports = {
  theme: {
    extend: {
      colors: {
        'excel-green': '#107C41',
        'excel-dark': '#185C37',
      }
    }
  }
}
```

### Comunicação com Backend

```typescript
// Importar bindings gerados
import { SendMessage } from '@/wailsjs/go/app/App'
import { EventsOn, EventsEmit } from '@/wailsjs/runtime/runtime'

// Chamar função Go
async function sendChatMessage(message: string) {
  try {
    await SendMessage(message)
  } catch (error) {
    console.error('Error:', error)
  }
}

// Escutar eventos do backend
EventsOn('message:stream', (data: string) => {
  console.log('Received:', data)
})
```

### Testes Frontend

```bash
cd frontend

# Executar testes (quando implementados)
npm test

# Linting
npm run lint

# Type checking
npx tsc --noEmit
```

## Desenvolvimento Backend

### Estrutura de um Handler

```go
// internal/app/example_handlers.go
package app

func (a *App) ExampleMethod(param string) (string, error) {
    // Validação
    if param == "" {
        return "", fmt.Errorf("param cannot be empty")
    }
    
    // Lógica de negócio (delegar para service)
    result, err := a.someService.DoSomething(param)
    if err != nil {
        return "", fmt.Errorf("failed to do something: %w", err)
    }
    
    return result, nil
}
```

### Adicionar Nova Funcionalidade

#### 1. Criar Serviço

```go
// internal/services/myfeature/service.go
package myfeature

type Service struct {
    // dependências
}

func NewService() *Service {
    return &Service{}
}

func (s *Service) DoSomething(input string) (string, error) {
    // implementação
    return "result", nil
}
```

#### 2. Adicionar ao App

```go
// internal/app/app.go
type App struct {
    // ...
    myFeatureService *myfeature.Service
}

func NewApp() *App {
    // ...
    myFeatureSvc := myfeature.NewService()
    
    return &App{
        // ...
        myFeatureService: myFeatureSvc,
    }
}
```

#### 3. Criar Handler

```go
// internal/app/myfeature_handlers.go
func (a *App) UseMyFeature(input string) (string, error) {
    return a.myFeatureService.DoSomething(input)
}
```

#### 4. Usar no Frontend

Após `wails dev`, bindings são gerados automaticamente:

```typescript
import { UseMyFeature } from '@/wailsjs/go/app/App'

const result = await UseMyFeature("test")
```

### Trabalhando com Excel COM

```go
// Exemplo de operação COM
func (s *ExcelService) ReadRange(workbook, sheet, rangeAddr string) ([][]interface{}, error) {
    // Operações COM devem ser thread-safe
    return s.client.GetRangeValues(workbook, sheet, rangeAddr)
}
```

**Importante**:
- Sempre libere recursos COM: `defer obj.Release()`
- Use retry logic para Excel ocupado
- Mantenha referências enquanto necessário

### Logging

O Excel-AI utiliza um sistema de logging estruturado e centralizado através do pacote `pkg/logger`.

**Características**:
- 11 Componentes: APP, CHAT, EXCEL, AI, CACHE, STORAGE, LICENSE, HTTP, STREAM, TOOLS, UNDO
- 5 Níveis de Log: DEBUG, INFO, WARN, ERROR, FATAL
- Configuração via arquivo JSON (`logger-config.json`)
- Saída para console e/ou arquivo
- Thread-safe com singleton pattern

**Uso Básico**:
```go
import "excel-ai/pkg/logger"

// Log informativo por componente
logger.AppInfo("Aplicação iniciada com sucesso")
logger.ChatInfo("Nova mensagem recebida")
logger.ExcelInfo("Workbook aberto: Vendas.xlsx")
logger.AIInfo("Resposta recebida com sucesso")

// Log de aviso
logger.ExcelWarn("Excel ocupado, aguardando...")

// Log de erro
logger.AIError("Falha ao conectar com API da IA")

// Log de debug (detalhado)
logger.CacheDebug("Cache hit: chave encontrada")
```

**Formatação Avançada**:
```go
import (
    "fmt"
    "excel-ai/pkg/logger"
)

logger.ChatInfo(fmt.Sprintf("Usuário %s enviou mensagem", username))
logger.ExcelDebug(fmt.Sprintf("Range %s contém %d células", range, count))
```

**Configuração** (arquivo `logger-config.json`):
```json
{
  "level": "INFO",
  "output": "console",
  "file_path": "excel-ai.log",
  "components": {
    "APP": "INFO",
    "CHAT": "INFO",
    "EXCEL": "INFO",
    "AI": "INFO",
    "CACHE": "WARN",
    "STORAGE": "WARN"
  }
}
```

**Componentes Disponíveis**:
- `APP` - Aplicação geral
- `CHAT` - Sistema de chat
- `EXCEL` - Integração com Excel
- `AI` - Inteligência artificial
- `CACHE` - Sistema de cache
- `STORAGE` - Persistência de dados
- `LICENSE` - Sistema de licença
- `HTTP` - Requisições HTTP
- `STREAM` - Streaming de respostas
- `TOOLS` - Ferramentas do sistema
- `UNDO` - Sistema de desfazer

**Níveis de Log**:
- `DEBUG` - Informações detalhadas para desenvolvimento
- `INFO` - Eventos normais e importantes
- `WARN` - Situações anormais mas não críticas
- `ERROR` - Erros que não param a aplicação
- `FATAL` - Erros que requerem encerramento imediato

**Boas Práticas**:
1. Escolha o nível adequado para cada mensagem
2. Use o componente correto para cada área do sistema
3. Seja específico e forneça contexto
4. Evite informações sensíveis (API keys, tokens, etc.)

Para mais detalhes, consulte o código em `pkg/logger/logger.go`.

### Testes Backend

```go
// internal/services/chat/service_test.go
package chat

import (
    "testing"
)

func TestServiceDoSomething(t *testing.T) {
    svc := NewService(nil)
    
    result, err := svc.DoSomething("input")
    
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if result != "expected" {
        t.Errorf("Expected 'expected', got '%s'", result)
    }
}
```

Executar testes:

```bash
# Todos os testes
go test ./...

# Teste específico
go test ./internal/services/chat -v

# Com coverage
go test -cover ./...
```

## Testes

### Testes Unitários

```bash
# Go
go test ./... -v

# Frontend (quando implementado)
cd frontend && npm test
```

### Testes de Integração

```go
// internal/services/chat/integration_test.go
//go:build integration

func TestIntegration(t *testing.T) {
    // Testes que requerem Excel, IA, etc.
}
```

Executar:
```bash
go test -tags=integration ./...
```

### Testes Manuais

Checklist para testar manualmente:

- [ ] Abrir Excel com dados de teste
- [ ] Iniciar Excel-ai
- [ ] Testar comandos básicos
- [ ] Testar criação de gráficos
- [ ] Testar formatação
- [ ] Testar salvamento de conversas
- [ ] Testar mudança de configurações
- [ ] Testar undo/redo

## Build e Deploy

### Build Local

```bash
# Debug build
wails build -debug

# Production build
wails build -ldflags="-w -s"

# Com compression
wails build -upx
```

### Build para Múltiplas Plataformas

```bash
# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/universal

# Linux
wails build -platform linux/amd64
```

### Gerar Bindings

Bindings TypeScript são gerados automaticamente com `wails dev` ou `wails build`. Para forçar geração:

```bash
wails generate module
```

### Assets

Assets frontend são embedados no binário:

```go
//go:embed all:frontend/dist
var assets embed.FS
```

Para atualizar assets:
```bash
cd frontend && npm run build && cd ..
```

## Convenções de Código

### Go

**Naming**:
- Packages: lowercase, single word
- Functions: CamelCase (exported), camelCase (private)
- Types: CamelCase (exported), camelCase (private)
- Constants: CamelCase ou UPPERCASE com underscore

**Estrutura**:
```go
// Imports agrupados
import (
    // stdlib
    "context"
    "fmt"
    
    // externos
    "github.com/wailsapp/wails/v2"
    
    // internos
    "excel-ai/internal/domain"
)

// Type declarations
type Service struct {
    client *http.Client
}

// Constructor
func NewService() *Service {
    return &Service{
        client: &http.Client{},
    }
}

// Methods
func (s *Service) Method() error {
    return nil
}
```

**Error Handling**:
```go
// Wrap errors com contexto
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Retornar early
if condition {
    return errors.New("invalid condition")
}
```

### TypeScript/React

**Naming**:
- Components: PascalCase
- Hooks: camelCase com prefixo `use`
- Utilities: camelCase
- Types/Interfaces: PascalCase

**Estrutura de Componente**:
```typescript
import { useState } from 'react'

interface MyComponentProps {
    title: string
    onAction?: () => void
}

export function MyComponent({ title, onAction }: MyComponentProps) {
    const [state, setState] = useState<string>('')
    
    // Handlers
    const handleClick = () => {
        onAction?.()
    }
    
    // Render
    return (
        <div className="flex flex-col gap-4">
            <h2>{title}</h2>
            <button onClick={handleClick}>Action</button>
        </div>
    )
}
```

**Custom Hook**:
```typescript
export function useMyHook(param: string) {
    const [state, setState] = useState<string>()
    
    useEffect(() => {
        // Side effects
    }, [param])
    
    return { state, setState }
}
```

### Commit Messages

Formato:
```
<type>: <subject>

<body>

<footer>
```

Types:
- `feat`: Nova funcionalidade
- `fix`: Bug fix
- `docs`: Documentação
- `style`: Formatação, sem mudança de código
- `refactor`: Refatoração
- `test`: Adicionar testes
- `chore`: Manutenção

Exemplo:
```
feat: add chart creation via AI commands

- Parse chart actions from AI response
- Implement chart creation in Excel service
- Add chart preview in UI

Closes #123
```

## Debugging

### Backend (Go)

#### VSCode Launch Configuration

`.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Wails Dev",
      "type": "go",
      "request": "launch",
      "mode": "exec",
      "program": "${workspaceFolder}/build/bin/excel-ai.exe",
      "preLaunchTask": "wails:build:debug"
    }
  ]
}
```

#### Print Debugging

```go
fmt.Printf("[DEBUG] Variable: %+v\n", variable)
fmt.Printf("[DEBUG] %s: %v\n", "label", value)
```

#### Delve Debugger

```bash
# Instalar
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug ./main.go
```

### Frontend (React)

#### Chrome DevTools

```typescript
// Console logs
console.log('Data:', data)
console.error('Error:', error)
console.table(arrayData)

// Breakpoint
debugger
```

#### React DevTools

Instale a extensão React DevTools no Chrome/Edge para inspeção de componentes.

### Wails DevTools

No modo `wails dev`, pressione `F12` na janela do app para abrir DevTools.

### Logs de Runtime

```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

// Log no console do frontend
runtime.LogInfo(a.ctx, "Info message")
runtime.LogError(a.ctx, "Error message")
runtime.LogDebug(a.ctx, "Debug message")
```

## Profiling

### Go

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### Frontend

Use o Chrome DevTools Performance tab para profiling de React.

## Hot Reload

### Backend

Mudanças em `.go` files automaticamente recompilam com `wails dev`.

### Frontend

Mudanças em frontend são hot-reloaded instantaneamente via Vite.

## Workflows Comuns

### Adicionar Nova API Endpoint

1. Criar método no service apropriado
2. Adicionar handler em `internal/app/`
3. Bindings gerados automaticamente
4. Usar no frontend

### Adicionar Novo Componente UI

1. Criar componente em `frontend/src/components/`
2. Adicionar tipos em `types/index.ts` se necessário
3. Importar e usar no componente pai

### Adicionar Nova Dependência

```bash
# Go
go get github.com/some/package
go mod tidy

# Frontend
cd frontend
npm install some-package
```

## Troubleshooting Desenvolvimento

### Bindings desatualizados

```bash
# Deletar e regenerar
rm -rf frontend/wailsjs
wails generate module
```

### Frontend build falha

```bash
cd frontend
rm -rf node_modules dist
npm install
npm run build
```

### Go module issues

```bash
go clean -modcache
go mod download
go mod verify
```

## Recursos

- [Wails Documentation](https://wails.io/docs/)
- [Go Documentation](https://go.dev/doc/)
- [React Documentation](https://react.dev)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Tailwind CSS](https://tailwindcss.com/docs)

## Próximos Passos

1. Explore o código existente
2. Escolha uma issue para trabalhar
3. Crie um branch de feature
4. Desenvolva e teste
5. Submeta um Pull Request

Veja também [CONTRIBUTING.md](CONTRIBUTING.md) para guidelines de contribuição.
