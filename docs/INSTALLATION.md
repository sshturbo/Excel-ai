# Guia de Instalação - Excel-ai

Este guia fornece instruções detalhadas para instalar e configurar o Excel-ai em seu sistema.

## Requisitos do Sistema

### Sistema Operacional

- **Windows 10/11** (recomendado) - Suporte completo COM
- **macOS** - Suporte parcial (sem integração COM)
- **Linux** - Suporte parcial (sem integração COM)

### Software Necessário

#### 1. Microsoft Excel
- **Versão**: Office 2016 ou superior
- **Tipo**: Instalação desktop (não Office Online)
- **Necessário para**: Integração COM completa

#### 2. Go
- **Versão**: 1.23 ou superior
- **Download**: https://golang.org/dl/
- **Verificar instalação**:
  ```bash
  go version
  # Deve exibir: go version go1.23.x ...
  ```

#### 3. Node.js e npm
- **Versão**: Node.js 18.x ou superior
- **Download**: https://nodejs.org/
- **Verificar instalação**:
  ```bash
  node --version  # v18.x.x ou superior
  npm --version   # 9.x.x ou superior
  ```

#### 4. Wails CLI
- **Instalação**:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- **Verificar instalação**:
  ```bash
  wails version
  # Deve exibir: Wails v2.11.0 ou superior
  ```

#### 5. Git
- **Download**: https://git-scm.com/downloads
- **Verificar instalação**:
  ```bash
  git --version
  ```

### Dependências do Sistema

#### Windows

**Compilador C++** (para dependências nativas):

Opção 1 - Visual Studio Build Tools:
```bash
# Baixar de: https://visualstudio.microsoft.com/downloads/
# Instalar "Desktop development with C++"
```

Opção 2 - MinGW-w64:
```bash
# Baixar de: https://www.mingw-w64.org/
```

**WebView2** (geralmente já instalado no Windows 10/11):
- Se necessário, baixe de: https://developer.microsoft.com/en-us/microsoft-edge/webview2/

#### macOS

```bash
# Instalar Xcode Command Line Tools
xcode-select --install
```

#### Linux

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk3-devel

# Arch Linux
sudo pacman -S gtk3 webkit2gtk
```

## Instalação Passo a Passo

### 1. Clonar o Repositório

```bash
# Via HTTPS
git clone https://github.com/sshturbo/Excel-ai.git

# Ou via SSH
git clone git@github.com:sshturbo/Excel-ai.git

# Entrar no diretório
cd Excel-ai
```

### 2. Instalar Dependências Go

```bash
# Baixar módulos Go
go mod download

# Verificar integridade
go mod verify
```

**Verificar que os módulos principais foram instalados**:
```bash
go list -m all | grep -E "wails|go-ole|sqlite"
```

### 3. Instalar Dependências Frontend

```bash
# Entrar no diretório frontend
cd frontend

# Instalar dependências npm
npm install

# Verificar instalação
npm list --depth=0

# Voltar ao diretório raiz
cd ..
```

### 4. Configurar Wails Doctor

Verifique se todos os requisitos foram atendidos:

```bash
wails doctor
```

**Saída esperada**:
```
Wails CLI v2.11.0

Scanning system - Please wait

System
------
OS:                     windows 10 amd64
Version:               ...
...

Go
--
Version:               1.23.x
...

Dependencies
------------
gcc:                   Available
npm:                   Available
...

✔ Your system is ready for Wails development!
```

Se houver problemas, siga as instruções fornecidas pelo `wails doctor`.

### 5. Configurar Chave de API

Antes de executar pela primeira vez, você precisará de uma chave de API de um provedor de IA.

#### Opção 1: Groq (Recomendado para começar)

1. Acesse https://console.groq.com/
2. Crie uma conta ou faça login
3. Vá para "API Keys"
4. Crie uma nova chave
5. Copie a chave

#### Opção 2: OpenRouter

1. Acesse https://openrouter.ai/
2. Crie uma conta
3. Vá para "Keys"
4. Crie uma nova chave
5. Copie a chave

#### Opção 3: Google Gemini

1. Acesse https://makersuite.google.com/app/apikey
2. Crie uma chave de API
3. Copie a chave

> **Nota**: A chave será configurada na interface da aplicação na primeira execução.

## Modos de Execução

### Modo de Desenvolvimento

Execute a aplicação com hot-reload para desenvolvimento:

```bash
wails dev
```

**O que acontece**:
- Backend Go é compilado e executado
- Frontend React inicia com Vite dev server
- Hot-reload ativo para mudanças no frontend
- Console de debug disponível

**Acessos**:
- Aplicação desktop abre automaticamente
- Dev server frontend: http://localhost:34115 (acesso pelo navegador para debug)

**Logs**:
```
DEB | [ExternalLogger] [Vite] Server started at http://localhost:5173
DEB | [ExternalLogger] [Vite] Ready in 250ms
DEB | [Window] Window created
```

### Modo de Produção

Compile a aplicação para distribuição:

```bash
wails build
```

**Flags úteis**:
```bash
# Build com console de debug (Windows)
wails build -debug

