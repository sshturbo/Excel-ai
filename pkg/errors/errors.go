package errors

import (
	"fmt"
)

// ErrorCode representa códigos de erro específicos
type ErrorCode string

const (
	// Erros gerais
	ErrCodeUnknown           ErrorCode = "UNKNOWN"
	ErrCodeInternal          ErrorCode = "INTERNAL"
	ErrCodeInvalidInput      ErrorCode = "INVALID_INPUT"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeRateLimit        ErrorCode = "RATE_LIMIT"
	ErrCodeTimeout          ErrorCode = "TIMEOUT"

	// Erros do Excel
	ErrCodeExcelNotConnected ErrorCode = "EXCEL_NOT_CONNECTED"
	ErrCodeExcelBusy        ErrorCode = "EXCEL_BUSY"
	ErrCodeExcelNotFound    ErrorCode = "EXCEL_NOT_FOUND"
	ErrCodeInvalidRange     ErrorCode = "INVALID_RANGE"
	ErrCodeInvalidSheet     ErrorCode = "INVALID_SHEET"

	// Erros de IA
	ErrCodeAIAPIKeyMissing ErrorCode = "AI_API_KEY_MISSING"
	ErrCodeAIAPIKeyInvalid ErrorCode = "AI_API_KEY_INVALID"
	ErrCodeAIQuotaExceeded ErrorCode = "AI_QUOTA_EXCEEDED"
	ErrCodeAIModelInvalid  ErrorCode = "AI_MODEL_INVALID"
	ErrCodeAIStreamError   ErrorCode = "AI_STREAM_ERROR"

	// Erros de Storage
	ErrCodeStorageError     ErrorCode = "STORAGE_ERROR"
	ErrCodeDatabaseError   ErrorCode = "DATABASE_ERROR"

	// Erros de Licença
	ErrCodeLicenseInvalid   ErrorCode = "LICENSE_INVALID"
	ErrCodeLicenseExpired  ErrorCode = "LICENSE_EXPIRED"
)

// AppError representa um erro estruturado da aplicação
type AppError struct {
	Code       ErrorCode
	Message    string
	Cause      error
	Component  string
	StatusCode int // Para respostas HTTP se necessário
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is verifica se um erro é de um tipo específico
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New cria um novo erro da aplicação
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Component: "APP",
	}
}

// NewWithCause cria um erro com causa
func NewWithCause(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Component: "APP",
	}
}

// NewWithComponent cria um erro com componente
func NewWithComponent(code ErrorCode, component, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Component: component,
	}
}

// Wrap envolve um erro existente com contexto adicional
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	
	// Se já for AppError, apenas adicione contexto
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Code:      code,
			Message:   message,
			Cause:     appErr,
			Component: appErr.Component,
		}
	}
	
	return &AppError{
		Code:      code,
		Message:   message,
		Cause:     err,
		Component: "APP",
	}
}

// Erros específicos do Excel
func ExcelNotConnected(msg string) *AppError {
	return NewWithComponent(ErrCodeExcelNotConnected, "EXCEL", msg)
}

func ExcelBusy(msg string) *AppError {
	return NewWithComponent(ErrCodeExcelBusy, "EXCEL", msg)
}

func ExcelNotFound(msg string) *AppError {
	return NewWithComponent(ErrCodeExcelNotFound, "EXCEL", msg)
}

func InvalidRange(msg string) *AppError {
	return NewWithComponent(ErrCodeInvalidRange, "EXCEL", msg)
}

func InvalidSheet(msg string) *AppError {
	return NewWithComponent(ErrCodeInvalidSheet, "EXCEL", msg)
}

// Erros específicos de IA
func AIAPIKeyMissing(msg string) *AppError {
	return NewWithComponent(ErrCodeAIAPIKeyMissing, "AI", msg)
}

func AIAPIKeyInvalid(msg string) *AppError {
	return NewWithComponent(ErrCodeAIAPIKeyInvalid, "AI", msg)
}

func AIQuotaExceeded(msg string) *AppError {
	return NewWithComponent(ErrCodeAIQuotaExceeded, "AI", msg)
}

func AIModelInvalid(msg string) *AppError {
	return NewWithComponent(ErrCodeAIModelInvalid, "AI", msg)
}

func AIStreamError(msg string) *AppError {
	return NewWithComponent(ErrCodeAIStreamError, "AI", msg)
}

// Erros específicos de Storage
func StorageError(msg string) *AppError {
	return NewWithComponent(ErrCodeStorageError, "STORAGE", msg)
}

func DatabaseError(msg string, cause error) *AppError {
	return NewWithCause(ErrCodeDatabaseError, msg, cause)
}

// Erros específicos de Licença
func LicenseInvalid(msg string) *AppError {
	return NewWithComponent(ErrCodeLicenseInvalid, "LICENSE", msg)
}

func LicenseExpired(msg string) *AppError {
	return NewWithComponent(ErrCodeLicenseExpired, "LICENSE", msg)
}

// Erros gerais
func InvalidInput(msg string) *AppError {
	return New(ErrCodeInvalidInput, msg)
}

func NotFound(msg string) *AppError {
	return New(ErrCodeNotFound, msg)
}

func Unauthorized(msg string) *AppError {
	return New(ErrCodeUnauthorized, msg)
}

func RateLimit(msg string) *AppError {
	return New(ErrCodeRateLimit, msg)
}

func Timeout(msg string) *AppError {
	return New(ErrCodeTimeout, msg)
}

func Internal(msg string) *AppError {
	return New(ErrCodeInternal, msg)
}

// IsAppError verifica se um erro é AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetCode retorna o código de erro se for AppError
func GetCode(err error) ErrorCode {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ErrCodeUnknown
}

// GetMessage retorna a mensagem de erro se for AppError
func GetMessage(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Message
	}
	return err.Error()
}

// GetUserFriendlyMessage retorna uma mensagem amigável para o usuário
func GetUserFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}

	if appErr, ok := err.(*AppError); ok {
		switch appErr.Code {
		case ErrCodeExcelNotConnected:
			return "Não foi possível conectar ao Excel. Verifique se o Excel está aberto."
		case ErrCodeExcelBusy:
			return "O Excel está ocupado. Tente novamente em instantes."
		case ErrCodeAIAPIKeyMissing:
			return "API key não configurada. Por favor, configure sua API key nas configurações."
		case ErrCodeAIAPIKeyInvalid:
			return "API key inválida. Verifique sua chave nas configurações."
		case ErrCodeAIQuotaExceeded:
			return "Limite de quota da API excedido. Tente novamente mais tarde."
		case ErrCodeInvalidInput:
			return "Entrada inválida. Verifique os dados informados."
		case ErrCodeRateLimit:
			return "Muitas requisições. Aguarde um momento antes de tentar novamente."
		case ErrCodeTimeout:
			return "Operação demorou muito tempo e expirou. Tente novamente."
		case ErrCodeLicenseInvalid:
			return "Licença inválida. Entre em contato com o suporte."
		case ErrCodeLicenseExpired:
			return "Sua licença expirou. Entre em contato com o suporte para renovar."
		default:
			return appErr.Message
		}
	}

	return err.Error()
}
