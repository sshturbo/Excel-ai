package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Logger struct {
	level  LogLevel
	mu     sync.Mutex
	file   *os.File
	useFile bool
}

type Config struct {
	Level      string           `json:"level"`
	Output     string           `json:"output"`
	FilePath   string           `json:"file_path"`
	Components map[string]string `json:"components"`
}

var (
	instance      *Logger
	once          sync.Once
	componentLevels map[string]LogLevel
)

func init() {
	componentLevels = make(map[string]LogLevel)
	for _, comp := range []string{
		ComponentApp, ComponentChat, ComponentExcel, ComponentAI,
		ComponentCache, ComponentStorage, ComponentLicense,
		ComponentHTTP, ComponentStream, ComponentTools, ComponentUndo,
	} {
		componentLevels[comp] = INFO
	}
}

func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			level:  INFO,
			useFile: false,
		}
	})
	return instance
}

func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func SetComponentLevel(component string, level LogLevel) {
	componentLevels[component] = level
}

func getComponentLevel(component string) LogLevel {
	if level, exists := componentLevels[component]; exists {
		return level
	}
	return INFO
}

func (l *Logger) SetFileOutput(filepath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.file = file
	l.useFile = true
	return nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) log(level LogLevel, component, message string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	componentLevel := getComponentLevel(component)
	if level < componentLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := l.levelToString(level)

	var fieldsStr string
	if len(fields) > 0 {
		fieldsStr = " |"
		for k, v := range fields {
			fieldsStr += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	logLine := fmt.Sprintf("[%s] %s [%s] %s%s\n", timestamp, levelStr, component, message, fieldsStr)
	fmt.Print(logLine)

	if l.useFile && l.file != nil {
		l.file.WriteString(logLine)
	}

	if level == FATAL {
		if l.useFile && l.file != nil {
			l.file.Close()
		}
		os.Exit(1)
	}
}

func (l *Logger) levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO "
	case WARN:
		return "WARN "
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKN "
	}
}

func (l *Logger) Debug(component, message string) {
	l.log(DEBUG, component, message, nil)
}

