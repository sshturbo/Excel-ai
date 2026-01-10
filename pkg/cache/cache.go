package cache

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"

	"excel-ai/pkg/logger"
)

const (
	cacheBucketName = "cache_entries"
)

// CacheEntry representa uma entrada no cache persistente
type CacheEntry struct {
	Key         string
	Result      string
	Error       string
	StoredAt    time.Time
	AccessCount int
	TTL         time.Duration
	Tags        []string
}

// CacheStatus status do cache
type CacheStatus struct {
	TotalEntries  int
	HitRate       float64
	Invalidations int64
	LastCleanup   time.Time
	DatabaseSize  int64
}

// PersistentCache implementa cache persistente em BoltDB
type PersistentCache struct {
	db                 *bolt.DB
	mu                 sync.RWMutex
	cacheHits          int64
	cacheMisses        int64
	cacheInvalidations int64
	cacheTTL           time.Duration
	dbPath             string
}

// NewPersistentCache cria um novo cache persistente em BoltDB
func NewPersistentCache(dbPath string, ttl time.Duration) (*PersistentCache, error) {
	// Abrir conexão com BoltDB
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir banco de dados BoltDB: %w", err)
	}

	// Criar bucket se não existir
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(cacheBucketName))
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao criar bucket: %w", err)
	}

	cache := &PersistentCache{
		db:       db,
		cacheTTL: ttl,
		dbPath:   dbPath,
	}

	// Iniciar limpeza automática de entradas expiradas
	go cache.startCleanupRoutine()

	return cache, nil
}

// Get tenta obter resultado do cache
func (c *PersistentCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result string
	var entry CacheEntry

	err := c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		data := bucket.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("entrada não encontrada")
		}

		return json.Unmarshal(data, &entry)
	})

	if err != nil {
		c.cacheMisses++
		return "", false
	}

	// Verificar se expirou
	if time.Since(entry.StoredAt) > entry.TTL {
		// Entrada expirada, remover
		go c.Delete(key)
		c.cacheMisses++
		return "", false
	}

	// Atualizar contador de acessos
	err = c.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		data := bucket.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("entrada não encontrada")
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			return err
		}

		entry.AccessCount++
		data, err = json.Marshal(entry)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), data)
	})

	if err != nil {
		logger.CacheError(fmt.Sprintf("[CACHE DB] Erro ao atualizar contador: %v", err))
	}

	c.cacheHits++
	logger.CacheDebug(fmt.Sprintf("[CACHE DB] Hit: %s (acessos: %d)", key, entry.AccessCount+1))

	return result, true
}

// Set armazena resultado no cache
func (c *PersistentCache) Set(key string, result string, tags []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := CacheEntry{
		Key:         key,
		Result:      result,
		StoredAt:    time.Now(),
		AccessCount: 1,
		TTL:         c.cacheTTL,
		Tags:        tags,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("erro ao serializar entrada: %w", err)
	}

	err = c.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		return bucket.Put([]byte(key), data)
	})

	if err != nil {
		c.cacheMisses++
		return fmt.Errorf("erro ao inserir no cache: %w", err)
	}

	c.cacheMisses++
	logger.CacheDebug(fmt.Sprintf("[CACHE DB] Set: %s (TTL: %v, tags: %v)", key, c.cacheTTL, tags))
	return nil
}

// Delete remove entrada do cache
func (c *PersistentCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		return bucket.Delete([]byte(key))
	})

	if err != nil {
		return fmt.Errorf("erro ao deletar do cache: %w", err)
	}

	logger.CacheDebug(fmt.Sprintf("[CACHE DB] Delete: %s", key))
	return nil
}

// Invalidate invalida cache baseado em tags
func (c *PersistentCache) Invalidate(tags []string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(tags) == 0 {
		return 0, nil
	}

	var count int64
	var keysToDelete []string

	err := c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var entry CacheEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}

			// Verificar se alguma das tags corresponde
			for _, tag := range tags {
				for _, entryTag := range entry.Tags {
					if entryTag == tag {
						keysToDelete = append(keysToDelete, string(k))
						break
					}
				}
			}
			return nil
		})
	})

	if err != nil {
		return 0, fmt.Errorf("erro ao buscar entradas para invalidação: %w", err)
	}

	// Deletar chaves encontradas
	err = c.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		for _, key := range keysToDelete {
			if err := bucket.Delete([]byte(key)); err != nil {
				return err
			}
			count++
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("erro ao invalidar cache: %w", err)
	}

	c.cacheInvalidations += count

	if count > 0 {
		logger.CacheInfo(fmt.Sprintf("[CACHE DB] Invalidação: %d entradas removidas (tags: %v)", count, tags))
	}

	return count, nil
}

