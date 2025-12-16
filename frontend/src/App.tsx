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
    EndUndoBatch,
    IsLicenseValid,
    GetLicenseMessage
} from "../wailsjs/go/app/App"

export default function App() {
    // Theme
    const { theme, toggleTheme } = useTheme()

    // Settings state
    const [showSettings, setShowSettings] = useState(false)
    const [askBeforeApply, setAskBeforeApply] = useState(true)
    const [apiKey, setApiKey] = useState('')
    const [model, setModel] = useState('openai/gpt-4o-mini')

    // License state
    const [licenseValid, setLicenseValid] = useState(true)
    const [licenseMessage, setLicenseMessage] = useState('')

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
            checkLicense()
        } else {
            console.warn('Wails runtime não detectado. Rodando fora do app (Vite puro).')
        }
    }, [])

    // Check license status
    const checkLicense = async () => {
        try {
            const valid = await IsLicenseValid()
            const message = await GetLicenseMessage()
            setLicenseValid(valid)
            setLicenseMessage(message)
            console.log('[LICENSE]', valid ? '✅' : '❌', message)
        } catch (err) {
            console.error('Error checking license:', err)
            setLicenseValid(false)
            setLicenseMessage('Erro ao verificar licença')
        }
    }

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

    // Handle applying pending actions - now calls backend to execute and resume AI
    const handleApplyActions = async () => {
        console.log('[DEBUG] handleApplyActions called - calling backend ConfirmPendingAction')

        try {
            // Import dynamically to avoid circular dependency
            const { ConfirmPendingAction } = await import("../wailsjs/go/app/App")

            toast.info("Executando ação...")

            // Clear local pending actions since backend will handle everything
            chat.setPendingActions([])

            // Call backend - this will execute the action and resume AI
            const response = await ConfirmPendingAction()

            console.log('[DEBUG] ConfirmPendingAction response:', response?.substring(0, 100))

            if (response.startsWith("Error:")) {
                toast.error(response)
            } else {
                toast.success("Ação executada!")
                // Refresh workbooks to show any new sheets/data
                await excel.refreshWorkbooks()
            }
        } catch (err) {
            console.error('Error confirming action:', err)
            toast.error('Erro ao confirmar ação')
            chat.setPendingActions([])
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

    // License blocked view
    if (!licenseValid) {
        return (
            <div className="flex items-center justify-center h-screen bg-background">
                <div className="text-center p-8 max-w-md">
                    <div className="text-6xl mb-4">🔒</div>
                    <h1 className="text-2xl font-bold text-red-500 mb-4">
                        Licença Inválida
                    </h1>
                    <p className="text-muted-foreground mb-6">
                        {licenseMessage || 'Sua licença expirou ou foi revogada.'}
                    </p>
                    <p className="text-sm text-muted-foreground">
                        Entre em contato com o suporte para obter uma nova licença.
                    </p>
                    <button
                        onClick={checkLicense}
                        className="mt-6 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:opacity-90"
                    >
                        Verificar Novamente
                    </button>
                </div>
            </div>
        )
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
                        validationMode={chat.validationMode}
                        onApply={handleApplyActions}
                        onDiscard={chat.handleDiscardActions}
                        onKeep={chat.handleKeepChanges}
                        onUndo={chat.handleUndoChanges}
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

                    {/* Continue Button - appears when agent reaches step limit */}
                    {chat.showContinueButton && !chat.isLoading && (
                        <div className="flex justify-center py-3 px-6 border-t border-gray-200 dark:border-gray-700 bg-linear-to-r from-blue-50 to-indigo-50 dark:from-gray-800 dark:to-gray-900">
                            <button
                                onClick={chat.handleContinue}
                                className="flex items-center gap-2 px-6 py-3 bg-linear-to-r from-blue-500 to-indigo-500 hover:from-blue-600 hover:to-indigo-600 text-white font-medium rounded-lg shadow-lg hover:shadow-xl transition-all duration-200 transform hover:scale-105"
                            >
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                Continuar Execução
                            </button>
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
