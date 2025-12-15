package app

import (
	"excel-ai/pkg/license"
)

// URL do arquivo com lista de hashes no GitHub (raw content)
const LicenseHashesURL = "https://raw.githubusercontent.com/sshturbo/licenshiposistem/refs/heads/main/key"

// CheckLicense verifica se a licença é válida
func (a *App) CheckLicense() (bool, string) {
	if a.storage == nil {
		return false, "Storage não disponível"
	}

	return license.QuickValidate(a.storage, LicenseHashesURL)
}

// GetLicenseStatus retorna status detalhado da licença
func (a *App) GetLicenseStatus() license.LicenseStatus {
	if a.storage == nil {
		return license.LicenseStatus{
			Valid:   false,
			Message: "Storage não disponível",
		}
	}

	return license.GetStatus(a.storage, LicenseHashesURL)
}

// ActivateLicense ativa uma nova licença
func (a *App) ActivateLicense() (bool, string) {
	if a.storage == nil {
		return false, "Storage não disponível"
	}

	adapter := license.NewStorageAdapter(a.storage)
	validator := license.NewValidator(LicenseHashesURL, adapter)

	_, err := validator.ActivateLicense()
	if err != nil {
		return false, "Falha ao ativar: " + err.Error()
	}

	// Atualizar status interno
	a.licenseValid = true
	a.licenseMessage = "Licença ativada com sucesso!"

	return true, a.licenseMessage
}

// IsLicenseValid retorna se a licença está válida (para frontend)
func (a *App) IsLicenseValid() bool {
	return a.licenseValid
}

// GetLicenseMessage retorna a mensagem da licença
func (a *App) GetLicenseMessage() string {
	return a.licenseMessage
}
