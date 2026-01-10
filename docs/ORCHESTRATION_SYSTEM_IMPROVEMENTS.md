# Melhorias Sistêmicas do Orquestrador - Excel-AI

## Visão Geral

Este documento descreve as 4 melhorias sistêmicas críticas implementadas no orquestrador para aumentar a confiabilidade e estabilidade do sistema.

## 1. Invalidation de Cache para Ações Mutáveis

### Problema Original
O cache era usado apenas para consultas, mas quando uma ação `write_*` ocorria, dados cacheados poderiam ficar obsoletos, levando o LLM a tomar decisões baseadas em informações incorretas.

### Solução Implementada

#### Sistema de Tags de Invalidation
Cada entrada no cache agora possui tags que permitem invalidação seletiva:

```go
type CacheEntry struct {
    Result      string
    Error       error
    StoredAt    time.Time
    AccessCount int
    TTL         time.Duration
    Tags        []string // Novo: Tags para invalidação
}
```

#### Tags Geradas Automaticamente
- `tool:nome_da_ferramenta` - Tag genérica da ferramenta
- `sheet:NomeDaPlanilha` - Se o argumento `sheet` estiver presente
- `workbook:NomeDoWorkbook` - Se o argumento `workbook` estiver presente
- `range:A1:B10` - Se o argumento `range` estiver presente

#### Lógica de Invalidation
Quando uma ação mutável (`write_*`, `create_*`, `delete_*`) é executada:
1. Gera tags da ação (ex: `write_range` com `sheet:Dados` → `["tool:write_range", "sheet:Dados"]`)
2. Verifica todas as entradas do cache
3. Remove entradas que têm tags correspondentes (exceto tags genéricas de `tool:`)

#### Exemplo Prático
```go
// Consulta cacheada
get_range_values(sheet: "Dados", range: "A1:B10")
→ Cacheado com tags: ["tool:get_range_values", "sheet:Dados", "range:A1:B10"]

// Ação mutável executada
write_range(sheet: "Dados", range: "A1:B10", values: [...])
→ Invalida cache com tag "sheet:Dados" ou "range:A1:B10"

// Próxima consulta
get_range_values(sheet: "Dados", range: "A1:B10")
→ Não está mais no cache (foi invalidado)
→ Dados atualizados são buscados do Excel
```

#### Métodos Implementados
- `generateCacheTags(toolName, args) []string` - Gera tags de invalidação
- `invalidateCacheForAction(toolName, args)` - Invalida cache relacionado a ação
- `shouldInvalidate(entryTags, actionTags) bool` - Determina se deve invalidar
- `GetCacheStatus() CacheStatus` - Retorna status do cache

#### Benefícios
✅ **Consistência de Dados**: LLM sempre toma decisões baseadas em dados atualizados
✅ **Invalidação Inteligente**: Remove apenas entradas relacionadas, não todo o cache
✅ **Tags Flexíveis**: Suporta múltiplos níveis de granularidade (tool, sheet, range)
✅ **Log Detalhado**: Registra invalidações para debugging

#### Logs do Sistema
```
[CACHE] Set: get_range_values (TTL: 5m0s, tags: [tool:get_range_values sheet:Dados range:A1:B10])
[CACHE] Hit: get_range_values (acessos: 5)
[Worker] Executando write_range: task-003
[CACHE] Invalidação: 3 entradas removidas (tags: [tool:write_range sheet:Dados range:A1:B10])
```

---

## 2. Memoização de Falhas no Recovery

### Problema Original
O recovery automático tentava executar tarefas falhas repetidamente sem memória, podendo causar loops infinitos se a falha fosse estrutural (ex: Excel bloqueado).

### Solução Implementada

#### Sistema de Memoização de Falhas
```go
type FailureRecord struct {
    TaskID      string
    FailCount   int
    LastFailure time.Time
    LastError   error
    IsRecurrent bool // True se falhou 3+ vezes
}
```

