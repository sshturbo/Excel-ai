// AI Response Processing Service
// Handles parsing and formatting of AI responses.
// Updated for native function calling system.

import type { AIProcessingResult, ExcelAction, Workbook } from '@/types'

interface ProcessAIResponseOptions {
    askBeforeApply: boolean
    maxRetries?: number
    onWorkbooksUpdate?: (workbooks: Workbook[]) => void
    onPendingAction?: (action: ExcelAction) => void
    onMessageUpdate?: (content: string) => void
    onStreamingUpdate?: (content: string) => void
}

/**
 * Processes AI response for display
 * Native function calling - no more :::excel-*::: blocks to parse
 */
export async function processAIResponse(
    response: string,
    options: ProcessAIResponseOptions
): Promise<AIProcessingResult> {

    // Detectar ação pendente via marcador do backend
    const pendingActionMatch = response.match(/🛑 \*\[Ação Pendente: ([^\]]+)\]\*/)
    const hasPendingAction = pendingActionMatch !== null

    // Detectar ações executadas (contar ✅)
    const successMatches = response.match(/✅/g)
    const actionsExecuted = successMatches ? successMatches.length : 0

    // Detectar se agente pausou
    const agentPaused = response.includes(':::agent-paused:::') ||
        response.includes('Ação Pendente')

    // Se há ação pendente e callback fornecido, notificar
    if (hasPendingAction && options.onPendingAction && pendingActionMatch) {
        const toolName = pendingActionMatch[1]
        // Criar ação pendente simplificada para o frontend
        const pendingAction: ExcelAction = {
            op: toolName,
            _pending: true
        } as any
        options.onPendingAction(pendingAction)
    }

    // Limpar resposta para display
    let displayContent = response

    // Remover marcadores antigos se ainda existirem (compatibilidade)
    displayContent = displayContent.replace(/:::excel-query\s*[\s\S]*?\s*:::/g, '\n*[Consultando Excel...]*\n')
    displayContent = displayContent.replace(/:::excel-action\s*[\s\S]*?\s*:::/g, '\n*[Executando Ação...]*\n')
    displayContent = displayContent.replace(/:::agent-paused:::/g, '')

    return {
        displayContent: displayContent.trim(),
        actionsExecuted,
        agentPaused
    }
}

/**
 * Extracts actions from AI response (legacy support)
 */
export function extractActionsFromResponse(response: string): ExcelAction[] {
    // Com function calling nativo, não há mais blocos :::excel-action:::
    // Manter para compatibilidade, mas retornar vazio
    const actionRegex = /:::excel-action\s*([\s\S]*?)\s*:::/g
    const matches = [...response.matchAll(actionRegex)]
    const actions: ExcelAction[] = []

    for (const match of matches) {
        try {
            const jsonStr = match[1]
            const cleanJson = jsonStr.replace(/```json/g, '').replace(/```/g, '').trim()
            const action: ExcelAction = JSON.parse(cleanJson)
            actions.push(action)
        } catch (e) {
            console.error("Erro ao parsear ação:", e)
        }
    }

    return actions
}

/**
 * Processes query blocks (legacy support)
 */
export async function processQueryBlocks(response: string): Promise<string[]> {
    // Com function calling nativo, queries são executadas pelo backend
    return []
}
