// Service to export conversations in various formats
import type { Message } from '@/types'

export interface ExportFormat {
    type: 'markdown' | 'json' | 'txt'
    filename: string
    content: string
}

/**
 * Export conversation to Markdown format
 */
export function exportToMarkdown(
    messages: Message[],
    title: string = 'Conversation'
): ExportFormat {
    let markdown = `# ${title}\n\n`
    markdown += `**Exportado em:** ${new Date().toLocaleString('pt-BR')}\n\n`
    markdown += `---\n\n`

    messages.forEach((msg, idx) => {
        if (msg.hidden || msg.role === 'system') return

        const roleLabel = msg.role === 'user' ? '👤 Usuário' : '🤖 IA'
        markdown += `## ${roleLabel}\n\n`

        // Remove reasoning tags from markdown export
        const cleanContent = msg.content
            .replace(/:::reasoning:::[\s\S]*?:::\/reasoning:::/g, '')
            .replace(/Resultados das ferramentas[\s\S]*?$/g, '')
            .trim()

        if (cleanContent) {
            markdown += `${cleanContent}\n\n`
        }

        if (msg.hasActions) {
            markdown += `✅ *Esta mensagem executou ações no Excel*\n\n`
        }

        markdown += `---\n\n`
    })

    return {
        type: 'markdown',
        filename: `${sanitizeFilename(title)}.md`,
        content: markdown
    }
}

/**
 * Export conversation to JSON format
 */
export function exportToJSON(
    messages: Message[],
    title: string = 'Conversation'
): ExportFormat {
    const data = {
        metadata: {
            title,
            exportedAt: new Date().toISOString(),
            messageCount: messages.filter(m => !m.hidden && m.role !== 'system').length
        },
        messages: messages
            .filter(m => !m.hidden && m.role !== 'system')
            .map(msg => ({
                role: msg.role,
                content: msg.content,
                hasActions: msg.hasActions,
                timestamp: new Date().toISOString()
            }))
    }

    return {
        type: 'json',
        filename: `${sanitizeFilename(title)}.json`,
        content: JSON.stringify(data, null, 2)
    }
}

/**
 * Export conversation to plain text format
 */
export function exportToText(
    messages: Message[],
    title: string = 'Conversation'
): ExportFormat {
    let text = `=${title}=\n\n`
    text += `Exportado em: ${new Date().toLocaleString('pt-BR')}\n`
    text += `${'='.repeat(title.length)}\n\n`

    messages.forEach((msg, idx) => {
        if (msg.hidden || msg.role === 'system') return

        const roleLabel = msg.role === 'user' ? 'USUÁRIO:' : 'IA:'
        text += `[${roleLabel}]\n\n`

        // Remove reasoning tags from text export
        const cleanContent = msg.content
            .replace(/:::reasoning:::[\s\S]*?:::\/reasoning:::/g, '')
            .replace(/Resultados das ferramentas[\s\S]*?$/g, '')
            .trim()

        if (cleanContent) {
            text += `${cleanContent}\n\n`
        }

        if (msg.hasActions) {
            text += `[✅ Ações executadas no Excel]\n\n`
        }

        text += `${'─'.repeat(50)}\n\n`
    })

    return {
        type: 'txt',
        filename: `${sanitizeFilename(title)}.txt`,
        content: text
    }
}

/**
 * Download a file to the user's computer
 */
export function downloadFile(exportData: ExportFormat): void {
    const blob = new Blob([exportData.content], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = exportData.filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
}

/**
 * Export selected messages only
 */
export function exportSelectedMessages(
    messages: Message[],
    selectedIndices: number[],
    format: 'markdown' | 'json' | 'txt' = 'markdown'
): ExportFormat {
    const selectedMessages = messages.filter((_, idx) => selectedIndices.includes(idx))
    
    switch (format) {
        case 'markdown':
            return exportToMarkdown(selectedMessages, 'Seleção de Mensagens')
        case 'json':
            return exportToJSON(selectedMessages, 'Seleção de Mensagens')
        case 'txt':
            return exportToText(selectedMessages, 'Seleção de Mensagens')
    }
}

/**
 * Sanitize filename for filesystem compatibility
 */
function sanitizeFilename(name: string): string {
    return name
        .replace(/[^a-zA-Z0-9\s\-_]/g, '')
        .replace(/\s+/g, '-')
        .replace(/-+/g, '-')
        .trim()
        .substring(0, 100)
}

/**
 * Get export format options
 */
export const EXPORT_FORMATS = [
    { value: 'markdown', label: 'Markdown (.md)', icon: '📝' },
    { value: 'json', label: 'JSON (.json)', icon: '📋' },
    { value: 'txt', label: 'Texto (.txt)', icon: '📄' }
] as const
