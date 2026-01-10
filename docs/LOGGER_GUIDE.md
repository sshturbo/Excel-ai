# Guia do Sistema de Logging

## Visão Geral

O Excel-AI utiliza um sistema de logging estruturado e centralizado para monitorar e diagnosticar problemas em tempo real. O logger suporta múltiplos níveis de severidade, saída para console e arquivo, e configuração por componente.

## Características

- ✅ **11 Componentes**: APP, CHAT, EXCEL, AI, CACHE, STORAGE, LICENSE, HTTP, STREAM, TOOLS, UNDO
- ✅ **5 Níveis de Log**: DEBUG, INFO, WARN, ERROR, FATAL
- ✅ **Configuração via JSON**: Facilidade de ajustar níveis por componente
- ✅ **Saída Dupla**: Console e arquivo simultâneos
- ✅ **Thread-Safe**: Mutex para evitar race conditions
- ✅ **Singleton Pattern**: Uma única instância para toda aplicação

## Arquivo de Configuração

O arquivo `logger-config.json` controla o comportamento do logger:

```json
{
  "level": "INFO",
  "output": "console",
  "file_path": "excel-ai.log",
  "components": {
    "APP": "INFO",
    "CHAT": "INFO",
    "EXCEL": "INFO",
    "AI": "INFO",
    "CACHE": "WARN",
    "STORAGE": "WARN",
    "LICENSE": "INFO",
    "HTTP": "INFO",
    "STREAM": "INFO",
    "TOOLS": "INFO",
    "UNDO": "INFO"
  }
}
```

### Parâmetros de Configuração

#### `level` (Opcional)
Nível global de log. Se omitido, usa INFO como padrão.
- `DEBUG`: Mostra todos os logs (muito verboso)
- `INFO`: Mostra logs informativos e acima
- `WARN`: Mostra apenas avisos e erros
- `ERROR`: Mostra apenas erros
- `FATAL`: Mostra apenas erros fatais

#### `output` (Opcional)
Tipo de saída do logger.
- `console`: Apenas console (padrão)
- `file`: Apenas arquivo
- `both`: Console e arquivo (não implementado, use `file`)

#### `file_path` (Opcional)
Caminho do arquivo de log quando `output` é `file`.
- Padrão: `excel-ai.log`

#### `components` (Opcional)
Níveis de log específicos por componente. Se um componente não estiver listado, usa o nível global.

## Como Usar

### Uso Básico

```go
import "excel-ai/pkg/logger"

// Log informativo
logger.AppInfo("Aplicação iniciada com sucesso")

// Log de aviso
logger.ExcelWarn("Excel ocupado, aguardando...")

// Log de erro
logger.AIError("Falha ao conectar com API da IA")

// Log de debug (detalhado)
logger.CacheDebug("Cache hit: chave encontrada")
```

### Formatação Avançada

```go
import (
    "fmt"
    "excel-ai/pkg/logger"
)

// Com formatação
logger.ChatInfo(fmt.Sprintf("Usuário %s enviou mensagem", username))

logger.ExcelDebug(fmt.Sprintf("Range %s contém %d células", range, count))
```

### Níveis por Componente

#### APP - Aplicação Geral
```go
logger.AppDebug("Debug geral")
logger.AppInfo("Informação geral")
logger.AppWarn("Aviso geral")
logger.AppError("Erro geral")
logger.AppFatal("Erro fatal - encerra app")
```

#### CHAT - Sistema de Chat
```go
logger.ChatInfo("Nova mensagem recebida")
logger.ChatError("Erro ao processar mensagem")
```

#### EXCEL - Integração com Excel
```go
logger.ExcelDebug("Conectando ao Excel...")
logger.ExcelInfo("Workbook aberto: Custo de 2024")
logger.ExcelWarn("Range expandido automaticamente")
logger.ExcelError("Falha ao criar tabela dinâmica")
```

#### AI - Inteligência Artificial
```go
logger.AIDebug("Enviando requisição para API")
logger.AIInfo("Resposta recebida com sucesso")
logger.AIError("Timeout na chamada da API")
```

#### CACHE - Sistema de Cache
```go
logger.CacheDebug("Cache hit: consulta_de_vendas")
logger.CacheInfo("Cache miss: calculando resultado")
logger.CacheWarn("Cache invalidado")
logger.CacheError("Falha ao salvar no cache")
```

#### STORAGE - Persistência de Dados
```go
logger.StorageInfo("Dados salvos no banco")
logger.StorageError("Falha ao ler conversações")
```

#### LICENSE - Sistema de Licença
```go
logger.LicenseInfo("Licença válida expira em 30 dias")
logger.LicenseError("Licença inválida ou expirada")
```

#### HTTP - Requisições HTTP
```go
logger.HTTPDebug("GET https://api.openai.com/v1/chat")
logger.HTTPInfo("Resposta 200 OK")
logger.HTTPError("Erro 500 Server Error")
```

#### STREAM - Streaming de Respostas
```go
logger.StreamDebug("Iniciando stream de resposta")
logger.StreamInfo("Chunk recebido: 150 bytes")
logger.StreamError("Stream interrompido")
```

