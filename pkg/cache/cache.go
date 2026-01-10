package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// CacheEntry representa uma entrada no cache persistente
type CacheEntry struct {
	ID           int64
	Key          string
	Result       string
	Error         string
	StoredAt     time.Time
	AccessCount  int
	TTL          time.Duration
	TagsJSON     string // Tags armazenadas como JSON
	Tags         []string // Tags em memória
}

// CacheStatus status do cache
type CacheStatus struct {
	TotalEntries  int     // Total de entradas no cache
	HitRate       float64 // Taxa de acerto do cache (%)
	Invalidations int64   // Total de invalidações
	LastCleanup   time.Time
	DatabaseSize  int64 // Tamanho do banco em bytes
}

// PersistentCache implementa cache persistente em SQLite
type PersistentCache struct {
	db               *sql.DB
	mu               sync.RWMutex
	cacheHits        int64
	cacheMisses      int64
	cacheInvalidations int64
	cacheTTL         time.Duration
	dbPath           string
}

// NewPersistentCache cria um novo cache persistente em SQLite
func NewPersistentCache(dbPath string, ttl time.Duration) (*PersistentCache, error) {
	// Abrir conexão com SQLite
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir banco de dados: %w", err)
	}

	// Criar tabela se não existir
	err = createTables(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao criar tabelas: %w", err)
	}

	// Configurar conexão
	db.SetMaxOpenConns(1) // SQLite não suporta múltiplas escritas simultâneas
	db.SetMaxIdleConns(1)

	cache := &PersistentCache{
		db:       db,
		cacheTTL: ttl,
		dbPath:   dbPath,
	}

	// Iniciar limpeza automática de entradas expiradas
	go cache.startCleanupRoutine()

	return cache, nil
}

