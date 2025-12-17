// AI Response Processing Service
// Handles parsing and formatting of AI responses.
// Native function calling system - simplified.

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
 * Native function calling - simplified processing
 */
export async function processAIResponse(
    response: string,
    options: ProcessAIResponseOptions
): Promise<AIProcessingResult> {

    // Detectar ações executadas (contar ✅)
    const successMatches = response.match(/✅/g)
    const actionsExecuted = successMatches ? successMatches.length : 0

    // Limpar resposta para display (remover marcadores antigos se existirem)
    let displayContent = response
        .replace(/:::excel-query\s*[\s\S]*?\s*:::/g, '')
        .replace(/:::excel-action\s*[\s\S]*?\s*:::/g, '')
        .replace(/:::agent-paused:::/g, '')
        .trim()

    return {
        displayContent,
        actionsExecuted
    }
}

/**
 * Extracts actions from AI response (legacy - returns empty)
 */
export function extractActionsFromResponse(_response: string): ExcelAction[] {
    // Com function calling nativo, não há mais blocos :::excel-action:::
    return []
}

/**
 * Processes query blocks (legacy - returns empty)
 */
export async function processQueryBlocks(_response: string): Promise<string[]> {
    // Com function calling nativo, queries são executadas pelo backend
    return []
}
