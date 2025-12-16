# Configuração - Excel-ai

Este documento detalha todas as opções de configuração disponíveis no Excel-ai.

## Índice

- [Visão Geral](#visão-geral)
- [Configurações de API](#configurações-de-api)
- [Configurações de Dados](#configurações-de-dados)
- [Configurações Avançadas](#configurações-avançadas)
- [Arquivo de Configuração](#arquivo-de-configuração)
- [Variáveis de Ambiente](#variáveis-de-ambiente)

## Visão Geral

O Excel-ai oferece duas formas de configuração:

1. **Interface Gráfica**: Através do menu de configurações (⚙️)
2. **Arquivo Local**: SQLite database armazenado localmente

As configurações são salvas automaticamente e persistem entre sessões.

## Configurações de API

### Provider

**Tipo**: String  
**Valores**: `openrouter`, `groq`, `google`, `custom`  
**Padrão**: `groq`  
**Descrição**: Provedor de API de IA a ser usado

#### OpenRouter
- **Acesso a**: GPT-4, Claude, Llama, Mixtral, e muitos outros
- **Website**: https://openrouter.ai/
- **Pricing**: Varia por modelo (pay-as-you-go)
- **Vantagem**: Maior variedade de modelos

**Configuração**:
```json
{
  "provider": "openrouter",
  "apiKey": "sk-or-v1-...",
  "model": "openai/gpt-4-turbo",
  "baseUrl": "https://openrouter.ai/api/v1"
}
```

#### Groq
- **Acesso a**: Llama 3.1, Mixtral, Gemma
- **Website**: https://console.groq.com/
- **Pricing**: Grátis (com rate limits)
- **Vantagem**: Extremamente rápido, ótimo para começar

**Configuração**:
```json
{
  "provider": "groq",
  "apiKey": "gsk_...",
  "model": "llama-3.1-70b-versatile",
  "baseUrl": "https://api.groq.com/openai/v1"
}
```

#### Google Gemini
- **Acesso a**: Gemini Pro, Gemini Ultra (quando disponível)
- **Website**: https://makersuite.google.com/
- **Pricing**: Grátis com limites, pago para uso intenso
- **Vantagem**: Multimodal, bom para dados complexos

**Configuração**:
```json
{
  "provider": "google",
  "apiKey": "AIza...",
  "model": "gemini-pro",
  "baseUrl": ""
}
```

#### Custom
- **Descrição**: Qualquer API compatível com OpenAI
- **Requisito**: Endpoint deve seguir o padrão OpenAI Chat Completions
- **Uso**: Ideal para modelos self-hosted ou outros provedores

**Configuração**:
```json
{
  "provider": "custom",
  "apiKey": "seu-token",
  "model": "nome-do-modelo",
  "baseUrl": "https://sua-api.com/v1"
}
```

---

### API Key

**Tipo**: String (sensível)  
**Obrigatório**: Sim  
**Descrição**: Chave de autenticação para a API do provedor

**Segurança**:
- Armazenada localmente em SQLite
- Nunca enviada para servidores do Excel-ai
- Enviada apenas para o provedor de IA selecionado

**Como obter**:

| Provider | URL | Passos |
|----------|-----|--------|
| OpenRouter | https://openrouter.ai/keys | Login → API Keys → Create Key |
| Groq | https://console.groq.com/keys | Login → API Keys → Create API Key |
| Google | https://makersuite.google.com/app/apikey | Login → Get API Key |

---

### Model

**Tipo**: String  
**Descrição**: Identificador do modelo de IA a ser usado

#### Modelos Recomendados

**Para Análise de Dados**:
- `anthropic/claude-3.5-sonnet` (OpenRouter) - Excelente raciocínio
- `google/gemini-pro` (Google) - Bom com números
- `llama-3.1-70b-versatile` (Groq) - Rápido e eficaz

**Para Uso Geral**:
- `openai/gpt-3.5-turbo` (OpenRouter) - Bom custo-benefício
- `mixtral-8x7b-32768` (Groq) - Contexto grande
- `gemini-pro` (Google) - Gratuito

**Para Máxima Qualidade**:
- `openai/gpt-4-turbo` (OpenRouter) - Melhor da classe
- `anthropic/claude-3-opus` (OpenRouter) - Análise profunda

**Para Velocidade**:
- `llama-3.1-8b-instant` (Groq) - Ultra rápido
- `mixtral-8x7b-instant` (Groq) - Rápido e capaz

#### Listagem Completa

**OpenRouter**:
```
openai/gpt-4-turbo
openai/gpt-4
openai/gpt-3.5-turbo
anthropic/claude-3.5-sonnet
anthropic/claude-3-opus
anthropic/claude-3-haiku
meta-llama/llama-3.1-405b-instruct
meta-llama/llama-3.1-70b-instruct
mistralai/mixtral-8x22b-instruct
google/gemini-pro-1.5
... (100+ modelos disponíveis)
```

**Groq**:
```
llama-3.1-405b-reasoning
llama-3.1-70b-versatile
llama-3.1-8b-instant
mixtral-8x7b-32768
gemma2-9b-it
```

**Google**:
```
gemini-pro
gemini-pro-vision
gemini-1.5-pro
gemini-1.5-flash
```

---

### Base URL

**Tipo**: String (URL)  
**Obrigatório**: Apenas para provider `custom`  
**Descrição**: URL base do endpoint da API

**Padrões**:
- OpenRouter: `https://openrouter.ai/api/v1`
- Groq: `https://api.groq.com/openai/v1`
- Google: Não aplicável (usa SDK próprio)
- Custom: Configurado pelo usuário

**Formato esperado**:
O endpoint deve ter `/chat/completions` disponível, ex:
```
https://sua-api.com/v1/chat/completions
```

---

## Configurações de Dados

### Auto-refresh Workbooks

**Tipo**: Boolean  
**Padrão**: `true`  
**Descrição**: Atualiza automaticamente a lista de workbooks abertos

Quando ativado:
- Excel-ai detecta quando você abre/fecha workbooks
- Lista é atualizada automaticamente
- Não precisa clicar em "Atualizar"

Quando desativado:
- Lista só atualiza manualmente
- Pode economizar recursos em máquinas lentas

---

### Preview Rows

**Tipo**: Number  
**Padrão**: `10`  
**Intervalo**: `5-100`  
**Descrição**: Número de linhas a mostrar em previews de dados

Valores recomendados:
- `5-10`: Planilhas grandes (performance)
- `20-50`: Planilhas médias (balanço)
- `50-100`: Planilhas pequenas (visão completa)

---

### Max History Messages

**Tipo**: Number  
**Padrão**: `20`  
**Intervalo**: `5-100`  
**Descrição**: Quantas mensagens manter no contexto da IA

**Impacto**:
- **Mais mensagens**: 
  - ✅ IA tem mais contexto
  - ✅ Melhor continuidade da conversa
  - ❌ Mais tokens usados (mais caro)
  - ❌ Respostas mais lentas
  
- **Menos mensagens**:
  - ✅ Mais rápido
  - ✅ Mais barato
  - ❌ IA "esquece" contexto antigo
  - ❌ Pode precisar repetir informações

**Recomendações**:
- Conversas curtas: `5-10`
- Uso normal: `15-20`
- Análises complexas: `30-50`

---

### Data Send Mode

**Tipo**: String  
**Valores**: `auto`, `preview`, `full`, `none`  
**Padrão**: `auto`  
**Descrição**: Como dados do Excel são enviados à IA

#### Auto
- IA decide quanto dado precisa
- Geralmente envia preview (primeiras linhas)
- Pode pedir mais se necessário

#### Preview
- Sempre envia apenas preview
- Mais rápido, menos tokens
- Pode não ter informação completa

#### Full
- Envia todos os dados da sheet/range
- Mais preciso
- Mais caro e lento

#### None
- Não envia dados automaticamente
- IA trabalha apenas com descrições
- Mais limitado

---

## Configurações Avançadas

### Temperatura

**Tipo**: Number  
**Padrão**: `0.7`  
**Intervalo**: `0.0-2.0`  
**Descrição**: Controla aleatoriedade das respostas

- **0.0-0.3**: Muito determinístico, sempre mesma resposta
- **0.4-0.7**: Balanço (recomendado para Excel)
- **0.8-1.2**: Mais criativo
- **1.3-2.0**: Muito aleatório (não recomendado)

**Quando ajustar**:
- Análises precisas: `0.2-0.4`
- Uso geral: `0.6-0.8`
- Brainstorming: `1.0-1.2`

---

### Max Tokens

**Tipo**: Number  
**Padrão**: `4096`  
**Descrição**: Máximo de tokens na resposta

**Considerações**:
- Respostas mais longas = mais custo
- Alguns modelos têm limites diferentes
- Para código/análises: `2000-4096`
- Para respostas curtas: `500-1000`

---

### Timeout

**Tipo**: Number (segundos)  
**Padrão**: `60`  
**Descrição**: Tempo máximo de espera por resposta

**Ajustar se**:
- Respostas lentas frequentes: Aumentar para `90-120`
- Quer falhar rápido: Reduzir para `30-45`

---

### Stream Responses

**Tipo**: Boolean  
**Padrão**: `true`  
**Descrição**: Mostrar resposta em tempo real (streaming)

**Ativado** (recomendado):
- Vê a resposta sendo gerada
- Melhor UX
- Pode cancelar no meio

**Desativado**:
- Aguarda resposta completa
- Simples, mas menos interativo

---

## Arquivo de Configuração

### Localização

O Excel-ai usa SQLite para armazenar configurações.

**Caminho do arquivo**:
- **Windows**: `%APPDATA%\excel-ai\storage.db`
- **macOS**: `~/Library/Application Support/excel-ai/storage.db`
- **Linux**: `~/.local/share/excel-ai/storage.db`

### Estrutura

```sql
-- Tabela config
CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Exemplos de chaves
INSERT INTO config VALUES ('api_provider', 'groq');
INSERT INTO config VALUES ('api_key', 'gsk_...');
INSERT INTO config VALUES ('model', 'llama-3.1-70b-versatile');
INSERT INTO config VALUES ('base_url', 'https://api.groq.com/openai/v1');
```

### Backup Manual

Para fazer backup das configurações:

```bash
# Windows (PowerShell)
Copy-Item "$env:APPDATA\excel-ai\storage.db" "backup-$(Get-Date -Format 'yyyyMMdd').db"

# macOS/Linux
cp ~/Library/Application\ Support/excel-ai/storage.db ~/backup-$(date +%Y%m%d).db
```

### Restaurar Configuração

Para restaurar backup:

```bash
# Feche o Excel-ai primeiro!

# Windows (PowerShell)
Copy-Item "backup-20240101.db" "$env:APPDATA\excel-ai\storage.db" -Force

# macOS/Linux
cp ~/backup-20240101.db ~/Library/Application\ Support/excel-ai/storage.db
```

### Resetar Configuração

Para voltar às configurações padrão:

```bash
# Feche o Excel-ai primeiro!

# Windows (PowerShell)
Remove-Item "$env:APPDATA\excel-ai\storage.db"

# macOS/Linux
rm ~/Library/Application\ Support/excel-ai/storage.db
```

O arquivo será recriado com padrões na próxima execução.

---

## Variáveis de Ambiente

### EXCEL_AI_CONFIG_PATH

**Descrição**: Caminho customizado para o arquivo de configuração  
**Uso**: Para usuários avançados que querem controlar onde configs são salvas

```bash
# Windows
set EXCEL_AI_CONFIG_PATH=C:\MyConfigs\excel-ai.db

# macOS/Linux
export EXCEL_AI_CONFIG_PATH=/custom/path/excel-ai.db
```

---

### EXCEL_AI_LOG_LEVEL

**Descrição**: Nível de logging  
**Valores**: `debug`, `info`, `warn`, `error`  
**Padrão**: `info`

```bash
# Para debug detalhado
set EXCEL_AI_LOG_LEVEL=debug

# Para apenas erros
set EXCEL_AI_LOG_LEVEL=error
```

---

## Perfis de Configuração

### Perfil: Desenvolvedor

```json
{
  "provider": "groq",
  "model": "llama-3.1-70b-versatile",
  "maxHistoryMessages": 10,
  "previewRows": 20,
  "temperature": 0.3,
  "streamResponses": true
}
```

### Perfil: Analista de Dados

```json
{
  "provider": "openrouter",
  "model": "anthropic/claude-3.5-sonnet",
  "maxHistoryMessages": 30,
  "previewRows": 50,
  "temperature": 0.5,
  "dataSendMode": "full"
}
```

### Perfil: Uso Casual

```json
{
  "provider": "groq",
  "model": "llama-3.1-8b-instant",
  "maxHistoryMessages": 10,
  "previewRows": 10,
  "temperature": 0.7,
  "dataSendMode": "preview"
}
```

### Perfil: Máxima Qualidade

```json
{
  "provider": "openrouter",
  "model": "openai/gpt-4-turbo",
  "maxHistoryMessages": 50,
  "previewRows": 100,
  "temperature": 0.6,
  "dataSendMode": "full",
  "maxTokens": 8192
}
```

---

## Troubleshooting

### Configurações não salvam

**Possíveis causas**:
- Permissões de arquivo
- Disco cheio
- Arquivo corrompido

**Solução**:
1. Verifique permissões do diretório de dados
2. Verifique espaço em disco
3. Tente resetar configuração (veja acima)

### API Key não funciona

**Verificar**:
- Copiou corretamente (sem espaços)
- Chave do provedor correto
- Chave não expirou/foi revogada
- Billing configurado (alguns provedores)

### Respostas inconsistentes

**Ajustar**:
- Reduzir temperatura (0.3-0.5)
- Usar modelo mais capaz
- Aumentar max tokens se respostas são cortadas

---

## Melhores Práticas

### ✅ Faça

- Mantenha API keys seguras
- Teste diferentes modelos para seu uso
- Ajuste temperatura baseado na tarefa
- Faça backup de configurações importantes
- Use streaming para melhor UX

### ❌ Não Faça

- Compartilhar API keys
- Usar temperatura > 1.0 para trabalho sério
- Enviar dados sensíveis desnecessariamente
- Ignorar limites de rate dos provedores

---

## Referências

- [API Documentation](API.md)
- [User Guide](USER_GUIDE.md)
- [OpenRouter Models](https://openrouter.ai/models)
- [Groq Documentation](https://console.groq.com/docs)
- [Google AI Documentation](https://ai.google.dev/)
