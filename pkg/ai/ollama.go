package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"excel-ai/pkg/logger"
)

// OllamaClient cliente para API nativa do Ollama (/api/chat)
// A API nativa tem melhor suporte a tools do que o endpoint OpenAI-compatible
type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// OllamaMessage representa uma mensagem no formato Ollama nativo
type OllamaMessage struct {
	Role       string                   `json:"role"`
	Content    string                   `json:"content"`
	ToolCalls  []OllamaToolCallResponse `json:"tool_calls,omitempty"`
	ToolCallID string                   `json:"tool_call_id,omitempty"` // Para respostas de tools
	ToolName   string                   `json:"tool_name,omitempty"`    // Algumas versões usam isso
}

// OllamaToolCallResponse representa um tool call na resposta do Ollama
type OllamaToolCallResponse struct {
	Function OllamaFunctionCall `json:"function"`
}

// OllamaFunctionCall representa a chamada de função do Ollama
type OllamaFunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"` // Ollama retorna como objeto, não string!
}

// OllamaChatRequest requisição para /api/chat
type OllamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Tools    []Tool          `json:"tools,omitempty"`
	Options  *OllamaOptions  `json:"options,omitempty"`
}

// OllamaOptions opções adicionais para o modelo
type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"` // max_tokens equivalente
	NumCtx      int     `json:"num_ctx,omitempty"`     // context length
}

// OllamaChatResponse resposta do /api/chat
type OllamaChatResponse struct {
	Model      string        `json:"model"`
	CreatedAt  string        `json:"created_at"`
	Message    OllamaMessage `json:"message"`
	Done       bool          `json:"done"`
	DoneReason string        `json:"done_reason,omitempty"`
	// Estatísticas
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
}

// NewOllamaClient cria um novo cliente Ollama nativo
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	// Remover /v1 se presente (queremos usar API nativa)
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	baseURL = strings.TrimSuffix(baseURL, "/")

	if model == "" {
		model = "llama3.2:latest"
	}

	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Ollama pode demorar mais na primeira execução (carregando modelo)
		},
	}
}

// SetBaseURL altera a URL base
func (c *OllamaClient) SetBaseURL(url string) {
	if url == "" {
		url = "http://localhost:11434"
	}
	url = strings.TrimSuffix(url, "/v1")
	url = strings.TrimSuffix(url, "/")
	c.baseURL = url
}

// SetModel altera o modelo
func (c *OllamaClient) SetModel(model string) {
	c.model = model
}

// GetBaseURL retorna a URL base
func (c *OllamaClient) GetBaseURL() string {
	return c.baseURL
}

// GetModel retorna o modelo atual
func (c *OllamaClient) GetModel() string {
	return c.model
}

// convertToOllamaMessages converte mensagens padrão para formato Ollama nativo
func convertToOllamaMessages(messages []Message) []OllamaMessage {
	result := make([]OllamaMessage, 0, len(messages))
	for _, m := range messages {
		msg := OllamaMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}

		// Se for resposta de ferramenta, o Ollama espera Role "tool"
		if m.Role == "tool" {
			msg.Role = "tool"
		}

		// Se a mensagem do assistente teve tool calls, precisamos incluí-las para manter o contexto
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			msg.ToolCalls = make([]OllamaToolCallResponse, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				var args map[string]interface{}
				// Tentar parsear argumentos como JSON
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					// Se falhar, passar como string (alguns modelos preferem)
					args = map[string]interface{}{"raw_args": tc.Function.Arguments}
				}
				msg.ToolCalls[j] = OllamaToolCallResponse{
					Function: OllamaFunctionCall{
						Name:      tc.Function.Name,
						Arguments: args,
					},
				}
			}
		}

		result = append(result, msg)
	}
	return result
}

