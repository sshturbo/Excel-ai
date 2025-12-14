// App.tsx - Main application component (refactored)
// Uses modular components and hooks for better maintainability

import { useState, useEffect, useRef } from 'react'
import { toast } from 'sonner'

// Hooks
import { useExcelConnection } from '@/hooks/useExcelConnection'
import { useChat } from '@/hooks/useChat'
import { useConversations } from '@/hooks/useConversations'
import { useStreamingMessage } from '@/hooks/useStreamingMessage'
import { useTheme } from '@/hooks/useTheme'

// Components
import { Header } from '@/components/layout/Header'
import { Sidebar } from '@/components/layout/Sidebar'
import { ChatInput } from '@/components/chat/ChatInput'
import { MessageBubble } from '@/components/chat/MessageBubble'
import { EmptyState } from '@/components/chat/EmptyState'
import { MarkdownRenderer } from '@/components/markdown/MarkdownRenderer'
import { DataPreview } from '@/components/excel/DataPreview'
import { ChartViewer } from '@/components/excel/ChartViewer'
import { PendingActions } from '@/components/excel/PendingActions'
import { Toolbar } from '@/components/excel/Toolbar'

// Services
import { executeExcelAction } from '@/services/excelActions'

// Types
import type { ExcelAction } from '@/types'

// Settings component
import Settings from './Settings'

// Wails bindings
import {
    GetSavedConfig,
    UndoLastChange,
    GetPreviewData,
    RefreshWorkbooks,
    StartUndoBatch,
    EndUndoBatch
} from "../wailsjs/go/app/App"

