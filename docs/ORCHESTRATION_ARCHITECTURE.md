t# Arquitetura de Orquestração Paralela - Excel-AI

## Visão Geral

O sistema de orquestração paralela permite que múltiplos modelos de IA trabalhem simultaneamente para executar tarefas do Excel de forma mais rápida e eficiente.

## Conceito

### Arquitetura Tradicional vs Orquestrada

**Tradicional (Sequencial):**
```
Usuário → Modelo → Função A → Modelo → Função B → Modelo → Função C
         ↑                                      ↓
         └────────────────────────────────────────┘
```

**Orquestrada (Paralela):**
```
Usuário → Orquestrador → Divide tarefas
                          ↓
                    ┌─────┴─────┐
                    ↓             ↓
                 Worker 1      Worker 2
                    ↓             ↓
                 Função A       Função B
                    ↓             ↓
                    └─────┬─────┘
                          ↓
                    Compila resultados
                          ↓
                    Gera resposta final
```

## Componentes

### 1. Orquestrador (Modelo A)

**Responsabilidade:**
- Analisa a solicitação do usuário
- Identifica tarefas independentes
- Divide em subtarefas executáveis
- Coordena a execução paralela
- Compila resultados de múltiplas tarefas
- Gera resposta final

**Funções:**
- `AnalyzeRequest()`: Analisa e divide a solicitação
- `ParseTasks()`: Extrai tarefas da resposta do orquestrador
- `ExecuteTask()`: Executa uma tarefa individual
- `GenerateFinalResponse()`: Cria resposta final baseada nos resultados

### 2. Workers Paralelos (Goroutines)

**Quantidade:** 5 workers por padrão (configurável)

**Responsabilidade:**
- Executam tarefas simultaneamente
- Reportam progresso em tempo real
- Retornam resultados para o coletor

**Características:**
- Não bloqueantes (non-blocking)
- Canal de comunicação (`taskChan`)
- Buffer de 100 tarefas
- Processamento assíncrono

### 3. Buffer de Mensagens

**Finalidade:**
- Maném o usuário informado enquanto tarefas executam
- Envia mensagens de progresso
- Heartbeat de tarefas pendentes

**Canais:**
- `messageChan`: Buffer de mensagens para UI
- `resultChan`: Resultados das tarefas
- `taskChan`: Fila de tarefas para workers

## Fluxo de Execução

### Passo 1: Análise da Solicitação

```
Usuário: "Analise as vendas, crie um gráfico e salve em nova planilha"

Orquestrador (Modelo A):
1. Identifica 3 tarefas independentes:
   - Tarefa 1: Consultar dados de vendas
   - Tarefa 2: Criar gráfico
   - Tarefa 3: Criar nova planilha
2. Verifica se podem ser paralelas
3. Gera JSON de tarefas
```

### Passo 2: Execução Paralela

```
Tarefas enviadas para 5 workers:

Worker 1: Executa "list_sheets" ✅
Worker 2: Executa "get_range_values" ✅  
Worker 3: Executa "create_chart" ✅
Worker 4: Executa "create_sheet" ✅
Worker 5: Aguardando...

Tempo total: ~3 segundos (vs 12 segundos sequencial)
```

### Passo 3: Buffer de Mensagens

Enquanto workers executam:

```
🎯 [Orquestrador] Analisando solicitação...
📋 [Orquestrador] 3 tarefas identificadas para execução paralela
⚙️ [Worker] Executando list_sheets: task-001
⚙️ [Worker] Executando get_range_values: task-002
⚙️ [Worker] Executando create_chart: task-003
✅ task-001 (0.8s)
✅ task-002 (2.1s)
✅ task-003 (2.5s)
📊 [Orquestrador] Compilando resultados...
🎉 [Orquestrator] 3/3 tarefas concluídas com sucesso
```

### Passo 4: Resposta Final

