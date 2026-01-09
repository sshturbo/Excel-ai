package chat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"excel-ai/pkg/ai"
)

// CacheEntry representa uma entrada no cache
type CacheEntry struct {
	Result      string
	Error       error
	StoredAt    time.Time
	AccessCount int
	TTL         time.Duration
}

// Orchestrator gerencia múltiplos modelos trabalhando em paralelo
type Orchestrator struct {
	service      *Service
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	
	// Canais de comunicação
	taskChan     chan *Task
	resultChan   chan *TaskResult
	messageChan  chan string
	
	// Estado
	running      bool
	pendingTasks map[string]*Task
	
	// Balanceamento dinâmico
	activeWorkers int
	totalTasks     int64
	successTasks   int64
	failedTasks    int64
	avgTaskTime   time.Duration
	muStats       sync.RWMutex
	
	// Cache de resultados
	cache         map[string]*CacheEntry
	muCache       sync.RWMutex
	cacheHits     int64
	cacheMisses   int64
	cacheTTL      time.Duration
	
	// Recovery automático
	workerTimeouts map[int]time.Time // Worker ID -> Timeout time
	muWorkerTimeout sync.RWMutex
	recoveryMode   bool
	
	// Priorização inteligente
	priorityQueue  []*Task
	muPriority    sync.Mutex
}

// Task representa uma tarefa a ser executada
type Task struct {
	ID         string
	Type       TaskType
	ToolName   string
	Arguments  map[string]interface{}
	Priority   int // Menor = maior prioridade
	CreatedAt  time.Time
}

// TaskType define o tipo da tarefa
type TaskType int

const (
	TaskTypeQuery TaskType = iota
	TaskTypeAction
	TaskTypeOrchestration
)

// TaskResult representa o resultado de uma tarefa
type TaskResult struct {
	TaskID   string
	Success  bool
	Result   string
	Error    error
	Duration time.Duration
}

// NewOrchestrator cria um novo orquestrador
func NewOrchestrator(service *Service) *Orchestrator {
	return &Orchestrator{
		service:       service,
		taskChan:      make(chan *Task, 100),
		resultChan:    make(chan *TaskResult, 100),
		messageChan:   make(chan string, 100),
		pendingTasks:  make(map[string]*Task),
		cache:         make(map[string]*CacheEntry),
		cacheTTL:      5 * time.Minute, // TTL padrão: 5 minutos
		workerTimeouts: make(map[int]time.Time),
		priorityQueue: make([]*Task, 0),
	}
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
		ActiveWorkers:  o.activeWorkers,
		AvgTaskTime:   o.avgTaskTime,
		SuccessRate:    successRate,
		IsRunning:      o.running,
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
	Issues       []string
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
	
	o.muCache.RLock()
	defer o.muCache.RUnlock()
	
	entry, exists := o.cache[key]
	if !exists {
		return "", false
	}
	
	// Verificar se expirou
	if time.Since(entry.StoredAt) > entry.TTL {
		o.muCache.RUnlock()
		o.muCache.Lock()
		delete(o.cache, key)
		o.muCache.Unlock()
		o.muCache.RLock()
		return "", false
	}
	
	// Atualizar contador de acessos
	o.muCache.RUnlock()
	o.muCache.Lock()
	entry.AccessCount++
	o.muCache.Unlock()
	
	o.muCache.Lock()
	o.cacheHits++
	o.muCache.Unlock()
	
	fmt.Printf("[CACHE] Hit: %s (acessos: %d)\n", toolName, entry.AccessCount)
	
	return entry.Result, true
}

// setInCache armazena resultado no cache
func (o *Orchestrator) setInCache(toolName string, args map[string]interface{}, result string) {
	key := o.generateCacheKey(toolName, args)
	
	entry := &CacheEntry{
		Result:      result,
		StoredAt:    time.Now(),
		AccessCount: 1,
		TTL:         o.cacheTTL,
	}
	
	o.muCache.Lock()
	defer o.muCache.Unlock()
	
	o.cache[key] = entry
	o.cacheMisses++
	
	fmt.Printf("[CACHE] Set: %s (TTL: %v)\n", toolName, o.cacheTTL)
}

// cacheCleanup remove entradas expiradas do cache
func (o *Orchestrator) cacheCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			o.muCache.Lock()
			now := time.Now()
			expired := 0
			
			for key, entry := range o.cache {
				if now.Sub(entry.StoredAt) > entry.TTL {
					delete(o.cache, key)
					expired++
				}
			}
			
			if expired > 0 {
				fmt.Printf("[CACHE] Limpeza: %d entradas expiradas removidas\n", expired)
			}
			
			o.muCache.Unlock()
		}
	}
}

// ClearCache limpa todo o cache (método público)
func (o *Orchestrator) ClearCache() {
	o.muCache.Lock()
	defer o.muCache.Unlock()
	
	count := len(o.cache)
	o.cache = make(map[string]*CacheEntry)
	
	fmt.Printf("[CACHE] Limpo: %d entradas removidas\n", count)
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
