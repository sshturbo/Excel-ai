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
)

// ZAIClient cliente específico para Z.AI (GLM Models)
type ZAIClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	maxTokens  int
}

// NewZAIClient cria um novo cliente Z.AI
func NewZAIClient(apiKey, model string) *ZAIClient {
	if model == "" {
		model = "glm-4.7" // Modelo padrão flagship
	}
	return &ZAIClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.z.ai/api/coding/paas/v4/",
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		maxTokens: 128000, // GLM models têm 128k de contexto
	}
}

// SetAPIKey define a API key
func (c *ZAIClient) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

// SetModel define o modelo
func (c *ZAIClient) SetModel(model string) {
	c.model = model
}

// SetBaseURL define a URL base
func (c *ZAIClient) SetBaseURL(url string) {
	c.baseURL = url
	// Garantir que termina com /
	if !strings.HasSuffix(c.baseURL, "/") {
		c.baseURL += "/"
	}
}

// SetMaxInputTokens define o limite de tokens
func (c *ZAIClient) SetMaxInputTokens(limit int) {
	c.maxTokens = limit
}

// GetBaseURL retorna a URL base
func (c *ZAIClient) GetBaseURL() string {
	return c.baseURL
}

// ChatStream envia mensagens com streaming
func (c *ZAIClient) ChatStream(ctx context.Context, messages []Message, onChunk func(string) error) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	// Aplicar pruning
	prunedMessages := PruneMessages(messages, c.maxTokens)

	// Check if model supports thinking/reasoning
	var thinkingConfig *struct {
		Type string `json:"type"`
	}

	// Models known to support thinking/reasoning in Z.AI
	if strings.Contains(c.model, "glm-4.7") || strings.Contains(c.model, "thinking") {
		thinkingConfig = &struct {
			Type string `json:"type"`
		}{Type: "enabled"}
	}

	reqBody := struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		Stream      bool      `json:"stream"`
		Temperature float64   `json:"temperature,omitempty"`
		Thinking    *struct {
			Type string `json:"type"`
		} `json:"thinking,omitempty"`
	}{
		Model:       c.model,
		Messages:    prunedMessages,
		Stream:      true,
		Temperature: 1.0,
		Thinking:    thinkingConfig,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := c.baseURL + "chat/completions"
	fmt.Printf("[Z.AI STREAM] POST %s\n", url)
	fmt.Printf("[Z.AI STREAM] Model: %s, Messages: %d\n", c.model, len(prunedMessages))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// Headers específicos para Z.AI
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept-Language", "en-US,en")

	// Debug: verificar API Key
	apiKeyPreview := "(vazia)"
	if len(c.apiKey) > 10 {
		apiKeyPreview = c.apiKey[:10] + "..." + c.apiKey[len(c.apiKey)-4:]
	} else if c.apiKey != "" {
		apiKeyPreview = "***"
	}
	fmt.Printf("[Z.AI STREAM] API Key: %s (len=%d)\n", apiKeyPreview, len(c.apiKey))
	fmt.Printf("[Z.AI STREAM] Headers: Content-Type=%s, Accept-Language=%s\n",
		req.Header.Get("Content-Type"), req.Header.Get("Accept-Language"))
	fmt.Printf("[Z.AI STREAM] Request Body: %s\n", string(jsonData))

	fmt.Printf("[Z.AI STREAM] Enviando requisição...\n")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[Z.AI STREAM] Erro na requisição: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("[Z.AI STREAM] Status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorMsg := string(body)
		fmt.Printf("[Z.AI ERROR] Status %d, Body: %s\n", resp.StatusCode, errorMsg)

		// Tratamento específico de erros Z.AI
		if resp.StatusCode == 401 {
			return "", fmt.Errorf("API key inválida ou expirada")
		}
		if resp.StatusCode == 429 {
			return "", fmt.Errorf("quota excedida - verifique seu saldo em https://z.ai/manage-apikey/apikey-list")
		}
		if resp.StatusCode == 400 {
			return "", fmt.Errorf("requisição inválida: %s", errorMsg)
		}

		return "", fmt.Errorf("erro Z.AI (%d): %s", resp.StatusCode, errorMsg)
	}

	reader := bufio.NewReader(resp.Body)
	var fullResponse strings.Builder

	fmt.Printf("[Z.AI STREAM] Iniciando leitura de chunks...\n")
	chunkCount := 0

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[Z.AI STREAM] EOF alcançado. Total de chunks: %d\n", chunkCount)
				break
			}
			return "", err
		}

		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		data := bytes.TrimPrefix(line, []byte("data: "))
		if string(data) == "[DONE]" {
			fmt.Printf("[Z.AI STREAM] [DONE] recebido. Total de chunks: %d\n", chunkCount)
			break
		}

		chunkCount++
		if chunkCount%10 == 0 {
			fmt.Printf("[Z.AI STREAM] Processados %d chunks...\n", chunkCount)
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content,omitempty"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(data, &chunk); err != nil {
			continue
		}

		// Debug: log chunk structure
		if len(chunk.Choices) > 0 {
			fmt.Printf("[Z.AI DEBUG] Chunk - Content: %q, Reasoning: %q\n",
				chunk.Choices[0].Delta.Content,
				chunk.Choices[0].Delta.ReasoningContent)
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			reasoningContent := chunk.Choices[0].Delta.ReasoningContent

			// Send reasoning content first if present (but don't add to fullResponse)
			if reasoningContent != "" {
				// Wrap reasoning in special tags for frontend parsing
				reasoningChunk := ":::reasoning:::" + reasoningContent
				if err := onChunk(reasoningChunk); err != nil {
					return "", err
				}
			}

			if content != "" {
				fullResponse.WriteString(content)
				if err := onChunk(content); err != nil {
					return "", err
				}
			}
		}
	}

	return fullResponse.String(), nil
}

