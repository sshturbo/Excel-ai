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
)

// parseRetryAfter tenta extrair o tempo de espera do header
func parseRetryAfter(resp *http.Response) time.Duration {
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

// formatGeminiError formata erros da API Gemini de forma amigável
func formatGeminiError(statusCode int, body string, model string) error {
	// Quota exceeded
	if strings.Contains(body, "RESOURCE_EXHAUSTED") || strings.Contains(body, "Quota exceeded") {
		return fmt.Errorf(`limite de quota atingido para o modelo "%s"

Possíveis soluções:
1. Aguarde alguns minutos e tente novamente
2. Use outro modelo como "gemini-1.5-flash" (mais disponível no tier gratuito)
3. Verifique sua quota em aistudio.google.com

Dica: modelos com sufixo "-exp" ou "-flash" geralmente têm limites mais altos`, model)
	}

	// Invalid API key
	if statusCode == 401 || strings.Contains(body, "API_KEY_INVALID") || strings.Contains(body, "UNAUTHENTICATED") {
		return fmt.Errorf(`API key inválida ou não autorizada

Verifique:
1. Se a chave foi copiada corretamente
2. Se a chave é do Google AI Studio (aistudio.google.com/apikey)
3. Se o projeto tem a API Gemini habilitada`)
	}

	// Model not found
	if strings.Contains(body, "NOT_FOUND") || strings.Contains(body, "model not found") {
		return fmt.Errorf(`modelo "%s" não encontrado

Tente usar um destes modelos:
• gemini-1.5-flash (recomendado)
• gemini-1.5-pro
• gemini-2.0-flash-exp`, model)
	}

	// Permission denied
	if strings.Contains(body, "PERMISSION_DENIED") {
		return fmt.Errorf(`acesso negado ao modelo "%s"

Este modelo pode não estar disponível para sua conta.
Tente usar "gemini-1.5-flash" que está disponível para todos`, model)
	}

	// Safety block
	if strings.Contains(body, "SAFETY") || strings.Contains(body, "blocked") {
		return fmt.Errorf(`resposta bloqueada pelo filtro de segurança do Gemini

Tente reformular sua pergunta de forma diferente`)
	}

	// Generic error with original message
	return fmt.Errorf("erro da API Gemini: %s", body)
}

// GeminiClient cliente para Google Gemini API
type GeminiClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// GeminiMessage formato de mensagem do Gemini
type GeminiMessage struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart parte de uma mensagem
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiRequest requisição para Gemini API
type GeminiRequest struct {
	Contents          []GeminiMessage   `json:"contents"`
	SystemInstruction *GeminiMessage    `json:"systemInstruction,omitempty"`
	GenerationConfig  *GenerationConfig `json:"generationConfig,omitempty"`
}

// GenerationConfig configurações de geração
type GenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

// GeminiResponse resposta do Gemini
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// NewGeminiClient cria um novo cliente Gemini
func NewGeminiClient(apiKey, model string) *GeminiClient {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &GeminiClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{},
	}
}

// SetAPIKey altera a chave da API
func (c *GeminiClient) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

// SetModel altera o modelo
func (c *GeminiClient) SetModel(model string) {
	c.model = model
}

// SetBaseURL altera a URL base (para compatibilidade)
func (c *GeminiClient) SetBaseURL(url string) {
	if url != "" {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// GetBaseURL retorna a URL base
func (c *GeminiClient) GetBaseURL() string {
	return c.baseURL
}

// GetAPIKey retorna a API key
func (c *GeminiClient) GetAPIKey() string {
	return c.apiKey
}

// convertToGeminiFormat converte mensagens OpenAI para formato Gemini
func (c *GeminiClient) convertToGeminiFormat(messages []Message) ([]GeminiMessage, *GeminiMessage) {
	var contents []GeminiMessage
	var systemInstruction *GeminiMessage

	for _, msg := range messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		// System message vai para systemInstruction
		if msg.Role == "system" {
			systemInstruction = &GeminiMessage{
				Role:  "user", // Gemini não tem role "system", usa user
				Parts: []GeminiPart{{Text: msg.Content}},
			}
			continue
		}

		contents = append(contents, GeminiMessage{
			Role:  role,
			Parts: []GeminiPart{{Text: msg.Content}},
		})
	}

	return contents, systemInstruction
}

// Chat envia mensagens para o Gemini e retorna a resposta
func (c *GeminiClient) Chat(messages []Message) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	// Strategy:
	// Attempt 0: Original Model
	// Wait (Retry-After or 60s)
	// Attempt 1: Original Model (Retry)
	// Wait (Retry-After or 60s)
	// Attempt 2: Flash Model (Last Resort)

	maxRetries := 2
	var lastErr error

	// Guardar modelo original para restaurar se precisarmos trocar
	originalModel := c.model
	defer func() { c.model = originalModel }()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Preparar request
		contents, systemInstruction := c.convertToGeminiFormat(messages)

		reqBody := GeminiRequest{
			Contents:          contents,
			SystemInstruction: systemInstruction,
			GenerationConfig: &GenerationConfig{
				MaxOutputTokens: 8192,
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", err
		}

		// Se for a última tentativa (índice 2), forçar Flash como fallback
		currentModel := c.model
		if attempt == 2 {
			currentModel = "gemini-1.5-flash"
			c.model = currentModel // Atualizar struct para erro formatado sair correto
		}

		url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, currentModel, c.apiKey)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			// Network err, maybe retry?
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			errStr := string(body)

			// Check for Quota/Rate Limit
			if resp.StatusCode == 429 || strings.Contains(errStr, "RESOURCE_EXHAUSTED") || strings.Contains(errStr, "Quota exceeded") {
				lastErr = formatGeminiError(resp.StatusCode, errStr, currentModel)

				// Se ainda temos tentativas, aguardar e tentar de novo
				if attempt < maxRetries {
					waitDuration := parseRetryAfter(resp)
					time.Sleep(waitDuration)
					continue
				}
			}
			return "", formatGeminiError(resp.StatusCode, errStr, currentModel)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return "", err
		}

		var geminiResp GeminiResponse
		if err := json.Unmarshal(body, &geminiResp); err != nil {
			return "", fmt.Errorf("erro ao parsear resposta: %w", err)
		}

		if geminiResp.Error != nil {
			// Se erro vier no corpo JSON
			errMsg := geminiResp.Error.Message
			if strings.Contains(errMsg, "RESOURCE_EXHAUSTED") || strings.Contains(errMsg, "Quota exceeded") {
				lastErr = formatGeminiError(resp.StatusCode, errMsg, currentModel)
				if attempt < maxRetries {
					time.Sleep(60 * time.Second) // JSON error geralmente não tem header, safe fallback
					continue
				}
			}
			return "", formatGeminiError(resp.StatusCode, errMsg, currentModel)
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			return "", fmt.Errorf("nenhuma resposta recebida do Gemini")
		}

		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", lastErr
}

