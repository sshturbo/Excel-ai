package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ValidationError representa um erro de validação
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validator struct principal
type Validator struct {
	errors []ValidationError
}

// NewValidator cria um novo validador
func NewValidator() *Validator {
	return &Validator{
		errors: make([]ValidationError, 0),
	}
}

// AddError adiciona um erro de validação
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors retorna se há erros de validação
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors retorna todos os erros de validação
func (v *Validator) Errors() []ValidationError {
	return v.errors
}

// Error retorna o primeiro erro ou nil
func (v *Validator) Error() error {
	if !v.HasErrors() {
		return nil
	}
	return v.errors[0]
}

// Clear limpa todos os erros
func (v *Validator) Clear() {
	v.errors = make([]ValidationError, 0)
}

// ValidateString valida uma string genérica
func (v *Validator) ValidateString(field, value string, required bool, minLength, maxLength int) {
	if required && strings.TrimSpace(value) == "" {
		v.AddError(field, "é obrigatório")
		return
	}

	if len(value) > maxLength {
		v.AddError(field, fmt.Sprintf("deve ter no máximo %d caracteres", maxLength))
	}

	if minLength > 0 && len(value) < minLength {
		v.AddError(field, fmt.Sprintf("deve ter no mínimo %d caracteres", minLength))
	}
}

// ValidateAPIKey valida uma chave de API
func (v *Validator) ValidateAPIKey(field, value string) {
	v.ValidateString(field, value, true, 10, 500)

	// Verificar se contém apenas caracteres válidos para API keys
	if value != "" {
		// API keys geralmente contém alfanuméricos e alguns caracteres especiais
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-\._]+$`, value)
		if !matched {
			v.AddError(field, "contém caracteres inválidos")
		}
	}
}

// ValidateExcelRange valida um range de Excel (ex: A1:B10)
func (v *Validator) ValidateExcelRange(field, value string) {
	if value == "" {
		return // Range opcional
	}

	// Padrão para range de Excel: A1, A1:B10, Sheet1!A1:B10
	pattern := `^([a-zA-Z0-9_]+!)?[A-Z]+[0-9]+(:[A-Z]+[0-9]+)?$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.AddError(field, "formato de range inválido (ex: A1, A1:B10, Sheet1!A1:B10)")
	}
}

// ValidateSheetName valida um nome de planilha
func (v *Validator) ValidateSheetName(field, value string) {
	v.ValidateString(field, value, true, 1, 31)

	if value != "" {
		// Nomes de planilhas não podem conter: \ / ? * [ ] :
		invalidChars := `/\?*[]:`
		for _, char := range invalidChars {
			if strings.Contains(value, string(char)) {
				v.AddError(field, fmt.Sprintf("não pode conter o caractere '%c'", char))
				return
			}
		}

		// Não pode começar ou terminar com apóstrofo
		if strings.HasPrefix(value, "'") || strings.HasSuffix(value, "'") {
			v.AddError(field, "não pode começar ou terminar com apóstrofo")
		}
	}
}

// ValidateCellValue valida o valor de uma célula
func (v *Validator) ValidateCellValue(field, value string) {
	// Limitar tamanho de valor de célula (Excel limita a 32,767 caracteres)
	v.ValidateString(field, value, false, 0, 32767)
}

// ValidateEmail valida um endereço de e-mail
func (v *Validator) ValidateEmail(field, value string) {
	v.ValidateString(field, value, false, 0, 255)

	if value != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(value) {
			v.AddError(field, "formato de e-mail inválido")
		}
	}
}

