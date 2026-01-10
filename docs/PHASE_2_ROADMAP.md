# Roadmap Fase 2 - Próximo Salto Lógico

## Visão Geral

A Fase 1 implementou as 4 melhorias críticas de confiabilidade e resiliência. A Fase 2 foca em otimização de performance, eficiência cognitiva e arquitetura avançada.

---

## 1. Versionamento de Snapshots com Replay de Decisão 🔄

### Descrição
Sistema de versionamento de snapshots com capacidade de replay de decisões passadas.

### Funcionalidades

#### Snapshot ID Incremental
```go
type Snapshot struct {
    ID          int64     // ID incremental único
    Timestamp   time.Time
    Decision    string     // Decisão tomada pelo LLM
    Result      string     // Resultado da execução
    Success     bool       // Se a decisão foi bem-sucedida
    Mode        OperationMode
}

// Gerar snapshot versionado
snapshot := &Snapshot{
    ID:        nextSnapshotID(),
    Timestamp: time.Now(),
    Decision:  llmDecision,
    // ...
}
```

#### Replay de Decisão
```go
// Replay de decisão específica
func (o *Orchestrator) ReplayDecision(snapshotID int64) error {
    snapshot := o.getSnapshot(snapshotID)
    
    // Validar se contexto ainda é válido
    if !o.validateSnapshotContext(snapshot) {
        return errors.New("contexto inválido para replay")
    }
    
    // Executar mesma decisão
    return o.executeDecision(snapshot.Decision)
}

// Replay automático em falhas recorrentes
if failure.IsRecurrent {
    // Tentar replay de decisão bem-sucedida anterior
    lastSuccessful := o.getLastSuccessfulSnapshot(taskKey)
    if lastSuccessful != nil {
        return o.ReplayDecision(lastSuccessful.ID)
    }
}
```

### Benefícios
✅ **Auditoria completa**: Histórico completo de todas as decisões
✅ **Aprendizado automático**: Replay de decisões bem-sucedidas em situações similares
✅ **Debugging facilitado**: É possível reproduzir exatamente a mesma decisão
✅ **Rollback**: Voltar para snapshot anterior em caso de problemas
✅ **Análise de padrões**: Identificar decisões que sempre falham

### Prioridade: **Média** (Útil mas não crítico)
- Complexidade: Alta
- Impacto: Alto (longo prazo)
- Custo de implementação: Alto

---

## 2. Classificador Rápido Antes do LLM ⚡

### Descrição
Sistema de classificação rápida que reduz chamadas ao modelo principal usando heurísticas e regras determinísticas.

### Arquitetura

#### Três Camadas de Classificação

```go
type DecisionType int

const (
    DecisionTypeHeuristic DecisionType = iota // Regra determinística
    DecisionTypeCache                      // Do cache/histórico
    DecisionTypeLLM                        // Precisa de LLM
)

// Classificador rápido
func (o *Orchestrator) ClassifyRequest(message string) DecisionType {
    // Camada 1: Timeout rápido (50ms)
    if o.quickTimeoutCheck(message) {
        return DecisionTypeHeuristic
    }
    
    // Camada 2: Permissão rápida (100ms)
    if !o.quickPermissionCheck(message) {
        return DecisionTypeHeuristic
    }
    
    // Camada 3: Cache de decisões (150ms)
    if cached, found := o.getDecisionCache(message); found {
        return DecisionTypeCache
    }
    
    // Camada 4: Lógica simples (200ms)
    if o.simpleLogicCheck(message) {
        return DecisionTypeHeuristic
    }
    
    // Camada 5: LLM completo
    return DecisionTypeLLM
}
```

#### Timeout ≠ Permissão ≠ Lógica

```go
// Timeout rápido: operações muito simples
func (o *Orchestrator) quickTimeoutCheck(message string) bool {
    // Exemplos de decisões que não precisam de LLM:
    // - "Qual sheet está ativa?" → sheet_exists (cache)
    // - "Quantas células?" → get_range_values (cache)
    
    quickPatterns := []string{
        "qual sheet", "quais sheets", "lista sheets",
        "quantas células", "quantas linhas",
    }
    
    for _, pattern := range quickPatterns {
        if strings.Contains(strings.ToLower(message), pattern) {
            return true
        }
    }
    return false
}

// Permissão rápida: verificações de segurança
func (o *Orchestrator) quickPermissionCheck(message string) bool {
    // Verificar operações perigosas sem confirmação
    dangerousOps := []string{
        "deletar", "apagar", "remover", "destroy",
        "formatar tudo", "limpar tudo",
    }
    
    for _, op := range dangerousOps {
        if strings.Contains(strings.ToLower(message), op) {
            return false // Precisa de confirmação humana
        }
    }
    return true
}

// Lógica simples: regras determinísticas
func (o *Orchestrator) simpleLogicCheck(message string) bool {
    // Exemplos:
    // - "Criar gráfico dos dados atuais" → create_chart
    // - "Pivot dos dados da sheet" → create_pivot
    
    if strings.Contains(message, "gráfico") {
        return true // Decisão determinística
    }
    
    return false
}
```