// ChatStreamWithTools envia mensagens com tools usando API nativa do Ollama
func (c *OllamaClient) ChatStreamWithTools(ctx context.Context, messages []Message, tools []Tool, onChunk func(string) error) (string, []ToolCall, error) {
	// Converter mensagens para formato Ollama
	ollamaMessages := convertToOllamaMessages(messages)

	reqBody := OllamaChatRequest{
		Model:    c.model,
		Messages: ollamaMessages,
		Stream:   true,
		Tools:    tools,
		Options: &OllamaOptions{
			Temperature: 0.7,
			NumPredict:  4096,
			NumCtx:      8192,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao serializar request: %w", err)
	}

	// Usar API nativa /api/chat (não /v1/chat/completions)
	url := c.baseURL + "/api/chat"

	logger.AIDebug(fmt.Sprintf("[OLLAMA] URL: %s", url))
	logger.AIDebug(fmt.Sprintf("[OLLAMA] Model: %s", c.model))
	logger.AIDebug(fmt.Sprintf("[OLLAMA] Messages count: %d", len(ollamaMessages)))
	logger.AIDebug(fmt.Sprintf("[OLLAMA] Tools count: %d", len(tools)))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("erro na requisição para Ollama: %w", err)
	}
	defer resp.Body.Close()

	logger.AIDebug(fmt.Sprintf("[OLLAMA] Response status: %d", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("Ollama retornou status %d: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	var fullResponse strings.Builder
	var toolCalls []ToolCall
	var rawToolCalls int

	// Processar stream de respostas (uma linha JSON por chunk)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", nil, fmt.Errorf("erro ao ler stream: %w", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var chunk OllamaChatResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			// Log mas continue (pode ser linha inválida)
			logger.AIError(fmt.Sprintf("[OLLAMA] Erro ao parsear chunk: %v - raw: %s", err, string(line[:min(100, len(line))])))
			continue
		}

		// DEBUG: Log do chunk
		if chunk.Message.Content != "" || len(chunk.Message.ToolCalls) > 0 {
			logger.AIDebug(fmt.Sprintf("[OLLAMA] Chunk - Content: %q, ToolCalls: %d, Done: %v",
				truncate(chunk.Message.Content, 50), len(chunk.Message.ToolCalls), chunk.Done))
		}

		// Processar conteúdo de texto
		if chunk.Message.Content != "" {
			content := chunk.Message.Content

			// Limpar JSON de tool calls que o modelo pode ter incluído no texto
			// Isso acontece quando modelos menores não entendem o formato correto
			content = cleanToolCallJSON(content)

			if content != "" {
				fullResponse.WriteString(content)
				if err := onChunk(content); err != nil {
					return "", nil, err
				}
			}
		}

		// Processar tool calls (Ollama envia tudo de uma vez quando done=false ainda)
		if len(chunk.Message.ToolCalls) > 0 {
			rawToolCalls += len(chunk.Message.ToolCalls)
			for _, tc := range chunk.Message.ToolCalls {
				logger.AIDebug(fmt.Sprintf("[OLLAMA] Raw tool call: name=%q, args=%v", tc.Function.Name, tc.Function.Arguments))

				// Validar que temos nome válido
				if tc.Function.Name == "" {
					logger.AIWarn("[OLLAMA] Ignorando tool call com nome vazio")
					continue
				}

				// Converter arguments de map para JSON string (formato esperado pelo resto do código)
				argsJSON, err := json.Marshal(tc.Function.Arguments)
				if err != nil {
					logger.AIError(fmt.Sprintf("[OLLAMA] Erro ao serializar arguments: %v", err))
					continue
				}

				toolCall := ToolCall{
					ID:   fmt.Sprintf("ollama_%s_%d", tc.Function.Name, time.Now().UnixNano()),
					Type: "function",
					Function: FunctionCall{
						Name:      tc.Function.Name,
						Arguments: string(argsJSON),
					},
				}
				toolCalls = append(toolCalls, toolCall)

				logger.AIDebug(fmt.Sprintf("[OLLAMA] Tool call válido: %s", tc.Function.Name))
			}
		}

		// Se done, terminamos
		if chunk.Done {
			logger.AIDebug(fmt.Sprintf("[OLLAMA] Stream finalizado. RawToolCalls: %d, ValidToolCalls: %d, ResponseLen: %d",
				rawToolCalls, len(toolCalls), fullResponse.Len()))
			break
		}
	}

	return fullResponse.String(), toolCalls, nil
}

// Helpers para logging
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetAvailableModels lista os modelos disponíveis no Ollama
func (c *OllamaClient) GetAvailableModels() ([]ModelInfo, error) {
	url := c.baseURL + "/api/tags"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama retornou status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Models []struct {
			Name       string `json:"name"`
			Model      string `json:"model"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
			Digest     string `json:"digest"`
			Details    struct {
				Family            string   `json:"family"`
				Families          []string `json:"families"`
				ParameterSize     string   `json:"parameter_size"`
				QuantizationLevel string   `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("erro ao parsear resposta: %w", err)
	}

	var models []ModelInfo
	for _, m := range response.Models {
		name := m.Name
		if name == "" {
			name = m.Model
		}

		// Adicionar info de parâmetros se disponível
		description := ""
		if m.Details.ParameterSize != "" {
			description = fmt.Sprintf("%s - %s", m.Details.ParameterSize, m.Details.Family)
		}

		models = append(models, ModelInfo{
			ID:          name,
			Name:        name,
			Description: description,
		})
	}

	return models, nil
}

// RecommendedModelsForTools retorna lista de modelos recomendados para function calling
func RecommendedModelsForTools() []string {
	return []string{
		"qwen2.5:7b",              // Excelente para tools
		"qwen2.5-coder:7b",        // Ótimo para código e tools
		"llama3.1:8b",             // Bom suporte geral
		"llama3.2:3b",             // Mínimo recomendado
		"mistral:7b",              // Bom suporte
		"mistral-nemo:12b",        // Excelente
		"hermes3:8b",              // Otimizado para tools
		"llama3-groq-tool-use:8b", // Específico para tools
	}
}

// IsModelSuitableForTools verifica se um modelo é adequado para function calling
func IsModelSuitableForTools(modelName string) (bool, string) {
	modelLower := strings.ToLower(modelName)

	// Modelos muito pequenos (1B ou menos)
	if strings.Contains(modelLower, ":1b") || strings.Contains(modelLower, "-1b") {
		return false, "Modelo muito pequeno para function calling. Recomendamos modelos com pelo menos 3B parâmetros."
	}

	// Modelos conhecidos por bom suporte a tools
	goodModels := []string{
		"qwen2.5", "qwen2", "qwen3",
		"llama3.1", "llama3.2", "llama3.3",
		"mistral", "mixtral",
		"hermes", "command-r", "granite",
		"groq-tool-use", "firefunction",
	}

	for _, good := range goodModels {
		if strings.Contains(modelLower, good) {
			return true, ""
		}
	}

	// Modelos desconhecidos - avisar mas permitir
	return true, "Modelo não testado para function calling. Se encontrar problemas, tente qwen2.5:7b ou llama3.1:8b."
}

// cleanToolCallJSON remove JSON de tool calls e definições de schema que o modelo incluiu no texto
func cleanToolCallJSON(content string) string {
	// Se o conteúdo contém definições de schema (comum em modelos que se confundem)
	if strings.Contains(content, "\"parameters\":") && strings.Contains(content, "\"properties\":") {
		return "" // Provavelmente é apenas o modelo repetindo o schema
	}

	// Padrões comuns que modelos escrevem no texto em vez de usar tool calling
	patterns := []string{
		`{"name":`,
		`{"type":"function"`,
		`{"tool":`,
		`{"function":`,
		`{"parameters":`,
		`[{"name":`,
		`[{"type":`,
		`"query_batch":`,
		`"execute_macro":`,
		`"list_sheets":`,
	}

	// Se o conteúdo parece ser apenas um fragmento de JSON solto
	trimmed := strings.TrimSpace(content)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) ||
		(strings.HasPrefix(trimmed, ",") && strings.Contains(trimmed, "{")) {

		// Verificar se é JSON válido ou quase isso
		if strings.Contains(trimmed, "\"name\"") || strings.Contains(trimmed, "\"function\"") {
			return ""
		}
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return extractTextFromMixedContent(content)
		}
	}

	return content
}

// extractTextFromMixedContent tenta extrair apenas texto de conteúdo misturado com JSON
func extractTextFromMixedContent(content string) string {
	var result strings.Builder
	inJSON := false
	braceCount := 0
	bracketCount := 0

	for _, char := range content {
		if char == '{' {
			if braceCount == 0 && bracketCount == 0 {
				inJSON = true
			}
			braceCount++
		} else if char == '[' && !inJSON {
			// Início de array JSON
			if bracketCount == 0 {
				inJSON = true
			}
			bracketCount++
		} else if char == '}' {
			braceCount--
			if braceCount <= 0 && bracketCount <= 0 {
				inJSON = false
				braceCount = 0
			}
		} else if char == ']' {
			bracketCount--
			if bracketCount <= 0 && braceCount <= 0 {
				inJSON = false
				bracketCount = 0
			}
		}

		if !inJSON && braceCount == 0 && bracketCount == 0 {
			result.WriteRune(char)
		}
	}

	cleaned := strings.TrimSpace(result.String())

	// Se ficou vazio ou só pontuação, retornar vazio
	if len(cleaned) < 3 {
		return ""
	}

	return cleaned
}
