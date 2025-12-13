package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Message representa uma mensagem do chat
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Conversation representa uma conversa salva
type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Messages  []Message `json:"messages"`
	Context   string    `json:"context,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ConversationSummary resumo para listagem
type ConversationSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Preview   string    `json:"preview"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Storage gerencia persistência de dados
type Storage struct {
	basePath string
}

// NewStorage cria nova instância do storage
func NewStorage() (*Storage, error) {
	// Usar pasta do usuário para armazenamento
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	basePath := filepath.Join(homeDir, ".excel-ai")

	// Criar diretórios necessários
	dirs := []string{
		basePath,
		filepath.Join(basePath, "conversations"),
		filepath.Join(basePath, "config"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	return &Storage{basePath: basePath}, nil
}

// SaveConversation salva uma conversa
func (s *Storage) SaveConversation(conv *Conversation) error {
	conv.UpdatedAt = time.Now()
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}

	// Gerar título automático se não existir
	if conv.Title == "" && len(conv.Messages) > 0 {
		for _, msg := range conv.Messages {
			if msg.Role == "user" {
				conv.Title = truncateString(msg.Content, 50)
				break
			}
		}
	}

	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(s.basePath, "conversations", conv.ID+".json")
	return os.WriteFile(filePath, data, 0644)
}

// LoadConversation carrega uma conversa pelo ID
func (s *Storage) LoadConversation(id string) (*Conversation, error) {
	filePath := filepath.Join(s.basePath, "conversations", id+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, err
	}

	return &conv, nil
}

// ListConversations lista todas as conversas
func (s *Storage) ListConversations() ([]ConversationSummary, error) {
	convDir := filepath.Join(s.basePath, "conversations")

	files, err := os.ReadDir(convDir)
	if err != nil {
		return nil, err
	}

	var summaries []ConversationSummary

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		conv, err := s.LoadConversation(file.Name()[:len(file.Name())-5])
		if err != nil {
			continue
		}

		preview := ""
		if len(conv.Messages) > 0 {
			lastMsg := conv.Messages[len(conv.Messages)-1]
			preview = truncateString(lastMsg.Content, 100)
		}

		summaries = append(summaries, ConversationSummary{
			ID:        conv.ID,
			Title:     conv.Title,
			Preview:   preview,
			UpdatedAt: conv.UpdatedAt,
		})
	}

	// Ordenar por data de atualização (mais recente primeiro)
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].UpdatedAt.After(summaries[j].UpdatedAt)
	})

	return summaries, nil
}

// DeleteConversation remove uma conversa
func (s *Storage) DeleteConversation(id string) error {
	filePath := filepath.Join(s.basePath, "conversations", id+".json")
	return os.Remove(filePath)
}

// Config armazena configurações do app
type Config struct {
	APIKey         string `json:"apiKey,omitempty"`
	Model          string `json:"model"`
	MaxRowsContext int    `json:"maxRowsContext"` // Máximo de linhas enviadas para IA
	MaxRowsPreview int    `json:"maxRowsPreview"` // Máximo de linhas no preview
	IncludeHeaders bool   `json:"includeHeaders"` // Incluir cabeçalhos no contexto
	DetailLevel    string `json:"detailLevel"`    // "minimal", "normal", "detailed"
	CustomPrompt   string `json:"customPrompt"`   // Prompt personalizado adicional
	Language       string `json:"language"`       // Idioma das respostas
	LastUsedWb     string `json:"lastUsedWorkbook,omitempty"`
}

// SaveConfig salva configurações
func (s *Storage) SaveConfig(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(s.basePath, "config", "settings.json")
	return os.WriteFile(filePath, data, 0644)
}

// LoadConfig carrega configurações
func (s *Storage) LoadConfig() (*Config, error) {
	filePath := filepath.Join(s.basePath, "config", "settings.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Model:          "openai/gpt-4o-mini",
				MaxRowsContext: 50,
				MaxRowsPreview: 100,
				IncludeHeaders: true,
				DetailLevel:    "normal",
				Language:       "pt-BR",
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// truncateString trunca uma string para o tamanho máximo
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GenerateID gera um ID único para conversas
func GenerateID() string {
	return time.Now().Format("20060102-150405")
}
