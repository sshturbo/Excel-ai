// Hook for managing chat state and operations
// Cleaned up for native function calling system
import { useState, useRef, useCallback, useEffect } from 'react'
import { toast } from 'sonner'
import type { Message, ExcelAction } from '@/types'
import { processAIResponse } from '@/services/aiProcessor'

import {
    SendMessage,
    ClearChat,
    DeleteLastMessages,
    CancelChat
} from "../../wailsjs/go/app/App"

interface UseChatOptions {
    askBeforeApply: boolean
    onWorkbooksUpdate?: (workbooks: any[]) => void
}

interface UseChatReturn {
    messages: Message[]
    setMessages: React.Dispatch<React.SetStateAction<Message[]>>
    inputMessage: string
    setInputMessage: (value: string) => void
    isLoading: boolean
    pendingActions: ExcelAction[]
    setPendingActions: React.Dispatch<React.SetStateAction<ExcelAction[]>>
    editingMessageIndex: number | null
    editContent: string
    messagesEndRef: React.RefObject<HTMLDivElement>
    inputRef: React.RefObject<HTMLTextAreaElement>
    handleSendMessage: () => Promise<void>
    handleCancelChat: () => Promise<void>
    handleRegenerate: () => Promise<void>
    handleCopy: (text: string) => void
    handleShare: (text: string) => void
    handleEditMessage: (index: number, content: string) => void
    handleCancelEdit: () => void
    handleSaveEdit: (index: number) => Promise<void>
    handleClearChat: () => Promise<void>
    handleApplyActions: () => Promise<void>
    handleDiscardActions: () => void
    processMessage: (text: string) => Promise<void>
    validationMode: boolean
    setValidationMode: React.Dispatch<React.SetStateAction<boolean>>
    handleKeepChanges: () => void
    handleUndoChanges: () => Promise<void>
}