// ValidateURL valida uma URL
func (v *Validator) ValidateURL(field, value string) {
	v.ValidateString(field, value, false, 0, 2048)

	if value != "" {
		urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9.\-]+(/.*)?$`)
		if !urlRegex.MatchString(value) {
			v.AddError(field, "URL deve começar com http:// ou https://")
		}
	}
}

// ValidateNumber valida se uma string representa um número
func (v *Validator) ValidateNumber(field, value string) {
	if value == "" {
		return
	}

	_, err := fmt.Sscanf(value, "%f", new(float64))
	if err != nil {
		v.AddError(field, "deve ser um número válido")
	}
}

// SanitizeInput sanitiza uma string para prevenir injeção
func SanitizeInput(input string) string {
	// Remover caracteres de controle perigosos
	var result strings.Builder
	for _, r := range input {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	
	return strings.TrimSpace(result.String())
}

// SanitizeExcelCommand sanitiza comandos para o Excel
func SanitizeExcelCommand(command string) string {
	// Remover caracteres que podem ser usados em injeção de fórmulas
	dangerousPatterns := []string{
		"=HYPERLINK(", "=EXEC(", "=CALL(", "=REGISTER(",
		"=", "+", "-", "@", "'", // Prevenir injeção de fórmulas no início
	}

	sanitized := command
	for _, pattern := range dangerousPatterns {
		sanitized = strings.ReplaceAll(sanitized, pattern, "")
	}

	return SanitizeInput(sanitized)
}

// ValidateModelName valida o nome de um modelo de IA
func (v *Validator) ValidateModelName(field, value string) {
	v.ValidateString(field, value, false, 0, 100)

	if value != "" {
		// Modelos geralmente seguem o formato: provider/model-name
		pattern := `^[a-zA-Z0-9\-_/]+$`
		matched, _ := regexp.MatchString(pattern, value)
		if !matched {
			v.AddError(field, "formato de nome de modelo inválido")
		}
	}
}

// ValidateConversationID valida um ID de conversação
func (v *Validator) ValidateConversationID(field, value string) {
	v.ValidateString(field, value, false, 0, 100)

	if value != "" {
		// IDs devem ser alfanuméricos e hífens
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-]+$`, value)
		if !matched {
			v.AddError(field, "formato de ID inválido")
		}
	}
}

// ValidateCellReference valida uma referência de célula (ex: A1, B10, Z99)
func (v *Validator) ValidateCellReference(field, value string) {
	if value == "" {
		return
	}

	pattern := `^[A-Z]+[0-9]+$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.AddError(field, "formato de célula inválido (ex: A1, B10)")
	}
}

// ValidateFileName valida um nome de arquivo
func (v *Validator) ValidateFileName(field, value string) {
	v.ValidateString(field, value, true, 1, 255)

	if value != "" {
		// Caracteres inválidos em nomes de arquivo Windows
		invalidChars := `<>"\/|?*`
		for _, char := range invalidChars {
			if strings.Contains(value, string(char)) {
				v.AddError(field, fmt.Sprintf("não pode conter o caractere '%c'", char))
				return
			}
		}

		// Nomes reservados no Windows
		reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
		upperValue := strings.ToUpper(value)
		for _, reserved := range reservedNames {
			if upperValue == reserved {
				v.AddError(field, fmt.Sprintf("'%s' é um nome reservado", value))
				return
			}
		}
	}
}

// ValidateInteger valida se uma string representa um inteiro
func (v *Validator) ValidateInteger(field, value string, min, max int) {
	if value == "" {
		return
	}

	var num int
	_, err := fmt.Sscanf(value, "%d", &num)
	if err != nil {
		v.AddError(field, "deve ser um número inteiro válido")
		return
	}

	if min != 0 && num < min {
		v.AddError(field, fmt.Sprintf("deve ser maior ou igual a %d", min))
	}

	if max != 0 && num > max {
		v.AddError(field, fmt.Sprintf("deve ser menor ou igual a %d", max))
	}
}

// ValidateEnum valida se um valor está em uma lista de valores permitidos
func (v *Validator) ValidateEnum(field, value string, allowedValues []string, caseSensitive bool) {
	if value == "" {
		return
	}

	for _, allowed := range allowedValues {
		if caseSensitive {
			if value == allowed {
				return
			}
		} else {
			if strings.EqualFold(value, allowed) {
				return
			}
		}
	}

	v.AddError(field, fmt.Sprintf("deve ser um dos seguintes valores: %s", strings.Join(allowedValues, ", ")))
}

// ValidateMaxLength valida o tamanho máximo de uma string
func (v *Validator) ValidateMaxLength(field, value string, maxLength int) {
	if len(value) > maxLength {
		v.AddError(field, fmt.Sprintf("deve ter no máximo %d caracteres", maxLength))
	}
}

// ValidateMinLength valida o tamanho mínimo de uma string
func (v *Validator) ValidateMinLength(field, value string, minLength int) {
	if len(value) < minLength {
		v.AddError(field, fmt.Sprintf("deve ter no mínimo %d caracteres", minLength))
	}
}