#### TOOLS - Ferramentas do Sistema
```go
logger.ToolsInfo("Ferramenta execute-tool chamada")
logger.ToolsError("Ferramenta desconhecida")
```

#### UNDO - Sistema de Desfazer
```go
logger.UndoInfo("Ação registrada no histórico")
logger.UndoError("Erro ao desfazer ação")
```

## Exemplos de Configuração

### Desenvolvimento (Muito Verboso)
```json
{
  "level": "DEBUG",
  "output": "console",
  "components": {
    "AI": "DEBUG",
    "EXCEL": "DEBUG",
    "CACHE": "DEBUG"
  }
}
```

### Produção (Equilibrado)
```json
{
  "level": "INFO",
  "output": "file",
  "file_path": "logs/excel-ai.log",
  "components": {
    "CACHE": "WARN",
    "STORAGE": "WARN"
  }
}
```

### Produção Crítica (Apenas Erros)
```json
{
  "level": "ERROR",
  "output": "file",
  "file_path": "logs/excel-ai-errors.log"
}
```

### Depuração de Excel
```json
{
  "level": "INFO",
  "output": "console",
  "components": {
    "EXCEL": "DEBUG",
    "AI": "INFO",
    "CHAT": "INFO",
    "CACHE": "WARN",
    "STORAGE": "WARN"
  }
}
```

## Boas Práticas

### 1. Escolha o Nível Adequado
- **DEBUG**: Informações detalhadas para desenvolvimento
- **INFO**: Eventos normais e importantes
- **WARN**: Situações anormais mas não críticas
- **ERROR**: Erros que não param a aplicação
- **FATAL**: Erros que requerem encerramento imediato

### 2. Use o Componente Correto
```go
// ✅ CORRETO
logger.ExcelInfo("Workbook aberto")
logger.AIInfo("Resposta recebida")

// ❌ INCORRETO
logger.AppInfo("Excel: Workbook aberto") // Não use APP para Excel
```

### 3. Seja Específico
```go
// ✅ BOM
logger.ExcelError(fmt.Sprintf("Falha ao criar tabela dinâmica: %v", err))

// ❌ RUIM
logger.AppError("Erro") // Sem contexto
```

### 4. Evite Informações Sensíveis
```go
// ❌ NÃO FAÇA
logger.ChatInfo("Token: sk-xxxxxxxxxxxxxxxxxxxx")

// ✅ FAÇA
logger.ChatInfo("Token configurado com sucesso")
```

## Troubleshooting

### Logger não aparece no console
1. Verifique se `logger-config.json` está no diretório correto
2. Confirme que o arquivo JSON é válido (use um JSON validator)
3. Verifique se o nível de log não está muito alto (ERROR/FATAL)

### Arquivo de log não é criado
1. Verifique permissões de escrita no diretório
2. Confirme que `output` está configurado como `file`
3. Verifique se `file_path` é um caminho válido

### Logs não aparecem para um componente
1. Verifique se o componente está configurado no JSON
2. Confirme que o nível do componente não está muito alto
3. Verifique se está usando a função correta (ex: ExcelInfo vs AppInfo)

## Integração com Panic Recovery

O `main.go` já inclui captura de panic com logging:

```go
defer func() {
    if r := recover(); r != nil {
        logger.AppFatal(fmt.Sprintf("Panic recuperado: %v", r))
    }
}()
```

Isso garante que qualquer panic será logado antes da aplicação encerrar.

## Graceful Shutdown

A aplicação fecha o logger corretamente ao receber SIGINT ou SIGTERM:

```go
go func() {
    sig := <-sigChan
    logger.AppInfo(fmt.Sprintf("Sinal recebido: %v, encerrando graciosamente", sig))
    logger.GetLogger().Close()
    os.Exit(0)
}()
```

## API Completa

### Funções Principais

```go
// Inicialização
logger.InitializeFromFile("logger-config.json")
logger.InitializeWithDefaults(logger.INFO)

// Controle de Níveis
logger.SetComponentLevel("EXCEL", logger.DEBUG)
logger.GetLogger().SetLevel(logger.DEBUG)

// Output em Arquivo
logger.GetLogger().SetFileOutput("app.log")
logger.GetLogger().Close() // Fecha o arquivo
```

### Componentes Disponíveis

- `ComponentApp`
- `ComponentChat`
- `ComponentExcel`
- `ComponentAI`
- `ComponentCache`
- `ComponentStorage`
- `ComponentLicense`
- `ComponentHTTP`
- `ComponentStream`
- `ComponentTools`
- `ComponentUndo`

### Níveis Disponíveis

- `logger.DEBUG`
- `logger.INFO`
- `logger.WARN`
- `logger.ERROR`
- `logger.FATAL`

## Suporte

Para problemas ou dúvidas sobre o sistema de logging, consulte:
1. Este documento (docs/LOGGER_GUIDE.md)
2. O código fonte (pkg/logger/logger.go)
3. A configuração exemplo (logger-config.json)
