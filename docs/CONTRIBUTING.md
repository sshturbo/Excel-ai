# Guia de Contribuição - Excel-ai

Obrigado por considerar contribuir com o Excel-ai! Este documento fornece diretrizes para contribuir com o projeto.

## Índice

- [Código de Conduta](#código-de-conduta)
- [Como Contribuir](#como-contribuir)
- [Processo de Desenvolvimento](#processo-de-desenvolvimento)
- [Padrões de Código](#padrões-de-código)
- [Commits e Pull Requests](#commits-e-pull-requests)
- [Reportando Bugs](#reportando-bugs)
- [Sugerindo Funcionalidades](#sugerindo-funcionalidades)
- [Primeiros Passos](#primeiros-passos)

## Código de Conduta

### Nossa Promessa

No interesse de promover um ambiente aberto e acolhedor, nós, como contribuidores e mantenedores, nos comprometemos a tornar a participação em nosso projeto e comunidade uma experiência livre de assédio para todos.

### Nossos Padrões

**Comportamento Aceitável**:
- Usar linguagem acolhedora e inclusiva
- Respeitar pontos de vista e experiências diferentes
- Aceitar críticas construtivas com graciosidade
- Focar no que é melhor para a comunidade
- Mostrar empatia com outros membros

**Comportamento Inaceitável**:
- Uso de linguagem ou imagens sexualizadas
- Trolling, comentários insultuosos/depreciativos
- Assédio público ou privado
- Publicar informações privadas de outros
- Conduta não profissional

## Como Contribuir

### Tipos de Contribuição

Aceitamos vários tipos de contribuição:

1. **Código**: Implementar funcionalidades, corrigir bugs
2. **Documentação**: Melhorar ou traduzir docs
3. **Testes**: Adicionar ou melhorar testes
4. **Design**: Melhorar UI/UX
5. **Issues**: Reportar bugs, sugerir features
6. **Revisão**: Revisar Pull Requests
7. **Traduções**: Internacionalizar a aplicação

### Onde Começar

#### Para Iniciantes

Procure por issues com labels:
- `good first issue`: Bom para começar
- `help wanted`: Precisamos de ajuda
- `documentation`: Melhorias em docs
- `bug`: Correção de bugs

#### Para Desenvolvedores Experientes

- `enhancement`: Novas funcionalidades
- `performance`: Otimizações
- `refactor`: Melhorias arquiteturais
- `research`: Investigação necessária

## Processo de Desenvolvimento

### 1. Fork e Clone

```bash
# Fork o repositório no GitHub (clique em Fork)

# Clone seu fork
git clone https://github.com/SEU-USUARIO/Excel-ai.git
cd Excel-ai

# Adicione o repositório original como upstream
git remote add upstream https://github.com/sshturbo/Excel-ai.git
```

### 2. Crie um Branch

```bash
# Atualize seu fork
git checkout main
git pull upstream main

# Crie um branch para sua contribuição
git checkout -b feature/minha-feature

# Ou para bug fix
git checkout -b fix/nome-do-bug
```

**Nomenclatura de Branches**:
- `feature/`: Nova funcionalidade
- `fix/`: Correção de bug
- `docs/`: Mudanças em documentação
- `refactor/`: Refatoração de código
- `test/`: Adicionar/melhorar testes
- `perf/`: Melhorias de performance

### 3. Faça Suas Mudanças

```bash
# Instale dependências (primeira vez)
go mod download
cd frontend && npm install && cd ..

# Desenvolva com hot-reload
wails dev

# Execute testes
go test ./...
cd frontend && npm test && cd ..
```

### 4. Commit

Siga o [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git add .
git commit -m "feat: adiciona suporte a gráficos de dispersão"
git commit -m "fix: corrige erro ao salvar conversa vazia"
git commit -m "docs: atualiza guia de instalação"
```

**Formato**:
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat`: Nova funcionalidade
- `fix`: Correção de bug
- `docs`: Documentação
- `style`: Formatação (não afeta lógica)
- `refactor`: Refatoração
- `perf`: Melhoria de performance
- `test`: Testes
- `chore`: Manutenção
- `ci`: CI/CD

**Exemplos**:
```
feat(excel): add support for pivot table creation

Implemented full pivot table support via COM automation.
Added parser for pivot table actions from AI responses.

Closes #123
```

```
fix(chat): prevent memory leak in streaming

Fixed goroutine leak when canceling message stream.
Now properly closes channels and cleans up resources.

Fixes #456
```

### 5. Push e Pull Request

```bash
# Push para seu fork
git push origin feature/minha-feature

# Abra Pull Request no GitHub
# Vá para: https://github.com/sshturbo/Excel-ai/pulls
```

**Template de PR**:
```markdown
## Descrição
Breve descrição das mudanças

## Motivação
Por que essa mudança é necessária?

## Mudanças
- [ ] Mudança 1
- [ ] Mudança 2
- [ ] Mudança 3

## Screenshots (se aplicável)
[Adicionar capturas de tela]

## Testes
- [ ] Testes unitários passam
- [ ] Testes manuais realizados
- [ ] Documentação atualizada

## Checklist
- [ ] Código segue os padrões do projeto
- [ ] Comentários adicionados onde necessário
- [ ] Documentação atualizada
- [ ] Sem warnings de lint
- [ ] Testes adicionados/atualizados
- [ ] Build passa sem erros

## Issues Relacionadas
Closes #123
```

### 6. Code Review

Seu PR será revisado por mantenedores. Esteja preparado para:
- Responder perguntas
- Fazer ajustes solicitados
- Discutir decisões de design
- Iterar sobre a solução

### 7. Merge

Após aprovação, um mantenedor fará o merge do seu PR.

## Padrões de Código

### Go

#### Style Guide

Siga as convenções padrão de Go:
- `gofmt` para formatação
- `golint` para linting
- [Effective Go](https://golang.org/doc/effective_go.html)

**Exemplo**:
```go
package mypackage

import (
    "context"
    "fmt"
    
    "github.com/external/package"
    
    "excel-ai/internal/domain"
)

// Service descreve o que o serviço faz.
type Service struct {
    client *http.Client
    config Config
}

// NewService cria uma nova instância do Service.
func NewService(cfg Config) *Service {
    return &Service{
        client: &http.Client{Timeout: 30 * time.Second},
        config: cfg,
    }
}

// DoSomething executa uma operação.
func (s *Service) DoSomething(ctx context.Context, param string) error {
    if param == "" {
        return fmt.Errorf("param cannot be empty")
    }
    
    // Lógica aqui
    return nil
}
```

#### Comentários

```go
// Package-level comment
package mypackage

// Exported types/functions devem ter comentários
// que comecem com o nome do elemento.

// Service provides functionality for X.
type Service struct {
    // campos privados não precisam de comentário (opcional)
    client *http.Client
}

// NewService creates and initializes a Service.
func NewService() *Service {
    return &Service{}
}
```

#### Error Handling

```go
// ✅ Bom: Wrap errors com contexto
if err != nil {
    return fmt.Errorf("failed to process data: %w", err)
}

// ❌ Ruim: Perder contexto
if err != nil {
    return err
}

// ✅ Bom: Retornar cedo
if invalid {
    return errors.New("invalid input")
}
// continua normalmente

// ❌ Ruim: Else desnecessário
if valid {
    // código
} else {
    return errors.New("invalid")
}
```

### TypeScript/React

#### Style Guide

- Use TypeScript estrito
- Componentes funcionais com hooks
- Props tipadas com interfaces
- Prettier para formatação

**Exemplo**:
```typescript
import { useState, useEffect } from 'react'

interface MyComponentProps {
    title: string
    count?: number
    onAction?: (value: string) => void
}

export function MyComponent({ 
    title, 
    count = 0, 
    onAction 
}: MyComponentProps) {
    const [value, setValue] = useState<string>('')
    
    useEffect(() => {
        // Side effect
        console.log('Component mounted')
        
        return () => {
            // Cleanup
            console.log('Component unmounted')
        }
    }, [])
    
    const handleClick = () => {
        onAction?.(value)
    }
    
    return (
        <div className="flex flex-col gap-4">
            <h2 className="text-xl font-bold">{title}</h2>
            <p>Count: {count}</p>
            <button onClick={handleClick}>
                Action
            </button>
        </div>
    )
}
```

#### Hooks Customizados

```typescript
// hooks/useMyFeature.ts
export function useMyFeature(param: string) {
    const [data, setData] = useState<Data | null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    
    useEffect(() => {
        async function fetchData() {
            setLoading(true)
            try {
                const result = await fetchFromAPI(param)
                setData(result)
                setError(null)
            } catch (err) {
                setError(err.message)
            } finally {
                setLoading(false)
            }
        }
        
        fetchData()
    }, [param])
    
    return { data, loading, error }
}
```

### CSS/Tailwind

```typescript
// ✅ Bom: Classes organizadas, semânticas
<div className="flex flex-col items-center gap-4 p-6 bg-white rounded-lg shadow-md">
    <h1 className="text-2xl font-bold text-gray-900">Title</h1>
    <p className="text-sm text-gray-600">Description</p>
</div>

// ❌ Ruim: Classes bagunçadas
<div className="flex p-6 bg-white flex-col gap-4 rounded-lg items-center shadow-md">
```

## Commits e Pull Requests

### Commits

**Bons commits**:
- Atômicos (uma mudança lógica)
- Descritivos
- Testáveis individualmente

**Exemplo**:
```bash
git commit -m "feat(excel): add chart creation"
git commit -m "test(excel): add tests for chart creation"
git commit -m "docs(api): document chart creation API"
```

### Pull Requests

**Título**:
```
feat: add support for Excel pivot tables
fix: resolve memory leak in streaming
docs: update installation guide for macOS
```

**Descrição**:
- O quê: O que foi mudado
- Por quê: Por que a mudança é necessária
- Como: Como foi implementado (se complexo)
- Testes: Como foi testado

**Tamanho**:
- Prefira PRs pequenos e focados
- Se grande, explique por que não pode ser quebrado
- Considere múltiplos PRs sequenciais

## Reportando Bugs

### Antes de Reportar

1. Verifique se já não existe issue similar
2. Tente reproduzir com versão mais recente
3. Colete informações necessárias

### Template de Bug Report

```markdown
## Descrição
Descrição clara do bug

## Passos para Reproduzir
1. Abrir Excel com planilha X
2. Executar comando Y
3. Observar erro Z

## Comportamento Esperado
O que deveria acontecer

## Comportamento Atual
O que acontece atualmente

## Screenshots
[Se aplicável]

## Ambiente
- OS: Windows 10
- Excel: Office 2021
- Excel-ai: v1.2.3
- Provider: Groq
- Model: llama-3.1-70b

## Logs
```
[Colar logs relevantes]
```

## Informações Adicionais
Qualquer contexto adicional
```

## Sugerindo Funcionalidades

### Template de Feature Request

```markdown
## Funcionalidade
Descrição clara da funcionalidade desejada

## Motivação
Por que essa funcionalidade seria útil?
Que problema resolve?

## Solução Proposta
Como imagina que funcionaria?

## Alternativas Consideradas
Outras formas de resolver o problema

## Impacto
- Usuários afetados: ...
- Complexidade estimada: ...
- Breaking changes: Sim/Não

## Exemplos
Exemplos de uso da funcionalidade
```

## Primeiros Passos

### Issues para Iniciantes

Procure por:
- `good first issue`
- `documentation`
- `help wanted`

### Áreas que Precisam de Ajuda

1. **Documentação**
   - Traduzir para outros idiomas
   - Adicionar mais exemplos
   - Melhorar explicações

2. **Testes**
   - Aumentar cobertura de testes
   - Adicionar testes de integração
   - Testes de UI

3. **Features**
   - Suporte a mais tipos de gráficos
   - Melhorias na UI
   - Novos comandos de Excel

4. **Performance**
   - Otimizar operações COM
   - Melhorar streaming
   - Reduzir uso de memória

### Mentoria

Se você é novo no projeto:
- Comente na issue que quer trabalhar
- Peça orientação se necessário
- Não tenha medo de fazer perguntas

## Processo de Review

### O que Revisores Procuram

- **Correção**: O código faz o que promete?
- **Testes**: Há testes adequados?
- **Documentação**: Está documentado?
- **Estilo**: Segue os padrões?
- **Performance**: É eficiente?
- **Segurança**: Há vulnerabilidades?

### Respondendo a Feedback

- Seja respeitoso e profissional
- Faça perguntas se não entender
- Explique suas decisões quando apropriado
- Implemente mudanças solicitadas
- Marque conversas como resolvidas

## Licença

Ao contribuir, você concorda que suas contribuições serão licenciadas sob a mesma licença do projeto.

## Agradecimentos

Obrigado por contribuir com o Excel-ai! Cada contribuição, não importa o tamanho, é valiosa e apreciada.

## Contato

- **Issues**: https://github.com/sshturbo/Excel-ai/issues
- **Discussions**: https://github.com/sshturbo/Excel-ai/discussions
- **Email**: jefferson@hiposystem.com.br

## Recursos Adicionais

- [Development Guide](DEVELOPMENT.md)
- [Architecture](ARCHITECTURE.md)
- [API Documentation](API.md)