#### Lógica de Memoização
1. **Registro de Falha**: Cada falha é registrada com timestamp
2. **Contagem de Tentativas**: Incrementa contador de falhas por tarefa
3. **Marcação como Recorrente**: Se falhar 3+ vezes, marca como recorrente
4. **Prevenção de Execução**: Tarefas recorrentes não são executadas novamente

#### Exemplo Prático
```go
// Primeira tentativa
executeTask("get_range_values", ...)
→ Erro: Excel bloqueado
→ FailureRecord: {FailCount: 1, IsRecurrent: false}

// Segunda tentativa
executeTask("get_range_values", ...)
→ Erro: Excel bloqueado
→ FailureRecord: {FailCount: 2, IsRecurrent: false}

// Terceira tentativa
executeTask("get_range_values", ...)
→ Erro: Excel bloqueado
→ FailureRecord: {FailCount: 3, IsRecurrent: true}
→ Log: "Tarefa marcada como falha recorrente (3 tentativas)"

// Quarta tentativa
executeTask("get_range_values", ...)
→ Verifica isRecurrentFailure() → true
→ Retorna imediatamente sem executar
→ Log: "Falha recorrente detectada, ignorando"
```

#### Limpeza em Sucesso
Se uma tarefa falhou anteriormente mas agora tem sucesso, o registro é limpo:

```go
if err == nil {
    o.clearFailureMemo(task) // Limpa registro de falhas
}
```

#### Métodos Implementados
- `isRecurrentFailure(task) bool` - Verifica se tarefa falhou recorrentemente
- `recordFailure(task, err)` - Registra falha da tarefa
- `clearFailureMemo(task)` - Limpa registro de falha (chamado em sucesso)
- `getFailureCount(task) int` - Retorna número de falhas
- `generateTaskKey(task) string` - Gera chave única para tarefa
- `GetFailureStats() map[string]interface{}` - Retorna estatísticas de falhas

#### Benefícios
✅ **Prevenção de Loops Infinitos**: Tarefas com falha estrutural não são reexecutadas
✅ **Backoff Implícito**: 3 tentativas antes de marcar como recorrente
✅ **Logging Detalhado**: Registra todas as falhas e tentativas
✅ **Auto-recovery**: Limpa registros quando tarefa tem sucesso
✅ **Estatísticas**: Acompanha total de falhas e recorrentes

#### Logs do Sistema
```
[FAILURE MEMO] Tarefa task-001 marcada como falha recorrente (3 tentativas)
[Failure Memo] Falha recorrente detectada, ignorando
[Failure Memo] Registro limpo após sucesso
```

---

## 3. Modo Degradado do Sistema

### Problema Original
Sistema era apenas "saudável" ou "com problemas", não havendo estado intermediário para situações onde ainda é funcional mas com limitações.

### Solução Implementada

#### Três Modos de Operação
```go
type OperationMode int

const (
    ModeNormal   OperationMode = iota // 100% funcional
    ModeDegraded                      // 50-75% funcional, reduzir paralelismo
    ModeCritical                      // < 50% funcional, modo emergencial
)
```

#### Critérios de Modo
Baseado na taxa de sucesso das tarefas:

| Taxa de Sucesso | Modo           | Ações                                          |
|------------------|----------------|-------------------------------------------------|
| ≥ 75%            | Normal         | 5 workers, TTL 5min, todas as tarefas         |
| 50-75%           | Degradado      | 3 workers, TTL 10min, apenas tarefas essenciais |
| < 50%            | Crítico        | 1 worker, TTL 30min, apenas tarefas urgentes    |

#### Ajustes Automáticos por Modo

**Modo Normal (100% funcional):**
```go
cacheTTL = 5 * time.Minute
workers = 5 (todos ativos)
availableTasks = [
    "list_sheets", "get_range_values", "query_batch",
    "write_cell", "write_range", "create_sheet",
    "format_range", "create_chart", "create_pivot_table"
]
```

**Modo Degradado (50-75% funcional):**
```go
cacheTTL = 10 * time.Minute
workers = 3 (paralelismo reduzido)
availableTasks = [
    "list_sheets", "get_range_values",
    "write_cell", "write_range"
]
```

