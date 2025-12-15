# Guia do Usuário - Excel-ai

Bem-vindo ao Excel-ai! Este guia irá ajudá-lo a aproveitar ao máximo a aplicação.

## Índice

- [Introdução](#introdução)
- [Primeiros Passos](#primeiros-passos)
- [Interface do Usuário](#interface-do-usuário)
- [Comandos Básicos](#comandos-básicos)
- [Operações Avançadas](#operações-avançadas)
- [Gerenciamento de Conversas](#gerenciamento-de-conversas)
- [Configurações](#configurações)
- [Dicas e Truques](#dicas-e-truques)
- [Resolução de Problemas](#resolução-de-problemas)

## Introdução

Excel-ai é um assistente inteligente que permite interagir com o Microsoft Excel usando linguagem natural. Você pode fazer perguntas, solicitar análises, criar gráficos e executar operações complexas simplesmente conversando com a aplicação.

### O que você pode fazer

- 📊 **Analisar dados**: "Qual é a média da coluna de vendas?"
- 📈 **Criar gráficos**: "Crie um gráfico de pizza com os dados da coluna B"
- ✏️ **Editar planilhas**: "Preencha a coluna C com a soma de A e B"
- 🎨 **Formatar células**: "Formate a primeira linha em negrito e azul"
- 📉 **Tabelas dinâmicas**: "Crie uma tabela dinâmica agrupando por categoria"
- 🔍 **Consultas complexas**: "Mostre os 10 produtos mais vendidos no último mês"

## Primeiros Passos

### Passo 1: Abrir o Excel

Antes de usar o Excel-ai, você precisa ter o Microsoft Excel aberto com uma planilha.

1. Abra o Microsoft Excel
2. Abra uma planilha existente ou crie uma nova
3. Certifique-se de que há dados na planilha (se quiser fazer análises)

### Passo 2: Iniciar o Excel-ai

1. Execute o Excel-ai
2. A aplicação será aberta em uma janela dedicada

### Passo 3: Configurar a API

Na primeira vez que usar:

1. Clique no ícone de **engrenagem** (⚙️) no canto superior direito
2. Selecione a aba **"API"**
3. Escolha um provedor (recomendamos **Groq** para começar)
4. Cole sua **chave de API**
5. Selecione um **modelo** (ex: `llama-3.1-70b-versatile`)
6. Clique em **"Salvar Configurações"**

### Passo 4: Primeira Conversa

1. Digite uma mensagem no campo de entrada: "Olá! Você pode me ajudar?"
2. Pressione **Enter** ou clique no botão enviar
3. Aguarde a resposta da IA

🎉 **Pronto!** Você está usando o Excel-ai!

## Interface do Usuário

### Layout Principal

```
┌─────────────────────────────────────────────────────────┐
│  [Logo] Excel-ai                    [🌓] [⚙️] [📁]    │  Header
├──────────────┬──────────────────────────────────────────┤
│              │                                          │
│  Conversas   │         Área de Chat                    │
│  Anteriores  │                                          │
│              │  [Mensagem da IA]                        │
│  • Conv 1    │                                          │
│  • Conv 2    │  [Sua mensagem]                          │
│  • Conv 3    │                                          │
│              │  [Mensagem da IA]                        │
│              │                                          │
│  [+ Nova]    │                                          │
│              ├──────────────────────────────────────────┤
│              │  [Digite sua mensagem aqui...] [Enviar] │
└──────────────┴──────────────────────────────────────────┘
```

### Componentes

#### 1. Header (Cabeçalho)
- **Logo**: Identifica a aplicação
- **Tema** (🌓): Alterna entre modo claro/escuro
- **Configurações** (⚙️): Acessa configurações
- **Conversas** (📁): Gerencia conversas salvas

#### 2. Sidebar (Barra Lateral)
- Lista de conversas anteriores
- Botão "Nova Conversa"
- Contador de workbooks detectados

#### 3. Área de Chat
- Mensagens do usuário (alinhadas à direita)
- Respostas da IA (alinhadas à esquerda)
- Suporte a Markdown para formatação rica
- Code blocks com syntax highlighting

#### 4. Campo de Entrada
- Digite suas mensagens aqui
- Pressione **Enter** para enviar
- **Shift+Enter** para nova linha

## Comandos Básicos

### Análise de Dados

#### Consultar Valores

```
"Qual é o valor da célula A1?"
"Mostre-me os dados do range A1:C10"
"Quais são os valores únicos na coluna B?"
```

#### Cálculos Simples

```
"Qual é a soma da coluna A?"
"Calcule a média dos valores em B2:B20"
"Qual é o maior valor na coluna de vendas?"
```

#### Estatísticas

```
"Faça uma análise estatística da coluna de preços"
"Qual é o desvio padrão dos dados?"
"Mostre a mediana e a moda da coluna C"
```

### Manipulação de Dados

#### Inserir Dados

```
"Escreva 'Total' na célula A10"
"Preencha o range B1:B5 com valores de 1 a 5"
"Adicione uma nova linha com [Nome, 25, São Paulo]"
```

#### Modificar Dados

```
"Multiplique todos os valores da coluna B por 1.1"
"Converta os textos da coluna A para maiúsculas"
"Substitua todos os valores 'N/A' por 0"
```

#### Copiar e Mover

```
"Copie os dados de A1:A10 para C1:C10"
"Mova a coluna B para a posição D"
```

### Formatação

#### Formatação de Texto

```
"Deixe a linha 1 em negrito"
"Formate a coluna A em itálico e vermelho"
"Centralize o texto da célula B2"
```

#### Formatação de Números

```
"Formate a coluna de preços como moeda (R$)"
"Exiba a coluna de porcentagens com 2 casas decimais"
"Formate as datas no padrão DD/MM/AAAA"
```

#### Cores e Bordas

```
"Pinte a célula A1 de amarelo"
"Adicione bordas ao range A1:D10"
"Destaque em verde as células com valores acima de 100"
```

### Criação de Gráficos

#### Gráficos Básicos

```
"Crie um gráfico de barras com os dados de A1:B10"
"Faça um gráfico de pizza usando a coluna B"
"Gere um gráfico de linhas para visualizar a tendência"
```

#### Gráficos Personalizados

```
"Crie um gráfico de dispersão comparando colunas A e B"
"Faça um gráfico de área empilhada com as últimas 3 colunas"
"Gere um gráfico de barras horizontal, título 'Vendas por Região'"
```

### Fórmulas

#### Aplicar Fórmulas

```
"Adicione uma fórmula em C1 que some A1 e B1"
"Crie uma fórmula para calcular 10% de desconto"
"Use PROCV para buscar dados de outra planilha"
```

#### Fórmulas Condicionais

```
"Adicione uma fórmula SE: se A1>100, escreva 'Alto', senão 'Baixo'"
"Conte quantas células têm valores acima de 50"
"Some apenas os valores que atendem à condição X"
```

## Operações Avançadas

### Tabelas Dinâmicas

```
"Crie uma tabela dinâmica com os dados de A1:E100"
"Agrupe por categoria e some os valores"
"Faça uma tabela dinâmica mostrando vendas por mês e região"
```

### Filtros e Ordenação

```
"Filtre para mostrar apenas valores acima de 1000"
"Ordene a coluna A em ordem alfabética"
"Mostre apenas as linhas onde a coluna B = 'Ativo'"
```

### Análise de Tendências

```
"Identifique tendências nos dados de vendas"
"Preveja os próximos 3 meses baseado no histórico"
"Mostre a correlação entre as colunas X e Y"
```

### Limpeza de Dados

```
"Remova linhas duplicadas"
"Elimine células vazias da coluna A"
"Padronize os formatos de data"
```

### Validação de Dados

```
"Adicione validação: apenas números entre 1 e 100"
"Crie uma lista suspensa com opções [Sim, Não, Talvez]"
"Valide emails na coluna de contatos"
```

## Gerenciamento de Conversas

### Salvar Conversa

1. Durante uma conversa, clique no ícone de **salvar** (💾)
2. Digite um título descritivo (ex: "Análise de Vendas Q1 2024")
3. Clique em **"Salvar"**

A conversa é salva localmente e pode ser recuperada mais tarde.

### Carregar Conversa

1. Clique no ícone de **conversas** (📁) no header
2. Ou use a **sidebar** para ver a lista
3. Clique em uma conversa para carregá-la
4. Todo o histórico será restaurado

### Excluir Conversa

1. Na lista de conversas, hover sobre a conversa
2. Clique no ícone de **lixeira** (🗑️)
3. Confirme a exclusão

### Nova Conversa

1. Clique em **"+ Nova Conversa"** na sidebar
2. Ou pressione **Ctrl+N** (Windows) / **Cmd+N** (Mac)
3. O histórico atual será limpo

## Configurações

### Aba API

#### Provider
Escolha o provedor de IA:
- **OpenRouter**: Acesso a GPT-4, Claude, e outros
- **Groq**: Inferência rápida, ótimo para começar
- **Google**: Gemini Pro/Ultra
- **Custom**: API própria compatível com OpenAI

#### API Key
Sua chave de API do provedor escolhido.

**Como obter**:
- [OpenRouter](https://openrouter.ai/)
- [Groq](https://console.groq.com/)
- [Google](https://makersuite.google.com/)

#### Model
O modelo de IA a ser usado:
- **GPT-4 Turbo**: Melhor qualidade, mais caro
- **GPT-3.5 Turbo**: Bom custo-benefício
- **Claude 3.5 Sonnet**: Excelente para análise
- **Llama 3.1 70B**: Grátis via Groq, muito bom
- **Gemini Pro**: Gratuito, bom para uso geral

#### Base URL
URL base da API (apenas para custom).

### Aba Dados

#### Auto-refresh Workbooks
Atualiza automaticamente a lista de workbooks abertos.

#### Preview Rows
Número de linhas a mostrar em previews (padrão: 10).

#### Max History Messages
Quantas mensagens manter no histórico de contexto (padrão: 20).

## Dicas e Truques

### 💡 Dica 1: Seja Específico

**Ruim**: "Faça um gráfico"
**Bom**: "Crie um gráfico de barras verticais usando os dados de A1:B10, com título 'Vendas por Mês'"

### 💡 Dica 2: Contexto é Importante

Mencione o workbook e sheet se houver múltiplos abertos:

```
"No workbook Vendas.xlsx, aba Resumo, some a coluna B"
```

### 💡 Dica 3: Comandos em Etapas

Para operações complexas, divida em etapas:

```
1. "Primeiro, filtre os dados onde Status = 'Concluído'"
2. "Agora, calcule a média dos valores filtrados"
3. "Por fim, crie um gráfico com esses dados"
```

### 💡 Dica 4: Use o Histórico

A IA mantém contexto da conversa. Você pode se referir a comandos anteriores:

```
Você: "Some a coluna A"
IA: "A soma é 1250"
Você: "Agora divida esse valor por 10"
```

### 💡 Dica 5: Desfazer é Seu Amigo

Se algo der errado, use o botão **Desfazer** (↩️) ou:
```
"Desfaça a última operação"
```

### 💡 Dica 6: Exploração de Dados

Peça para a IA explorar e sugerir:

```
"Analise esses dados e sugira insights interessantes"
"Que tipo de gráfico seria melhor para visualizar isso?"
"Há alguma anomalia nos dados?"
```

### 💡 Dica 7: Templates

Crie suas próprias templates de comandos frequentes:

```
"Formato padrão de relatório: título em negrito, azul, centralizado, 
dados com bordas, totais em amarelo"
```

Salve isso em uma conversa e reutilize.

### 💡 Dica 8: Atalhos de Teclado

- **Enter**: Enviar mensagem
- **Shift+Enter**: Nova linha
- **Ctrl+N**: Nova conversa
- **Ctrl+S**: Salvar conversa
- **Ctrl+Z**: Desfazer (na interface)
- **Esc**: Cancelar mensagem em streaming

## Resolução de Problemas

### "Excel não está respondendo"

**Problema**: Excel está ocupado ou em modo de edição.

**Solução**:
1. Pressione **ESC** no Excel para sair da edição
2. Feche qualquer diálogo aberto
3. Certifique-se de que nenhuma célula está sendo editada
4. Tente novamente

### "Workbook não encontrado"

**Problema**: Excel-ai não está detectando seu workbook.

**Solução**:
1. Verifique se o Excel está aberto
2. Certifique-se de que o arquivo está salvo (tem um nome)
3. Clique em **"Atualizar"** na interface
4. Se necessário, reinicie o Excel-ai

### "API Key inválida"

**Problema**: Chave de API não está funcionando.

**Solução**:
1. Verifique se copiou a chave corretamente (sem espaços)
2. Confirme que a chave é do provedor correto
3. Teste a chave no site do provedor
4. Gere uma nova chave se necessário

### "Resposta muito lenta"

**Problema**: IA demora para responder.

**Solução**:
1. Experimente um modelo mais rápido (ex: Groq)
2. Verifique sua conexão de internet
3. Reduza o contexto (menos mensagens no histórico)
4. Simplifique sua pergunta

### "Gráfico não foi criado"

**Problema**: Comando não executou a ação esperada.

**Solução**:
1. Seja mais específico sobre os dados
2. Verifique se o range existe
3. Tente pedir de forma diferente
4. Use "Desfazer" e tente novamente

### "Erro ao salvar conversa"

**Problema**: Não consegue salvar a conversa.

**Solução**:
1. Verifique permissões de escrita
2. Certifique-se de que há espaço em disco
3. Tente com um título diferente
4. Reinicie a aplicação se persistir

## Exemplos Práticos

### Exemplo 1: Análise de Vendas

```
Usuário: "Olá! Tenho uma planilha de vendas. Pode me ajudar a analisá-la?"
IA: "Claro! Posso ver que você tem dados de vendas. O que gostaria de saber?"
Usuário: "Qual foi o total de vendas no mês?"
IA: "O total de vendas foi R$ 125.450,00"
Usuário: "E qual produto vendeu mais?"
IA: "O produto mais vendido foi 'Notebook Dell', com 45 unidades"
Usuário: "Crie um gráfico mostrando as vendas por produto"
IA: "Gráfico criado! Adicionei um gráfico de barras na sua planilha"
```

### Exemplo 2: Formatação de Relatório

```
Usuário: "Preciso formatar um relatório"
IA: "Como gostaria de formatar?"
Usuário: "Linha 1 deve ser cabeçalho: negrito, fundo azul, texto branco"
IA: "Formatação aplicada à linha 1"
Usuário: "Agora, adicione bordas em toda a tabela de A1 a F20"
IA: "Bordas adicionadas"
Usuário: "Perfeito! Formate os valores da coluna E como moeda"
IA: "Coluna E formatada como R$"
```

### Exemplo 3: Criação de Dashboard

```
Usuário: "Quero criar um dashboard"
IA: "Ótimo! Que visualizações você precisa?"
Usuário: "Um gráfico de pizza com distribuição por categoria, 
         um gráfico de linhas mostrando tendência mensal, 
         e um resumo com totais"
IA: "Vou criar isso para você..."
[IA cria os gráficos e tabela de resumo]
IA: "Dashboard criado! Adicionei 2 gráficos e uma tabela de resumo"
```

## Melhores Práticas

### ✅ Faça

- Seja claro e específico nos comandos
- Mantenha o Excel aberto e responsivo
- Salve conversas importantes
- Use desfazer quando necessário
- Experimente diferentes formas de pedir

### ❌ Não Faça

- Editar células manualmente enquanto a IA está trabalhando
- Fechar o Excel durante operações
- Usar comandos ambíguos
- Esperar que a IA "adivinhe" dados não visíveis
- Ignorar mensagens de erro

## Glossário

- **Workbook**: Arquivo do Excel (.xlsx)
- **Sheet/Aba**: Planilha dentro de um workbook
- **Range**: Intervalo de células (ex: A1:B10)
- **Célula**: Interseção de linha e coluna (ex: A1)
- **Streaming**: Resposta da IA em tempo real
- **Contexto**: Histórico de mensagens mantido pela IA

## Recursos Adicionais

- [Instalação](INSTALLATION.md)
- [Configuração](CONFIGURATION.md)
- [API Documentation](API.md)

## Suporte

Encontrou um problema ou tem uma sugestão?

1. Verifique este guia primeiro
2. Consulte [Resolução de Problemas](#resolução-de-problemas)
3. Abra uma issue no [GitHub](https://github.com/sshturbo/Excel-ai/issues)

---

**Divirta-se usando o Excel-ai! 🚀**
