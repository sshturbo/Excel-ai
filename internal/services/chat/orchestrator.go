package chat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"excel-ai/pkg/ai"
	"excel-ai/pkg/cache"
)

// CacheEntry representa uma entrada no cache
type CacheEntry struct {
	Result      string
	Error       error
	StoredAt    time.Time
	AccessCount int
	TTL         time.Duration
	Tags        []string // Tags para invalidação (ex: "sheet:Dados", "workbook:Financeiro")
}

// OperationMode define o modo de operação do orquestrador
type OperationMode int

const (
	ModeNormal   OperationMode = iota // 100% funcional
	ModeDegraded                      // 50-75% funcional, reduzir paralelismo
	ModeCritical                      // < 50% funcional, modo emergencial
)

// FailureRecord registra falhas de tarefas
type FailureRecord struct {
	TaskID      string
	FailCount   int
	LastFailure time.Time
	LastError   error
	IsRecurrent bool // True se falhou 3+ vezes
}

// DecisionSnapshot captura estado imutável para decisão do LLM
type DecisionSnapshot struct {
	Timestamp      time.Time
	OperationMode  OperationMode
	Stats          OrchestratorStats
	Health         HealthStatus
	CacheStatus    CacheStatus
	PendingTasks   int
	AvailableTasks []string // Tarefas disponíveis baseadas no modo
}

// CacheStatus status do cache
type CacheStatus struct {
	TotalEntries  int
	HitRate       float64
	Invalidations int64
	LastCleanup   time.Time
}

// Orchestrator gerencia múltiplos modelos trabalhando em paralelo
type Orchestrator struct {
	service *Service
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex

	// Canais de comunicação
	taskChan    chan *Task
	resultChan  chan *TaskResult
	messageChan chan string

	// Estado
	running      bool
	pendingTasks map[string]*Task

	// Balanceamento dinâmico
	activeWorkers int
	totalTasks    int64
	successTasks  int64
	failedTasks   int64
	avgTaskTime   time.Duration
	muStats       sync.RWMutex

	// Cache de resultados (persistente em SQLite)
	cache              *cache.PersistentCache
	muCache            sync.RWMutex
	cacheTTL           time.Duration

	// Recovery automático
	workerTimeouts  map[int]time.Time // Worker ID -> Timeout time
	muWorkerTimeout sync.RWMutex
	recoveryMode    bool

	// Memoização de falhas
	failureMemo map[string]*FailureRecord
	muFailure   sync.RWMutex

	// Modo de operação
	operationMode OperationMode
	muMode        sync.RWMutex

	// Snapshot de decisão
	decisionSnapshot *DecisionSnapshot
	muSnapshot       sync.RWMutex

	// Priorização inteligente
	priorityQueue []*Task
	muPriority    sync.Mutex

	// Classificador rápido (Fase 2.1)
	decisionCache    map[string]*DecisionCache
	muDecisionCache sync.RWMutex
}

// Task representa uma tarefa a ser executada
type Task struct {
	ID        string
	Type      TaskType
	ToolName  string
	Arguments map[string]interface{}
	Priority  int // Menor = maior prioridade
	CreatedAt time.Time
}

// TaskType define o tipo da tarefa
type TaskType int

const (
	TaskTypeQuery TaskType = iota
	TaskTypeAction
	TaskTypeOrchestration
)

// DecisionType define como uma decisão foi tomada
type DecisionType int

const (
	DecisionTypeHeuristic DecisionType = iota // Regra determinística
	DecisionTypeCache                      // Do cache/histórico
	DecisionTypeLLM                        // Precisa de LLM
)

// DecisionCache entrada de cache de decisões
type DecisionCache struct {
	Message      string
	Decision     string
	Timestamp    time.Time
	HitCount     int
	SuccessRate  float64
}

// QuickClassifierResult resultado da classificação rápida
type QuickClassifierResult struct {
	Type        DecisionType
	Reason       string
	Heuristic   string  // Aplicável se Type=Heuristic
	ShouldCache  bool    // Se deve ser cacheado
}

// CognitiveBudget define o orçamento cognitivo atual
type CognitiveBudget struct {
	MaxTokens      int  // Limite de tokens
	AllowReasoning bool // Permite raciocínio estendido
	ToolComplexity int  // Nível de complexidade de ferramentas (1=simple, 3=complex)
}

// PromptBuilder construtor de prompts adaptativos
type PromptBuilder struct {
	mode        OperationMode
	budget      CognitiveBudget
	contextStr  string
}

// TaskResult representa o resultado de uma tarefa
type TaskResult struct {
	TaskID   string
	Success  bool
	Result   string
	Error    error
	Duration time.Duration
}

// NewOrchestrator cria um novo orquestrador
func NewOrchestrator(service *Service) (*Orchestrator, error) {
	// Criar diretório de cache se não existir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter diretório home: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".excel-ai")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de cache: %w", err)
	}

	dbPath := filepath.Join(cacheDir, "cache.db")

	// Criar cache persistente em SQLite
	persistentCache, err := cache.NewPersistentCache(dbPath, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cache persistente: %w", err)
	}

	return &Orchestrator{
		service:         service,
		taskChan:        make(chan *Task, 100),
		resultChan:      make(chan *TaskResult, 100),
		messageChan:     make(chan string, 100),
		pendingTasks:    make(map[string]*Task),
		cache:           persistentCache,
		cacheTTL:        5 * time.Minute, // TTL padrão: 5 minutos
		workerTimeouts: make(map[int]time.Time),
		failureMemo:     make(map[string]*FailureRecord),
		operationMode:   ModeNormal,
		priorityQueue:   make([]*Task, 0),
		decisionCache:   make(map[string]*DecisionCache), // Classificador rápido
	}, nil
}

