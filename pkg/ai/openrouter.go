package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Config configuração do cliente OpenRouter
type Config struct {
	APIKey string
	Model  string
}

// Client cliente para API OpenRouter
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
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
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

// NewClient cria um novo cliente OpenRouter
func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = "openai/gpt-4o-mini" // modelo padrão econômico
	}
	return &Client{
		config: Config{
			APIKey: apiKey,
			Model:  model,
		},
		httpClient: &http.Client{},
	}
}

// SetModel altera o modelo utilizado
func (c *Client) SetModel(model string) {
	c.config.Model = model
}

// SetAPIKey altera a chave da API
func (c *Client) SetAPIKey(apiKey string) {
	c.config.APIKey = apiKey
}

// Chat envia mensagens para a IA e retorna a resposta
func (c *Client) Chat(messages []Message) (string, error) {
	if c.config.APIKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	reqBody := ChatRequest{
		Model:    c.config.Model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
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
