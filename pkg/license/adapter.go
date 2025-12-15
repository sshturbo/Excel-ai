package license

import (
	"time"

	"excel-ai/pkg/storage"
)

// StorageAdapter adapta o storage para a interface LicenseStorage
type StorageAdapter struct {
	storage *storage.Storage
}

// NewStorageAdapter cria um novo adapter
func NewStorageAdapter(s *storage.Storage) *StorageAdapter {
	return &StorageAdapter{storage: s}
}

// SaveLicense implementa LicenseStorage
func (a *StorageAdapter) SaveLicense(license *LicenseInfo) error {
	// Converter para o tipo do storage
	storageLicense := &storage.LicenseInfo{
		Hash:          license.Hash,
		MachineID:     license.MachineID,
		ActivatedAt:   license.ActivatedAt,
		LastValidated: license.LastValidated,
	}
	return a.storage.SaveLicense(storageLicense)
}

// LoadLicense implementa LicenseStorage
func (a *StorageAdapter) LoadLicense() (*LicenseInfo, error) {
	storageLicense, err := a.storage.LoadLicense()
	if err != nil || storageLicense == nil {
		return nil, err
	}

	return &LicenseInfo{
		Hash:          storageLicense.Hash,
		MachineID:     storageLicense.MachineID,
		ActivatedAt:   storageLicense.ActivatedAt,
		LastValidated: storageLicense.LastValidated,
	}, nil
}

// QuickValidate faz validação rápida sem criar todo o validator
func QuickValidate(s *storage.Storage, githubURL string) (bool, string) {
	adapter := NewStorageAdapter(s)
	validator := NewValidator(githubURL, adapter)

	valid, msg, err := validator.ValidateLicense()
	if err != nil {
		return false, msg
	}
	return valid, msg
}

// GetLicenseStatus retorna status da licença para UI
type LicenseStatus struct {
	Valid       bool      `json:"valid"`
	Message     string    `json:"message"`
	Hash        string    `json:"hash,omitempty"`
	ActivatedAt time.Time `json:"activatedAt,omitempty"`
	MachineID   string    `json:"machineId,omitempty"`
}

// GetStatus retorna o status completo da licença
func GetStatus(s *storage.Storage, githubURL string) LicenseStatus {
	adapter := NewStorageAdapter(s)
	validator := NewValidator(githubURL, adapter)

	valid, msg, _ := validator.ValidateLicense()

	status := LicenseStatus{
		Valid:   valid,
		Message: msg,
	}

	// Carregar dados da licença para mostrar
	if license, err := adapter.LoadLicense(); err == nil && license != nil {
		status.Hash = license.Hash[:8] + "..." // Mostrar só início
		status.ActivatedAt = license.ActivatedAt
		status.MachineID = license.MachineID[:8] + "..."
	}

	return status
}