# Build otimizado (sem debug)
wails build -ldflags="-w -s"

# Build para plataforma específica
wails build -platform windows/amd64

# Build com UPX compression
wails build -upx
```

**Saída**:
- **Windows**: `build/bin/Excel-ai.exe`
- **macOS**: `build/bin/Excel-ai.app`
- **Linux**: `build/bin/Excel-ai`

### Executar Build

```bash
# Windows
.\build\bin\Excel-ai.exe

# macOS
open build/bin/Excel-ai.app

# Linux
./build/bin/Excel-ai
```

## Configuração Inicial

### 1. Primeira Execução

1. Inicie o Excel-ai
2. Abra o Microsoft Excel com uma planilha de teste
3. Na interface do Excel-ai, clique em "Configurações" (ícone de engrenagem)

### 2. Configurar API

**Na aba "API"**:

1. **Provider**: Selecione o provedor
   - OpenRouter (para GPT-4, Claude, etc.)
   - Groq (rápido e gratuito)
   - Google (Gemini)
   - Custom (API própria)

2. **API Key**: Cole sua chave de API

3. **Model**: Selecione o modelo
   - OpenRouter: `openai/gpt-4-turbo`, `anthropic/claude-3.5-sonnet`, etc.
   - Groq: `llama-3.1-70b-versatile`, `mixtral-8x7b`, etc.
   - Google: `gemini-pro`, `gemini-pro-vision`

4. **Base URL** (opcional para custom):
   - OpenRouter: `https://openrouter.ai/api/v1`
   - Groq: `https://api.groq.com/openai/v1`
   - Custom: Seu endpoint

5. Clique em "Salvar Configurações"

### 3. Testar Conexão

1. Certifique-se de que o Excel está aberto
2. Digite uma mensagem de teste: "Olá, você está conectado?"
3. Verifique se recebe uma resposta da IA

### 4. Testar Integração com Excel

1. No Excel, crie uma planilha simples com dados na coluna A
2. No Excel-ai, digite: "Qual é a soma da coluna A?"
3. Verifique se a IA consegue ler e processar os dados

## Verificação de Instalação

Execute esta checklist para garantir que tudo está funcionando:

- [ ] `wails doctor` mostra sistema pronto
- [ ] `go version` retorna 1.23 ou superior
- [ ] `node --version` retorna 18.x ou superior
- [ ] `wails version` retorna 2.11.0 ou superior
- [ ] Microsoft Excel instalado e funcionando
- [ ] `wails dev` inicia sem erros
- [ ] Interface abre e exibe corretamente
- [ ] Configurações de API salvam com sucesso
- [ ] Consegue enviar mensagem e receber resposta
- [ ] Excel-ai detecta workbooks abertos no Excel
- [ ] Comandos executam ações no Excel

## Solução de Problemas

### Erro: "Excel Application not found"

**Causa**: Excel não está aberto ou não foi detectado

**Solução**:
1. Abra o Microsoft Excel
2. Abra ou crie uma planilha
3. Tente novamente no Excel-ai

### Erro: "Call was rejected by callee"

**Causa**: Excel está ocupado (editando célula, diálogo aberto)

**Solução**:
1. Saia do modo de edição de célula (pressione ESC)
2. Feche qualquer diálogo aberto no Excel
3. Tente novamente

### Erro: "Failed to initialize COM"

**Causa**: Problema com inicialização COM (Windows)

**Solução**:
1. Reinicie o Excel-ai
2. Execute como Administrador se necessário
3. Verifique se não há outro processo interferindo

