package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

// Message representa uma mensagem do chat
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Hidden    bool      `json:"hidden,omitempty"`
}

// Conversation representa uma conversa salva
type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Messages  []Message `json:"messages"`
	Context   string    `json:"context,omitempty"`
	ExcelPath string    `json:"excelPath,omitempty"`
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

// ProviderConfig armazena configurações específicas de um provedor
type ProviderConfig struct {
	APIKey    string `json:"apiKey,omitempty"`
	Model     string `json:"model,omitempty"`
	ToolModel string `json:"toolModel,omitempty"`
	BaseURL   string `json:"baseUrl,omitempty"`
}

// Config armazena configurações do app
type Config struct {
	// Configurações atuais (provider selecionado)
	Provider  string `json:"provider,omitempty"` // "openrouter", "groq", "custom"
	APIKey    string `json:"apiKey,omitempty"`   // API key do provider atual
	Model     string `json:"model"`              // Modelo do provider atual
	ToolModel string `json:"toolModel"`          // Modelo secundário para execução de ferramentas (opcional)
	BaseURL   string `json:"baseUrl,omitempty"`  // URL base do provider atual

	// Configurações salvas por provedor
	ProviderConfigs map[string]ProviderConfig `json:"providerConfigs,omitempty"`

	// Configurações gerais
	MaxRowsContext  int    `json:"maxRowsContext"`  // Máximo de linhas enviadas para IA
	MaxContextChars int    `json:"maxContextChars"` // Máximo de caracteres no contexto (Token Limit Control)
	MaxRowsPreview  int    `json:"maxRowsPreview"`  // Máximo de linhas no preview
	IncludeHeaders  bool   `json:"includeHeaders"`  // Incluir cabeçalhos no contexto
	AskBeforeApply  bool   `json:"askBeforeApply"`  // Modo Seguro vs YOLO
	DetailLevel     string `json:"detailLevel"`     // "minimal", "normal", "detailed"
	CustomPrompt    string `json:"customPrompt"`    // Prompt personalizado adicional
	Language        string `json:"language"`        // Idioma das respostas
	LastUsedWb      string `json:"lastUsedWorkbook,omitempty"`
}

// Storage gerencia persistência de dados
type Storage struct {
	db *sql.DB
}