**Modo Crítico (< 50% funcional):**
```go
cacheTTL = 30 * time.Minute
workers = 1 (paralelismo mínimo)
availableTasks = [
    "list_sheets",
    "write_cell"
]
```

#### Monitor Automático
O modo é avaliado a cada 10 segundos automaticamente:

```go
func (o *Orchestrator) operationModeMonitor() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-o.ctx.Done():
            return
        case <-ticker.C:
            o.evaluateOperationMode()
        }
    }
}
```

#### Métodos Implementados
- `evaluateOperationMode()` - Avalia e ajusta modo baseado em métricas
- `applyOperationMode(mode)` - Aplica configurações do modo
- `GetOperationMode() OperationMode` - Retorna modo atual
- `GetOperationModeName() string` - Retorna nome do modo
- `operationModeMonitor()` - Monitora e ajusta modo automaticamente

#### Benefícios
✅ **Adaptabilidade**: Sistema se ajusta automaticamente a problemas
✅ **Estabilidade Percebida**: Mesmo com problemas, sistema continua funcional
✅ **Redução de Carga**: Menos workers em modo degradado = menos conflitos
✅ **Foco em Essencial**: Modo crítico prioriza apenas tarefas urgentes
✅ **Transparência**: Logs indicam mudanças de modo

#### Logs do Sistema
```
[MODE] Modo de operação alterado: Normal -> Degradado
[MODE] Modo Degradado: paralelismo reduzido, TTL aumentado
[MODE] Modo de operação alterado: Degradado -> Crítico
[MODE] Modo Crítico: paralelismo mínimo, cache desativado
```

---

## 4. Snapshot de Decisão para o LLM

### Problema Original
O LLM tomava decisões baseadas em stats que mudavam durante execução paralela, podendo levar a decisões inconsistentes.

### Solução Implementada

#### Snapshot Imutável por Ciclo
```go
type DecisionSnapshot struct {
    Timestamp      time.Time
    OperationMode  OperationMode
    Stats          OrchestratorStats
    Health         HealthStatus
    CacheStatus    CacheStatus
    PendingTasks   int
    AvailableTasks []string // Tarefas disponíveis baseadas no modo
}
```

#### Como Funciona
1. **Captura no Início**: No início de cada ciclo de orquestração, captura snapshot
2. **Imutável Durante Ciclo**: Snapshot não muda durante execução das tarefas
3. **LLM Decide com Snapshot**: LLM usa snapshot consolidado para decisões
4. **Novo Snapshot no Próximo Ciclo**: Apenas atualiza no próximo ciclo

#### Exemplo de Uso
```go
// Início do ciclo de orquestração
snapshot := o.captureDecisionSnapshot()

// Snapshot contém estado consolidado:
// - OperationMode: Normal
// - Stats: {TotalTasks: 100, SuccessRate: 92%}
// - Health: {IsHealthy: true, WorkersActive: 5}
// - CacheStatus: {TotalEntries: 47, HitRate: 78%}
// - PendingTasks: 12
// - AvailableTasks: ["list_sheets", "get_range_values", ...]

// LLM decide baseado no snapshot
// Mesmo que stats mudem durante execução paralela,
// LLM usa o snapshot imutável
```

#### Tarefas Disponíveis por Modo
O snapshot inclui lista de tarefas disponíveis baseada no modo:

```go
switch snapshot.OperationMode {
case ModeNormal:
    snapshot.AvailableTasks = [
        "list_sheets", "get_range_values", "query_batch",
        "write_cell", "write_range", "create_sheet",
        "format_range", "create_chart", "create_pivot_table"
    ]
case ModeDegraded:
    snapshot.AvailableTasks = [
        "list_sheets", "get_range_values",
        "write_cell", "write_range"
    ]
case ModeCritical:
    snapshot.AvailableTasks = [
        "list_sheets",
        "write_cell"
    ]
}
```

#### Métodos Implementados
- `captureDecisionSnapshot() DecisionSnapshot` - Captura snapshot imutável
- `GetDecisionSnapshot() DecisionSnapshot` - Retorna snapshot atual

