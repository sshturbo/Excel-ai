# Excel-ai

<div align="center">

**Uma aplicação de desktop inteligente que integra Excel com IA para análise e manipulação de dados através de linguagem natural**

[![Wails](https://img.shields.io/badge/Wails-v2.11.0-blue)](https://wails.io)
[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18.2-61DAFB?logo=react)](https://reactjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript)](https://www.typescriptlang.org)

</div>

---

## 📋 Índice

- [Sobre o Projeto](#sobre-o-projeto)
- [Características Principais](#características-principais)
- [Tecnologias Utilizadas](#tecnologias-utilizadas)
- [Arquitetura](#arquitetura)
- [Instalação](#instalação)
- [Uso](#uso)
- [Documentação](#documentação)
- [Contribuindo](#contribuindo)
- [Licença](#licença)

---

## 🚀 Sobre o Projeto

Excel-ai é uma aplicação desktop desenvolvida com Wails que permite aos usuários interagir com planilhas do Microsoft Excel usando comandos em linguagem natural. A aplicação utiliza modelos de IA (como GPT, Groq, Google Gemini) para interpretar comandos e executar operações complexas no Excel de forma intuitiva.

### Por que Excel-ai?

- **Interface Natural**: Converse com suas planilhas como se estivesse falando com um assistente
- **Automação Inteligente**: Execute operações complexas com comandos simples
- **Multiplataforma**: Funciona no Windows com suporte para macOS e Linux
- **Integração Nativa**: Conecta-se diretamente ao Microsoft Excel via COM
- **Histórico de Conversas**: Mantenha contexto e histórico de todas as suas interações
- **Visualizações**: Crie gráficos e visualizações automaticamente

---

## ✨ Características Principais

### 🗣️ Comandos em Linguagem Natural
- Análise de dados através de perguntas simples
- Criação e modificação de planilhas por comando de voz
- Formatação automática de dados
- Geração de gráficos e tabelas dinâmicas

### 📊 Manipulação de Excel
- Leitura e escrita de células
- Criação de gráficos (pizza, barras, linhas, etc.)
- Tabelas dinâmicas
- Formatação condicional
- Fórmulas complexas

### 🤖 Integração com IA
- Suporte para múltiplos provedores:
  - OpenRouter (GPT-4, Claude, etc.)
  - Groq (modelos rápidos)
  - Google Gemini
  - APIs personalizadas compatíveis com OpenAI
- Streaming de respostas em tempo real
- Contexto de conversa mantido

### 💾 Gerenciamento de Conversas
- Salvar e carregar conversas anteriores
- Histórico completo de mensagens
- Contexto preservado entre sessões
- Exportar conversas

### 🎨 Interface Moderna
- Design responsivo com Tailwind CSS
- Tema claro/escuro
- Componentes do shadcn/ui
- Suporte a Markdown nas respostas
- Syntax highlighting para código

---

## 🛠️ Tecnologias Utilizadas

### Backend (Go)
- **Wails v2** - Framework para aplicações desktop
- **go-ole** - Integração COM com Microsoft Excel
- **go-sqlite** - Armazenamento local de configurações e conversas

### Frontend (React)
- **React 18** - Biblioteca UI
- **TypeScript** - Tipagem estática
- **Vite** - Build tool e dev server
- **Tailwind CSS** - Framework CSS
- **shadcn/ui** - Componentes UI
- **Radix UI** - Primitivos acessíveis
- **React Markdown** - Renderização de Markdown
- **Chart.js** - Visualizações de dados

### IA
- **OpenRouter API** - Gateway para múltiplos modelos
- **Google Gemini API** - Modelos do Google
- **Groq API** - Inferência rápida

---

## 🏗️ Arquitetura

```
excel-ai/
├── main.go                 # Ponto de entrada da aplicação
├── internal/               # Código interno da aplicação
│   ├── app/               # Lógica principal e handlers
│   ├── domain/            # Modelos de domínio
│   ├── dto/               # Data Transfer Objects
│   └── services/          # Serviços de negócio
│       ├── chat/          # Serviço de chat com IA
│       └── excel/         # Serviço de integração com Excel
├── pkg/                   # Pacotes reutilizáveis
│   ├── ai/               # Clientes de IA
│   ├── excel/            # Cliente COM do Excel
│   ├── license/          # Sistema de licenciamento
│   └── storage/          # Persistência de dados
└── frontend/             # Interface React
    └── src/
        ├── components/   # Componentes React
        ├── hooks/        # Custom hooks
        ├── services/     # Serviços frontend
        └── types/        # Tipos TypeScript
```

Para mais detalhes, veja [ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## 📦 Instalação

### Pré-requisitos

- **Go 1.23+** instalado
- **Node.js 18+** e npm
- **Wails CLI** instalado (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- **Microsoft Excel** instalado (Windows)
- **Chave de API** de um provedor de IA (OpenRouter, Groq, ou Google)

### Passos de Instalação

1. Clone o repositório:
```bash
git clone https://github.com/sshturbo/Excel-ai.git
cd Excel-ai
```

2. Instale as dependências do Go:
```bash
go mod download
```

3. Instale as dependências do frontend:
```bash
cd frontend
npm install
cd ..
```

4. Execute em modo de desenvolvimento:
```bash
wails dev
```

5. Ou compile a aplicação:
```bash
wails build
```

Para instruções detalhadas, consulte [INSTALLATION.md](docs/INSTALLATION.md).

---

## 🎯 Uso

### Iniciando a Aplicação

1. Abra o Microsoft Excel com uma planilha
2. Inicie o Excel-ai
3. Configure sua chave de API nas configurações
4. Comece a conversar com sua planilha!

### Exemplos de Comandos

```
"Qual é a soma da coluna A?"
"Crie um gráfico de pizza com os dados da coluna B"
"Formate a linha 1 em negrito"
"Adicione uma tabela dinâmica com os dados da aba Vendas"
"Calcule a média dos últimos 10 valores"
```

Para mais exemplos e guia completo, veja [USER_GUIDE.md](docs/USER_GUIDE.md).

---

## 📚 Documentação

A documentação completa está organizada em:

- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Arquitetura detalhada do sistema
- **[INSTALLATION.md](docs/INSTALLATION.md)** - Guia completo de instalação
- **[DEVELOPMENT.md](docs/DEVELOPMENT.md)** - Guia para desenvolvedores
- **[API.md](docs/API.md)** - Documentação da API e métodos backend
- **[FRONTEND.md](docs/FRONTEND.md)** - Estrutura e componentes do frontend
- **[USER_GUIDE.md](docs/USER_GUIDE.md)** - Manual do usuário
- **[CONFIGURATION.md](docs/CONFIGURATION.md)** - Opções de configuração
- **[CONTRIBUTING.md](docs/CONTRIBUTING.md)** - Como contribuir
- **[LICENSE_INFO.md](docs/LICENSE_INFO.md)** - Informações sobre licenciamento

---

## 🤝 Contribuindo

Contribuições são bem-vindas! Por favor, leia [CONTRIBUTING.md](docs/CONTRIBUTING.md) para detalhes sobre nosso código de conduta e processo de submissão de pull requests.

---

## 👨‍💻 Autor

**Jefferson Hipolito de Oliveira**
- Email: jefferson@hiposystem.com.br
- GitHub: [@sshturbo](https://github.com/sshturbo)

---

## 📄 Licença

Este projeto está sob uma licença proprietária. Veja [LICENSE_INFO.md](docs/LICENSE_INFO.md) para mais informações.

---

## 🙏 Agradecimentos

- [Wails](https://wails.io) - Framework incrível para aplicações desktop
- [shadcn/ui](https://ui.shadcn.com) - Componentes UI elegantes
- Comunidade Go e React

---

<div align="center">

**⭐ Se este projeto foi útil, considere dar uma estrela! ⭐**

</div>
