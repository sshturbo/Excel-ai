// Hook for managing conversations/history
import { useState, useCallback } from 'react'
import { toast } from 'sonner'
import type { ConversationItem, Message } from '@/types'
import { cleanAssistantContent, extractUserQuestion, isInternalUserMessage } from '@/services/contentCleaner'

import {
    ListConversations,
    LoadConversation,
    DeleteConversation,
    NewConversation
} from "../../wailsjs/go/app/App"

interface UseConversationsReturn {
    conversations: ConversationItem[]
    loadConversations: () => Promise<void>
    handleLoadConversation: (convId: string) => Promise<Message[]>
    handleDeleteConversation: (convId: string, e: React.MouseEvent) => Promise<void>
    handleNewConversation: () => Promise<void>
}

export function useConversations(): UseConversationsReturn {
    const [conversations, setConversations] = useState<ConversationItem[]>([])

    const loadConversations = useCallback(async () => {
        try {
            const list = await ListConversations()
            console.log('[DEBUG] Conversas carregadas:', list)
            if (list && list.length > 0) {
                setConversations(list)
            } else {
                setConversations([])
            }
        } catch (err) {
            console.error('Erro ao carregar conversas:', err)
        }
    }, [])

    const handleLoadConversation = useCallback(async (convId: string): Promise<Message[]> => {
        try {
            const messages = await LoadConversation(convId)
            if (messages && messages.length > 0) {
                const loadedMessages: Message[] = []

                for (const m of messages) {
                    // Ignore system messages
                    if (m.role === 'system') continue

                    // Ignore internal user messages
                    if (m.role === 'user') {
                        if (isInternalUserMessage(m.content)) continue

                        // Extract actual user question from context message
                        const userQuestion = extractUserQuestion(m.content)
                        if (userQuestion) {
                            loadedMessages.push({
                                role: 'user',
                                content: userQuestion
                            })
                            continue
                        }
                    }

                    // Clean assistant message content
                    let content = m.content
                    if (m.role === 'assistant') {
                        content = cleanAssistantContent(content)
                        // If empty after cleaning, skip
                        if (!content) continue
                    }

                    loadedMessages.push({
                        role: m.role as 'user' | 'assistant',
                        content: content
                    })
                }

                if (loadedMessages.length > 0) {
                    toast.success('Conversa carregada!')
                    return loadedMessages
                } else {
                    toast.info('Conversa vazia')
                }
            } else {
                toast.info('Conversa vazia')
            }
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao carregar conversa: ' + errorMessage)
        }
        return []
    }, [])

    const handleDeleteConversation = useCallback(async (convId: string, e: React.MouseEvent) => {
        e.stopPropagation() // Prevent triggering parent onClick
        try {
            await DeleteConversation(convId)
            setConversations(prev => prev.filter(c => c.id !== convId))
            toast.success('Conversa excluída!')
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao excluir: ' + errorMessage)
        }
    }, [])

    const handleNewConversation = useCallback(async () => {
        try {
            await NewConversation()
            // Reload conversation list
            const list = await ListConversations()
            if (list) setConversations(list)
            toast.success('Nova conversa criada')
        } catch (err) {
            console.error(err)
        }
    }, [])

    return {
        conversations,
        loadConversations,
        handleLoadConversation,
        handleDeleteConversation,
        handleNewConversation
    }
}