```
Modelo A recebe resultados:
- Planilhas: ["Dados", "Relatório"]
- Valores: 1500 linhas de dados
- Gráfico criado: "Vendas_2024"

Gera resposta coerente:
"Análise completa! Encontrei 1500 registros de vendas.
Criei o gráfico 'Vendas_2024' na planilha 'Dados'
e preparei a nova planilha 'Relatório' para o relatório final."
```

## Vantagens

### 1. Velocidade

**Exemplo Prático:**
- **Tradicional:** 10 tarefas × 2 segundos = 20 segundos
- **Orquestrado:** 10 tarefas / 5 workers = 4 segundos
- **Ganho:** 5x mais rápido

### 2. Eficiência de Recursos

**Tradicional:**
- Modelo ocioso enquanto aguarda resultados
- CPU/GPU subutilizados

**Orquestrado:**
- Workers sempre ativos
- Recursos otimizados
- Múltiplas requisições simultâneas

### 3. Melhor Experiência do Usuário

**Benefícios:**
- Feedback em tempo real
- Progresso visível
- Menor tempo de espera
- Respostas mais completas

### 4. Escalabilidade

**Ajustáveis:**
- Número de workers (padrão: 5)
- Tamanho do buffer (padrão: 100)
- Timeout por tarefa
- Priorização de tarefas

## Estrutura de Tarefas

### Formato JSON

```json
[
  {
    "tool": "get_range_values",
    "args": {
      "sheet": "Dados",
      "range": "A1:Z1000"
    },
    "priority": 1,
    "description": "Consultar dados de vendas"
  },
  {
    "tool": "create_chart",
    "args": {
      "sheet": "Dados",
      "range": "A1:B100",
      "chart_type": "bar",
      "title": "Vendas 2024"
    },
    "priority": 2,
    "description": "Criar gráfico de vendas"
  }
]
```

### Tipos de Tarefas

```go
type TaskType int

const (
    TaskTypeQuery TaskType = iota    // Consultas: get_range_values, list_sheets
    TaskTypeAction                    // Ações: write_range, create_sheet
    TaskTypeOrchestration             // Orquestração: análise de tarefas
)
```

### Prioridade

- **1:** Urgente - executar primeiro
- **2:** Normal - padrão
- **3:** Baixa - executar por último

## Configuração

### Backend (Go)

```go
// No serviço de chat
service.SetOrchestration(true)    // Habilitar orquestração
service.StartOrchestrator(ctx)     // Iniciar workers
service.StopOrchestrator()        // Parar workers
```

### Frontend (TypeScript)

```typescript
// Via Wails
await window.go.main.SetOrchestration(true)
await window.go.main.StartOrchestrator()
await window.go.main.StopOrchestrator()
const isEnabled = await window.go.main.GetOrchestration()
```

### Arquivo de Configuração

```json
{
  "useOrchestration": true,
  "orchestrationWorkers": 5,
  "orchestrationBufferSize": 100,
  "orchestrationTimeout": 30
}
```

## Casos de Uso

### Caso 1: Consultas Múltiplas

**Solicitação:**
"Mostre as vendas, os produtos e as planilhas disponíveis"

**Execução Paralela:**
```
Worker 1: list_sheets → ["Dados", "Relatório", "Resumo"]
Worker 2: get_range_values → 500 linhas de vendas
Worker 3: list_tables → ["Vendas", "Produtos"]
```

**Resultado:** 1.5 segundos (vs 4.5 segundos sequencial)

### Caso 2: Criação de Relatórios

**Solicitação:**
"Crie um relatório completo com dados, gráficos e resumo"

**Execução Paralela:**
```
Worker 1: Copiar dados para nova planilha
Worker 2: Criar gráfico de vendas
Worker 3: Criar gráfico de produtos
Worker 4: Criar tabela dinâmica
Worker 5: Formatar cabeçalhos
```

**Resultado:** 5 segundos (vs 20 segundos sequencial)

### Caso 3: Análise de Dados

**Solicitação:**
"Analise as vendas por região e crie um dashboard"