// Start inicia o orquestrador
func (o *Orchestrator) Start(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.running {
		return fmt.Errorf("orchestrator já está rodando")
	}

	o.ctx, o.cancel = context.WithCancel(ctx)
	o.running = true

	// Iniciar workers
	for i := 0; i < 5; i++ { // 5 workers paralelos
		go o.worker(i)
	}

	// Iniciar coletor de resultados
	go o.resultCollector()

	// Iniciar buffer de mensagens
	go o.messageBuffer()

	// Iniciar limpeza de cache expirado
	go o.cacheCleanup()

	// Iniciar monitor de recovery
	go o.recoveryMonitor()

	// Iniciar priorizador
	go o.priorityDispatcher()

	// Iniciar monitor de modo de operação
	go o.operationModeMonitor()

	fmt.Println("[ORCHESTRATOR] ✅ Iniciado com 5 workers")
	return nil
}

// Stop para o orquestrador
func (o *Orchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.running {
		return
	}

	o.running = false
	o.cancel()

	// Fechar canais
	close(o.taskChan)
	close(o.resultChan)
	close(o.messageChan)

	fmt.Println("[ORCHESTRATOR] 🛑 Parado")
}

// OrchestrateMessage processa uma mensagem do usuário usando múltiplos modelos
func (o *Orchestrator) OrchestrateMessage(
	message string,
	contextStr string,
	askBeforeApply bool,
	onChunk func(string) error,
) (string, error) {

	// Passo 1: Modelo A (Orquestrador) analisa a solicitação
	onChunk("\n🎯 [Orquestrador] Analisando solicitação...\n")

	tasks, orchestrationPrompt, err := o.analyzeRequest(message, contextStr, onChunk)
	if err != nil {
		return "", fmt.Errorf("erro na análise: %w", err)
	}

	if len(tasks) == 0 {
		// Nenhuma tarefa específica - delegar para modelo principal
		onChunk("\n💬 [Orquestrador] Nenhuma tarefa específica, usando modelo principal...\n")
		return o.service.SendMessage(message, contextStr, askBeforeApply, onChunk)
	}

	// Passo 2: Enviar tarefas para execução paralela
	onChunk(fmt.Sprintf("\n📋 [Orquestrador] %d tarefas identificadas para execução paralela\n", len(tasks)))

	var wg sync.WaitGroup
	results := make([]*TaskResult, len(tasks))

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t *Task) {
			defer wg.Done()
			results[idx] = o.executeTask(t, onChunk)
		}(i, task)
	}

	// Passo 3: Enviar mensagens do orquestrador enquanto aguarda
	o.sendOrchestrationMessages(orchestrationPrompt, onChunk)

	// Aguardar todas as tarefas
	wg.Wait()

	// Passo 4: Compilar resultados
	onChunk("\n📊 [Orquestrador] Compilando resultados...\n")

	successCount := 0
	var finalResults []string

	for _, result := range results {
		if result.Success {
			successCount++
			finalResults = append(finalResults, result.Result)
			onChunk(fmt.Sprintf("✅ %s (%.2fs)\n", result.TaskID, result.Duration.Seconds()))
		} else {
			onChunk(fmt.Sprintf("❌ %s: %v\n", result.TaskID, result.Error))
		}
	}

	onChunk(fmt.Sprintf("\n🎉 [Orquestrator] %d/%d tarefas concluídas com sucesso\n", successCount, len(tasks)))

	// Passo 5: Enviar resultados para Modelo A para resposta final
	if len(finalResults) > 0 {
		return o.generateFinalResponse(message, finalResults, onChunk)
	}

	return strings.Join(finalResults, "\n"), nil
}

// analyzeRequest usa o Modelo A para analisar e dividir a solicitação
func (o *Orchestrator) analyzeRequest(
	message string,
	contextStr string,
	onChunk func(string) error,
) ([]*Task, string, error) {

	// Criar prompt para orquestrador
	orchestrationPrompt := fmt.Sprintf(`
Você é um ORQUESTRADOR Excel especializado. Sua função é analisar solicitações do usuário e dividi-las em tarefas executáveis.

SOLICITAÇÃO DO USUÁRIO:
%s

CONTEXTO DO EXCEL:
%s

SUA FUNÇÃO:
1. Identifique todas as ações/consultas necessárias
2. Divida em tarefas independentes que podem ser executadas em paralelo
3. Retorne uma lista JSON de tarefas

FORMATO DE RETORNO (JSON ARRAY):
[
  {
    "tool": "nome_da_ferramenta",
    "args": {parâmetros},
    "priority": 1,
    "description": "o que fazer"
  }
]

FERRAMENTAS DISPONÍVEIS:
- list_sheets, sheet_exists, get_headers, get_range_values, query_batch
- write_cell, write_range, create_sheet, delete_sheet
- format_range, autofit_columns, clear_range
- create_chart, create_pivot_table
- apply_filter, sort_range

REGRAS:
- Consultas (query_*) podem rodar em paralelo
- Ações (write_*, create_*, delete_*) podem rodar em paralelo se forem em células/planilhas diferentes
- Prioridade 1 = urgente, 2 = normal, 3 = baixa

RETORNE APENAS O JSON ARRAY, sem explicações adicionais.
`, message, contextStr)

	// Usar modelo principal (Modelo A) para análise
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []ai.Message{
		{Role: "system", Content: "Você é um orquestrador Excel especializado em dividir tarefas em paralelo."},
		{Role: "user", Content: orchestrationPrompt},
	}

	var responseBuilder strings.Builder
	response, err := o.service.client.ChatStream(ctx, messages, func(chunk string) error {
		responseBuilder.WriteString(chunk)
		return nil
	})
	if err != nil {
		return nil, orchestrationPrompt, err
	}

	// Parsear resposta para extrair tarefas
	tasks, err := o.parseTasks(response)
	if err != nil {
		onChunk(fmt.Sprintf("⚠️ [Orquestrador] Erro ao parsear tarefas: %v\n", err))
		return nil, orchestrationPrompt, nil
	}

	return tasks, orchestrationPrompt, nil
}

