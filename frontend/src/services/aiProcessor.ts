// AI Response Processing Service
// Handles parsing and formatting of AI responses.
// Execution logic has been moved to Backend (Autonomous Agent).

import type { AIProcessingResult, ExcelAction, Workbook } from '@/types'

// Removed imports related to execution (QueryExcel, SendErrorFeedback, etc) to prevent client-side execution.

interface ProcessAIResponseOptions {
    askBeforeApply: boolean
    maxRetries?: number
    onWorkbooksUpdate?: (workbooks: Workbook[]) => void
    onPendingAction?: (action: ExcelAction) => void
    onMessageUpdate?: (content: string) => void
    onStreamingUpdate?: (content: string) => void
}

/**
 * Processes excel-query blocks (now handled by backend, this acts as cleaner/logger)
 */
export async function processQueryBlocks(response: string): Promise<string[]> {
    // Backend handles queries autonomously. We just parse for debug/display if needed.
    const queryRegex = /:::excel-query\s*([\s\S]*?)\s*:::/g
    const queryResults: string[] = []

    // Log queries found for debugging
    const matches = [...response.matchAll(queryRegex)]
    if (matches.length > 0) {
        console.log('[Frontend] Queries detected (executed by backend):', matches.length)
    }

    return queryResults
}

/**
 * Processes AI response (Display Only)
 * Execution is now Server-Side.
 */
export async function processAIResponse(
    response: string,
    options: ProcessAIResponseOptions
): Promise<AIProcessingResult> {

    // We just need to clean the response for display if desired.
    // Ideally we show "Executing..." placeholders, but for now let's just show text.

    const actionRegex = /:::excel-action\s*([\s\S]*?)\s*:::/g
    const queryRegex = /:::excel-query\s*([\s\S]*?)\s*:::/g

    // Extract actions
    const actions = extractActionsFromResponse(response)

    // If we found actions...
    if (actions.length > 0) {
        // ...and we are in "Ask Before Apply" mode (which implies backend paused and returned them)
        if (options.askBeforeApply && options.onPendingAction) {
            console.log('[aiProcessor] Pending actions detected:', actions.length)
            actions.forEach(action => options.onPendingAction!(action))
        } else {
            // Autonomous mode: Backend (likely) executed them, or we are just displaying history.
            // If this is a live response in autonomous mode, actionsExecuted should be > 0 (handled by backend return?)
            // Actually, backend now returns execution logs.
        }
    }

    let displayContent = response

    // Hide the raw JSON blocks from display
    displayContent = displayContent.replace(queryRegex, '\n*[Consultando Excel...]*\n')

    // If pending, show a standard message. If executed, show Executing.
    // For simplicity, we just hide the JSON. The UI banner will handle the "Pending" notification.
    displayContent = displayContent.replace(actionRegex, options.askBeforeApply ? '\n*(Ação aguardando aprovação)*\n' : '\n*[Executando Ação...]*\n')

    // Detectar se o agente pausou por limite de passos
    const agentPaused = displayContent.includes(':::agent-paused:::')
    displayContent = displayContent.replace(/:::agent-paused:::/g, '')

    return { displayContent: displayContent.trim(), actionsExecuted: actions.length, agentPaused }
}

/**
 * Extracts actions from AI response without executing them
 */
export function extractActionsFromResponse(response: string): ExcelAction[] {
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