export default function App() {
    // Theme
    const { theme, toggleTheme } = useTheme()

    // Settings state
    const [showSettings, setShowSettings] = useState(false)
    const [askBeforeApply, setAskBeforeApply] = useState(true)
    const [apiKey, setApiKey] = useState('')
    const [model, setModel] = useState('openai/gpt-4o-mini')

    // View state
    const [showPreview, setShowPreview] = useState(false)
    const [showChart, setShowChart] = useState(false)
    const [chartType, setChartType] = useState<'bar' | 'line' | 'pie'>('bar')
    const [chartData, setChartData] = useState<any>(null)

    // Custom hooks
    const excel = useExcelConnection()
    const conversations = useConversations()

    const chat = useChat({
        askBeforeApply,
        onWorkbooksUpdate: excel.setWorkbooks
    })

    const { streamingContent } = useStreamingMessage(chat.isLoading)

    // Update streaming content to messages
    const lastContentRef = useRef('')
    useEffect(() => {
        if (!chat.isLoading || !streamingContent) return
        if (streamingContent === lastContentRef.current) return

        lastContentRef.current = streamingContent
        chat.setMessages(prev => {
            const newMsgs = [...prev]
            const lastIndex = newMsgs.length - 1
            if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                if (newMsgs[lastIndex].content !== streamingContent) {
                    newMsgs[lastIndex] = { ...newMsgs[lastIndex], content: streamingContent }
                    return newMsgs
                }
            }
            return prev
        })
    }, [streamingContent, chat.isLoading])

    // Load saved config on mount
    useEffect(() => {
        const hasWailsRuntime = typeof (window as any)?.go !== 'undefined'
        if (hasWailsRuntime) {
            excel.handleConnect()
            loadConfig()
            conversations.loadConversations()
        } else {
            console.warn('Wails runtime não detectado. Rodando fora do app (Vite puro).')
        }
    }, [])

    const loadConfig = async () => {
        try {
            const config = await GetSavedConfig()
            if (config) {
                if (config.apiKey) setApiKey(config.apiKey)
                if (config.model) setModel(config.model)
            }
        } catch (err) {
            console.error('Error loading config:', err)
        }
    }

    // Prepare chart data when preview data changes
    useEffect(() => {
        if (excel.previewData) {
            const data = excel.prepareChartData(excel.previewData)
            setChartData(data)
        }
    }, [excel.previewData])

    // Handle applying pending actions
    const handleApplyActions = async () => {
        if (chat.pendingActions.length === 0) return

        try {
            await StartUndoBatch()

            let successCount = 0
            for (const action of chat.pendingActions) {
                const result = await executeExcelAction(action as ExcelAction, excel.setWorkbooks)
                if (result.success) {
                    successCount++
                } else {
                    toast.error(`Erro: ${result.error}`)
                    break
                }
            }

            await EndUndoBatch()

            if (successCount > 0) {
                toast.success(`${successCount} alteração(ões) aplicada(s)!`)

                // Mark last message as having actions
                chat.setMessages(prev => {
                    const newMsgs = [...prev]
                    const lastIndex = newMsgs.length - 1
                    if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                        newMsgs[lastIndex] = { ...newMsgs[lastIndex], hasActions: true }
                    }
                    return newMsgs
                })

                // Refresh preview
                if (excel.selectedWorkbook && excel.selectedSheets.length > 0) {
                    const preview = await GetPreviewData(excel.selectedWorkbook, excel.selectedSheets[0])
                    if (preview) {
                        // This would need setPreviewData exposed from useExcelConnection
                        // For now we'll refresh workbooks
                        await excel.refreshWorkbooks()
                    }
                }
            }

            chat.setPendingActions([])
        } catch (err) {
            console.error('Error applying actions:', err)
            toast.error('Erro ao aplicar alterações')
        }
    }

    // Handle undo
    const handleUndo = async () => {
        try {
            await UndoLastChange()
            toast.success('Alteração desfeita!')
            if (excel.selectedWorkbook && excel.selectedSheets.length > 0) {
                const preview = await GetPreviewData(excel.selectedWorkbook, excel.selectedSheets[0])
                // Would need to update preview data in hook
            }
        } catch (err) {
            toast.error('Nada para desfazer')
        }
    }

    // Handle new conversation
    const handleNewConversation = async () => {
        await conversations.handleNewConversation()
        chat.setMessages([])
        // Could also reset excel selection here if desired
    }

    // Handle loading a conversation
    const handleLoadConversation = async (convId: string) => {
        const messages = await conversations.handleLoadConversation(convId)
        if (messages.length > 0) {
            chat.setMessages(messages)
        }
    }

    // Render message content using MarkdownRenderer
    const renderMessageContent = (content: string) => {
        return <MarkdownRenderer content={content} />
    }

    // Settings view
    if (showSettings) {
        return (
            <Settings
                onClose={() => setShowSettings(false)}
                askBeforeApply={askBeforeApply}
                onAskBeforeApplyChange={setAskBeforeApply}
            />
        )
    }

    return (
        <div className="flex flex-col h-screen bg-background text-foreground">
            {/* Header */}
            <Header
                connected={excel.connected}
                theme={theme}
                onNewConversation={handleNewConversation}
                onOpenSettings={() => setShowSettings(true)}
                onConnect={excel.handleConnect}
                onToggleTheme={toggleTheme}
            />

            {/* Main */}
            <main className="flex flex-1 overflow-hidden">
                {/* Sidebar */}
                <Sidebar
                    workbooks={excel.workbooks}
                    connected={excel.connected}
                    selectedWorkbook={excel.selectedWorkbook}
                    selectedSheets={excel.selectedSheets}
                    expandedWorkbook={excel.expandedWorkbook}
                    contextLoaded={excel.contextLoaded}
                    onExpandWorkbook={excel.setExpandedWorkbook}
                    onSelectSheet={excel.handleSelectSheet}
                    conversations={conversations.conversations}
                    onLoadConversations={conversations.loadConversations}
                    onLoadConversation={handleLoadConversation}
                    onDeleteConversation={conversations.handleDeleteConversation}
                />

                {/* Chat Area */}
                <section className="flex-1 flex flex-col bg-linear-to-b from-background to-muted/20">
                    {/* Toolbar */}
                    {excel.previewData && (
                        <Toolbar
                            showPreview={showPreview}
                            showChart={showChart}
                            chartType={chartType}
                            onTogglePreview={() => { setShowPreview(!showPreview); setShowChart(false); }}
                            onToggleChart={() => { setShowChart(!showChart); setShowPreview(false); }}
                            onChartTypeChange={setChartType}
                        />
                    )}

                    {/* Data Preview */}
                    {showPreview && excel.previewData && (
                        <DataPreview previewData={excel.previewData} />
                    )}

                    {/* Chart View */}
                    {showChart && chartData && (
                        <ChartViewer chartType={chartType} chartData={chartData} />
                    )}

                    {/* Pending Actions Banner */}
                    <PendingActions
                        count={chat.pendingActions.length}
                        onApply={handleApplyActions}
                        onDiscard={chat.handleDiscardActions}
                    />

                    {/* Chat Messages */}
                    {!showPreview && !showChart && (
                        <div className="flex-1 overflow-y-auto p-6 space-y-4">
                            {chat.messages.length === 0 ? (
                                <EmptyState selectedSheets={excel.selectedSheets} />
                            ) : (
                                chat.messages.map((msg, idx) => (
                                    <MessageBubble
                                        key={idx}
                                        message={msg}
                                        index={idx}
                                        isLastAssistant={msg.role === 'assistant' && idx === chat.messages.length - 1}
                                        isLoading={chat.isLoading}
                                        isEditing={chat.editingMessageIndex === idx}
                                        editContent={chat.editContent}
                                        onEditContentChange={(value) => chat.handleEditMessage(idx, value)}
                                        onStartEdit={chat.handleEditMessage}
                                        onCancelEdit={chat.handleCancelEdit}
                                        onSaveEdit={chat.handleSaveEdit}
                                        onCopy={chat.handleCopy}
                                        onShare={chat.handleShare}
                                        onRegenerate={chat.handleRegenerate}
                                        onUndo={msg.hasActions ? handleUndo : undefined}
                                        renderContent={renderMessageContent}
                                    />
                                ))
                            )}
                            <div ref={chat.messagesEndRef} />
                        </div>
                    )}

                    {/* Chat Input */}
                    <ChatInput
                        inputMessage={chat.inputMessage}
                        isLoading={chat.isLoading}
                        inputRef={chat.inputRef}
                        onInputChange={chat.setInputMessage}
                        onSend={chat.handleSendMessage}
                        onCancel={chat.handleCancelChat}
                    />
                </section>
            </main>
        </div>
    )
}