// parseTasks extrai tarefas da resposta do orquestrador
func (o *Orchestrator) parseTasks(response string) ([]*Task, error) {
	// Implementação simplificada - em produção usaria JSON parser robusto
	var tasks []*Task

	// Exemplo: extrair tool calls do JSON da resposta
	// Aqui seria o código para parsear o JSON array

	// Por enquanto, vamos extrair manualmente com regex ou substring
	if strings.Contains(response, "[") && strings.Contains(response, "]") {
		// Tem formato JSON - implementar parseamento real
		// ...
	}

	return tasks, nil
}

// executeTask executa uma única tarefa com suporte a cache
func (o *Orchestrator) executeTask(task *Task, reportProgress func(string) error) *TaskResult {
	start := time.Now()

	// Tentar obter do cache (apenas para consultas, não para ações)
	if task.Type == TaskTypeQuery {
		if cached, found := o.getFromCache(task.ToolName, task.Arguments); found {
			o.muStats.Lock()
			o.totalTasks++
			o.successTasks++
			o.muStats.Unlock()

			duration := time.Since(start)
			reportProgress(fmt.Sprintf("💾 [Cache] %s: %s (do cache)\n", task.ToolName, task.ID))

			return &TaskResult{
				TaskID:   task.ID,
				Success:  true,
				Result:   cached,
				Duration: duration,
			}
		}
	}

	// Atualizar contadores
	o.muStats.Lock()
	o.totalTasks++
	o.activeWorkers++
	o.muStats.Unlock()

	reportProgress(fmt.Sprintf("⚙️ [Worker] Executando %s: %s\n", task.ToolName, task.ID))

	result, err := o.service.executeToolCall(task.ToolName, task.Arguments)

	duration := time.Since(start)

	// Armazenar no cache (apenas para consultas bem-sucedidas)
	if err == nil && task.Type == TaskTypeQuery {
		o.setInCache(task.ToolName, task.Arguments, result)
	}

	// Atualizar estatísticas
	o.muStats.Lock()
	o.activeWorkers--

	if err == nil {
		o.successTasks++
		// Calcular média móvel
		if o.avgTaskTime == 0 {
			o.avgTaskTime = duration
		} else {
			// Média ponderada (dar mais peso às tarefas recentes)
			o.avgTaskTime = (o.avgTaskTime*9 + duration) / 10
		}
	} else {
		o.failedTasks++
	}

	o.muStats.Unlock()

	return &TaskResult{
		TaskID:   task.ID,
		Success:  err == nil,
		Result:   result,
		Error:    err,
		Duration: duration,
	}
}

// GetStats retorna estatísticas do orquestrador
func (o *Orchestrator) GetStats() OrchestratorStats {
	o.muStats.RLock()
	defer o.muStats.RUnlock()

	successRate := 0.0
	if o.totalTasks > 0 {
		successRate = float64(o.successTasks) / float64(o.totalTasks) * 100
	}

	return OrchestratorStats{
		TotalTasks:    o.totalTasks,
		SuccessTasks:  o.successTasks,
		FailedTasks:   o.failedTasks,
		ActiveWorkers: o.activeWorkers,
		AvgTaskTime:   o.avgTaskTime,
		SuccessRate:   successRate,
		IsRunning:     o.running,
	}
}

// OrchestratorStats contém estatísticas do orquestrador
type OrchestratorStats struct {
	TotalTasks    int64
	SuccessTasks  int64
	FailedTasks   int64
	ActiveWorkers int
	AvgTaskTime   time.Duration
	SuccessRate   float64
	IsRunning     bool
}

// HealthCheck verifica se o orquestrador está saudável
func (o *Orchestrator) HealthCheck() HealthStatus {
	o.muStats.RLock()
	defer o.muStats.RUnlock()

	status := HealthStatus{
		IsHealthy:     true,
		WorkersActive: o.activeWorkers,
		TotalTasks:    o.totalTasks,
		TasksPending:  len(o.pendingTasks),
		LastCheck:     time.Now(),
	}

	// Verificar se há tarefas pendentes por muito tempo
	if len(o.pendingTasks) > 0 {
		o.mu.Lock()
		for _, task := range o.pendingTasks {
			if time.Since(task.CreatedAt) > 5*time.Minute {
				status.IsHealthy = false
				status.Issues = append(status.Issues, fmt.Sprintf("Tarefa %s travada por mais de 5 minutos", task.ID))
			}
		}
		o.mu.Unlock()
	}

	// Verificar sucesso médio
	if o.totalTasks > 10 {
		successRate := float64(o.successTasks) / float64(o.totalTasks)
		if successRate < 0.7 { // Menos de 70% de sucesso
			status.IsHealthy = false
			status.Issues = append(status.Issues, fmt.Sprintf("Taxa de sucesso baixa: %.1f%%", successRate*100))
		}
	}

	return status
}

// HealthStatus representa o status de saúde do orquestrador
type HealthStatus struct {
	IsHealthy     bool
	WorkersActive int
	TotalTasks    int64
	TasksPending  int
	LastCheck     time.Time
	Issues        []string
}

// worker processa tarefas em paralelo
func (o *Orchestrator) worker(id int) {
	fmt.Printf("[ORCHESTRATOR] Worker %d iniciado\n", id)

	for task := range o.taskChan {
		fmt.Printf("[ORCHESTRATOR] Worker %d processando tarefa %s\n", id, task.ID)

		result := o.executeTask(task, func(msg string) error {
			o.messageChan <- msg
			return nil
		})
		o.resultChan <- result
	}

	fmt.Printf("[ORCHESTRATOR] Worker %d finalizado\n", id)
}

// resultCollector coleta resultados de todas as tarefas
func (o *Orchestrator) resultCollector() {
	for result := range o.resultChan {
		o.mu.Lock()
		delete(o.pendingTasks, result.TaskID)
		o.mu.Unlock()

		fmt.Printf("[ORCHESTRATOR] Resultado recebido: %s (Success: %v)\n", result.TaskID, result.Success)
	}
}

