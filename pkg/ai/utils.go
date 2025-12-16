package ai

import (
	"fmt"
)

// EstimateTokens estima o número de tokens em uma string (aproximação simples)
// Regra geral: ~4 caracteres por token, mas usamos 3 para ser conservador e evitar 413
func EstimateTokens(text string) int {
	return len(text) / 3
}

// PruneMessages remove mensagens antigas para manter o contexto dentro do limite de tokens
// Mantém sempre a mensagem de sistema (se existir e for a primeira)
// Mantém sempre a última mensagem (user input)
func PruneMessages(messages []Message, maxInputTokens int) []Message {
	if len(messages) == 0 {
		return messages
	}

	// Se limite não definido ou muito baixo, usar padrão seguro
	if maxInputTokens <= 0 {
		maxInputTokens = 4000
	}

	totalTokens := 0
	for _, m := range messages {
		totalTokens += EstimateTokens(m.Content)
	}

	// Se estiver dentro do limite, retorna original
	if totalTokens <= maxInputTokens {
		return messages
	}

	fmt.Printf("[Prune] Total tokens (%d) exceeds limit (%d). Pruning...\n", totalTokens, maxInputTokens)

	var pruned []Message
	var systemMessage *Message

	// Identificar e separar System Prompt (geralmente o primeiro)
	startIndex := 0
	if messages[0].Role == "system" {
		systemMessage = &messages[0]
		totalTokens -= EstimateTokens(systemMessage.Content) // Remover contagem temporariamente para recalcular
		startIndex = 1
	}

	// Identificar a última mensagem (que deve ser preservada)
	lastMessage := messages[len(messages)-1]
	lastMsgTokens := EstimateTokens(lastMessage.Content)

	// Tokens disponíveis para histórico (menos system e last msg)
	availableTokens := maxInputTokens - lastMsgTokens
	if systemMessage != nil {
		availableTokens -= EstimateTokens(systemMessage.Content)
	}

	// Se o system + last message já estourarem o limite,
	// retornamos apenas eles (é o melhor que podemos fazer)
	if availableTokens < 0 {
		result := []Message{}
		if systemMessage != nil {
			result = append(result, *systemMessage)
		}
		result = append(result, lastMessage)
		return result
	}

	// Coletar mensagens do final para o início até encher o bucket
	var history []Message
	currentTokens := 0

	// Iterar de trás para frente, pulando a última (já salva) e parando antes do system
	for i := len(messages) - 2; i >= startIndex; i-- {
		msg := messages[i]
		tokens := EstimateTokens(msg.Content)

		if currentTokens+tokens > availableTokens {
			break // Não cabe mais nada
		}

		history = append([]Message{msg}, history...) // Prepend
		currentTokens += tokens
	}

	// Reconstruir lista final
	if systemMessage != nil {
		pruned = append(pruned, *systemMessage)
	}
	pruned = append(pruned, history...)
	pruned = append(pruned, lastMessage)

	fmt.Printf("[Prune] Pruned from %d to %d messages\n", len(messages), len(pruned))
	return pruned
}
