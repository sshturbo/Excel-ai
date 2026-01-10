package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"excel-ai/pkg/logger"
)

// Config configuração do cliente AI
type Config struct {
	APIKey         string
	Model          string
	BaseURL        string
	MaxInputTokens int
}

// Client cliente para API AI (OpenAI Compatible)
type Client struct {
	config     Config
	httpClient *http.Client
}

// Message representa uma mensagem no chat
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ChatRequest requisição para a API
type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

// ChatResponse resposta da API
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewClient cria um novo cliente AI
func NewClient(apiKey, model, baseURL string) *Client {
	if model == "" {
		model = "openai/gpt-4o-mini" // modelo padrão econômico
	}
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	return &Client{
		config: Config{
			APIKey:         apiKey,
			Model:          model,
			BaseURL:        baseURL,
			MaxInputTokens: 4000, // Limite padrão seguro
		},
		httpClient: &http.Client{
			Timeout: 2 * time.Minute, // Timeout para evitar travamento
		},
	}
}

// SetBaseURL altera a URL base da API
func (c *Client) SetBaseURL(url string) {
	if url == "" {
		url = "https://openrouter.ai/api/v1"
	}
	// Não remover barra final - algumas APIs como Z.AI requerem
	c.config.BaseURL = url
}

// SetModel altera o modelo utilizado
func (c *Client) SetModel(model string) {
	c.config.Model = model
}

// SetAPIKey altera a chave da API
func (c *Client) SetAPIKey(apiKey string) {
	c.config.APIKey = apiKey
}

// GetBaseURL retorna a URL base configurada
func (c *Client) GetBaseURL() string {
	return c.config.BaseURL
}

// GetAPIKey retorna a API key configurada
func (c *Client) GetAPIKey() string {
	return c.config.APIKey
}

// SetMaxInputTokens define o limite de tokens de entrada
func (c *Client) SetMaxInputTokens(limit int) {
	c.config.MaxInputTokens = limit
}

// buildURL constrói a URL completa evitando barras duplas
func (c *Client) buildURL(path string) string {
	baseURL := c.config.BaseURL
	// Se a base URL termina com / e o path começa com /, remover uma barra
	if strings.HasSuffix(baseURL, "/") && strings.HasPrefix(path, "/") {
		return baseURL + path[1:]
	}
	// Se a base URL não termina com / e o path não começa com /, adicionar barra
	if !strings.HasSuffix(baseURL, "/") && !strings.HasPrefix(path, "/") {
		return baseURL + "/" + path
	}
	// Caso contrário, concatenar diretamente
	return baseURL + path
}

