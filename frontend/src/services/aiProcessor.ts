// AI Response Processing Service
// Handles parsing and execution of AI responses with excel-action and excel-query blocks

import { toast } from 'sonner'
import type { ExcelAction, AIProcessingResult, Workbook } from '@/types'
import { executeExcelAction } from './excelActions'

import {
    QueryExcel,
    SendErrorFeedback,
    StartUndoBatch,
    EndUndoBatch
} from "../../wailsjs/go/app/App"

interface ProcessAIResponseOptions {
    askBeforeApply: boolean
    maxRetries?: number
    onWorkbooksUpdate?: (workbooks: Workbook[]) => void
    onPendingAction?: (action: ExcelAction) => void
    onMessageUpdate?: (content: string) => void
    onStreamingUpdate?: (content: string) => void
}

/**
 * Processes excel-query blocks and returns results
 * @param response AI response containing excel-query blocks
 * @returns Array of query result strings
 */
export async function processQueryBlocks(response: string): Promise<string[]> {
    const queryRegex = /:::excel-query\s*([\s\S]*?)\s*:::/g
    const queryMatches = [...response.matchAll(queryRegex)]
    const queryResults: string[] = []

    for (const match of queryMatches) {
        try {
            const jsonStr = match[1]
            // Support multiple queries in the same block
            const lines = jsonStr.split('\n').filter(l => l.trim().startsWith('{'))

            for (const line of lines) {
                const query = JSON.parse(line.trim())
                console.log('[QUERY]', query)

                const result = await QueryExcel(query.type, {
                    name: query.name || '',
                    sheet: query.sheet || '',
                    range: query.range || ''
                })

                if (result.success) {
                    queryResults.push(`Query "${query.type}": ${JSON.stringify(result.data)}`)
                } else {
                    queryResults.push(`Query "${query.type}" falhou: ${result.error}`)
                }
            }
        } catch (err) {
            console.error('Erro ao processar query:', err)
        }
    }

    if (queryResults.length > 0) {
        console.log('[QUERY RESULTS]', queryResults)
    }

    return queryResults
}

/**
 * Processes AI response and executes excel-action blocks
 * @param response Raw AI response
 * @param options Processing options
 * @returns Processing result with display content and action count
 */
export async function processAIResponse(
    response: string,
    options: ProcessAIResponseOptions
): Promise<AIProcessingResult> {
    const {
        askBeforeApply,
        maxRetries = 2,
        onWorkbooksUpdate,
        onPendingAction,
        onMessageUpdate,
        onStreamingUpdate
    } = options

    // Process queries first
    await processQueryBlocks(response)

    const actionRegex = /:::excel-action\s*([\s\S]*?)\s*:::/g
    let matches = [...response.matchAll(actionRegex)]
    let actionsExecuted = 0
    let currentResponse = response
    let retryCount = 0
    let undoBatchStarted = false

    while (matches.length > 0 && retryCount <= maxRetries) {
        if (!askBeforeApply && retryCount === 0) {
            await StartUndoBatch()
            undoBatchStarted = true
        }

        let hasError = false
        let errorMessage = ''

        for (const match of matches) {
            try {
                const jsonStr = match[1]
                const cleanJson = jsonStr.replace(/```json/g, '').replace(/```/g, '').trim()
                const action: ExcelAction = JSON.parse(cleanJson)

                if (askBeforeApply) {
                    if (onPendingAction) {
                        onPendingAction(action)
                    }
                } else {
                    const result = await executeExcelAction(action, onWorkbooksUpdate)
                    if (result.success) {
                        actionsExecuted++
                    } else {
                        hasError = true
                        errorMessage = result.error || 'Erro desconhecido'
                        toast.warning(`Erro: ${errorMessage}. Solicitando correção...`)
                        break // Stop at first error to request correction
                    }
                }
            } catch (e: unknown) {
                console.error("Erro ao parsear ação", e)
                hasError = true
                errorMessage = `Erro ao processar comando: ${e instanceof Error ? e.message : String(e)}`
                break
            }
        }

        // If there was an error, send feedback to AI
        if (hasError && retryCount < maxRetries) {
            retryCount++
            console.log(`[DEBUG] Enviando erro para IA (tentativa ${retryCount}):`, errorMessage)

            if (onStreamingUpdate) {
                onStreamingUpdate('')
            }
            toast.info(`Solicitando correção à IA (tentativa ${retryCount})...`)

            try {
                const correctedResponse = await SendErrorFeedback(errorMessage)
                currentResponse = correctedResponse
                matches = [...correctedResponse.matchAll(actionRegex)]

                // Update message with new response
                const newDisplayContent = correctedResponse.replace(actionRegex, '').trim()
                if (onMessageUpdate) {
                    onMessageUpdate(newDisplayContent)
                }
            } catch (feedbackErr) {
                console.error("Erro ao enviar feedback:", feedbackErr)
                const msg = feedbackErr instanceof Error ? feedbackErr.message : String(feedbackErr)
                toast.error('Erro ao solicitar correção à IA: ' + msg)
                break
            }
        } else {
            break
        }
    }

    // Always finalize the batch if it was started
    if (!askBeforeApply && undoBatchStarted) {
        try {
            await EndUndoBatch()
        } catch (e) {
            console.warn('Falha ao finalizar lote de Undo:', e)
        }
    }

    const displayContent = currentResponse.replace(actionRegex, '').trim()
    return { displayContent, actionsExecuted }
}

/**
 * Extracts actions from AI response without executing them
 * @param response AI response containing excel-action blocks
 * @returns Array of parsed actions
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