// Clear limpa todo o cache
func (c *PersistentCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var count int
	err := c.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		err := bucket.ForEach(func(k, v []byte) error {
			if err := bucket.Delete(k); err != nil {
				return err
			}
			count++
			return nil
		})
		return err
	})

	if err != nil {
		return fmt.Errorf("erro ao limpar cache: %w", err)
	}

	logger.CacheInfo(fmt.Sprintf("[CACHE DB] Limpo: %d entradas removidas", count))
	return nil
}

// Cleanup remove entradas expiradas
func (c *PersistentCache) Cleanup() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var count int
	var keysToDelete []string

	err := c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var entry CacheEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}

			// Verificar se expirou
			if time.Since(entry.StoredAt) > entry.TTL {
				keysToDelete = append(keysToDelete, string(k))
			}
			return nil
		})
	})

	if err != nil {
		return 0, fmt.Errorf("erro ao buscar entradas expiradas: %w", err)
	}

	// Deletar chaves expiradas
	err = c.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		for _, key := range keysToDelete {
			if err := bucket.Delete([]byte(key)); err != nil {
				return err
			}
			count++
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("erro ao limpar entradas expiradas: %w", err)
	}

	if count > 0 {
		logger.CacheInfo(fmt.Sprintf("[CACHE DB] Limpeza: %d entradas expiradas removidas", count))
	}

	return count, nil
}

// GetStatus retorna status do cache
func (c *PersistentCache) GetStatus() CacheStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalEntries int

	err := c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return nil
		}

		stats := bucket.Stats()
		totalEntries = stats.KeyN
		return nil
	})

	if err != nil {
		logger.CacheError(fmt.Sprintf("[CACHE DB] Erro ao contar entradas: %v", err))
	}

	// Obter tamanho do banco
	dbSize := c.getDatabaseSize()

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

	err := c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		data := bucket.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("entrada não encontrada")
		}

		return json.Unmarshal(data, &entry)
	})

	if err != nil {
		if err.Error() == "entrada não encontrada" {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao obter entrada: %w", err)
	}

	return &entry, nil
}

// GetAllEntries retorna todas as entradas do cache
func (c *PersistentCache) GetAllEntries() ([]*CacheEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var entries []*CacheEntry

	err := c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var entry CacheEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}

			entries = append(entries, &entry)
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao obter entradas: %w", err)
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
func (c *PersistentCache) getDatabaseSize() int64 {
	// BoltDB não tem método Stat() para obter tamanho
	// Retornar 0 por enquanto
	return 0
}

// Vacuum compacta o banco de dados (recria o bucket sem dados expirados)
func (c *PersistentCache) Vacuum() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var count int
	var validEntries []struct {
		key  string
		data []byte
	}

	// Primeiro, coletar entradas válidas
	err := c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cacheBucketName))
		if bucket == nil {
			return fmt.Errorf("bucket não existe")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var entry CacheEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}

			// Manter apenas entradas não expiradas
			if time.Since(entry.StoredAt) <= entry.TTL {
				validEntries = append(validEntries, struct {
					key  string
					data []byte
				}{string(k), v})
				count++
			}
			return nil
		})
	})

	if err != nil {
		return fmt.Errorf("erro ao ler entradas: %w", err)
	}

	// Deletar bucket antigo e criar novo
	err = c.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(cacheBucketName))
		if err != nil {
			return err
		}

		bucket, err := tx.CreateBucket([]byte(cacheBucketName))
		if err != nil {
			return err
		}

		// Inserir entradas válidas
		for _, entry := range validEntries {
			if err := bucket.Put([]byte(entry.key), entry.data); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("erro ao compactar banco: %w", err)
	}

	logger.CacheInfo(fmt.Sprintf("[CACHE DB] Banco compactado (mantidas %d entradas)", count))
	return nil
}