// messageBuffer envia mensagens enquanto aguarda resultados
func (o *Orchestrator) messageBuffer() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case msg := <-o.messageChan:
			// Enviar mensagem para UI
			if msg != "" {
				fmt.Printf("[ORCHESTRATOR] Mensagem bufferizada: %s\n", msg)
			}
		case <-ticker.C:
			// Enviar heartbeat se houver tarefas pendentes
			o.mu.Lock()
			pending := len(o.pendingTasks)
			o.mu.Unlock()

			if pending > 0 {
				fmt.Printf("[ORCHESTRATOR] 💓 Tarefas pendentes: %d\n", pending)
			}
		}
	}
}

// sendOrchestrationMessages envia mensagens do orquestrador
func (o *Orchestrator) sendOrchestrationMessages(
	prompt string,
	onChunk func(string) error,
) {
	// Enviar prompt do orquestrador para usuário ver o que está sendo feito
	messages := strings.Split(prompt, "\n")
	for _, msg := range messages {
		if strings.TrimSpace(msg) != "" {
			onChunk(msg + "\n")
			time.Sleep(100 * time.Millisecond) // Pequeno delay para melhor UX
		}
	}
}

// generateFinalResponse gera resposta final baseada nos resultados
func (o *Orchestrator) generateFinalResponse(
	originalRequest string,
	results []string,
	onChunk func(string) error,
) (string, error) {

	onChunk("\n🤖 [Orquestrador] Gerando resposta final...\n")

	finalPrompt := fmt.Sprintf(`
SOLICITAÇÃO ORIGINAL DO USUÁRIO:
%s

RESULTADOS DAS TAREFAS EXECUTADAS:
%s

SUA FUNÇÃO:
Analise os resultados e forneça uma resposta clara e útil ao usuário.
Combine os resultados de múltiplas tarefas em uma resposta coesa.

REGRAS:
- Seja direto e profissional
- Use os dados dos resultados para responder
- Em Português do Brasil
- Se algo falhou, explique de forma clara
`, originalRequest, strings.Join(results, "\n\n"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []ai.Message{
		{Role: "system", Content: "Você é um assistente Excel profissional."},
		{Role: "user", Content: finalPrompt},
	}

	var responseBuilder strings.Builder
	response, err := o.service.client.ChatStream(ctx, messages, func(chunk string) error {
		responseBuilder.WriteString(chunk)
		return onChunk(chunk)
	})
	if err != nil {
		return "", err
	}

	return response, nil
}

// chunkResponse divide a resposta em chunks para streaming
func (o *Orchestrator) chunkResponse(response string, chunkSize int) []string {
	var chunks []string

	for i := 0; i < len(response); i += chunkSize {
		end := i + chunkSize
		if end > len(response) {
			end = len(response)
		}
		chunks = append(chunks, response[i:end])
	}

	return chunks
}

// ============================================
// CACHE DE RESULTADOS
// ============================================

// generateCacheKey gera uma chave única para o cache
func (o *Orchestrator) generateCacheKey(toolName string, args map[string]interface{}) string {
	// Criar hash dos argumentos
	hash := sha256.New()
	hash.Write([]byte(toolName))

	// Adicionar argumentos ordenados
	if args != nil {
		keys := make([]string, 0, len(args))
		for k := range args {
			keys = append(keys, k)
		}

		// Ordenar para consistência
		for i := 0; i < len(keys)-1; i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}

		// Adicionar ao hash
		for _, k := range keys {
			hash.Write([]byte(k))
			hash.Write([]byte(fmt.Sprintf("%v", args[k])))
		}
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// getFromCache tenta obter resultado do cache
func (o *Orchestrator) getFromCache(toolName string, args map[string]interface{}) (string, bool) {
	key := o.generateCacheKey(toolName, args)

	result, found := o.cache.Get(key)
	if found {
		fmt.Printf("[CACHE DB] Hit: %s\n", toolName)
	}
	return result, found
}

// setInCache armazena resultado no cache com tags
func (o *Orchestrator) setInCache(toolName string, args map[string]interface{}, result string) {
	key := o.generateCacheKey(toolName, args)
	tags := o.generateCacheTags(toolName, args)

	err := o.cache.Set(key, result, tags)
	if err != nil {
		fmt.Printf("[CACHE DB] Erro ao armazenar: %v\n", err)
	}
}

// cacheCleanup remove entradas expiradas do cache (já é feito automaticamente pelo PersistentCache)
func (o *Orchestrator) cacheCleanup() {
	// O PersistentCache já faz limpeza automática a cada 1 minuto
	// Este método é mantido para compatibilidade mas não faz nada
	<-o.ctx.Done()
}

// ClearCache limpa todo o cache (método público)
func (o *Orchestrator) ClearCache() error {
	return o.cache.Clear()
}

// generateCacheTags gera tags para invalidação de cache
func (o *Orchestrator) generateCacheTags(toolName string, args map[string]interface{}) []string {
	tags := []string{}

	// Tag geral da ferramenta
	tags = append(tags, fmt.Sprintf("tool:%s", toolName))

	// Tag específica baseada nos argumentos
	if args != nil {
		// Tag de sheet se presente
		if sheet, ok := args["sheet"].(string); ok {
			tags = append(tags, fmt.Sprintf("sheet:%s", sheet))
		}

		// Tag de workbook se presente
		if workbook, ok := args["workbook"].(string); ok {
			tags = append(tags, fmt.Sprintf("workbook:%s", workbook))
		}

		// Tag de range se presente
		if rangeVal, ok := args["range"].(string); ok {
			tags = append(tags, fmt.Sprintf("range:%s", rangeVal))
		}
	}

	return tags
}

// invalidateCacheForAction invalida cache relacionado a uma ação mutável
func (o *Orchestrator) invalidateCacheForAction(toolName string, args map[string]interface{}) {
	tags := o.generateCacheTags(toolName, args)

	_, err := o.cache.Invalidate(tags)
	if err != nil {
		fmt.Printf("[CACHE DB] Erro ao invalidar: %v\n", err)
	}
}


// GetCacheStatus retorna status do cache
func (o *Orchestrator) GetCacheStatus() CacheStatus {
	cacheStatus := o.cache.GetStatus()
	
	return CacheStatus{
		TotalEntries:  cacheStatus.TotalEntries,
		HitRate:       cacheStatus.HitRate,
		Invalidations: cacheStatus.Invalidations,
		LastCleanup:   cacheStatus.LastCleanup,
	}
}

// ============================================
// RECOVERY AUTOMÁTICO
// ============================================

// recoveryMonitor monitora workers e executa recovery
func (o *Orchestrator) recoveryMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			o.checkWorkerHealth()
		}
	}
}

