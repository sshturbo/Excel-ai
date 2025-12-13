package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Config configuração do cliente AI
type Config struct {
	APIKey  string
	Model   string
	BaseURL string
}

// Client cliente para API AI (OpenAI Compatible)
type Client struct {
	config     Config
	httpClient *http.Client
}

// Message representa uma mensagem no chat
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
	// Remover barra final se existir
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		config: Config{
			APIKey:  apiKey,
			Model:   model,
			BaseURL: baseURL,
		},
		httpClient: &http.Client{},
	}
}

// SetBaseURL altera a URL base da API
func (c *Client) SetBaseURL(url string) {
	if url == "" {
		url = "https://openrouter.ai/api/v1"
	}
	c.config.BaseURL = strings.TrimRight(url, "/")
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

// Chat envia mensagens para a IA e retorna a resposta
func (c *Client) Chat(messages []Message) (string, error) {
	if c.config.APIKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	reqBody := ChatRequest{
		Model:     c.config.Model,
		Messages:  messages,
		MaxTokens: 4096, // Limite razoável para evitar erro de crédito insuficiente
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("HTTP-Referer", "https://excel-ai-app.local")
	req.Header.Set("X-Title", "Excel-AI")

	resp, err := c.httpClient.Do(req)
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
		ID            string `json:"id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		ContextLength int    `json:"context_length"`
		ContextWindow int    `json:"context_window"` // Groq usa context_window
		Pricing       struct {
			Prompt     string `json:"prompt"`
			Completion string `json:"completion"`
		} `json:"pricing"`
		// Campos Groq
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// GetAvailableModels busca os modelos disponíveis na API
func (c *Client) GetAvailableModels() ([]ModelInfo, error) {
	url := c.config.BaseURL + "/models"

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

	var models []ModelInfo
	for _, m := range modelsResp.Data {
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

// ChatStream envia mensagens para a IA e processa a resposta via streaming
func (c *Client) ChatStream(messages []Message, onChunk func(string) error) (string, error) {
	if c.config.APIKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	reqBody := struct {
		Model     string    `json:"model"`
		Messages  []Message `json:"messages"`
		Stream    bool      `json:"stream"`
		MaxTokens int       `json:"max_tokens,omitempty"`
	}{
		Model:     c.config.Model,
		Messages:  messages,
		Stream:    true,
		MaxTokens: 4096, // Limite razoável para evitar erro de crédito insuficiente
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("HTTP-Referer", "https://excel-ai-app.local")
	req.Header.Set("X-Title", "Excel-AI")
	// Evita reutilização de conexão após streaming (reduz risco de bytes pendentes em keep-alive).
	req.Close = true

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorMsg := string(body)

		// Tratamento amigável para erro de política de dados em modelos gratuitos
		if strings.Contains(errorMsg, "No endpoints found matching your data policy") {
			return "", fmt.Errorf("Para usar modelos gratuitos, você precisa habilitar a coleta de dados no OpenRouter.\nAcesse: https://openrouter.ai/settings/privacy")
		}

		// Tratamento amigável para erro de autenticação
		if resp.StatusCode == 401 || strings.Contains(errorMsg, "Unauthorized") || strings.Contains(errorMsg, "cookie auth") {
			return "", fmt.Errorf("API Key inválida ou não autorizada. Verifique:\n1. A chave está correta para o provedor selecionado\n2. Se mudou de Groq para OpenRouter (ou vice-versa), atualize a API Key")
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
					return "", err
				}
			}
		}
	}

	return fullResponse.String(), nil
}