// createTables cria as tabelas necessárias no banco de dados
func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cache_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			result TEXT NOT NULL,
			error TEXT,
			stored_at DATETIME NOT NULL,
			access_count INTEGER NOT NULL DEFAULT 1,
			ttl_seconds INTEGER NOT NULL,
			tags TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_cache_key ON cache_entries(key);
		CREATE INDEX IF NOT EXISTS idx_cache_stored_at ON cache_entries(stored_at);
		CREATE INDEX IF NOT EXISTS idx_cache_tags ON cache_entries(tags);
	`)
	return err
}

// Get tenta obter resultado do cache
func (c *PersistentCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result string
	var storedAt time.Time
	var ttlSeconds int64
	var accessCount int
	var errStr string

	err := c.db.QueryRow(
		"SELECT result, stored_at, ttl_seconds, access_count, error FROM cache_entries WHERE key = ?",
		key,
	).Scan(&result, &storedAt, &ttlSeconds, &accessCount, &errStr)

	if err != nil {
		if err == sql.ErrNoRows {
			c.cacheMisses++
			return "", false
		}
		fmt.Printf("[CACHE DB] Erro ao consultar: %v\n", err)
		c.cacheMisses++
		return "", false
	}

	// Verificar se expirou
	ttl := time.Duration(ttlSeconds) * time.Second
	if time.Since(storedAt) > ttl {
		// Entrada expirada, remover
		go c.Delete(key)
		c.cacheMisses++
		return "", false
	}

	// Atualizar contador de acessos
	_, err = c.db.Exec(
		"UPDATE cache_entries SET access_count = access_count + 1 WHERE key = ?",
		key,
	)
	if err != nil {
		fmt.Printf("[CACHE DB] Erro ao atualizar contador: %v\n", err)
	}

	c.cacheHits++
	fmt.Printf("[CACHE DB] Hit: %s (acessos: %d)\n", key, accessCount+1)

	return result, true
}

// Set armazena resultado no cache
func (c *PersistentCache) Set(key string, result string, tags []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Converter tags para JSON
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("erro ao serializar tags: %w", err)
	}

	ttlSeconds := int64(c.cacheTTL.Seconds())
	now := time.Now()

	// Inserir ou atualizar
	_, err = c.db.Exec(`
		INSERT INTO cache_entries (key, result, stored_at, access_count, ttl_seconds, tags)
		VALUES (?, ?, ?, 1, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			result = excluded.result,
			stored_at = excluded.stored_at,
			ttl_seconds = excluded.ttl_seconds,
			tags = excluded.tags,
			access_count = 1
	`, key, result, now, ttlSeconds, string(tagsJSON))

	if err != nil {
		c.cacheMisses++
		return fmt.Errorf("erro ao inserir no cache: %w", err)
	}

	c.cacheMisses++
	fmt.Printf("[CACHE DB] Set: %s (TTL: %v, tags: %v)\n", key, c.cacheTTL, tags)
	return nil
}

// Delete remove entrada do cache
func (c *PersistentCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec("DELETE FROM cache_entries WHERE key = ?", key)
	if err != nil {
		return fmt.Errorf("erro ao deletar do cache: %w", err)
	}

	fmt.Printf("[CACHE DB] Delete: %s\n", key)
	return nil
}

// Invalidate invalida cache baseado em tags
func (c *PersistentCache) Invalidate(tags []string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(tags) == 0 {
		return 0, nil
	}

	// Construir query dinâmica para invalidação
	query := "DELETE FROM cache_entries WHERE tags LIKE ?"
	args := []interface{}{"%" + tags[0] + "%"}

	for i := 1; i < len(tags); i++ {
		query += " OR tags LIKE ?"
		args = append(args, "%"+tags[i]+"%")
	}

	// Não invalidar pela tag genérica da ferramenta
	for i, tag := range tags {
		if tag == "tool:" {
			// Remover esta tag da query
			query = "DELETE FROM cache_entries WHERE 0=1"
			args = []interface{}{}
			break
		}
		args[i] = "%invalid:" + tag + "%"
	}

	result, err := c.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("erro ao invalidar cache: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	c.cacheInvalidations += rowsAffected

	if rowsAffected > 0 {
		fmt.Printf("[CACHE DB] Invalidação: %d entradas removidas (tags: %v)\n", rowsAffected, tags)
	}

	return rowsAffected, nil
}

// Clear limpa todo o cache
func (c *PersistentCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	result, err := c.db.Exec("DELETE FROM cache_entries")
	if err != nil {
		return fmt.Errorf("erro ao limpar cache: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("[CACHE DB] Limpo: %d entradas removidas\n", rowsAffected)
	return nil
}

// Cleanup remove entradas expiradas
func (c *PersistentCache) Cleanup() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove entradas onde stored_at + ttl < agora
	result, err := c.db.Exec(`
		DELETE FROM cache_entries 
		WHERE datetime(stored_at, '+' || CAST(ttl_seconds AS TEXT) || ' seconds') < datetime('now')
	`)

	if err != nil {
		return 0, fmt.Errorf("erro ao limpar entradas expiradas: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("[CACHE DB] Limpeza: %d entradas expiradas removidas\n", rowsAffected)
	}

	return int(rowsAffected), nil
}

// GetStatus retorna status do cache
func (c *PersistentCache) GetStatus() CacheStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalEntries int
	err := c.db.QueryRow("SELECT COUNT(*) FROM cache_entries").Scan(&totalEntries)
	if err != nil {
		fmt.Printf("[CACHE DB] Erro ao contar entradas: %v\n", err)
		totalEntries = 0
	}

	// Obter tamanho do banco
	var dbSize int64
	dbSize, _ = c.getDatabaseSize()

	hitRate := 0.0
	if c.cacheHits+c.cacheMisses > 0 {
		hitRate = float64(c.cacheHits) / float64(c.cacheHits+c.cacheMisses) * 100
	}

	return CacheStatus{
		TotalEntries:  totalEntries,
		HitRate:       hitRate,
		Invalidations: c.cacheInvalidations,
		LastCleanup:   time.Now(),
		DatabaseSize:  dbSize,
	}
}

// GetEntry retorna uma entrada completa do cache
func (c *PersistentCache) GetEntry(key string) (*CacheEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var entry CacheEntry
	var tagsJSON string

	err := c.db.QueryRow(
		"SELECT id, key, result, error, stored_at, access_count, ttl_seconds, tags FROM cache_entries WHERE key = ?",
		key,
	).Scan(&entry.ID, &entry.Key, &entry.Result, &entry.Error, &entry.StoredAt, &entry.AccessCount, &entry.TTL, &tagsJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao obter entrada: %w", err)
	}

	// Parse tags JSON
	entry.TTL = time.Duration(int64(entry.TTL.Seconds())) * time.Second
	err = json.Unmarshal([]byte(tagsJSON), &entry.Tags)
	if err != nil {
		fmt.Printf("[CACHE DB] Erro ao parsear tags: %v\n", err)
		entry.Tags = []string{}
	}

	return &entry, nil
}

// GetAllEntries retorna todas as entradas do cache
func (c *PersistentCache) GetAllEntries() ([]*CacheEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.db.Query("SELECT id, key, result, error, stored_at, access_count, ttl_seconds, tags FROM cache_entries ORDER BY stored_at DESC")
	if err != nil {
		return nil, fmt.Errorf("erro ao obter entradas: %w", err)
	}
	defer rows.Close()

	var entries []*CacheEntry
	for rows.Next() {
		var entry CacheEntry
		var tagsJSON string

		err := rows.Scan(&entry.ID, &entry.Key, &entry.Result, &entry.Error, &entry.StoredAt, &entry.AccessCount, &entry.TTL, &tagsJSON)
		if err != nil {
			fmt.Printf("[CACHE DB] Erro ao scan linha: %v\n", err)
			continue
		}

		// Parse tags JSON
		entry.TTL = time.Duration(int64(entry.TTL.Seconds())) * time.Second
		err = json.Unmarshal([]byte(tagsJSON), &entry.Tags)
		if err != nil {
			fmt.Printf("[CACHE DB] Erro ao parsear tags: %v\n", err)
			entry.Tags = []string{}
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// SetTTL define o TTL padrão do cache
func (c *PersistentCache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheTTL = ttl
}

// GetTTL retorna o TTL padrão do cache
func (c *PersistentCache) GetTTL() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cacheTTL
}

// Close fecha a conexão com o banco de dados
func (c *PersistentCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Close()
}

// startCleanupRoutine inicia rotina de limpeza automática
func (c *PersistentCache) startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.Cleanup()
	}
}

// getDatabaseSize retorna o tamanho do arquivo do banco de dados
func (c *PersistentCache) getDatabaseSize() (int64, error) {
	var size int64
	err := c.db.QueryRow("SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()").Scan(&size)
	if err != nil {
		return 0, err
	}
	return size, nil
}

// Vacuum compacta o banco de dados (libera espaço)
func (c *PersistentCache) Vacuum() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("erro ao compactar banco: %w", err)
	}

	fmt.Printf("[CACHE DB] Banco compactado\n")
	return nil
}