#### Redução de Chamadas ao Modelo Principal

```go
func (o *Orchestrator) ProcessMessage(message string) (string, error) {
    decisionType := o.ClassifyRequest(message)
    
    switch decisionType {
    case DecisionTypeHeuristic:
        // Resposta instantânea (< 100ms)
        return o.applyHeuristic(message)
        
    case DecisionTypeCache:
        // Do cache (< 50ms)
        return o.getCachedDecision(message)
        
    case DecisionTypeLLM:
        // Chamada completa ao LLM (2-10s)
        return o.processWithLLM(message)
    }
}
```

### Métricas de Eficiência

```
Sem Classificador:
- 100% das chamadas → LLM
- Tempo médio: 5s
- Custo: $0.05/mensagem

Com Classificador (estimado):
- 40% heurística → < 100ms
- 30% cache → < 50ms
- 30% LLM → 5s
- Tempo médio: 1.6s (68% mais rápido)
- Custo: $0.015/mensagem (70% economia)
```

### Benefícios
✅ **Performance drástica**: 70% das requisições respondidas em < 100ms
✅ **Economia de custos**: Redução de 70% nas chamadas de API
✅ **Experiência do usuário**: Respostas quase instantâneas para casos comuns
✅ **Escalabilidade**: Sistema suporta muito mais usuários com mesmo hardware
✅ **Latência zero** para operações simples

### Prioridade: **Alta** (Impacto imediato em performance e custo)
- Complexidade: Média
- Impacto: Muito Alto
- Custo de implementação: Médio

---

## 3. Orçamento Cognitivo Adaptativo 🧠

### Descrição
Sistema que ajusta a complexidade dos prompts baseada no modo de operação e contexto.

### Arquitetura

#### Modo Crítico → Prompts Menores

```go
func (o *Orchestrator) buildPrompt(message string) string {
    mode := o.GetOperationMode()
    
    switch mode {
    case ModeCritical:
        // Prompt minimalista (< 200 tokens)
        return o.buildMinimalPrompt(message)
        
    case ModeDegraded:
        // Prompt enxuto (200-500 tokens)
        return o.buildLeanPrompt(message)
        
    case ModeNormal:
        // Prompt completo com raciocínio (500-1000 tokens)
        return o.buildFullPrompt(message)
    }
}

// Prompt minimalista para modo crítico
func (o *Orchestrator) buildMinimalPrompt(message string) string {
    return fmt.Sprintf(`Ação: %s
Contexto: %s

Responda apenas com a ferramenta a usar.
Formato: tool_name(args)
`, message, o.getMinimalContext())
}
```

#### Modo Normal → Raciocínio Completo

```go
func (o *Orchestrator) buildFullPrompt(message string) string {
    return fmt.Sprintf(`Você é um assistente Excel especializado.

SOLICITAÇÃO:
%s

CONTEXTO COMPLETO:
%s

CONSIDERAÇÕES:
- Analise os dados disponíveis
- Considere múltiplas abordagens
- Explique seu raciocínio
- Suger melhorias se aplicável

RESPOSTA:
1. Análise da situação
2. Ferramentas necessárias
3. Explicação do processo
4. Resultado esperado
`, message, o.getFullContext())
}
```

#### Orçamento Dinâmico