**Execução Paralela:**
```
Worker 1: Filtrar por região Norte
Worker 2: Filtrar por região Sul
Worker 3: Filtrar por região Leste
Worker 4: Filtrar por região Oeste
Worker 5: Calcular totais
```

**Resultado:** 3 segundos (vs 12 segundos sequencial)

## Limitações e Considerações

### Quando Usar Orquestração

**✅ Use quando:**
- Múltiplas tarefas independentes
- Tarefas podem ser paralelas
- Velocidade é crítica
- Usuário quer feedback rápido

**❌ Não use quando:**
- Tarefas sequenciais dependentes
- Ações que modificam as mesmas células
- Requer confirmação manual
- Tarefa simples única

### Dependências de Tarefas

**Independentes (Paralelo):**
```
✅ Consultar planilhas
✅ Criar gráfico A
✅ Criar gráfico B
✅ Listar tabelas
```

**Dependentes (Sequencial):**
```
❌ Criar planilha → Escrever dados (requer planilha criada)
❌ Deletar planilha → Criar nova (conflito)
❌ Filtrar dados → Criar gráfico (depende dos dados filtrados)
```

### Conflitos de Recursos

**Problema:**
```
Worker 1: Escreve em A1:B10
Worker 2: Escreve em A1:B10
```

**Solução:**
- Orquestrador detecta conflitos
- Serializa tarefas conflitantes
- Executa em ordem de prioridade

## Monitoramento e Debug

### Logs do Sistema

```bash
[ORCHESTRATOR] ✅ Iniciado com 5 workers
[ORCHESTRATOR] Worker 0 processando tarefa task-001
[ORCHESTRATOR] Resultado recebido: task-001 (Success: true)
[ORCHESTRATOR] 💓 Tarefas pendentes: 3
[ORCHESTRATOR] 🛑 Parado
```

### Métricas

- **Tarefas Executadas:** Total de tarefas processadas
- **Taxa de Sucesso:** % de tarefas concluídas
- **Tempo Médio:** Tempo médio por tarefa
- **Workers Ativos:** Número de workers ocupados
- **Tarefas Pendentes:** Tarefas na fila

### Performance

**Benchmarks:**
- 10 tarefas paralelas: ~4 segundos
- 10 tarefas sequenciais: ~20 segundos
- Ganho de performance: **5x**

## Troubleshooting

### Problema: Workers não iniciam

**Solução:**
```go
// Verificar se orquestrador foi iniciado
err := service.StartOrchestrator(context.Background())
if err != nil {
    log.Fatal("Erro ao iniciar orquestrador:", err)
}
```

### Problema: Tarefas travam

**Solução:**
```go
// Verificar context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

service.StartOrchestrator(ctx)
```

### Problema: Resultados inconsistentes

**Solução:**
- Verificar dependências de tarefas
- Usar mutex para recursos compartilhados
- Implementar sistema de prioridades

## Melhorias Futuras

### Planejado

1. **Balanceamento Dinâmico de Carga**
   - Ajustar número de workers automaticamente
   - Redistribuir tarefas entre workers

2. **Cache de Resultados**
   - Armazenar resultados de tarefas comuns
   - Evitar reexecuções desnecessárias

3. **Priorização Inteligente**
   - Analisar criticidade das tarefas
   - Priorizar tarefas do usuário

4. **Recovery Automático**
   - Detectar workers travados
   - Reiniciar workers automaticamente
   - Reprocessar tarefas falhas

5. **Dashboard de Monitoramento**
   - Visualizar workers ativos
   - Mostrar fila de tarefas
   - Gráficos de performance em tempo real

## Conclusão

A arquitetura de orquestração paralela transforma a interação com o Excel-AI, proporcionando:

- **Velocidade:** 5x mais rápido em tarefas múltiplas
- **Eficiência:** Melhor uso de recursos
- **Experiência:** Feedback em tempo real
- **Escalabilidade:** Ajustável conforme necessário

Esta arquitetura é ideal para usuários que precisam executar múltiplas ações no Excel simultaneamente, como criar relatórios, analisar dados e formatar planilhas.