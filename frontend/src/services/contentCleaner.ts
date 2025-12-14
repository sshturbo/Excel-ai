// Content cleaning utilities for AI responses
// Extracted from App.tsx cleanTechnicalBlocks function

/**
 * Removes technical blocks from AI response content
 * @param content Raw AI response content
 * @returns Cleaned content without excel-action and excel-query blocks
 */
export function cleanTechnicalBlocks(content: string): string {
    // Remove :::excel-action blocks completely
    let cleaned = content.replace(/:::excel-action\s*[\s\S]*?\s*:::/g, '')
    // Remove :::excel-query blocks completely
    cleaned = cleaned.replace(/:::excel-query\s*[\s\S]*?\s*:::/g, '')
    // Remove multiple empty lines
    cleaned = cleaned.replace(/\n{3,}/g, '\n\n')
    return cleaned.trim()
}

/**
 * Processes streaming content, hiding incomplete technical blocks
 * @param rawContent Raw streaming content
 * @returns Object with cleaned content and status flags
 */
export function processStreamingContent(rawContent: string): {
    cleanContent: string
    hasIncompleteAction: boolean
    hasIncompleteQuery: boolean
} {
    let cleanContent = rawContent

    // Remove complete excel-action blocks
    cleanContent = cleanContent.replace(/:::excel-action\s*[\s\S]*?\s*:::/g, '')
    // Remove complete excel-query blocks
    cleanContent = cleanContent.replace(/:::excel-query\s*[\s\S]*?\s*:::/g, '')

    // Check for incomplete technical blocks at the end (NOT thinking blocks)
    const incompleteActionMatch = cleanContent.match(/:::excel-action[\s\S]*$/)
    const incompleteQueryMatch = cleanContent.match(/:::excel-query[\s\S]*$/)

    if (incompleteActionMatch) {
        cleanContent = cleanContent.replace(/:::excel-action[\s\S]*$/, '')
    }
    if (incompleteQueryMatch) {
        cleanContent = cleanContent.replace(/:::excel-query[\s\S]*$/, '')
    }

    cleanContent = cleanContent.replace(/\n{3,}/g, '\n\n').trim()

    const hasIncompleteAction = /:::excel-action(?![\s\S]*:::)/.test(rawContent)
    const hasIncompleteQuery = /:::excel-query(?![\s\S]*:::)/.test(rawContent)

    return {
        cleanContent,
        hasIncompleteAction,
        hasIncompleteQuery
    }
}

/**
 * Cleans assistant message content when loading from conversation history
 * @param content Raw assistant message content
 * @returns Cleaned content
 */
export function cleanAssistantContent(content: string): string {
    // Remove technical blocks
    let cleaned = content.replace(/:::excel-action\s*[\s\S]*?\s*:::/g, '')
    cleaned = cleaned.replace(/:::excel-query\s*[\s\S]*?\s*:::/g, '')
    cleaned = cleaned.replace(/\n{3,}/g, '\n\n').trim()
    return cleaned
}

/**
 * Extracts the user's actual question from a formatted context message
 * @param content Message content with context
 * @returns The user's actual question or null
 */
export function extractUserQuestion(content: string): string | null {
    if (content.includes('Contexto do Excel:') && content.includes('Pergunta do usuário:')) {
        const match = content.match(/Pergunta do usuário:\s*([\s\S]+)$/)
        if (match) {
            return match[1].trim()
        }
    }
    return null
}

/**
 * Checks if a user message is an internal system message that should be hidden
 * @param content User message content
 * @returns True if the message should be hidden
 */
export function isInternalUserMessage(content: string): boolean {
    return (
        content.startsWith('Resultados das queries:') ||
        content.startsWith('[ERRO na ação') ||
        content.startsWith('A ação anterior falhou')
    )
}