```go
type CognitiveBudget struct {
    MaxTokens      int  // Limite de tokens
    AllowReasoning bool // Permite raciocínio estendido
    ToolComplexity int  // Nível de complexidade de ferramentas
}

func (o *Orchestrator) getCognitiveBudget() CognitiveBudget {
    stats := o.GetStats()
    mode := o.GetOperationMode()
    
    budget := CognitiveBudget{
        AllowReasoning: true,
        ToolComplexity: 3, // 1=simple, 3=complex
    }
    
    switch mode {
    case ModeCritical:
        budget.MaxTokens = 200
        budget.AllowReasoning = false
        budget.ToolComplexity = 1
        
    case ModeDegraded:
        budget.MaxTokens = 500
        budget.AllowReasoning = false
        budget.ToolComplexity = 2
        
    case ModeNormal:
        if stats.SuccessRate > 90 {
            // Sistema muito saudável → orçamento generoso
            budget.MaxTokens = 1500
            budget.AllowReasoning = true
            budget.ToolComplexity = 3
        } else {
            // Sistema saudável mas não perfeito
            budget.MaxTokens = 800
            budget.AllowReasoning = true
            budget.ToolComplexity = 2
        }
    }
    
    return budget
}
```

### Benefícios

#### Performance
```
Modo Crítico:
- Tokens: 200 (vs 1000 normal)
- Tempo: 0.5s (vs 3s normal)
- Economia: 80%

Modo Degradado:
- Tokens: 500 (vs 1000 normal)
- Tempo: 1.5s (vs 3s normal)
- Economia: 50%

Modo Normal:
- Tokens: 800-1500
- Qualidade: Máxima
```

#### Resiliência
✅ Sistema continua funcional mesmo com recursos limitados
✅ Adaptação automática à carga do sistema
✅ Priorização de tarefas críticas

### Prioridade: **Média-Alta** (Otimiza mas não é crítico)
- Complexidade: Média
- Impacto: Alto
- Custo de implementação: Médio

---

## Comparativo de Prioridades

| Recurso | Prioridade | Impacto | Complexidade | Custo |
|----------|-----------|----------|-------------|-------|
| Classificador Rápido | Alta | Muito Alto | Média | Médio |
| Orçamento Cognitivo | Média-Alta | Alto | Média | Médio |
| Versionamento de Snapshots | Média | Alto | Alta | Alto |

---

## Recomendação de Implementação

### Fase 2.1 - Ganho Rápido (2-3 semanas)
**Foco: Classificador Rápido**

1. Implementar camada de timeout (semana 1)
2. Implementar camada de permissão (semana 1)
3. Implementar cache de decisões (semana 2)
4. Implementar lógica simples (semana 2)
5. Testes A/B com métricas (semana 3)

**ROI Esperado**: 70% redução em custo e 68% melhoria em latência

### Fase 2.2 - Otimização Adaptativa (1-2 semanas)
**Foco: Orçamento Cognitivo**

1. Implementar prompts modais (semana 1)
2. Implementar orçamento dinâmico (semana 1-2)
3. Testar transição entre modos (semana 2)

**ROI Esperado**: 50-80% economia em modo crítico/degradado

### Fase 2.3 - Arquitetura Avançada (3-4 semanas)
**Foco: Versionamento de Snapshots**

1. Implementar sistema de IDs (semana 1)
2. Implementar replay de decisões (semana 2)
3. Implementar auditoria (semana 3)
4. Implementar rollback (semana 3-4)

**ROI Esperado**: Melhoria em debugging e aprendizado automático

---

## Métricas de Sucesso

### Classificador Rápido
- [ ] 70% das requisições sem LLM
- [ ] Latência média < 2s
- [ ] Economia > 60% em custos de API
- [ ] Acurácia de decisões > 85%

### Orçamento Cognitivo
- [ ] Modo crítico: tempo de resposta < 1s
- [ ] Modo degradado: tempo de resposta < 2s
- [ ] Economia > 50% em tokens
- [ ] Manutenção de qualidade > 90%

### Versionamento de Snapshots
- [ ] Replay bem-sucedido em 80% dos casos
- [ ] Auditoria 100% das decisões
- [ ] Tempo de replay < 500ms
- [ ] Redução em decisões recorrentes

---

## Conclusão

As três ideias propostas são excelentes e complementares:

1. **Classificador Rápido** - Maior impacto imediato, implementação mais rápida
2. **Orçamento Cognitivo** - Melhoria contínua, adaptação inteligente
3. **Versionamento de Snapshots** - Arquitetura avançada, benefícios de longo prazo

**Minha recomendação**: Começar com o Classificador Rápido (Fase 2.1), pois oferece o maior ROI no menor tempo. Depois, implementar o Orçamento Cognitivo (Fase 2.2) para otimização contínua. Versionamento de Snapshots pode ser implementado mais tarde quando houver mais histórico de decisões para analisar.

Essa abordagem permite obter benefícios incrementais rapidamente enquanto constrói uma arquitetura mais robusta ao longo do tempo.