// Chat envia mensagens para a IA e retorna a resposta
func (c *Client) Chat(messages []Message) (string, error) {
	if c.config.APIKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	// Aplicar pruning no histórico para respeitar limites
	prunedMessages := PruneMessages(messages, c.config.MaxInputTokens)

	reqBody := ChatRequest{
		Model:     c.config.Model,
		Messages:  prunedMessages,
		MaxTokens: 4096, // Limite de resposta
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := c.buildURL("/chat/completions")
	logger.AIDebug(fmt.Sprintf("[HTTP REQUEST] POST %s", url))
	logger.AIDebug(fmt.Sprintf("[HTTP REQUEST] Model: %s, Messages: %d", c.config.Model, len(prunedMessages)))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey[:10]+"...")
	req.Header.Set("Accept-Language", "en-US,en")
	req.Header.Set("HTTP-Referer", "https://excel-ai-app.local")
	req.Header.Set("X-Title", "Excel-AI")

	logger.AIDebug("[HTTP REQUEST] Enviando requisição...")
	resp, err := c.httpClient.Do(req)
	logger.AIDebug(fmt.Sprintf("[HTTP RESPONSE] Status: %d, Error: %v", resp.StatusCode, err))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("erro da API: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("nenhuma resposta recebida")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// BuildExcelContext cria um contexto formatado a partir de dados do Excel
func BuildExcelContext(headers []string, rows [][]string, maxRows int) string {
	var sb strings.Builder

	// Resumo dos dados
	sb.WriteString(fmt.Sprintf("Total de colunas: %d\n", len(headers)))
	sb.WriteString(fmt.Sprintf("Total de linhas de dados: %d\n\n", len(rows)))

	// Cabeçalhos com índice de coluna
	sb.WriteString("Colunas disponíveis:\n")
	for i, h := range headers {
		colLetter := getColumnLetter(i + 1)
		sb.WriteString(fmt.Sprintf("  %s: %s\n", colLetter, h))
	}
	sb.WriteString("\n")

	// Tabela de dados
	sb.WriteString("Dados (primeiras linhas):\n\n")
	sb.WriteString("| Linha | " + strings.Join(headers, " | ") + " |\n")
	sb.WriteString("|-------|" + strings.Repeat("---|", len(headers)) + "\n")

	// Dados (limitados)
	limit := len(rows)
	if limit > maxRows {
		limit = maxRows
	}

	for i := 0; i < limit; i++ {
		rowNum := i + 2 // Linha 1 é cabeçalho, dados começam na 2
		sb.WriteString(fmt.Sprintf("| %d | %s |\n", rowNum, strings.Join(rows[i], " | ")))
	}

	if len(rows) > maxRows {
		sb.WriteString(fmt.Sprintf("\n... e mais %d linhas não exibidas\n", len(rows)-maxRows))
	}

	return sb.String()
}

// getColumnLetter converte índice numérico para letra de coluna Excel (1=A, 2=B, ..., 27=AA)
func getColumnLetter(col int) string {
	result := ""
	for col > 0 {
		col--
		result = string(rune('A'+col%26)) + result
		col /= 26
	}
	return result
}

// ModelInfo representa informações de um modelo da OpenRouter
type ModelInfo struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	ContextLength int     `json:"context_length"`
	Pricing       Pricing `json:"pricing"`
}

// Pricing representa os preços de um modelo
type Pricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

// ModelsResponse resposta da API de modelos
type ModelsResponse struct {
	Data []struct {
		ID                  string   `json:"id"`
		Name                string   `json:"name"`
		Description         string   `json:"description"`
		ContextLength       int      `json:"context_length"`
		ContextWindow       int      `json:"context_window"` // Groq usa context_window
		SupportedParameters []string `json:"supported_parameters"`
		Pricing             struct {
			Prompt     string `json:"prompt"`
			Completion string `json:"completion"`
		} `json:"pricing"`
		// Campos Groq
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// GetAvailableModels busca os modelos disponíveis na API (apenas com suporte a function calling)
func (c *Client) GetAvailableModels() ([]ModelInfo, error) {
	url := c.buildURL("/models")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// API de modelos não requer autenticação, mas se tiver key, envia
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	req.Header.Set("HTTP-Referer", "https://excel-ai-app.local")
	req.Header.Set("X-Title", "Excel-AI")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	var modelsResp ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("erro ao parsear JSON: %w", err)
	}

	// Detectar se é Groq (todos os modelos suportam function calling)
	isGroq := strings.Contains(c.config.BaseURL, "groq.com")

	var models []ModelInfo
	for _, m := range modelsResp.Data {
		// Filtrar apenas modelos com suporte a function calling (tools)
		// Groq: todos suportam, então pular filtro
		if !isGroq {
			supportsTools := false
			for _, param := range m.SupportedParameters {
				if param == "tools" {
					supportsTools = true
					break
				}
			}
			if !supportsTools {
				continue
			}
		}

		name := m.Name
		if name == "" {
			name = m.ID // Groq não retorna Name, usa ID
		}

		// Usar context_window se context_length não estiver definido (Groq)
		contextLen := m.ContextLength
		if contextLen == 0 {
			contextLen = m.ContextWindow
		}

		models = append(models, ModelInfo{
			ID:            m.ID,
			Name:          name,
			Description:   m.Description,
			ContextLength: contextLen,
			Pricing: Pricing{
				Prompt:     m.Pricing.Prompt,
				Completion: m.Pricing.Completion,
			},
		})
	}

	return models, nil
}

// parseRetryAfterHeader tenta extrair o tempo de espera do header (OpenRouter)
func parseRetryAfterHeader(resp *http.Response) time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 60 * time.Second // Default fallback
	}

	// Tentar parsear como segundos
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Tentar parsear como data (RFC1123)
	if date, err := time.Parse(time.RFC1123, retryAfter); err == nil {
		return time.Until(date)
	}

	return 60 * time.Second
}