### Erro: "npm install" falha

**Causa**: Problemas de rede ou cache corrompido

**Solução**:
```bash
# Limpar cache npm
npm cache clean --force

# Deletar node_modules e package-lock
rm -rf frontend/node_modules
rm frontend/package-lock.json

# Reinstalar
cd frontend && npm install
```

### Erro: "go mod download" lento ou falha

**Causa**: Problemas de proxy ou rede

**Solução**:
```bash
# Configurar proxy Go (se necessário)
export GOPROXY=https://proxy.golang.org,direct

# Ou usar mirror
export GOPROXY=https://goproxy.io,direct

# Tentar novamente
go mod download
```

### Erro: "wails: command not found"

**Causa**: GOPATH/bin não está no PATH

**Solução**:
```bash
# Verificar GOPATH
go env GOPATH

# Adicionar ao PATH (Linux/macOS)
export PATH=$PATH:$(go env GOPATH)/bin

# Windows (PowerShell)
$env:Path += ";$(go env GOPATH)\bin"

# Adicionar permanentemente ao .bashrc, .zshrc, ou variáveis de ambiente do Windows
```

### Problema: Interface não carrega corretamente

**Causa**: Assets frontend não foram compilados

**Solução**:
```bash
# Compilar frontend manualmente
cd frontend
npm run build
cd ..

# Executar novamente
wails dev
```

### Erro: "API request failed"

**Causa**: Chave de API inválida ou problemas de rede

**Solução**:
1. Verifique se a chave de API está correta
2. Verifique sua conexão com a internet
3. Verifique se o provedor de API está online
4. Teste a chave usando curl:

```bash
# Teste OpenRouter
curl -X POST https://openrouter.ai/api/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-3.5-turbo","messages":[{"role":"user","content":"test"}]}'
```

### Porta em uso

**Causa**: Porta 34115 ou 5173 já está em uso

**Solução**:
```bash
# Encontrar e matar processo (Windows)
netstat -ano | findstr :34115
taskkill /PID <PID> /F

# Linux/macOS
lsof -ti:34115 | xargs kill -9
```

## Build para Distribuição

### Windows Installer

```bash
# Build com NSIS installer
wails build -nsis
```

Isso criará um instalador `.exe` em `build/bin/`.

### Assinatura de Código (Opcional)

Para distribuição profissional:

1. Obtenha um certificado de assinatura de código
2. Use `signtool` (Windows) ou `codesign` (macOS):

```bash
# Windows
signtool sign /f certificate.pfx /p password /tr http://timestamp.digicert.com build/bin/Excel-ai.exe

# macOS
codesign --deep --force --verify --verbose --sign "Developer ID" build/bin/Excel-ai.app
```

## Atualizações

Para atualizar o Excel-ai:

```bash
# Atualizar código
git pull origin main

# Atualizar dependências Go
go mod download

# Atualizar dependências frontend
cd frontend && npm install && cd ..

# Rebuild
wails build
```

## Desinstalação

### Remover Aplicação

```bash
# Deletar diretório
rm -rf Excel-ai/

# Windows: Deletar build/bin/Excel-ai.exe
```

### Limpar Dados do Usuário

Os dados do usuário ficam em:

- **Windows**: `%APPDATA%/excel-ai/`
- **macOS**: `~/Library/Application Support/excel-ai/`
- **Linux**: `~/.local/share/excel-ai/`

Remova esses diretórios para limpar completamente.

## Próximos Passos

Após a instalação bem-sucedida:

1. Leia o [USER_GUIDE.md](USER_GUIDE.md) para aprender a usar
2. Explore os exemplos de comandos
3. Configure modelos avançados se necessário
4. Consulte [DEVELOPMENT.md](DEVELOPMENT.md) se quiser contribuir

## Suporte

Se encontrar problemas não cobertos aqui:

1. Verifique as [Issues no GitHub](https://github.com/sshturbo/Excel-ai/issues)
2. Abra uma nova issue com detalhes do problema
3. Inclua logs de erro e versões do sistema

## Recursos Adicionais

- [Wails Installation Guide](https://wails.io/docs/gettingstarted/installation)
- [Go Installation](https://golang.org/doc/install)
- [Node.js Downloads](https://nodejs.org/en/download/)
- [Git Installation](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