// NewStorage cria nova instância do storage
func NewStorage() (*Storage, error) {
	// Usar pasta do usuário para armazenamento
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	basePath := filepath.Join(homeDir, ".excel-ai")
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(basePath, "excel-ai.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := initDB(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{db: db}, nil
}

func initDB(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			title TEXT,
			context TEXT,
			excel_path TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id TEXT,
			role TEXT,
			content TEXT,
			timestamp DATETIME,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS checkpoints (
			id TEXT PRIMARY KEY,
			conversation_id TEXT,
			name TEXT,
			messages TEXT,
			context TEXT,
			created_at DATETIME,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS undo_actions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id TEXT NOT NULL,
			batch_id INTEGER NOT NULL,
			operation_type TEXT NOT NULL DEFAULT 'write',
			workbook TEXT,
			sheet TEXT,
			cell TEXT,
			old_value TEXT,
			undo_data TEXT,
			approved BOOLEAN DEFAULT FALSE,
			created_at DATETIME,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("erro ao inicializar DB: %w", err)
		}
	}

	// Migrations (ignore errors if columns exist)
	migrations := []string{
		`ALTER TABLE undo_actions ADD COLUMN operation_type TEXT NOT NULL DEFAULT 'write'`,
		`ALTER TABLE undo_actions ADD COLUMN undo_data TEXT`,
		`ALTER TABLE conversations ADD COLUMN excel_path TEXT`,
		`UPDATE messages SET role = 'system' WHERE content LIKE 'TOOL RESULTS:%' AND role = 'user'`,
	}
	for _, m := range migrations {
		db.Exec(m)
	}

	return nil
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

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Salvar conversa
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO conversations (id, title, context, excel_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, conv.ID, conv.Title, conv.Context, conv.ExcelPath, conv.CreatedAt, conv.UpdatedAt)
	if err != nil {
		return err
	}

	// Limpar mensagens antigas (simples estratégia de replace)
	_, err = tx.Exec("DELETE FROM messages WHERE conversation_id = ?", conv.ID)
	if err != nil {
		return err
	}

	// Inserir mensagens
	stmt, err := tx.Prepare(`
		INSERT INTO messages (conversation_id, role, content, timestamp)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, msg := range conv.Messages {
		_, err = stmt.Exec(conv.ID, msg.Role, msg.Content, msg.Timestamp)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadConversation carrega uma conversa pelo ID
func (s *Storage) LoadConversation(id string) (*Conversation, error) {
	var conv Conversation
	err := s.db.QueryRow(`
		SELECT id, title, context, excel_path, created_at, updated_at 
		FROM conversations WHERE id = ?
	`, id).Scan(&conv.ID, &conv.Title, &conv.Context, &conv.ExcelPath, &conv.CreatedAt, &conv.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversa não encontrada")
		}
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT role, content, timestamp 
		FROM messages 
		WHERE conversation_id = ? 
		ORDER BY id ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Role, &msg.Content, &msg.Timestamp); err != nil {
			return nil, err
		}
		conv.Messages = append(conv.Messages, msg)
	}

	return &conv, nil
}

// ListConversations lista todas as conversas (resumo)
func (s *Storage) ListConversations() ([]ConversationSummary, error) {
	rows, err := s.db.Query(`
		SELECT id, title, updated_at 
		FROM conversations 
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []ConversationSummary
	for rows.Next() {
		var summary ConversationSummary
		if err := rows.Scan(&summary.ID, &summary.Title, &summary.UpdatedAt); err != nil {
			return nil, err
		}

		// Pegar preview (última mensagem)
		var preview string
		_ = s.db.QueryRow("SELECT content FROM messages WHERE conversation_id = ? ORDER BY id DESC LIMIT 1", summary.ID).Scan(&preview)
		summary.Preview = truncateString(preview, 100)

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// DeleteConversation remove uma conversa
func (s *Storage) DeleteConversation(id string) error {
	// Com ON DELETE CASCADE, deletar a conversa deleta as mensagens
	// Mas SQLite precisa de PRAGMA foreign_keys = ON para isso funcionar automaticamente
	// Vamos garantir deletando manualmente ou ativando PRAGMA

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM messages WHERE conversation_id = ?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM conversations WHERE id = ?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SaveConfig salva configurações
func (s *Storage) SaveConfig(cfg *Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)
	`, "main_config", string(data))
	return err
}

// LoadConfig carrega configurações
func (s *Storage) LoadConfig() (*Config, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", "main_config").Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			// Retornar configuração padrão Groq
			return &Config{
				Provider:        "groq",
				BaseURL:         "https://api.groq.com/openai/v1",
				APIKey:          "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",
				Model:           "openai/gpt-oss-120b",
				ProviderConfigs: make(map[string]ProviderConfig),
				MaxRowsContext:  50,
				MaxContextChars: 6000, // Defeito 8k TPM safe
				MaxRowsPreview:  100,
				IncludeHeaders:  true,
				AskBeforeApply:  true, // Modo seguro por padrão
				DetailLevel:     "normal",
				Language:        "pt-BR",
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal([]byte(value), &cfg); err != nil {
		return nil, err
	}

	// Inicializar mapa se não existir
	if cfg.ProviderConfigs == nil {
		cfg.ProviderConfigs = make(map[string]ProviderConfig)
	}

	// Default para novos campos (migration on-the-fly)
	if cfg.MaxContextChars == 0 {
		cfg.MaxContextChars = 6000
	}

	return &cfg, nil
}

// GetProviderConfig retorna configurações de um provedor específico
func (s *Storage) GetProviderConfig(providerName string) (*ProviderConfig, error) {
	cfg, err := s.LoadConfig()
	if err != nil {
		return nil, err
	}

	if providerCfg, exists := cfg.ProviderConfigs[providerName]; exists {
		return &providerCfg, nil
	}

	return nil, nil // Não existe configuração para este provedor
}

// SetProviderConfig salva configurações de um provedor específico
func (s *Storage) SetProviderConfig(providerName string, apiKey, model, toolModel, baseURL string) error {
	cfg, err := s.LoadConfig()
	if err != nil {
		return err
	}

	if cfg.ProviderConfigs == nil {
		cfg.ProviderConfigs = make(map[string]ProviderConfig)
	}

	cfg.ProviderConfigs[providerName] = ProviderConfig{
		APIKey:    apiKey,
		Model:     model,
		ToolModel: toolModel,
		BaseURL:   baseURL,
	}

	return s.SaveConfig(cfg)
}

// SwitchProvider muda para outro provedor, carregando suas configurações salvas
func (s *Storage) SwitchProvider(providerName string) (*Config, error) {
	cfg, err := s.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Atualizar provider atual
	cfg.Provider = providerName

	// Carregar configurações salvas deste provedor
	if providerCfg, exists := cfg.ProviderConfigs[providerName]; exists {
		cfg.APIKey = providerCfg.APIKey
		cfg.Model = providerCfg.Model
		cfg.ToolModel = providerCfg.ToolModel
		cfg.BaseURL = providerCfg.BaseURL
	} else {
		// Sem configuração salva, limpar credenciais e usar defaults
		cfg.APIKey = ""
		cfg.Model = ""
		cfg.ToolModel = ""

		// Definir BaseURL padrão por provedor
		switch providerName {
		case "groq":
			cfg.BaseURL = "https://api.groq.com/openai/v1"
		case "openrouter":
			cfg.BaseURL = "https://openrouter.ai/api/v1"
		case "google":
			cfg.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
			cfg.Model = "gemini-1.5-flash" // Default model for Google
		default:
			cfg.BaseURL = ""
		}
	}

	// Salvar a mudança
	if err := s.SaveConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
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

// ========== HIPO.md - Contexto por projeto ==========

// LoadProjectContext carrega o arquivo HIPO.md do diretório especificado
// Retorna o conteúdo do arquivo ou string vazia se não existir
func LoadProjectContext(projectDir string) (string, error) {
	hipoPath := filepath.Join(projectDir, "HIPO.md")

	// Tenta HIPO.md primeiro, depois .hipo.md (hidden file)
	if _, err := os.Stat(hipoPath); os.IsNotExist(err) {
		hipoPath = filepath.Join(projectDir, ".hipo.md")
		if _, err := os.Stat(hipoPath); os.IsNotExist(err) {
			return "", nil // Sem arquivo de contexto, não é erro
		}
	}

	content, err := os.ReadFile(hipoPath)
	if err != nil {
		return "", fmt.Errorf("erro ao ler HIPO.md: %w", err)
	}

	return string(content), nil
}

// ========== Checkpointing de conversas ==========

// Checkpoint representa um ponto salvo de uma conversa
type Checkpoint struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversationId"`
	Name           string    `json:"name"`
	Messages       []Message `json:"messages"`
	Context        string    `json:"context,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

// SaveCheckpoint salva um checkpoint de uma conversa
func (s *Storage) SaveCheckpoint(conversationID, name string, messages []Message, context string) error {
	checkpoint := Checkpoint{
		ID:             GenerateID() + "-cp",
		ConversationID: conversationID,
		Name:           name,
		Messages:       messages,
		Context:        context,
		CreatedAt:      time.Now(),
	}

	messagesJSON, err := json.Marshal(checkpoint.Messages)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO checkpoints (id, conversation_id, name, messages, context, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		checkpoint.ID, checkpoint.ConversationID, checkpoint.Name,
		string(messagesJSON), checkpoint.Context, checkpoint.CreatedAt)

	return err
}

// LoadCheckpoint carrega um checkpoint específico
func (s *Storage) LoadCheckpoint(checkpointID string) (*Checkpoint, error) {
	var cp Checkpoint
	var messagesJSON string

	err := s.db.QueryRow(`
		SELECT id, conversation_id, name, messages, context, created_at
		FROM checkpoints WHERE id = ?`, checkpointID).
		Scan(&cp.ID, &cp.ConversationID, &cp.Name, &messagesJSON, &cp.Context, &cp.CreatedAt)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(messagesJSON), &cp.Messages); err != nil {
		return nil, err
	}

	return &cp, nil
}

// ListCheckpoints lista todos os checkpoints de uma conversa
func (s *Storage) ListCheckpoints(conversationID string) ([]Checkpoint, error) {
	rows, err := s.db.Query(`
		SELECT id, conversation_id, name, created_at
		FROM checkpoints 
		WHERE conversation_id = ?
		ORDER BY created_at DESC`, conversationID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkpoints []Checkpoint
	for rows.Next() {
		var cp Checkpoint
		if err := rows.Scan(&cp.ID, &cp.ConversationID, &cp.Name, &cp.CreatedAt); err != nil {
			continue
		}
		checkpoints = append(checkpoints, cp)
	}

	return checkpoints, nil
}

// DeleteCheckpoint remove um checkpoint
func (s *Storage) DeleteCheckpoint(checkpointID string) error {
	_, err := s.db.Exec("DELETE FROM checkpoints WHERE id = ?", checkpointID)
	return err
}

// LicenseInfo estrutura da licença (compatível com pkg/license)
type LicenseInfo struct {
	Hash          string    `json:"hash"`
	MachineID     string    `json:"machine_id"`
	ActivatedAt   time.Time `json:"activated_at"`
	LastValidated time.Time `json:"last_validated"`
}

// SaveLicense salva a licença no banco
func (s *Storage) SaveLicense(license *LicenseInfo) error {
	data, err := json.Marshal(license)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO settings (key, value) VALUES ('license', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, string(data))

	return err
}

// LoadLicense carrega a licença do banco
func (s *Storage) LoadLicense() (*LicenseInfo, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", "license").Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Sem licença
		}
		return nil, err
	}

	var license LicenseInfo
	if err := json.Unmarshal([]byte(value), &license); err != nil {
		return nil, err
	}

	return &license, nil
}

// ========== Undo Actions por Conversa ==========

// UndoAction representa uma ação que pode ser desfeita
type UndoAction struct {
	ID             int64     `json:"id"`
	ConversationID string    `json:"conversationId"`
	BatchID        int64     `json:"batchId"`
	OperationType  string    `json:"operationType"` // write, create-sheet, rename-sheet, etc.
	Workbook       string    `json:"workbook"`
	Sheet          string    `json:"sheet"`
	Cell           string    `json:"cell"`
	OldValue       string    `json:"oldValue"`
	UndoData       string    `json:"undoData"` // JSON with operation-specific undo data
	Approved       bool      `json:"approved"`
	CreatedAt      time.Time `json:"createdAt"`
}

// SaveUndoAction salva uma ação de undo vinculada a uma conversa
// SaveUndoAction salva uma ação de undo simple (compatibilidade)
func (s *Storage) SaveUndoAction(convID string, batchID int64, workbook, sheet, cell, oldValue string) error {
	return s.SaveUndoActionFull(convID, batchID, "write", workbook, sheet, cell, oldValue, "")
}

// SaveUndoActionFull salva uma ação de undo com todos os detalhes
func (s *Storage) SaveUndoActionFull(convID string, batchID int64, opType, workbook, sheet, cell, oldValue, undoData string) error {
	_, err := s.db.Exec(`
		INSERT INTO undo_actions (conversation_id, batch_id, operation_type, workbook, sheet, cell, old_value, undo_data, approved, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, FALSE, ?)
	`, convID, batchID, opType, workbook, sheet, cell, oldValue, undoData, time.Now())
	return err
}

// GetPendingUndoActions retorna ações não aprovadas de uma conversa
func (s *Storage) GetPendingUndoActions(convID string) ([]UndoAction, error) {
	rows, err := s.db.Query(`
		SELECT id, conversation_id, batch_id, operation_type, workbook, sheet, cell, old_value, undo_data, approved, created_at
		FROM undo_actions 
		WHERE conversation_id = ? AND approved = FALSE
		ORDER BY id DESC
	`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []UndoAction
	for rows.Next() {
		var a UndoAction
		var opType, undoData sql.NullString

		if err := rows.Scan(&a.ID, &a.ConversationID, &a.BatchID, &opType, &a.Workbook, &a.Sheet, &a.Cell, &a.OldValue, &undoData, &a.Approved, &a.CreatedAt); err != nil {
			continue
		}

		// Handle potential NULLs for new columns existing rows
		a.OperationType = opType.String
		if a.OperationType == "" {
			a.OperationType = "write" // Default legacy
		}
		a.UndoData = undoData.String

		actions = append(actions, a)
	}
	return actions, nil
}

// ApproveUndoActions marca ações de uma conversa como aprovadas (não podem mais ser desfeitas)
func (s *Storage) ApproveUndoActions(convID string) error {
	_, err := s.db.Exec(`UPDATE undo_actions SET approved = TRUE WHERE conversation_id = ? AND approved = FALSE`, convID)
	return err
}

// DeleteUndoActions remove ações de undo de uma conversa (após desfazer)
func (s *Storage) DeleteUndoActions(convID string, batchID int64) error {
	_, err := s.db.Exec(`DELETE FROM undo_actions WHERE conversation_id = ? AND batch_id = ?`, convID, batchID)
	return err
}

// HasPendingUndoActions verifica se há ações pendentes (não aprovadas) para uma conversa
func (s *Storage) HasPendingUndoActions(convID string) (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM undo_actions WHERE conversation_id = ? AND approved = FALSE`, convID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetLastBatchID retorna o último batch_id de uma conversa
func (s *Storage) GetLastBatchID(convID string) (int64, error) {
	var batchID int64
	err := s.db.QueryRow(`
		SELECT batch_id FROM undo_actions 
		WHERE conversation_id = ? AND approved = FALSE
		ORDER BY id DESC LIMIT 1
	`, convID).Scan(&batchID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return batchID, nil
}

// SetConversationExcelPath salva o caminho do Excel vinculado à conversa
func (s *Storage) SetConversationExcelPath(convID, path string) error {
	_, err := s.db.Exec(`UPDATE conversations SET excel_path = ?, updated_at = ? WHERE id = ?`, path, time.Now(), convID)
	return err
}

// GetConversationExcelPath retorna o caminho do Excel vinculado à conversa
func (s *Storage) GetConversationExcelPath(convID string) (string, error) {
	var path sql.NullString
	err := s.db.QueryRow(`SELECT excel_path FROM conversations WHERE id = ?`, convID).Scan(&path)
	if err != nil {
		return "", err
	}
	return path.String, nil
}
