# Integração do Z.AI no Excel-AI

## Visão Geral

O Z.AI (智谱AI/GLM) foi integrado ao Excel-Ai como um provedor de IA nativo, permitindo o uso dos modelos GLM (General Language Model).

## Características

- **API Compatível com OpenAI**: O z.ai usa um endpoint compatível com OpenAI, facilitando a integração
- **Modelos GLM Disponíveis**:
  - **GLM-4.7**: Modelo flagship otimizado para coding e function calling
  - **GLM-4.6V**: Modelo multimodal com visão avançada
  - **GLM-4.6**: Modelo versátil com function calling nativo
  - **GLM-4.5**: Modelo econômico com suporte a function calling
  - **GLM-4.5 Air**: Modelo leve para respostas rápidas

- **Suporte a Function Calling**: Todos os modelos GLM têm suporte nativo a function calling
- **Contexto de 128K tokens**: Capacidade de contexto extensa

## Configuração

### 1. Obter API Key do Z.AI

1. Acesse: https://open.bigmodel.cn/developercenter/balance
2. Faça login ou crie uma conta
3. Gere uma nova API Key
4. Copie a chave para uso no Excel-AI

### 2. Configurar no Excel-AI

1. Abra as configurações do Excel-Ai
2. No campo "Provedor", selecione **"Z.AI (GLM Models)"**
3. Cole sua API Key no campo "API Key"
4. Clique em "Carregar Modelos" para ver os modelos disponíveis
5. Selecione o modelo desejado (recomendado: `glm-4.7` para coding)

### 3. Base URL

A URL base do Z.AI é configurada automaticamente como:
```
https://api.z.ai/api/paas/v4
```

## Modelos Disponíveis

| Modelo | Descrição | Contexto | Melhor Uso |
|--------|-----------|-----------|------------|
| `glm-4.7` | Modelo flagship - otimizado para coding | 128K | Programação e function calling |
| `glm-4.6v` | Modelo multimodal com visão | 128K | Análise de imagens e código |
| `glm-4.6` | Modelo versátil e equilibrado | 128K | Uso geral com function calling |
| `glm-4.5` | Modelo econômico | 128K | Tarefas gerais com custo reduzido |
| `glm-4.5-air` | Modelo ultraleve | 128K | Respostas rápidas |

## Preços

Os preços são em RMB (Yuan Chinês):
- GLM-4.7: ¥2.5/1M tokens (input), ¥10/1M tokens (output)
- GLM-4.6V: ¥2/1M tokens (input), ¥8/1M tokens (output)
- GLM-4.6: ¥1.5/1M tokens (input), ¥6/1M tokens (output)
- GLM-4.5: ¥1/1M tokens (input), ¥4/1M tokens (output)
- GLM-4.5 Air: ¥0.5/1M tokens (input), ¥2/1M tokens (output)

## Uso com Excel-AI

### Function Calling

O Z.AI tem suporte nativo a function calling, permitindo:
- **Query Data**: Consultar dados do Excel
- **List Sheets**: Listar planilhas disponíveis
- **Execute Macro**: Executar macros do Excel
- **Chart Operations**: Criar e modificar gráficos

### Exemplo de Uso

1. Conecte-se a um arquivo Excel
2. Selecione o provedor Z.AI e o modelo GLM-4.7
3. Faça perguntas como:
   - "Análise os dados da coluna A"
   - "Crie um gráfico de barras com estes dados"
   - "Filtra as vendas por região"
   - "Execute a macro 'RelatorioVendas'"

## Troubleshooting

### Erro de Autenticação
- Verifique se a API Key está correta
- Confirme que a conta tem saldo disponível
- Acesse: https://open.bigmodel.cn/developercenter/balance

### Modelos Não Aparecendo
- Verifique a conexão com a internet
- Confirme que a API Key tem permissão para acessar os modelos
- Tente recarregar os modelos nas configurações

### Timeout nas Requisições
- GLM-4.7 pode ser mais lento para tarefas complexas
- Tente GLM-4.6 ou GLM-4.5 para respostas mais rápidas
- Verifique se o firewall não está bloqueando conexões para api.z.ai

## Documentação Oficial

- **Documentação Principal**: https://docs.z.ai/
- **Quick Start**: https://docs.z.ai/guides/overview/quick-start
- **Function Calling**: https://docs.z.ai/guides/capabilities/function-calling
- **GLM-4.7**: https://docs.z.ai/guides/llm/glm-4.7
- **OpenAI Compatible**: https://docs.z.ai/guides/develop/openai/python

## Comparação com Outros Provedores

| Característica | Z.AI | OpenRouter | Groq | Ollama |
|--------------|------|-----------|------|--------|
| Custo | ¥ (RMB) | USD | Gratuito | Gratuito |
| Function Calling | ✅ Nativo | ✅ | ✅ | ✅ (Modelos específicos) |
| Contexto | 128K | Varia | 128K | Varia |
| Velocidade | Alta | Alta | Muito Alta | Local |
| Privacidade | Cloud | Cloud | Cloud | Local |

## Suporte

Para problemas específicos do Z.AI:
- Fórum da comunidade: https://open.bigmodel.cn/
- Documentação: https://docs.z.ai/
- Status da API: https://status.z.ai/

Para problemas de integração com o Excel-Ai:
- Verifique os logs do aplicativo
- Teste a API Key usando a documentação oficial do Z.AI
- Confirme que a URL base está correta: `https://api.z.ai/api/paas/v4`