#### Benefícios
✅ **Decisões Consistentes**: LLM sempre vê estado consolidado
✅ **Evita Race Conditions**: Não há estado mutável durante decisões
✅ **Previsível**: Comportamento determinístico por ciclo
✅ **Debugável**: Snapshot pode ser inspecionado para debugging
✅ **Contexto Rico**: Inclui stats, health, cache e tarefas disponíveis

---

## Integração com executeTask

Todas as 4 melhorias são integradas no método `executeTask`:

```go
func (o *Orchestrator) executeTask(task *Task, reportProgress func(string) error) *TaskResult {
    start := time.Now()

    // 1. Verificar memoização de falhas antes de executar
    if o.isRecurrentFailure(task) {
        return &TaskResult{
            Success: false,
            Error:   fmt.Errorf("tarefa falhou recorrentemente (%d tentativas)", o.getFailureCount(task)),
        }
    }

    // 2. Tentar obter do cache
    if task.Type == TaskTypeQuery {
        if cached, found := o.getFromCache(task.ToolName, task.Arguments); found {
            return &TaskResult{
                Success: true,
                Result:  cached,
            }
        }
    }

    // 3. Executar tarefa
    result, err := o.service.executeToolCall(task.ToolName, task.Arguments)

    // 4. Invalidar cache para ações mutáveis
    if err == nil && task.Type == TaskTypeAction {
        o.invalidateCacheForAction(task.ToolName, task.Arguments)
    }

    // 5. Armazenar no cache (apenas para consultas bem-sucedidas)
    if err == nil && task.Type == TaskTypeQuery {
        o.setInCache(task.ToolName, task.Arguments, result)
    }

    // 6. Atualizar estatísticas e memoização de falhas
    if err == nil {
        o.successTasks++
        o.clearFailureMemo(task) // Limpa registro de falha
    } else {
        o.failedTasks++
        o.recordFailure(task, err) // Registra falha
    }

    return &TaskResult{
        Success: err == nil,
        Result:  result,
        Error:   err,
    }
}
```

---

## API Pública Nova

### Métodos Adicionados

#### Cache
```go
// Retorna status do cache
func (o *Orchestrator) GetCacheStatus() CacheStatus

// Limpa todo o cache
func (o *Orchestrator) ClearCache()
```

#### Falhas
```go
// Retorna estatísticas de falhas
func (o *Orchestrator) GetFailureStats() map[string]interface{}
```

#### Modo de Operação
```go
// Retorna modo de operação atual
func (o *Orchestrator) GetOperationMode() OperationMode

// Retorna nome do modo de operação
func (o *Orchestrator) GetOperationModeName() string
```

#### Snapshot
```go
// Retorna snapshot de decisão atual
func (o *Orchestrator) GetDecisionSnapshot() DecisionSnapshot
```

### Estruturas de Dados

#### CacheStatus
```go
type CacheStatus struct {
    TotalEntries  int     // Total de entradas no cache
    HitRate       float64 // Taxa de acerto do cache (%)
    Invalidations int64   // Total de invalidações
    LastCleanup   time.Time // Timestamp da última limpeza
}
```

#### FailureStats (map)
```go
{
    "total_memoized": 10,  // Total de tarefas memoizadas
    "total_failures":  25,  // Total de falhas registradas
    "recurrent":       3     // Tarefas com falha recorrente
}
```

#### DecisionSnapshot
```go
type DecisionSnapshot struct {
    Timestamp      time.Time
    OperationMode  OperationMode
    Stats          OrchestratorStats
    Health         HealthStatus
    CacheStatus    CacheStatus
    PendingTasks   int
    AvailableTasks []string
}
```

---

## Casos de Uso

### Caso 1: Cache Consistente em Ações Mutáveis

**Cenário:**
1. Usuário consulta dados da planilha "Dados"
2. Sistema cacheia resultado com tag "sheet:Dados"
3. Usuário modifica dados da planilha "Dados" via `write_range`
4. Sistema invalida cache com tag "sheet:Dados"
5. Usuário consulta novamente
6. Sistema retorna dados atualizados (não do cache obsoleto)