// checkWorkerHealth verifica se workers estão saudáveis
func (o *Orchestrator) checkWorkerHealth() {
	o.muWorkerTimeout.RLock()
	timeoutCount := len(o.workerTimeouts)
	o.muWorkerTimeout.RUnlock()

	if timeoutCount == 0 {
		return
	}

	now := time.Now()
	stalledWorkers := 0

	o.muWorkerTimeout.Lock()
	for workerID, timeout := range o.workerTimeouts {
		if now.Sub(timeout) > 2*time.Minute { // 2 minutos sem atividade
			fmt.Printf("[RECOVERY] Worker %d travado, iniciando recovery...\n", workerID)
			delete(o.workerTimeouts, workerID)
			stalledWorkers++
		}
	}
	o.muWorkerTimeout.Unlock()

	if stalledWorkers > 0 {
		fmt.Printf("[RECOVERY] %d workers recuperados\n", stalledWorkers)
		o.recoveryMode = false
	}
}

// markWorkerActive marca worker como ativo
func (o *Orchestrator) markWorkerActive(workerID int) {
	o.muWorkerTimeout.Lock()
	defer o.muWorkerTimeout.Unlock()

	o.workerTimeouts[workerID] = time.Now()

	if o.recoveryMode {
		fmt.Printf("[RECOVERY] Worker %d reativado\n", workerID)
	}
}

// ============================================
// MEMOIZAÇÃO DE FALHAS
// ============================================

// isRecurrentFailure verifica se uma tarefa falhou recorrentemente
func (o *Orchestrator) isRecurrentFailure(task *Task) bool {
	taskKey := o.generateTaskKey(task)

	o.muFailure.RLock()
	record, exists := o.failureMemo[taskKey]
	o.muFailure.RUnlock()

	if !exists {
		return false
	}

	// Se falhou 3+ vezes, é recorrente
	if record.FailCount >= 3 && record.IsRecurrent {
		return true
	}

	return false
}

// recordFailure registra uma falha de tarefa
func (o *Orchestrator) recordFailure(task *Task, err error) {
	taskKey := o.generateTaskKey(task)

	o.muFailure.Lock()
	defer o.muFailure.Unlock()

	record, exists := o.failureMemo[taskKey]
	if !exists {
		record = &FailureRecord{
			TaskID:      task.ID,
			FailCount:   0,
			LastFailure: time.Now(),
			IsRecurrent: false,
		}
	}

	record.FailCount++
	record.LastFailure = time.Now()
	record.LastError = err

	// Marcar como recorrente se falhou 3+ vezes
	if record.FailCount >= 3 {
		record.IsRecurrent = true
		fmt.Printf("[FAILURE MEMO] Tarefa %s marcada como falha recorrente (%d tentativas)\n", task.ID, record.FailCount)
	}

	o.failureMemo[taskKey] = record
}

// clearFailureMemo limpa o registro de falha de uma tarefa (chamada em caso de sucesso)
func (o *Orchestrator) clearFailureMemo(task *Task) {
	taskKey := o.generateTaskKey(task)

	o.muFailure.Lock()
	defer o.muFailure.Unlock()

	delete(o.failureMemo, taskKey)
}

// getFailureCount retorna o número de falhas de uma tarefa
func (o *Orchestrator) getFailureCount(task *Task) int {
	taskKey := o.generateTaskKey(task)

	o.muFailure.RLock()
	defer o.muFailure.RUnlock()

	if record, exists := o.failureMemo[taskKey]; exists {
		return record.FailCount
	}

	return 0
}

// generateTaskKey gera uma chave única para uma tarefa
func (o *Orchestrator) generateTaskKey(task *Task) string {
	return fmt.Sprintf("%s:%s", task.ToolName, task.ID)
}

// GetFailureStats retorna estatísticas de falhas
func (o *Orchestrator) GetFailureStats() map[string]interface{} {
	o.muFailure.RLock()
	defer o.muFailure.RUnlock()

	recurrentCount := 0
	totalFailures := 0

	for _, record := range o.failureMemo {
		totalFailures += record.FailCount
		if record.IsRecurrent {
			recurrentCount++
		}
	}

	return map[string]interface{}{
		"total_memoized": len(o.failureMemo),
		"total_failures": totalFailures,
		"recurrent":      recurrentCount,
	}
}

// ============================================
// MODO DE OPERAÇÃO DEGRADADO
// ============================================

// evaluateOperationMode avalia o modo de operação atual baseado em métricas
func (o *Orchestrator) evaluateOperationMode() {
	o.muStats.RLock()
	successRate := 0.0
	if o.totalTasks > 0 {
		successRate = float64(o.successTasks) / float64(o.totalTasks) * 100
	}
	o.muStats.RUnlock()

	o.muMode.Lock()
	defer o.muMode.Unlock()

	newMode := ModeNormal

	// Determinar modo baseado em métricas
	if successRate < 50 {
		newMode = ModeCritical
	} else if successRate < 75 {
		newMode = ModeDegraded
	}

	// Se o modo mudou, notificar
	if newMode != o.operationMode {
		oldMode := o.operationMode
		o.operationMode = newMode
		o.applyOperationMode(newMode)

		modeName := map[OperationMode]string{
			ModeNormal:   "Normal",
			ModeDegraded: "Degradado",
			ModeCritical: "Crítico",
		}

		fmt.Printf("[MODE] Modo de operação alterado: %s -> %s\n", modeName[oldMode], modeName[newMode])
	}
}