func (l *Logger) Debugf(component, format string, args ...interface{}) {
	l.log(DEBUG, component, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Info(component, message string) {
	l.log(INFO, component, message, nil)
}

func (l *Logger) Infof(component, format string, args ...interface{}) {
	l.log(INFO, component, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Warn(component, message string) {
	l.log(WARN, component, message, nil)
}

func (l *Logger) Warnf(component, format string, args ...interface{}) {
	l.log(WARN, component, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Error(component, message string) {
	l.log(ERROR, component, message, nil)
}

func (l *Logger) Errorf(component, format string, args ...interface{}) {
	l.log(ERROR, component, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Fatal(component, message string) {
	l.log(FATAL, component, message, nil)
}

func (l *Logger) Fatalf(component, format string, args ...interface{}) {
	l.log(FATAL, component, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) WithFields(level LogLevel, component, message string, fields map[string]interface{}) {
	l.log(level, component, message, fields)
}

const (
	ComponentApp     = "APP"
	ComponentExcel   = "EXCEL"
	ComponentChat    = "CHAT"
	ComponentAI      = "AI"
	ComponentCache   = "CACHE"
	ComponentStorage = "STORAGE"
	ComponentLicense = "LICENSE"
	ComponentHTTP    = "HTTP"
	ComponentStream  = "STREAM"
	ComponentTools   = "TOOLS"
	ComponentUndo    = "UNDO"
)

func AppDebug(message string)   { GetLogger().Debug(ComponentApp, message) }
func AppInfo(message string)    { GetLogger().Info(ComponentApp, message) }
func AppWarn(message string)    { GetLogger().Warn(ComponentApp, message) }
func AppError(message string)   { GetLogger().Error(ComponentApp, message) }
func AppFatal(message string)   { GetLogger().Fatal(ComponentApp, message) }

func ExcelDebug(message string) { GetLogger().Debug(ComponentExcel, message) }
func ExcelInfo(message string)  { GetLogger().Info(ComponentExcel, message) }
func ExcelWarn(message string)  { GetLogger().Warn(ComponentExcel, message) }
func ExcelError(message string) { GetLogger().Error(ComponentExcel, message) }

func ChatDebug(message string) { GetLogger().Debug(ComponentChat, message) }
func ChatInfo(message string)  { GetLogger().Info(ComponentChat, message) }
func ChatWarn(message string)  { GetLogger().Warn(ComponentChat, message) }
func ChatError(message string) { GetLogger().Error(ComponentChat, message) }

func AIDebug(message string) { GetLogger().Debug(ComponentAI, message) }
func AIInfo(message string)  { GetLogger().Info(ComponentAI, message) }
func AIWarn(message string)  { GetLogger().Warn(ComponentAI, message) }
func AIError(message string) { GetLogger().Error(ComponentAI, message) }

func StorageDebug(message string) { GetLogger().Debug(ComponentStorage, message) }
func StorageInfo(message string)  { GetLogger().Info(ComponentStorage, message) }
func StorageWarn(message string)  { GetLogger().Warn(ComponentStorage, message) }
func StorageError(message string) { GetLogger().Error(ComponentStorage, message) }

func CacheDebug(message string) { GetLogger().Debug(ComponentCache, message) }
func CacheInfo(message string)  { GetLogger().Info(ComponentCache, message) }
func CacheWarn(message string)  { GetLogger().Warn(ComponentCache, message) }
func CacheError(message string) { GetLogger().Error(ComponentCache, message) }

func LicenseDebug(message string) { GetLogger().Debug(ComponentLicense, message) }
func LicenseInfo(message string)  { GetLogger().Info(ComponentLicense, message) }
func LicenseWarn(message string)  { GetLogger().Warn(ComponentLicense, message) }
func LicenseError(message string) { GetLogger().Error(ComponentLicense, message) }

func HTTPDebug(message string) { GetLogger().Debug(ComponentHTTP, message) }
func HTTPInfo(message string)  { GetLogger().Info(ComponentHTTP, message) }
func HTTPWarn(message string)  { GetLogger().Warn(ComponentHTTP, message) }
func HTTPError(message string) { GetLogger().Error(ComponentHTTP, message) }

func StreamDebug(message string) { GetLogger().Debug(ComponentStream, message) }
func StreamInfo(message string)  { GetLogger().Info(ComponentStream, message) }
func StreamWarn(message string)  { GetLogger().Warn(ComponentStream, message) }
func StreamError(message string) { GetLogger().Error(ComponentStream, message) }

func ToolsDebug(message string) { GetLogger().Debug(ComponentTools, message) }
func ToolsInfo(message string)  { GetLogger().Info(ComponentTools, message) }
func ToolsWarn(message string)  { GetLogger().Warn(ComponentTools, message) }
func ToolsError(message string) { GetLogger().Error(ComponentTools, message) }

func UndoDebug(message string) { GetLogger().Debug(ComponentUndo, message) }
func UndoInfo(message string)  { GetLogger().Info(ComponentUndo, message) }
func UndoWarn(message string)  { GetLogger().Warn(ComponentUndo, message) }
func UndoError(message string) { GetLogger().Error(ComponentUndo, message) }

func LoadConfig(configPath string) error {
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err == nil {
			configPath = absPath
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de configuração %s: %w", configPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("erro ao parsear arquivo de configuração: %w", err)
	}

	logger := GetLogger()
	logger.mu.Lock()

	if globalLevel, ok := parseLogLevel(config.Level); ok {
		logger.level = globalLevel
	}

	if config.Output == "file" && config.FilePath != "" {
		logger.mu.Unlock()
		if err := logger.SetFileOutput(config.FilePath); err != nil {
			fmt.Printf("[LOGGER] Aviso: não foi possível configurar arquivo de log: %v\n", err)
		}
		logger.mu.Lock()
	}

	for component, levelStr := range config.Components {
		if level, ok := parseLogLevel(levelStr); ok {
			componentLevels[component] = level
		}
	}

	logger.mu.Unlock()

	logger.Info("LOGGER", fmt.Sprintf("Configuração carregada de %s (nível: %s)", configPath, config.Level))
	return nil
}

func parseLogLevel(levelStr string) (LogLevel, bool) {
	switch levelStr {
	case "DEBUG":
		return DEBUG, true
	case "INFO":
		return INFO, true
	case "WARN":
		return WARN, true
	case "ERROR":
		return ERROR, true
	case "FATAL":
		return FATAL, true
	default:
		return INFO, false
	}
}

func InitializeFromFile(configPath string) error {
	GetLogger()
	return LoadConfig(configPath)
}

func InitializeWithDefaults(level LogLevel) {
	logger := GetLogger()
	logger.SetLevel(level)
	logger.Info("LOGGER", fmt.Sprintf("Logger inicializado com nível padrão: %s", LevelToString(level)))
}

func LevelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKN"
	}
}