**Resultado:** LLM sempre toma decisões baseadas em dados atualizados

---

### Caso 2: Prevenção de Loop Infinito

**Cenário:**
1. Excel está bloqueado/fechado
2. Sistema tenta executar `get_range_values`
3. Falha 1x: registrada, contagem = 1
4. Falha 2x: registrada, contagem = 2
5. Falha 3x: marcada como recorrente
6. Tentativa 4x: não executa, retorna imediatamente

**Resultado:** Sistema não entra em loop infinito, economiza recursos

---

### Caso 3: Adaptação a Problemas

**Cenário:**
1. Sistema funcionando normalmente (taxa de sucesso: 95%)
2. Começam a ocorrer falhas (taxa cai para 65%)
3. Monitor detecta mudança, ajusta para Modo Degradado
4. Reduz paralelismo de 5 para 3 workers
5. Aumenta TTL de cache de 5min para 10min
6. Limita tarefas disponíveis para apenas essenciais
7. Taxa de sucesso volta para 80%
8. Sistema se ajusta automaticamente para Modo Normal

**Resultado:** Sistema se adapta automaticamente a problemas, mantém funcionalidade

---

### Caso 4: Decisões Consistentes do LLM

**Cenário:**
1. Usuário solicita "Analise vendas e crie gráfico"
2. Sistema captura snapshot no início do ciclo
3. LLM decide executar `get_range_values` e `create_chart` em paralelo
4. Durante execução paralela, stats mudam (sucesso aumenta)
5. LLM continua usando snapshot original, não stats em tempo real
6. Gera resposta coerente baseada em snapshot consolidado

**Resultado:** Decisões consistentes, sem race conditions

---

## Monitoramento e Debugging

### Logs do Sistema

Todos os sistemas produzem logs detalhados:

```
[CACHE] Set: get_range_values (TTL: 5m0s, tags: [tool:get_range_values sheet:Dados])
[CACHE] Hit: get_range_values (acessos: 5)
[CACHE] Invalidação: 3 entradas removidas (tags: [sheet:Dados])
[FAILURE MEMO] Tarefa task-001 marcada como falha recorrente (3 tentativas)
[MODE] Modo de operação alterado: Normal -> Degradado
[MODE] Modo Degradado: paralelismo reduzido, TTL aumentado
[SNAPSHOT] Capturado: Mode=Normal, SuccessRate=92%, Workers=5
```

### Métricas Disponíveis

Via `GetStats()`:
- TotalTasks, SuccessTasks, FailedTasks
- ActiveWorkers, AvgTaskTime, SuccessRate

Via `GetCacheStatus()`:
- TotalEntries, HitRate, Invalidations

Via `GetFailureStats()`:
- total_memoized, total_failures, recurrent

Via `GetDecisionSnapshot()`:
- Estado consolidado completo do sistema

---

## Melhorias Futuras Planejadas

### 1. Ajuste Dinâmico de Workers
- Ajustar número de workers em tempo real baseado na carga
- Escalar horizontalmente quando necessário

### 2. Cache Distribuído
- Compartilhar cache entre sessões
- Persistir cache em disco para recuperação rápida

### 3. Machine Learning de Predição
- Prever falhas baseadas em padrões históricos
- Ajustar automaticamente estratégias de execução

### 4. Dashboard Avançado
- Visualizar modo de operação em tempo real
- Gráficos de transição entre modos
- Alertas de mudanças críticas

---

## Conclusão

As 4 melhorias sistêmicas implementadas transformaram o orquestrador em um sistema:

1. **Mais Confiável**: Cache consistente e prevenção de loops infinitos
2. **Mais Adaptável**: Ajusta automaticamente a problemas
3. **Mais Estável**: Modo degradado mantém funcionalidade mesmo com problemas
4. **Mais Consistente**: LLM toma decisões baseadas em snapshots imutáveis

O sistema agora é production-ready, com monitoramento completo e auto-recovery inteligente.