// ChatStream envia mensagens para o Gemini com streaming e retry logic
func (c *GeminiClient) ChatStream(ctx context.Context, messages []Message, onChunk func(string) error) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	maxRetries := 2
	var lastErr error

	// Guardar modelo original
	originalModel := c.model
	defer func() { c.model = originalModel }()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Logica de Fallback (Mudança de modelo APENAS na tentativa 3/índice 2)
		if attempt == 2 {
			c.model = "gemini-1.5-flash"
		}

		contents, systemInstruction := c.convertToGeminiFormat(messages)

		reqBody := GeminiRequest{
			Contents:          contents,
			SystemInstruction: systemInstruction,
			GenerationConfig: &GenerationConfig{
				MaxOutputTokens: 8192,
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", err
		}

		url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", c.baseURL, c.model, c.apiKey)

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			// Network err - retry immediate? Let's just continue
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			errStr := string(body)

			// Check for Quota/Rate Limit
			if resp.StatusCode == 429 || strings.Contains(errStr, "RESOURCE_EXHAUSTED") || strings.Contains(errStr, "Quota exceeded") {
				lastErr = formatGeminiError(resp.StatusCode, errStr, c.model)

				if attempt < maxRetries {
					waitDuration := parseRetryAfter(resp)

					// Logs Específicos do Plano
					var msg string
					switch attempt {
					case 0:
						msg = fmt.Sprintf("\n\n⏳ Quota excedida (%d). Aguardando %.0fs antes de nova tentativa com modelo '%s'...\n\n", resp.StatusCode, waitDuration.Seconds(), c.model)
					case 1:
						msg = fmt.Sprintf("\n\n⚠️ Nova falha de quota com modelo '%s'. Tentando fallback para 'gemini-1.5-flash'...\n\n", c.model)
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
					// Falha final após fallback
					onChunk("\n\n❌ Quota excedida mesmo após fallback. Possível limite diário atingido.\n\n")
				}
			}

			return "", formatGeminiError(resp.StatusCode, errStr, c.model)
		}

		reader := bufio.NewReader(resp.Body)
		var fullResponse strings.Builder

		// Read stream
		streamSuccess := true
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				resp.Body.Close()
				return fullResponse.String(), err
			}

			line = bytes.TrimSpace(line)
			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := bytes.TrimPrefix(line, []byte("data: "))
			if len(data) == 0 {
				continue
			}

			var chunk struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
			}

			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}

			if len(chunk.Candidates) > 0 {
				if len(chunk.Candidates[0].Content.Parts) > 0 {
					text := chunk.Candidates[0].Content.Parts[0].Text
					if text != "" {
						fullResponse.WriteString(text)
						if err := onChunk(text); err != nil {
							resp.Body.Close()
							return "", err
						}
					}
				}
			}
		}
		resp.Body.Close()

		if streamSuccess {
			return fullResponse.String(), nil
		}
	}

	return "", lastErr
}

// GetAvailableModels retorna modelos Gemini disponíveis
func (c *GeminiClient) GetAvailableModels() ([]ModelInfo, error) {
	if c.apiKey == "" {
		// Retorna lista padrão se não tiver API key
		return []ModelInfo{
			{ID: "gemini-2.0-flash-exp", Name: "Gemini 2.0 Flash (Experimental)", ContextLength: 1000000},
			{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", ContextLength: 1000000},
			{ID: "gemini-1.5-flash-8b", Name: "Gemini 1.5 Flash 8B", ContextLength: 1000000},
			{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", ContextLength: 2000000},
		}, nil
	}

	url := fmt.Sprintf("%s/models?key=%s", c.baseURL, c.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var modelsResp struct {
		Models []struct {
			Name                       string   `json:"name"`
			DisplayName                string   `json:"displayName"`
			InputTokenLimit            int      `json:"inputTokenLimit"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, err
	}

	var models []ModelInfo
	for _, m := range modelsResp.Models {
		// Filtrar apenas modelos que suportam generateContent
		supportsGenerate := false
		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				supportsGenerate = true
				break
			}
		}
		if !supportsGenerate {
			continue
		}

		// Extrair ID do nome (ex: "models/gemini-1.5-flash" -> "gemini-1.5-flash")
		id := strings.TrimPrefix(m.Name, "models/")

		models = append(models, ModelInfo{
			ID:            id,
			Name:          m.DisplayName,
			ContextLength: m.InputTokenLimit,
			// Excluir metadados extras por enquanto
		})
	}

	return models, nil
}