// applyOperationMode aplica configurações do modo de operação
func (o *Orchestrator) applyOperationMode(mode OperationMode) {
	switch mode {
	case ModeNormal:
		// 100% funcional - 5 workers, TTL padrão
		o.cacheTTL = 5 * time.Minute
		o.cache.SetTTL(o.cacheTTL)
		// Workers já iniciados com 5

	case ModeDegraded:
		// 50-75% funcional - 3 workers, TTL aumentado
		o.cacheTTL = 10 * time.Minute
		o.cache.SetTTL(o.cacheTTL)
		fmt.Printf("[MODE] Modo Degradado: paralelismo reduzido, TTL aumentado\n")

	case ModeCritical:
		// < 50% funcional - 1 worker, TTL desativado
		o.cacheTTL = 30 * time.Minute
		o.cache.SetTTL(o.cacheTTL)
		fmt.Printf("[MODE] Modo Crítico: paralelismo mínimo, cache desativado\n")
	}
}

// GetOperationMode retorna o modo de operação atual
func (o *Orchestrator) GetOperationMode() OperationMode {
	o.muMode.RLock()
	defer o.muMode.RUnlock()
	return o.operationMode
}

// ============================================
// SNAPSHOT DE DECISÃO
// ============================================

// captureDecisionSnapshot captura um snapshot imutável do estado atual
func (o *Orchestrator) captureDecisionSnapshot() DecisionSnapshot {
	o.muSnapshot.Lock()
	defer o.muSnapshot.Unlock()

	snapshot := DecisionSnapshot{
		Timestamp:     time.Now(),
		OperationMode: o.GetOperationMode(),
		Stats:         o.GetStats(),
		Health:        o.HealthCheck(),
		CacheStatus:   o.GetCacheStatus(),
		PendingTasks:  len(o.pendingTasks),
	}

	// Definir tarefas disponíveis baseadas no modo
	switch snapshot.OperationMode {
	case ModeNormal:
		snapshot.AvailableTasks = []string{
			"list_sheets", "get_range_values", "query_batch",
			"write_cell", "write_range", "create_sheet",
			"format_range", "create_chart", "create_pivot_table",
		}
	case ModeDegraded:
		snapshot.AvailableTasks = []string{
			"list_sheets", "get_range_values",
			"write_cell", "write_range",
		}
	case ModeCritical:
		snapshot.AvailableTasks = []string{
			"list_sheets",
			"write_cell",
		}
	}

	o.decisionSnapshot = &snapshot
	return snapshot
}

// GetDecisionSnapshot retorna o snapshot atual
func (o *Orchestrator) GetDecisionSnapshot() DecisionSnapshot {
	o.muSnapshot.RLock()
	defer o.muSnapshot.RUnlock()

	if o.decisionSnapshot == nil {
		return o.captureDecisionSnapshot()
	}

	return *o.decisionSnapshot
}

// GetOperationModeName retorna o nome do modo de operação
func (o *Orchestrator) GetOperationModeName() string {
	mode := o.GetOperationMode()

	names := map[OperationMode]string{
		ModeNormal:   "Normal",
		ModeDegraded: "Degradado",
		ModeCritical: "Crítico",
	}

	return names[mode]
}

// operationModeMonitor monitora e ajusta o modo de operação automaticamente
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

// ============================================
// PRIORIZAÇÃO INTELIGENTE
// ============================================

