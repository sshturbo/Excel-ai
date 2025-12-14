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

    // Backend takes care of everything.
    // We just need to clean the response for display if desired.
    // Ideally we show "Executing..." placeholders, but for now let's just show text.

    const actionRegex = /:::excel-action\s*([\s\S]*?)\s*:::/g
    const queryRegex = /:::excel-query\s*([\s\S]*?)\s*:::/g

    // Remove blocks from display content to keep UI clean
    // (Or keep them if we want transparency - User preference)
    // Let's hide them by default as they are verbose JSONs.
    // But maybe replace with a badge [Action executed]

    let displayContent = response

    displayContent = displayContent.replace(queryRegex, '\n*[Consultando Excel...]*\n')
    displayContent = displayContent.replace(actionRegex, '\n*[Executando Ação...]*\n')

    return { displayContent: displayContent.trim(), actionsExecuted: 0 }
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