export function useChat(options: UseChatOptions): UseChatReturn {
    const { askBeforeApply, onWorkbooksUpdate } = options

    const [messages, setMessages] = useState<Message[]>([])
    const [inputMessage, setInputMessage] = useState('')
    const [isLoading, setIsLoading] = useState(false)
    const [pendingActions, setPendingActions] = useState<ExcelAction[]>([])
    const [editingMessageIndex, setEditingMessageIndex] = useState<number | null>(null)
    const [editContent, setEditContent] = useState('')
    const [validationMode, setValidationMode] = useState(false)

    const messagesEndRef = useRef<HTMLDivElement>(null)
    const inputRef = useRef<HTMLTextAreaElement>(null)

    // Auto-scroll to bottom when messages change
    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, [messages])

    const processMessage = useCallback(async (text: string) => {
        setIsLoading(true)

        // Add placeholder for assistant message
        setMessages(prev => [...prev, { role: 'assistant', content: '' }])

        try {
            const response = await SendMessage(text, askBeforeApply)

            // Process response for display
            const { displayContent, actionsExecuted } = await processAIResponse(response, {
                askBeforeApply,
                onWorkbooksUpdate,
                onPendingAction: (action) => setPendingActions(prev => [...prev, action]),
                onMessageUpdate: (content) => {
                    setMessages(prev => {
                        const newMsgs = [...prev]
                        const lastIndex = newMsgs.length - 1
                        if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                            newMsgs[lastIndex] = { ...newMsgs[lastIndex], content }
                        }
                        return newMsgs
                    })
                }
            })

            // Update final message
            setMessages(prev => {
                const newMsgs = [...prev]
                const lastIndex = newMsgs.length - 1
                if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                    newMsgs[lastIndex] = {
                        ...newMsgs[lastIndex],
                        content: displayContent,
                        hasActions: actionsExecuted > 0
                    }
                }
                return newMsgs
            })
        } catch (err) {
            console.error('Erro ao processar mensagem:', err)
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro: ' + errorMessage)

            // Remove placeholder on error
            setMessages(prev => prev.slice(0, -1))
        } finally {
            setIsLoading(false)
        }
    }, [askBeforeApply, onWorkbooksUpdate])

    const handleSendMessage = useCallback(async () => {
        if (!inputMessage.trim() || isLoading) return

        const userMessage = inputMessage.trim()
        setInputMessage('')
        setMessages(prev => [...prev, { role: 'user', content: userMessage }])

        await processMessage(userMessage)
    }, [inputMessage, isLoading, processMessage])

    const handleCancelChat = useCallback(async () => {
        try {
            await CancelChat()
            setIsLoading(false)
            toast.info('Geração cancelada')
        } catch (err) {
            console.error('Erro ao cancelar:', err)
        }
    }, [])

    const handleRegenerate = useCallback(async () => {
        if (messages.length < 2) return

        // Get last user message
        let lastUserMessageIndex = -1
        for (let i = messages.length - 1; i >= 0; i--) {
            if (messages[i].role === 'user') {
                lastUserMessageIndex = i
                break
            }
        }

        if (lastUserMessageIndex === -1) return

        const userMessage = messages[lastUserMessageIndex].content

        try {
            await DeleteLastMessages(2)
            setMessages(prev => prev.slice(0, -2))
            setMessages(prev => [...prev, { role: 'user', content: userMessage }])
            await processMessage(userMessage)
        } catch (err) {
            toast.error('Erro ao regenerar')
        }
    }, [messages, processMessage])

    const handleCopy = useCallback((text: string) => {
        navigator.clipboard.writeText(text)
        toast.success('Copiado!')
    }, [])

    const handleShare = useCallback((text: string) => {
        navigator.clipboard.writeText(text)
        toast.success('Copiado para compartilhar!')
    }, [])

    const handleEditMessage = useCallback((index: number, content: string) => {
        setEditingMessageIndex(index)
        setEditContent(content)
    }, [])

    const handleCancelEdit = useCallback(() => {
        setEditingMessageIndex(null)
        setEditContent('')
    }, [])

    const handleSaveEdit = useCallback(async (index: number) => {
        if (!editContent.trim()) return

        // Update local messages: cut to index and add edited
        const newMessages = messages.slice(0, index)
        newMessages.push({ role: 'user', content: editContent.trim() })
        setMessages(newMessages)
        setEditingMessageIndex(null)
        setEditContent('')

        // Call backend to edit message correctly
        try {
            const { EditMessage } = await import("../../wailsjs/go/app/App")
            await EditMessage(index, editContent.trim())
        } catch (err) {
            console.error('Erro ao editar mensagem:', err)
        }

        // Send edited message to AI
        await processMessage(editContent.trim())
    }, [editContent, messages, processMessage])

    const handleClearChat = useCallback(async () => {
        try {
            await ClearChat()
            setMessages([])
            toast.success('Chat limpo')
        } catch (err) {
            toast.error('Erro ao limpar')
        }
    }, [])

    const handleApplyActions = useCallback(async () => {
        // Implemented in App.tsx
    }, [])

    const handleDiscardActions = useCallback(() => {
        setPendingActions([])
        toast.info('Alterações descartadas')
    }, [])

    const handleKeepChanges = useCallback(() => {
        setValidationMode(false)
        setPendingActions([])
        toast.success('Alterações mantidas!')
    }, [])

    const handleUndoChanges = useCallback(async () => {
        try {
            const { UndoLastChange } = await import("../../wailsjs/go/app/App")
            await UndoLastChange()
            toast.info('Alterações desfeitas.')
            setValidationMode(false)
            setPendingActions([])
        } catch (e) {
            console.error(e)
            toast.error('Erro ao desfazer.')
        }
    }, [])

    return {
        messages,
        setMessages,
        inputMessage,
        setInputMessage,
        isLoading,
        pendingActions,
        setPendingActions,
        validationMode,
        setValidationMode,
        editingMessageIndex,
        editContent,
        messagesEndRef,
        inputRef,
        handleSendMessage,
        handleCancelChat,
        handleRegenerate,
        handleCopy,
        handleShare,
        handleEditMessage,
        handleCancelEdit,
        handleSaveEdit,
        handleClearChat,
        handleApplyActions,
        handleDiscardActions,
        handleKeepChanges,
        handleUndoChanges,
        processMessage
    }
}