// ChatStream envia mensagens para a IA e processa a resposta via streaming
// ctx pode ser usado para cancelar a requisição
func (c *Client) ChatStream(ctx context.Context, messages []Message, onChunk func(string) error) (string, error) {
	if c.config.APIKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	maxRetries := 2
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Aplicar pruning no histórico para respeitar limites
		prunedMessages := PruneMessages(messages, c.config.MaxInputTokens)

		reqBody := struct {
			Model     string    `json:"model"`
			Messages  []Message `json:"messages"`
			Stream    bool      `json:"stream"`
			MaxTokens int       `json:"max_tokens,omitempty"`
		}{
			Model:     c.config.Model,
			Messages:  prunedMessages,
			Stream:    true,
			MaxTokens: 4096,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", err
		}

		url := c.buildURL("/chat/completions")
		logger.AIDebug(fmt.Sprintf("[STREAM REQUEST] Tentativa %d/%d - POST %s", attempt+1, maxRetries+1, url))
		logger.AIDebug(fmt.Sprintf("[STREAM REQUEST] Model: %s, Messages: %d, Stream: true", c.config.Model, len(prunedMessages)))

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey[:10]+"...")
		req.Header.Set("Accept-Language", "en-US,en")
		req.Header.Set("HTTP-Referer", "https://excel-ai-app.local")
		req.Header.Set("X-Title", "Excel-AI")
		req.Close = true

		logger.AIDebug("[STREAM REQUEST] Enviando requisição...")
		resp, err := c.httpClient.Do(req)
		logger.AIDebug(fmt.Sprintf("[STREAM RESPONSE] Status: %d, Error: %v", resp.StatusCode, err))
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			errorMsg := string(body)

			logger.AIError(fmt.Sprintf("[STREAM ERROR] Status %d, Body: %s", resp.StatusCode, errorMsg))

			// Rate Limit / Quota errors
			if resp.StatusCode == 429 || strings.Contains(errorMsg, "rate limit") || strings.Contains(errorMsg, "quota") {
				lastErr = fmt.Errorf("limite de quota atingido: %s", errorMsg)

				if attempt < maxRetries {
					waitDuration := parseRetryAfterHeader(resp)

					// Log para o usuário
					var msg string
					switch attempt {
					case 0:
						msg = fmt.Sprintf("\n\n⏳ Quota excedida (%d). Aguardando %.0fs antes de nova tentativa...\n\n", resp.StatusCode, waitDuration.Seconds())
					case 1:
						msg = "\n\n⚠️ Nova falha de quota. Tentando novamente...\n\n"
					}
					onChunk(msg)

					// Sleep respeitando Context
					select {
					case <-time.After(waitDuration):
						continue
					case <-ctx.Done():
						return "", ctx.Err()
					}
				} else {
					onChunk("\n\n❌ Quota excedida após múltiplas tentativas.\n\n")
				}

				return "", lastErr
			}

			// Tratamento amigável para erro de política de dados em modelos gratuitos
			if strings.Contains(errorMsg, "No endpoints found matching your data policy") {
				return "", fmt.Errorf("para usar modelos gratuitos, habilite a coleta de dados no OpenRouter: https://openrouter.ai/settings/privacy")
			}

			// Tratamento amigável para erro de autenticação
			if resp.StatusCode == 401 || strings.Contains(errorMsg, "Unauthorized") || strings.Contains(errorMsg, "cookie auth") {
				return "", fmt.Errorf("API key inválida ou não autorizada - verifique se a chave está correta para o provedor selecionado")
			}

			return "", fmt.Errorf("erro na API: %s - %s", resp.Status, errorMsg)
		}

		reader := bufio.NewReader(resp.Body)
		var fullResponse strings.Builder

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				resp.Body.Close()
				return "", err
			}

			line = bytes.TrimSpace(line)
			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := bytes.TrimPrefix(line, []byte("data: "))
			if string(data) == "[DONE]" {
				break
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					fullResponse.WriteString(content)
					if err := onChunk(content); err != nil {
						resp.Body.Close()
						return "", err
					}
				}
			}
		}
		resp.Body.Close()

		return fullResponse.String(), nil
	}

	return "", lastErr
}

