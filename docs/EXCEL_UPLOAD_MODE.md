# Modo de Importação de Arquivos Excel

Visão geral do modo de upload de arquivos integrado que permite trabalhar com arquivos Excel sem precisar ter o Excel instalado.

## Como Usar

### 1. Carregar um Arquivo Excel
**Use o botão "+" no ChatInput (recomendado):**

1. Clique no botão de "+" no canto inferior esquerdo do input de chat
2. Selecione um arquivo `.xlsx` do seu computador
3. O arquivo será carregado automaticamente
4. **O arquivo aparecerá na sidebar** como se fosse um workbook do Excel
5. **Você pode interagir com o chat normalmente** usando os dados do arquivo

**Alternativa: Toggle no Header**
1. No topo da aplicação, clique no botão com ícone de upload (📤) ao lado do botão de tema
2. Arraste ou selecione o arquivo `.xlsx` na área de upload
3. O sistema carregará e mostrará o arquivo na sidebar

### 2. Interface Unificada
Após carregar o arquivo, você terá **a mesma interface do modo COM**:

- **Sidebar**: O arquivo aparece na lista de workbooks
- **Planilhas**: Clique na planilha para ver seus dados
- **Toolbar**: Mostra visualizações de dados e gráficos
- **Chat**: Converse com a IA para manipular os dados
- **Preview**: Visualize os dados da planilha selecionada

### 3. Fluxo de Trabalho Integrado

```
1. Clique no botão "+" no chat
2. Selecione arquivo "relatorio_vendas.xlsx"
3. Sistema carrega e mostra na sidebar:
   - Workbook: relatorio_vendas.xlsx
   - Planilha: "Vendas 2024"
   - Planilha: "Resumo"
4. Clique em "Vendas 2024"
5. Sistema mostra preview dos dados
6. Peça à IA: "Mostre os top 10 produtos por vendas"
7. IA processa e aplica filtros
8. Baixe o arquivo modificado ou continue interagindo
```

## Diferenças Entre os Modos

### Modo COM (Excel Instalado) 💻
- Requer Microsoft Excel instalado
- Conecta ao Excel em tempo real
- Alterações aplicadas diretamente no Excel aberto
- Melhor para uso interativo contínuo

### Modo Upload (Sem Excel) 📤
- **Não requer Excel instalado**
- Usa biblioteca Excelize para manipulação
- Trabalha com arquivos `.xlsx` carregados
- **Interface idêntica ao modo COM**
- Mesma sidebar, toolbar e chat
- Download do arquivo modificado ao final

## Funcionalidades Disponíveis

### Atualmente ✅
- ✅ Upload de arquivos `.xlsx` via botão "+" no chat
- ✅ Arquivo aparece na sidebar como workbook (mesma aparência do modo COM)
- ✅ Preview da estrutura do arquivo (planilhas, dimensões)
- ✅ Visualização dos dados de cada planilha
- ✅ **Interface unificada** - mesma sidebar, toolbar, chat
- ✅ **Interação completa com o chat** usando dados do arquivo
- ✅ Download do arquivo modificado
- ✅ Gerenciamento de sessões

### Em Desenvolvimento 🚧
- 🚧 Integração completa com IA para modificar arquivos carregados
- 🚧 Aplicação de transformações e fórmulas via chat
- 🚧 Histórico de modificações
- 🚧 Undo/Redo de alterações

## Interface Visual

### Sidebar (Idêntica ao Modo COM)
```
┌─────────────────────────┐
│ 📁 relatorio.xlsx     │ ← Arquivo carregado
│   └─ Vendas 2024      │
│   └─ Resumo           │
│                        │
│ 💬 Conversas           │
│   └─ Conversa 1       │
│   └─ Conversa 2       │
└─────────────────────────┘
```

### Toolbar (Mesmas Funções)
- 🔍 **Preview**: Ver dados da planilha
- 📊 **Gráfico**: Ver visualização gráfica
- 🔄 **Refresh**: Recarregar dados

### Chat (Mesma Experiência)
- Input no final da tela
- Histórico de mensagens
- Respostas da IA
- Sugestões de ações

## Comando de Chat Exemplos

```
"Mostre as primeiras 10 linhas"
"Calcule o total da coluna B"
"Filtre as linhas onde coluna A > 100"
"Crie um gráfico com estes dados"
"Exporte esta tabela para CSV"
```

## Limitações Atuais

1. **Apenas arquivos .xlsx**: Suporta apenas o formato moderno do Excel
2. **Sem fórmulas dinâmicas**: Fórmulas são avaliadas no upload
3. **Tamanho de arquivo**: Arquivos muito grandes podem ter performance reduzida
4. **Macros e VBA**: Não suportados neste modo

## Troubleshooting

### Arquivo não aparece na sidebar
- Verifique se o arquivo é `.xlsx` válido
- Aguarde o processamento completar
- Verifique o console para erros

### Preview não mostra dados
- Clique na planilha na sidebar
- Aguarde o carregamento dos dados
- Tente recarregar a página

### Download não funciona
- Verifique permissões do navegador
- Desative bloqueadores de pop-up
- Tente usar outro navegador

### Não consigo interagir com o chat
- Verifique se a API key está configurada
- Verifique a conexão com o backend
- Recarregue a página

## Notas Técnicas

### Arquitetura
- **Backend**: Go com biblioteca `excelize`
- **Frontend**: React + TypeScript
- **Comunicação**: Wails bindings
- **Integração**: Hooks unificados `useExcelUpload` e `useExcelConnection`

### Estado Unificado
- Um único estado de aplicação controla ambos os modos
- `isUploadMode`: boolean indica qual modo está ativo
- Componentes compartilhados para ambos os modos

### Performance
- Upload até 10MB: < 2 segundos
- Arquivos até 50.000 linhas: Aceitável
- Preview carrega até 100 linhas por vez

## Comparação com Modo COM

| Característica | Modo COM | Modo Upload |
|---------------|-----------|-------------|
| Requer Excel | Sim | Não |
| Conexão | Tempo real | Arquivo estático |
| Sidebar | Identical | Identical |
| Toolbar | Identical | Identical |
| Chat | Identical | Identical |
| Download | Não necessário | Sim |
| Fórmulas | Ativas | Estáticas |
| Macros | Suportado | Não |

## Próximos Passos

1. **Integração completa com IA**: Modificar arquivos via chat
2. **Transformações avançadas**: Filtros, ordenação, agrupamento
3. **Visualizações**: Gráficos e dashboards dinâmicos
4. **Exportação multi-formato**: CSV, PDF, JSON
5. **Comparação de versões**: Diferenças entre arquivos

## Feedback

Se encontrar problemas ou tiver sugestões:
- Abra uma issue no GitHub
- Entre em contato com a equipe de desenvolvimento
- Envie feedback através do botão de feedback na aplicação
