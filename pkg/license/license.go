package license

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// LicenseInfo armazena informações da licença
type LicenseInfo struct {
	Hash          string    `json:"hash"`
	MachineID     string    `json:"machine_id"`
	ActivatedAt   time.Time `json:"activated_at"`
	LastValidated time.Time `json:"last_validated"`
}

// Validator gerencia a validação da licença
type Validator struct {
	githubURL string
	storage   LicenseStorage
}

// LicenseStorage interface para persistência
type LicenseStorage interface {
	SaveLicense(license *LicenseInfo) error
	LoadLicense() (*LicenseInfo, error)
}

// NewValidator cria um novo validador
func NewValidator(githubURL string, storage LicenseStorage) *Validator {
	return &Validator{
		githubURL: githubURL,
		storage:   storage,
	}
}

// GetMachineID obtém um identificador único da máquina Windows
func GetMachineID() (string, error) {
	// Usar WMIC para obter UUID da placa-mãe
	cmd := exec.Command("wmic", "csproduct", "get", "UUID")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: usar nome do computador + usuário
		cmd = exec.Command("hostname")
		hostname, _ := cmd.Output()
		return hashString(strings.TrimSpace(string(hostname))), nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "UUID" {
			return hashString(line), nil
		}
	}

	return "", fmt.Errorf("não foi possível obter ID da máquina")
}

// hashString cria hash SHA256 de uma string
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// FetchValidHashes busca lista de hashes válidos do GitHub
func (v *Validator) FetchValidHashes() ([]string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(v.githubURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com servidor de licenças: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("servidor de licenças indisponível: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parsear como JSON array ou linhas simples
	var hashes []string

	// Tentar JSON primeiro
	if err := json.Unmarshal(body, &hashes); err != nil {
		// Fallback: uma hash por linha
		lines := strings.Split(string(body), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				hashes = append(hashes, line)
			}
		}
	}

	return hashes, nil
}

// ActivateLicense ativa uma licença escolhendo um hash disponível
func (v *Validator) ActivateLicense() (*LicenseInfo, error) {
	// Buscar hashes válidos
	validHashes, err := v.FetchValidHashes()
	if err != nil {
		return nil, err
	}

	if len(validHashes) == 0 {
		return nil, fmt.Errorf("nenhuma licença disponível")
	}

	// Obter ID da máquina
	machineID, err := GetMachineID()
	if err != nil {
		return nil, err
	}

	// Escolher primeiro hash disponível (pode ser randomizado)
	selectedHash := validHashes[0]

	// Criar licença
	license := &LicenseInfo{
		Hash:          selectedHash,
		MachineID:     machineID,
		ActivatedAt:   time.Now(),
		LastValidated: time.Now(),
	}

	// Salvar
	if err := v.storage.SaveLicense(license); err != nil {
		return nil, err
	}

	return license, nil
}

// ValidateLicense verifica se a licença ainda é válida
func (v *Validator) ValidateLicense() (bool, string, error) {
	// Carregar licença salva
	license, err := v.storage.LoadLicense()
	if err != nil || license == nil {
		// Sem licença - tentar ativar automaticamente
		license, err = v.ActivateLicense()
		if err != nil {
			return false, "Falha ao ativar licença: " + err.Error(), err
		}
	}

	// Verificar se máquina é a mesma
	currentMachineID, err := GetMachineID()
	if err != nil {
		return false, "Erro ao verificar máquina", err
	}

	if license.MachineID != currentMachineID {
		return false, "Licença não válida para esta máquina", nil
	}

	// Buscar lista atualizada de hashes válidos
	validHashes, err := v.FetchValidHashes()
	if err != nil {
		// Se não conseguir conectar, usar cache (permitir offline por X dias)
		daysSinceValidation := time.Since(license.LastValidated).Hours() / 24
		if daysSinceValidation > 7 {
			return false, "Não foi possível validar licença online (offline por mais de 7 dias)", nil
		}
		// Permitir uso offline temporário
		return true, "Modo offline (última validação: " + license.LastValidated.Format("02/01/2006") + ")", nil
	}

	// Verificar se hash ainda está na lista
	hashValid := false
	for _, h := range validHashes {
		if h == license.Hash {
			hashValid = true
			break
		}
	}

	if !hashValid {
		return false, "Licença revogada ou expirada", nil
	}

	// Atualizar última validação
	license.LastValidated = time.Now()
	v.storage.SaveLicense(license)

	return true, "Licença válida", nil
}