// ChatStreamWithTools envia mensagens com tools e retorna texto e/ou function calls
func (c *Client) ChatStreamWithTools(ctx context.Context, messages []Message, tools []Tool, onChunk func(string) error) (string, []ToolCall, error) {
	if c.config.APIKey == "" {
		return "", nil, fmt.Errorf("API key não configurada")
	}

	maxRetries := 2
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Aplicar pruning no histórico
		prunedMessages := PruneMessages(messages, c.config.MaxInputTokens)

		reqBody := struct {
			Model      string    `json:"model"`
			Messages   []Message `json:"messages"`
			Stream     bool      `json:"stream"`
			MaxTokens  int       `json:"max_tokens,omitempty"`
			Tools      []Tool    `json:"tools,omitempty"`
			ToolChoice string    `json:"tool_choice,omitempty"`
		}{
			Model:      c.config.Model,
			Messages:   prunedMessages,
			Stream:     true,
			MaxTokens:  4096,
			Tools:      tools,
			ToolChoice: "auto",
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", nil, err
		}

		url := c.buildURL("/chat/completions")
		logger.AIDebug(fmt.Sprintf("[TOOLS REQUEST] Tentativa %d/%d - POST %s", attempt+1, maxRetries+1, url))
		logger.AIDebug(fmt.Sprintf("[TOOLS REQUEST] Model: %s, Messages: %d, Tools: %d", c.config.Model, len(prunedMessages), len(tools)))

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey[:10]+"...")
		req.Header.Set("Accept-Language", "en-US,en")
		req.Header.Set("HTTP-Referer", "https://excel-ai-app.local")
		req.Header.Set("X-Title", "Excel-AI")
		req.Close = true

		logger.AIDebug("[TOOLS REQUEST] Enviando requisição...")
		resp, err := c.httpClient.Do(req)
		logger.AIDebug(fmt.Sprintf("[TOOLS RESPONSE] Status: %d, Error: %v", resp.StatusCode, err))
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			errorMsg := string(body)

			logger.AIError(fmt.Sprintf("[TOOLS ERROR] Status %d, Body: %s", resp.StatusCode, errorMsg))

			// Rate Limit / Quota errors
			if resp.StatusCode == 429 || strings.Contains(errorMsg, "rate limit") || strings.Contains(errorMsg, "quota") {
				lastErr = fmt.Errorf("limite de quota atingido: %s", errorMsg)

				if attempt < maxRetries {
					waitDuration := parseRetryAfterHeader(resp)
					var msg string
					switch attempt {
					case 0:
						msg = fmt.Sprintf("\n\n⏳ Quota excedida (%d). Aguardando %.0fs antes de nova tentativa...\n\n", resp.StatusCode, waitDuration.Seconds())
					case 1:
						msg = "\n\n⚠️ Nova falha de quota. Tentando novamente...\n\n"
					}
					onChunk(msg)

					select {
					case <-time.After(waitDuration):
						continue
					case <-ctx.Done():
						return "", nil, ctx.Err()
					}
				} else {
					onChunk("\n\n❌ Quota excedida após múltiplas tentativas.\n\n")
				}

				return "", nil, lastErr
			}

			// Tratamento amigável para erro de política de dados
			if strings.Contains(errorMsg, "No endpoints found matching your data policy") {
				return "", nil, fmt.Errorf("para usar modelos gratuitos, habilite a coleta de dados no OpenRouter: https://openrouter.ai/settings/privacy")
			}

			// Tratamento amigável para erro de autenticação
			if resp.StatusCode == 401 || strings.Contains(errorMsg, "Unauthorized") || strings.Contains(errorMsg, "cookie auth") {
				return "", nil, fmt.Errorf("API key inválida ou não autorizada - verifique se a chave está correta para o provedor selecionado")
			}

			return "", nil, fmt.Errorf("erro na API: %s - %s", resp.Status, errorMsg)
		}

		reader := bufio.NewReader(resp.Body)
		var fullResponse strings.Builder
		var toolCalls []ToolCall

		// Map para acumular tool calls parciais (streaming envia em partes)
		toolCallsMap := make(map[int]*ToolCall)

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				resp.Body.Close()
				return "", nil, err
			}

			line = bytes.TrimSpace(line)
			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := bytes.TrimPrefix(line, []byte("data: "))
			if string(data) == "[DONE]" {
				break
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content   string `json:"content,omitempty"`
						ToolCalls []struct {
							Index    int    `json:"index"`
							ID       string `json:"id,omitempty"`
							Type     string `json:"type,omitempty"`
							Function struct {
								Name      string `json:"name,omitempty"`
								Arguments string `json:"arguments,omitempty"`
							} `json:"function,omitempty"`
						} `json:"tool_calls,omitempty"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason,omitempty"`
				} `json:"choices"`
			}

			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta

				// Handle text content
				if delta.Content != "" {
					fullResponse.WriteString(delta.Content)
					if err := onChunk(delta.Content); err != nil {
						resp.Body.Close()
						return "", nil, err
					}
				}

				// Handle tool calls (streaming - may come in parts)
				for _, tc := range delta.ToolCalls {
					if existing, ok := toolCallsMap[tc.Index]; ok {
						// Append to existing tool call
						existing.Function.Arguments += tc.Function.Arguments
					} else {
						// New tool call
						toolCallsMap[tc.Index] = &ToolCall{
							ID:   tc.ID,
							Type: tc.Type,
							Function: FunctionCall{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						}
					}
				}
			}
		}
		resp.Body.Close()

		// Convert toolCallsMap to slice
		for i := 0; i < len(toolCallsMap); i++ {
			if tc, ok := toolCallsMap[i]; ok {
				toolCalls = append(toolCalls, *tc)
			}
		}

		return fullResponse.String(), toolCalls, nil
	}

	return "", nil, lastErr
}