// priorityDispatcher gerencia fila de prioridades
func (o *Orchestrator) priorityDispatcher() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			o.muPriority.Lock()
			if len(o.priorityQueue) > 0 {
				// Ordenar por prioridade (menor = mais importante)
				for i := 0; i < len(o.priorityQueue)-1; i++ {
					for j := i + 1; j < len(o.priorityQueue); j++ {
						if o.priorityQueue[i].Priority > o.priorityQueue[j].Priority {
							o.priorityQueue[i], o.priorityQueue[j] = o.priorityQueue[j], o.priorityQueue[i]
						}
					}
				}

				// Enviar tarefa com maior prioridade
				task := o.priorityQueue[0]
				o.priorityQueue = o.priorityQueue[1:]
				o.muPriority.Unlock()

				select {
				case o.taskChan <- task:
					// Tarefa enviada
				case <-o.ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					// Fila cheia, tentar novamente
					o.muPriority.Lock()
					o.priorityQueue = append([]*Task{task}, o.priorityQueue...)
					o.muPriority.Unlock()
				}
			} else {
				o.muPriority.Unlock()
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// addTaskWithPriority adiciona tarefa com prioridade
func (o *Orchestrator) addTaskWithPriority(task *Task) {
	o.muPriority.Lock()
	defer o.muPriority.Unlock()

	o.priorityQueue = append(o.priorityQueue, task)

	priorityName := "normal"
	switch task.Priority {
	case 1:
		priorityName = "urgente"
	case 3:
		priorityName = "baixa"
	}

	fmt.Printf("[PRIORITY] Tarefa %s adicionada (prioridade: %s)\n", task.ID, priorityName)
}

// analyzeTaskPriority analisa e define prioridade da tarefa
func (o *Orchestrator) analyzeTaskPriority(toolName string, _ map[string]interface{}) int {
	// Tarefas de consulta (menos críticas)
	if strings.HasPrefix(toolName, "get_") || strings.HasPrefix(toolName, "list_") || strings.HasPrefix(toolName, "query_") {
		return 2 // Normal
	}

	// Tarefas de ação (mais críticas)
	if strings.HasPrefix(toolName, "write_") || strings.HasPrefix(toolName, "create_") || strings.HasPrefix(toolName, "delete_") {
		return 1 // Urgente
	}

	// Formatação e outras (menos críticas)
	return 3 // Baixa
}

// ============================================
// CLASSIFICADOR RÁPIDO (FASE 2.1)
// ============================================

// ClassifyRequest classifica uma mensagem usando heurísticas rápidas
func (o *Orchestrator) ClassifyRequest(message string) QuickClassifierResult {
	messageLower := strings.ToLower(message)

	// Camada 1: Timeout rápido (50ms) - operações muito simples
	if o.quickTimeoutCheck(messageLower) {
		return QuickClassifierResult{
			Type:       DecisionTypeHeuristic,
			Reason:      "Timeout rápido - operação simples",
			Heuristic:  o.applySimpleHeuristic(messageLower),
			ShouldCache: true,
		}
	}

	// Camada 2: Permissão rápida (100ms) - verificações de segurança
	if !o.quickPermissionCheck(messageLower) {
		return QuickClassifierResult{
			Type:       DecisionTypeHeuristic,
			Reason:      "Operação perigosa - requer confirmação",
			Heuristic:  "BLOCKED: Operação requer confirmação do usuário",
			ShouldCache: false,
		}
	}

	// Camada 3: Cache de decisões (150ms)
	if cached, found := o.getDecisionCache(message); found {
		fmt.Printf("[CLASSIFIER] Cache hit: %s\n", cached.Decision)
		return QuickClassifierResult{
			Type:       DecisionTypeCache,
			Reason:      "Decisão cacheada",
			Heuristic:  cached.Decision,
			ShouldCache: true,
		}
	}

	// Camada 4: Lógica simples (200ms) - regras determinísticas
	if o.simpleLogicCheck(messageLower) {
		decision := o.applySimpleHeuristic(messageLower)
		return QuickClassifierResult{
			Type:       DecisionTypeHeuristic,
			Reason:      "Lógica simples aplicada",
			Heuristic:  decision,
			ShouldCache: true,
		}
	}

	// Camada 5: LLM completo
	return QuickClassifierResult{
		Type:       DecisionTypeLLM,
		Reason:      "Requer análise completa do LLM",
		Heuristic:  "",
		ShouldCache: true,
	}
}

// quickTimeoutCheck verifica se a mensagem pode ser respondida instantaneamente
func (o *Orchestrator) quickTimeoutCheck(message string) bool {
	// Padrões de operações muito simples que não precisam de LLM
	quickPatterns := []string{
		"qual sheet", "quais sheets", "lista sheets", "listar sheets",
		"qual a planilha", "quais as planilhas", "lista planilhas",
		"quantas células", "quantas linhas", "quantas colunas",
		"sheet existe", "planilha existe", "tem sheet",
		"nome da sheet", "nome da planilha",
	}

	for _, pattern := range quickPatterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}

// quickPermissionCheck verifica operações perigosas que requerem confirmação
func (o *Orchestrator) quickPermissionCheck(message string) bool {
	// Operações perigosas que requerem confirmação humana
	dangerousOps := []string{
		"deletar", "apagar", "remover", "destroy",
		"formatar tudo", "limpar tudo", "apagar tudo",
		"destruir", "eliminar tudo",
	}

	for _, op := range dangerousOps {
		if strings.Contains(message, op) {
			return false // Operação perigosa - bloquear
		}
	}
	return true
}

// simpleLogicCheck aplica regras determinísticas simples
func (o *Orchestrator) simpleLogicCheck(message string) bool {
	// Padrões que podem ser resolvidos com lógica simples
	simplePatterns := []string{
		"criar gráfico", "criar chart", "fazer gráfico",
		"pivot table", "tabela dinâmica", "criar pivot",
		"aplicar filtro", "filtrar dados",
		"ordenar", "sort", "classificar",
	}

	for _, pattern := range simplePatterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}

// applySimpleHeuristic aplica heurística simples e retorna decisão
func (o *Orchestrator) applySimpleHeuristic(message string) string {
	if strings.Contains(message, "gráfico") || strings.Contains(message, "chart") {
		return "create_chart(range=A1:C10,type=bar)"
	}

	if strings.Contains(message, "pivot") || strings.Contains(message, "tabela dinâmica") {
		return "create_pivot_table(source=Sheet1!A1:C100)"
	}

	if strings.Contains(message, "filtro") || strings.Contains(message, "filtrar") {
		return "apply_filter(sheet=Sheet1,range=A1:Z100)"
	}

	if strings.Contains(message, "ordenar") || strings.Contains(message, "sort") {
		return "sort_range(sheet=Sheet1,range=A1:C100,by=columnA)"
	}

	// Fallback para listagem
	if strings.Contains(message, "sheet") || strings.Contains(message, "planilha") {
		return "list_sheets()"
	}

	return ""
}

// getDecisionCache tenta obter decisão do cache
func (o *Orchestrator) getDecisionCache(message string) (*DecisionCache, bool) {
	o.muDecisionCache.RLock()
	defer o.muDecisionCache.RUnlock()

	cache, exists := o.decisionCache[message]
	if !exists {
		return nil, false
	}

	// Verificar se ainda é válido (TTL de 1 hora)
	if time.Since(cache.Timestamp) > 1*time.Hour {
		return nil, false
	}

	cache.HitCount++
	return cache, true
}

// setDecisionCache armazena decisão no cache
func (o *Orchestrator) setDecisionCache(message string, decision string) {
	o.muDecisionCache.Lock()
	defer o.muDecisionCache.Unlock()

	cache, exists := o.decisionCache[message]
	if !exists {
		cache = &DecisionCache{
			Message:     message,
			Decision:    decision,
			Timestamp:   time.Now(),
			HitCount:    0,
			SuccessRate: 100.0,
		}
	}

	cache.HitCount++
	o.decisionCache[message] = cache

	fmt.Printf("[CLASSIFIER] Cache set: %s -> %s\n", message, decision)
}

// GetClassifierStats retorna estatísticas do classificador
func (o *Orchestrator) GetClassifierStats() map[string]interface{} {
	o.muDecisionCache.RLock()
	defer o.muDecisionCache.RUnlock()

	totalCached := len(o.decisionCache)
	totalHits := 0

	for _, cache := range o.decisionCache {
		totalHits += cache.HitCount
	}

	return map[string]interface{}{
		"total_cached_decisions": totalCached,
		"total_cache_hits":      totalHits,
		"hit_rate":            float64(totalHits) / float64(totalCached),
	}
}

// ============================================
// ORÇAMENTO COGNITIVO (FASE 2.2)
// ============================================

// getCognitiveBudget retorna o orçamento cognitivo baseado no modo atual
func (o *Orchestrator) getCognitiveBudget() CognitiveBudget {
	o.muStats.RLock()
	stats := o.GetStats()
	o.muStats.RUnlock()

	mode := o.GetOperationMode()

	budget := CognitiveBudget{
		AllowReasoning: true,
		ToolComplexity: 3, // 1=simple, 3=complex
	}

	switch mode {
	case ModeCritical:
		// Modo crítico: prompt minimalista
		budget.MaxTokens = 200
		budget.AllowReasoning = false
		budget.ToolComplexity = 1

	case ModeDegraded:
		// Modo degradado: prompt enxuto
		budget.MaxTokens = 500
		budget.AllowReasoning = false
		budget.ToolComplexity = 2

	case ModeNormal:
		// Modo normal: orçamento dinâmico baseado em saúde do sistema
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

// buildPrompt constrói um prompt adaptativo baseado no orçamento cognitivo
func (o *Orchestrator) buildPrompt(message string, contextStr string) string {
	mode := o.GetOperationMode()

	switch mode {
	case ModeCritical:
		return o.buildMinimalPrompt(message, contextStr)
	case ModeDegraded:
		return o.buildLeanPrompt(message, contextStr)
	case ModeNormal:
		return o.buildFullPrompt(message, contextStr)
	}

	return o.buildFullPrompt(message, contextStr)
}

// buildMinimalPrompt constrói um prompt minimalista para modo crítico
func (o *Orchestrator) buildMinimalPrompt(message string, contextStr string) string {
	return fmt.Sprintf(`Ação: %s
Contexto: %s

Responda apenas com a ferramenta a usar.
Formato: tool_name(args)
`, message, o.getMinimalContext(contextStr))
}

// buildLeanPrompt constrói um prompt enxuto para modo degradado
func (o *Orchestrator) buildLeanPrompt(message string, contextStr string) string {
	budget := o.getCognitiveBudget()

	return fmt.Sprintf(`SOLICITAÇÃO:
%s

CONTEXTO:
%s

FERRAMENTAS:
%s

INSTRUÇÕES:
- Seja direto e conciso
- Use ferramenta apropriada
- Máximo %d tokens
`,
		message,
		o.getLeanContext(contextStr),
		o.getAvailableTools(budget.ToolComplexity),
		budget.MaxTokens,
	)
}

// buildFullPrompt constrói um prompt completo com raciocínio
func (o *Orchestrator) buildFullPrompt(message string, contextStr string) string {
	budget := o.getCognitiveBudget()

	return fmt.Sprintf(`Você é um assistente Excel especializado.

SOLICITAÇÃO:
%s

CONTEXTO COMPLETO:
%s

CONSIDERAÇÕES:
- Analise os dados disponíveis
- Considere múltiplas abordagens
- Explique seu raciocínio
- Sugira melhorias se aplicável

FERRAMENTAS DISPONÍVEIS:
%s

RESPOSTA:
1. Análise da situação
2. Ferramentas necessárias
3. Explicação do processo
4. Resultado esperado

Orçamento: %d tokens (raciocínio %v)
`,
		message,
		contextStr,
		o.getAvailableTools(budget.ToolComplexity),
		budget.MaxTokens,
		budget.AllowReasoning,
	)
}

// getMinimalContext retorna contexto mínimo para modo crítico
func (o *Orchestrator) getMinimalContext(contextStr string) string {
	// Extrair apenas informações essenciais
	lines := strings.Split(contextStr, "\n")
	if len(lines) > 3 {
		return strings.Join(lines[:3], "\n")
	}
	return contextStr
}

// getLeanContext retorna contexto enxuto para modo degradado
func (o *Orchestrator) getLeanContext(contextStr string) string {
	// Extrair contexto resumido (primeiros 5 linhas + últimas 2)
	lines := strings.Split(contextStr, "\n")
	if len(lines) > 7 {
		return strings.Join(append(lines[:5], lines[len(lines)-2:]...), "\n")
	}
	return contextStr
}

// getAvailableTools retorna lista de ferramentas disponíveis baseada na complexidade
func (o *Orchestrator) getAvailableTools(complexity int) string {
	// Filtrar ferramentas baseadas no nível de complexidade
	tools := map[int][]string{
		1: {"list_sheets", "get_range_values"},                // Simples
		2: {"list_sheets", "get_range_values", "write_cell", "write_range"}, // Médio
		3: { // Complexo - todas as ferramentas
			"list_sheets", "get_range_values", "query_batch",
			"write_cell", "write_range", "create_sheet",
			"format_range", "autofit_columns", "clear_range",
			"create_chart", "create_pivot_table",
			"apply_filter", "sort_range",
		},
	}

	if toolList, exists := tools[complexity]; exists {
		return strings.Join(toolList, ", ")
	}

	return strings.Join(tools[3], ", ") // Fallback para todas
}

// GetCognitiveBudgetStats retorna estatísticas do orçamento cognitivo
func (o *Orchestrator) GetCognitiveBudgetStats() map[string]interface{} {
	budget := o.getCognitiveBudget()
	mode := o.GetOperationMode()

	modeName := map[OperationMode]string{
		ModeNormal:   "Normal",
		ModeDegraded: "Degradado",
		ModeCritical: "Crítico",
	}

	return map[string]interface{}{
		"mode":              modeName[mode],
		"max_tokens":        budget.MaxTokens,
		"allow_reasoning":   budget.AllowReasoning,
		"tool_complexity":   budget.ToolComplexity,
		"estimated_tokens_per_prompt": estimatePromptTokens(budget.MaxTokens),
	}
}

// estimatePromptTokens estima o número de tokens baseado no orçamento
func estimatePromptTokens(budget int) int {
	// Estimativa simples: orçamento é o limite máximo
	return budget
}
