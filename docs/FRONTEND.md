# Documentação do Frontend - Excel-ai

Este documento descreve a arquitetura, componentes e estrutura do frontend do Excel-ai.

## Índice

- [Visão Geral](#visão-geral)
- [Stack Tecnológico](#stack-tecnológico)
- [Estrutura de Diretórios](#estrutura-de-diretórios)
- [Componentes](#componentes)
- [Hooks Personalizados](#hooks-personalizados)
- [Gerenciamento de Estado](#gerenciamento-de-estado)
- [Serviços](#serviços)
- [Estilização](#estilização)
- [Integração com Backend](#integração-com-backend)

## Visão Geral

O frontend do Excel-ai é uma Single Page Application (SPA) construída com React e TypeScript, usando Vite como bundler e dev server.

### Características

- ⚛️ **React 18** com TypeScript
- 🎨 **Tailwind CSS** para estilização
- 🧩 **shadcn/ui** para componentes base
- 🔄 **Custom Hooks** para lógica reutilizável
- 📱 **Responsivo** e moderno
- 🌓 **Tema claro/escuro**
- ✨ **Animações** suaves
- 🔌 **Wails Bindings** para comunicação com Go

## Stack Tecnológico

### Core

| Tecnologia | Versão | Uso |
|------------|--------|-----|
| React | 18.2 | Framework UI |
| TypeScript | 5.9 | Tipagem estática |
| Vite | 7.2 | Build tool |
| Tailwind CSS | 4.1 | Framework CSS |

### UI Components

| Biblioteca | Uso |
|-----------|-----|
| shadcn/ui | Componentes base |
| Radix UI | Primitivos acessíveis |
| Lucide React | Ícones |
| next-themes | Gerenciamento de tema |

### Utilitários

| Biblioteca | Uso |
|-----------|-----|
| react-markdown | Renderização de Markdown |
| react-syntax-highlighter | Syntax highlighting |
| chart.js | Gráficos |
| class-variance-authority | Variantes de componentes |
| clsx / tailwind-merge | Merge de classes CSS |

## Estrutura de Diretórios

```
frontend/
├── src/
│   ├── components/          # Componentes React
│   │   ├── ui/             # Componentes base (shadcn)
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   ├── card.tsx
│   │   │   └── ...
│   │   │
│   │   ├── layout/         # Componentes de layout
│   │   │   ├── Header.tsx
│   │   │   └── Sidebar.tsx
│   │   │
│   │   ├── chat/           # Componentes de chat
│   │   │   ├── ChatInput.tsx
│   │   │   ├── MessageBubble.tsx
│   │   │   ├── EmptyState.tsx
│   │   │   └── MessageList.tsx
│   │   │
│   │   ├── excel/          # Componentes Excel
│   │   │   ├── DataPreview.tsx
│   │   │   ├── ChartViewer.tsx
│   │   │   ├── Toolbar.tsx
│   │   │   └── PendingActions.tsx
│   │   │
│   │   ├── markdown/       # Markdown
│   │   │   └── MarkdownRenderer.tsx
│   │   │
│   │   └── settings/       # Configurações
│   │       ├── ApiTab.tsx
│   │       ├── DataTab.tsx
│   │       └── SettingsHeader.tsx
│   │
│   ├── hooks/              # Custom Hooks
│   │   ├── useChat.ts
│   │   ├── useConversations.ts
│   │   ├── useExcelConnection.ts
│   │   ├── useStreamingMessage.ts
│   │   └── useTheme.ts
│   │
│   ├── services/           # Serviços
│   │   ├── excelActions.ts
│   │   ├── aiProcessor.ts
│   │   └── contentCleaner.ts
│   │
│   ├── types/              # Tipos TypeScript
│   │   └── index.ts
│   │
│   ├── lib/                # Utilitários
│   │   └── utils.ts
│   │
│   ├── assets/             # Assets estáticos
│   │   ├── fonts/
│   │   └── images/
│   │
│   ├── App.tsx             # Componente raiz
│   ├── Settings.tsx        # Tela de configurações
│   ├── main.tsx            # Entry point
│   └── index.css           # Estilos globais
│
├── wailsjs/                # Bindings Wails (gerado)
│   └── go/
│       └── app/
│           └── App.js/ts   # Bindings do backend
│
├── public/                 # Assets públicos
├── index.html              # HTML template
├── package.json            # Dependências
├── tsconfig.json           # Config TypeScript
├── vite.config.ts          # Config Vite
└── tailwind.config.js      # Config Tailwind
```

## Componentes

### Layout Components

#### Header

**Localização**: `src/components/layout/Header.tsx`

**Responsabilidade**: Cabeçalho da aplicação

```typescript
interface HeaderProps {
  theme: 'light' | 'dark'
  onToggleTheme: () => void
  onOpenSettings: () => void
  onOpenConversations: () => void
}
```

**Conteúdo**:
- Logo
- Toggle de tema
- Botão de configurações
- Botão de conversas

#### Sidebar

**Localização**: `src/components/layout/Sidebar.tsx`

**Responsabilidade**: Barra lateral com histórico

```typescript
interface SidebarProps {
  conversations: ConversationSummary[]
  currentConversationId?: string
  onSelectConversation: (id: string) => void
  onNewConversation: () => void
  onDeleteConversation: (id: string) => void
}
```

**Conteúdo**:
- Lista de conversas
- Botão "Nova Conversa"
- Informações de conexão com Excel

### Chat Components

#### ChatInput

**Localização**: `src/components/chat/ChatInput.tsx`

**Responsabilidade**: Campo de entrada de mensagens

```typescript
interface ChatInputProps {
  onSend: (message: string) => void
  disabled?: boolean
  placeholder?: string
}
```

**Features**:
- Textarea expansível
- Submit com Enter
- Nova linha com Shift+Enter
- Botão de envio
- Estado de loading

#### MessageBubble

**Localização**: `src/components/chat/MessageBubble.tsx`

**Responsabilidade**: Bolha de mensagem

```typescript
interface MessageBubbleProps {
  role: 'user' | 'assistant'
  content: string
  timestamp?: string
}
```

**Features**:
- Alinhamento diferente por role
- Renderização de Markdown
- Timestamp
- Avatar

#### EmptyState

**Localização**: `src/components/chat/EmptyState.tsx`

**Responsabilidade**: Estado vazio (sem mensagens)

**Conteúdo**:
- Mensagem de boas-vindas
- Sugestões de comandos
- Ícones ilustrativos

### Excel Components

#### DataPreview

**Localização**: `src/components/excel/DataPreview.tsx`

**Responsabilidade**: Preview de dados do Excel

```typescript
interface DataPreviewProps {
  data: any[][]
  title?: string
  maxRows?: number
}
```

**Features**:
- Tabela responsiva
- Scroll horizontal/vertical
- Formatação de células
- Loading state

#### ChartViewer

**Localização**: `src/components/excel/ChartViewer.tsx`

**Responsabilidade**: Visualização de gráficos

```typescript
interface ChartViewerProps {
  chartData: ChartData
  chartType: 'bar' | 'line' | 'pie' | 'scatter'
}
```

**Features**:
- Suporte a múltiplos tipos
- Interativo (Chart.js)
- Responsivo
- Exportar imagem

#### Toolbar

**Localização**: `src/components/excel/Toolbar.tsx`

**Responsabilidade**: Barra de ferramentas

```typescript
interface ToolbarProps {
  workbooks: string[]
  selectedWorkbook?: string
  onSelectWorkbook: (name: string) => void
  onRefresh: () => void
  onUndo: () => void
}
```

**Conteúdo**:
- Seletor de workbook
- Botão de refresh
- Botão de undo
- Status de conexão

### Markdown Components

#### MarkdownRenderer

**Localização**: `src/components/markdown/MarkdownRenderer.tsx`

**Responsabilidade**: Renderizar Markdown com features avançadas

```typescript
interface MarkdownRendererProps {
  content: string
  className?: string
}
```

**Features**:
- Syntax highlighting (código)
- Tabelas
- Listas
- Links
- Imagens
- GFM (GitHub Flavored Markdown)

### Settings Components

#### ApiTab

**Localização**: `src/components/settings/ApiTab.tsx`

**Responsabilidade**: Configurações de API

**Campos**:
- Provider
- API Key
- Model
- Base URL

#### DataTab

**Localização**: `src/components/settings/DataTab.tsx`

**Responsabilidade**: Configurações de dados

**Campos**:
- Auto-refresh
- Preview rows
- Max history messages

## Hooks Personalizados

### useChat

**Localização**: `src/hooks/useChat.ts`

**Responsabilidade**: Gerenciar estado e lógica de chat

```typescript
interface UseChatReturn {
  messages: Message[]
  isLoading: boolean
  error: string | null
  sendMessage: (content: string) => Promise<void>
  cancelMessage: () => void
  clearHistory: () => void
}

function useChat(): UseChatReturn
```

**Funcionalidades**:
- Enviar mensagens
- Receber streaming
- Cancelar mensagem
- Limpar histórico
- Gerenciar estado de loading

### useConversations

**Localização**: `src/hooks/useConversations.ts`

**Responsabilidade**: Gerenciar conversas salvas

```typescript
interface UseConversationsReturn {
  conversations: ConversationSummary[]
  currentConversation: string | null
  loading: boolean
  saveConversation: (title: string) => Promise<void>
  loadConversation: (id: string) => Promise<void>
  deleteConversation: (id: string) => Promise<void>
  newConversation: () => void
}

function useConversations(): UseConversationsReturn
```

### useExcelConnection

**Localização**: `src/hooks/useExcelConnection.ts`

**Responsabilidade**: Gerenciar conexão com Excel

```typescript
interface UseExcelConnectionReturn {
  workbooks: string[]
  sheets: string[]
  selectedWorkbook: string | null
  selectedSheet: string | null
  isConnected: boolean
  refreshWorkbooks: () => Promise<void>
  selectWorkbook: (name: string) => void
  selectSheet: (name: string) => void
}

function useExcelConnection(): UseExcelConnectionReturn
```

### useStreamingMessage

**Localização**: `src/hooks/useStreamingMessage.ts`

**Responsabilidade**: Gerenciar streaming de mensagens

```typescript
interface UseStreamingMessageReturn {
  streamingContent: string
  isStreaming: boolean
  startStreaming: () => void
  appendChunk: (chunk: string) => void
  endStreaming: () => void
}

function useStreamingMessage(): UseStreamingMessageReturn
```

### useTheme

**Localização**: `src/hooks/useTheme.ts`

**Responsabilidade**: Gerenciar tema claro/escuro

```typescript
interface UseThemeReturn {
  theme: 'light' | 'dark'
  toggleTheme: () => void
  setTheme: (theme: 'light' | 'dark') => void
}

function useTheme(): UseThemeReturn
```

## Gerenciamento de Estado

### Local State (useState)

Para estado local de componentes:
```typescript
const [value, setValue] = useState<string>('')
```

### Context (não usado atualmente)

O projeto não usa Context API extensivamente, preferindo props drilling para simplicidade.

### Estado Global (futuro)

Considerar Zustand ou Jotai se estado global se tornar complexo.

## Serviços

### excelActions

**Localização**: `src/services/excelActions.ts`

**Responsabilidade**: Executar ações no Excel

```typescript
export async function executeExcelAction(action: ExcelAction): Promise<void> {
  switch (action.type) {
    case 'read_data':
      return await readDataAction(action)
    case 'write_data':
      return await writeDataAction(action)
    case 'create_chart':
      return await createChartAction(action)
    // ...
  }
}
```

### aiProcessor

**Localização**: `src/services/aiProcessor.ts`

**Responsabilidade**: Processar respostas da IA

```typescript
export function parseAIResponse(response: string): {
  text: string
  actions: ExcelAction[]
} {
  // Extrai ações JSON da resposta
  // Retorna texto limpo e ações
}
```

### contentCleaner

**Localização**: `src/services/contentCleaner.ts`

**Responsabilidade**: Limpar e formatar conteúdo

```typescript
export function cleanMarkdown(content: string): string
export function sanitizeHTML(html: string): string
export function formatTimestamp(date: Date): string
```

## Estilização

### Tailwind CSS

**Configuração**: `tailwind.config.js`

```javascript
module.exports = {
  content: ['./src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        // ...
      },
    },
  },
  plugins: [],
}
```

### CSS Variables

**Localização**: `src/index.css`

```css
@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    /* ... */
  }
  
  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    /* ... */
  }
}
```

### Utility Classes

Criadas em `src/lib/utils.ts`:

```typescript
import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

Uso:
```typescript
<div className={cn("base-class", condition && "conditional-class", className)} />
```

## Integração com Backend

### Wails Bindings

Gerados automaticamente em `frontend/wailsjs/`.

**Uso**:
```typescript
import { SendMessage, GetWorkbooks } from '@/wailsjs/go/app/App'

// Chamar método Go
const workbooks = await GetWorkbooks()
await SendMessage("Hello")
```

### Event System

**Escutar eventos**:
```typescript
import { EventsOn, EventsOff } from '@/wailsjs/runtime/runtime'

useEffect(() => {
  const unsub = EventsOn('message:stream', (data: string) => {
    console.log('Received:', data)
  })
  
  return () => EventsOff('message:stream')
}, [])
```

**Emitir eventos** (do backend):
```go
runtime.EventsEmit(a.ctx, "workbooks:changed", workbooks)
```

## Build e Deploy

### Development

```bash
cd frontend
npm run dev
```

### Production Build

```bash
npm run build
# Saída: frontend/dist/
```

### Preview

```bash
npm run preview
```

## Boas Práticas

### ✅ Faça

- Use TypeScript para tudo
- Componentes pequenos e focados
- Props tipadas com interfaces
- Custom hooks para lógica reutilizável
- Tailwind para estilização
- Acessibilidade (aria labels)

### ❌ Evite

- any sem necessidade
- Componentes gigantes
- Lógica complexa em JSX
- Inline styles
- Duplicação de código

## Performance

### Otimizações

1. **Code Splitting**: Vite faz automaticamente
2. **Lazy Loading**: Para rotas/componentes grandes
3. **Memoization**: `useMemo` / `useCallback` quando necessário
4. **Debouncing**: Para inputs frequentes

### Exemplo de Memoization

```typescript
const expensiveValue = useMemo(() => {
  return computeExpensiveValue(data)
}, [data])

const handleChange = useCallback((value: string) => {
  onChange(value)
}, [onChange])
```

## Testes (Futuro)

### Estrutura Proposta

```
frontend/
├── src/
│   └── __tests__/
│       ├── components/
│       ├── hooks/
│       └── services/
```

### Ferramentas Recomendadas

- **Vitest**: Test runner
- **React Testing Library**: Testes de componentes
- **MSW**: Mock de API

## Referências

- [React Documentation](https://react.dev)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [shadcn/ui](https://ui.shadcn.com)
- [Wails Frontend](https://wails.io/docs/guides/frontend)