// ChatStreamWithTools envia mensagens com tools
func (c *ZAIClient) ChatStreamWithTools(ctx context.Context, messages []Message, tools []Tool, onChunk func(string) error) (string, []ToolCall, error) {
	if c.apiKey == "" {
		return "", nil, fmt.Errorf("API key não configurada")
	}

	// Aplicar pruning
	prunedMessages := PruneMessages(messages, c.maxTokens)

	// Check if model supports thinking/reasoning
	var thinkingConfig *struct {
		Type string `json:"type"`
	}
	if strings.Contains(c.model, "glm-4.7") || strings.Contains(c.model, "thinking") {
		thinkingConfig = &struct {
			Type string `json:"type"`
		}{Type: "enabled"}
	}

	reqBody := struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		Stream      bool      `json:"stream"`
		Tools       []Tool    `json:"tools,omitempty"`
		ToolChoice  string    `json:"tool_choice,omitempty"`
		Temperature float64   `json:"temperature,omitempty"`
		Thinking    *struct {
			Type string `json:"type"`
		} `json:"thinking,omitempty"`
	}{
		Model:       c.model,
		Messages:    prunedMessages,
		Stream:      true,
		Tools:       tools,
		ToolChoice:  "auto",
		Temperature: 1.0,
		Thinking:    thinkingConfig,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, err
	}

	url := c.baseURL + "chat/completions"
	fmt.Printf("[Z.AI TOOLS] POST %s\n", url)
	fmt.Printf("[Z.AI TOOLS] Model: %s, Messages: %d, Tools: %d\n", c.model, len(prunedMessages), len(tools))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, err
	}

	// Headers específicos para Z.AI
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept-Language", "en-US,en")

	// Debug: verificar API Key e Headers
	apiKeyPreview := "(vazia)"
	if len(c.apiKey) > 10 {
		apiKeyPreview = c.apiKey[:10] + "..." + c.apiKey[len(c.apiKey)-4:]
	} else if c.apiKey != "" {
		apiKeyPreview = "***"
	}
	fmt.Printf("[Z.AI TOOLS] API Key: %s (len=%d)\n", apiKeyPreview, len(c.apiKey))
	fmt.Printf("[Z.AI TOOLS] Headers: Content-Type=%s, Accept-Language=%s\n",
		req.Header.Get("Content-Type"), req.Header.Get("Accept-Language"))

	// Log do request body (limitado a 500 chars para não poluir)
	requestBodyPreview := string(jsonData)
	if len(requestBodyPreview) > 500 {
		requestBodyPreview = requestBodyPreview[:500] + "...[truncated]"
	}
	fmt.Printf("[Z.AI TOOLS] Request Body: %s\n", requestBodyPreview)

	fmt.Printf("[Z.AI TOOLS] Enviando requisição...\n")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[Z.AI TOOLS] Erro na requisição: %v\n", err)
		return "", nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("[Z.AI TOOLS] Status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorMsg := string(body)
		fmt.Printf("[Z.AI ERROR] Status %d, Body: %s\n", resp.StatusCode, errorMsg)

		// Tratamento específico de erros Z.AI
		if resp.StatusCode == 401 {
			return "", nil, fmt.Errorf("API key inválida ou expirada")
		}
		if resp.StatusCode == 429 {
			return "", nil, fmt.Errorf("quota excedida - verifique seu saldo em https://z.ai/manage-apikey/apikey-list")
		}
		if resp.StatusCode == 400 {
			return "", nil, fmt.Errorf("requisição inválida: %s", errorMsg)
		}

		return "", nil, fmt.Errorf("erro Z.AI (%d): %s", resp.StatusCode, errorMsg)
	}

	reader := bufio.NewReader(resp.Body)
	var fullResponse strings.Builder
	var toolCalls []ToolCall

	// Map para acumular tool calls parciais
	toolCallsMap := make(map[int]*ToolCall)

	fmt.Printf("[Z.AI TOOLS STREAM] Iniciando leitura de chunks...\n")
	chunkCount := 0

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[Z.AI TOOLS STREAM] EOF alcançado. Total de chunks: %d\n", chunkCount)
				break
			}
			return "", nil, err
		}

		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		data := bytes.TrimPrefix(line, []byte("data: "))
		if string(data) == "[DONE]" {
			fmt.Printf("[Z.AI TOOLS STREAM] [DONE] recebido. Total de chunks: %d\n", chunkCount)
			break
		}

		chunkCount++
		if chunkCount%10 == 0 {
			fmt.Printf("[Z.AI TOOLS STREAM] Processados %d chunks...\n", chunkCount)
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content          string `json:"content,omitempty"`
					ReasoningContent string `json:"reasoning_content,omitempty"`
					ToolCalls        []struct {
						Index    int    `json:"index"`
						ID       string `json:"id,omitempty"`
						Type     string `json:"type,omitempty"`
						Function struct {
							Name      string `json:"name,omitempty"`
							Arguments string `json:"arguments,omitempty"`
						} `json:"function,omitempty"`
					} `json:"tool_calls,omitempty"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(data, &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta

			// Handle reasoning content first (but don't add to fullResponse)
			if delta.ReasoningContent != "" {
				// Wrap reasoning in special tags for frontend parsing
				reasoningChunk := ":::reasoning:::" + delta.ReasoningContent
				if err := onChunk(reasoningChunk); err != nil {
					return "", nil, err
				}
			}

			// Handle text content (add to fullResponse)
			if delta.Content != "" {
				fullResponse.WriteString(delta.Content)
				if err := onChunk(delta.Content); err != nil {
					return "", nil, err
				}
			}

			// Handle tool calls
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

	// Convert toolCallsMap to slice
	for i := 0; i < len(toolCallsMap); i++ {
		if tc, ok := toolCallsMap[i]; ok {
			toolCalls = append(toolCalls, *tc)
		}
	}

	return fullResponse.String(), toolCalls, nil
}

// GetAvailableModels retorna os modelos disponíveis (hardcoded para Z.AI)
func (c *ZAIClient) GetAvailableModels() ([]ModelInfo, error) {
	// Z.AI não tem endpoint /models funcional, retornar lista hardcoded
	return []ModelInfo{
		{ID: "glm-4.7", Name: "GLM-4.7", Description: "Latest flagship model - optimized for coding", ContextLength: 128000, Pricing: Pricing{Prompt: "¥2.5/1M", Completion: "¥10/1M"}},
		{ID: "glm-4.6v", Name: "GLM-4.6V", Description: "Vision model with multimodal capabilities", ContextLength: 128000, Pricing: Pricing{Prompt: "¥2/1M", Completion: "¥8/1M"}},
		{ID: "glm-4.6", Name: "GLM-4.6", Description: "Balanced model with native function calling", ContextLength: 128000, Pricing: Pricing{Prompt: "¥1.5/1M", Completion: "¥6/1M"}},
		{ID: "glm-4.5", Name: "GLM-4.5", Description: "Cost-effective model with function calling", ContextLength: 128000, Pricing: Pricing{Prompt: "¥1/1M", Completion: "¥4/1M"}},
		{ID: "glm-4.5-air", Name: "GLM-4.5 Air", Description: "Lightweight model for fast responses", ContextLength: 128000, Pricing: Pricing{Prompt: "¥0.5/1M", Completion: "¥2/1M"}},
	}, nil
}
