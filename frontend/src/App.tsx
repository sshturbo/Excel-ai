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
import { useFullscreen } from '@/hooks/useFullscreen'
import { useExcelUpload } from '@/hooks/useExcelUpload'

// Components
import { Header } from '@/components/layout/Header'
import { Sidebar } from '@/components/layout/Sidebar'
import { ChatInput } from '@/components/chat/ChatInput'
import { MessageBubble } from '@/components/chat/MessageBubble'
import { EmptyState } from '@/components/chat/EmptyState'
import { MarkdownRenderer } from '@/components/markdown/MarkdownRenderer'
import { DataPreview } from '@/components/excel/DataPreview'
import { HandsontableViewer } from '@/components/excel/HandsontableViewer'
import { ChartViewer } from '@/components/excel/ChartViewer'
import { ActionState, PendingActions } from '@/components/excel/PendingActions'
import { Toolbar } from '@/components/excel/Toolbar'
import { ExportDialog } from '@/components/ExportDialog'
import { SaveConfirmationDialog } from '@/components/excel/SaveConfirmationDialog'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

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
    // Theme & Fullscreen
    const { theme, toggleTheme } = useTheme()
    const { isFullscreen, toggleFullscreen } = useFullscreen()

    // Sidebar state
    const [isSidebarOpen, setIsSidebarOpen] = useState(true)

    // Settings state
    const [showSettings, setShowSettings] = useState(false)
    const [showExportDialog, setShowExportDialog] = useState(false)
    const [askBeforeApply, setAskBeforeApply] = useState(true)
    const [apiKey, setApiKey] = useState('')
    const [model, setModel] = useState('openai/gpt-4o-mini')

    // License state
    const [licenseValid, setLicenseValid] = useState(true)
    const [licenseMessage, setLicenseMessage] = useState('')

    // View state
    const [showPreview, setShowPreview] = useState(false)
    const [showChart, setShowChart] = useState(false)
    const [showSaveConfirm, setShowSaveConfirm] = useState(false)

    // Action execution state
    const [actionState, setActionState] = useState<ActionState>('pending')
    const [actionError, setActionError] = useState<string | undefined>()
    const [hasPendingAction, setHasPendingAction] = useState(false)

    // Custom hooks
    const excel = useExcelConnection()
    const conversations = useConversations()
    const excelUpload = useExcelUpload()

    const chat = useChat({
        askBeforeApply,
        onWorkbooksUpdate: excel.setWorkbooks
    })

    const { streamingContent, reasoningContent } = useStreamingMessage(chat.isLoading)

    // Update streaming content to messages
    const lastContentRef = useRef('')
    const lastReasoningRef = useRef('')

    // Unified streaming effect: Updates message with both reasoning and content
    useEffect(() => {
        if (!chat.isLoading) return

        // Construct full visually formatted content
        let displayContent = streamingContent
        if (reasoningContent) {
            // ALWAYS prepend reasoning if it exists during streaming
            displayContent = `:::reasoning:::${reasoningContent}:::/reasoning:::\n\n${streamingContent}`
        }

        // If nothing to show yet, do nothing
        if (!displayContent) return

        lastContentRef.current = streamingContent

        chat.setMessages(prev => {
            const newMsgs = [...prev]
            const lastIndex = newMsgs.length - 1
            if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                // Only update if content changed
                if (newMsgs[lastIndex].content !== displayContent) {
                    newMsgs[lastIndex] = { ...newMsgs[lastIndex], content: displayContent }
                    return newMsgs
                }
            }
            return prev
        })
    }, [streamingContent, reasoningContent, chat.isLoading])

    // Save reasoning content to message when loading completes (persistence)
    useEffect(() => {
        if (!chat.isLoading && reasoningContent && reasoningContent !== lastReasoningRef.current) {
            lastReasoningRef.current = reasoningContent
            // Final update to ensure reasoning is saved
            chat.setMessages(prev => {
                const newMsgs = [...prev]
                const lastIndex = newMsgs.length - 1
                if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                    const currentContent = newMsgs[lastIndex].content
                    // If content doesn't already have reasoning tags (e.g. from streaming update above), add them
                    if (!currentContent.includes(':::reasoning:::')) {
                        newMsgs[lastIndex] = {
                            ...newMsgs[lastIndex],
                            content: `:::reasoning:::${reasoningContent}:::/reasoning:::\n\n${currentContent}`
                        }
                        return newMsgs
                    }
                }
                return prev
            })
        }
    }, [chat.isLoading, reasoningContent])

    // Load saved config on mount
    useEffect(() => {
        const hasWailsRuntime = typeof (window as any)?.go !== 'undefined'
        if (hasWailsRuntime) {
            loadConfig()
            conversations.loadConversations()
            checkLicense()
        } else {
            console.warn('Wails runtime não detectado. Rodando fora do app (Vite puro).')
        }
    }, [])

    // Keyboard shortcuts for action confirmation (Y/n like gemini-cli)
    useEffect(() => {
        if (chat.pendingActions.length === 0 || actionState !== 'pending') return

        const handleKeyDown = (e: KeyboardEvent) => {
            // Ignore if user is typing in an input
            if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return

            if (e.key.toLowerCase() === 'y') {
                e.preventDefault()
                handleApplyActions()
            } else if (e.key.toLowerCase() === 'n') {
                e.preventDefault()
                handleDiscardActions()
            }
        }

        window.addEventListener('keydown', handleKeyDown)
        return () => window.removeEventListener('keydown', handleKeyDown)
    }, [chat.pendingActions.length, actionState])

    // Check for pending action when loading state changes
    useEffect(() => {
        if (!chat.isLoading) {
            // After AI response completes, check if backend has pending action
            import("../wailsjs/go/app/App").then(({ HasPendingAction }) => {
                HasPendingAction().then(hasPending => {
                    setHasPendingAction(hasPending)
                    if (hasPending) {
                        setActionState('pending')
                    }
                }).catch(() => { })
            })
        }
    }, [chat.isLoading])

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
                // Load YOLO mode setting (askBeforeApply is true by default in backend)
                setAskBeforeApply(config.askBeforeApply !== false)
            }
        } catch (err) {
            console.error('Error loading config:', err)
        }
    }

    // Handle applying pending actions - now calls backend to execute and resume AI
    const handleApplyActions = async () => {
        console.log('[DEBUG] handleApplyActions called - calling backend ConfirmPendingAction')

        setActionState('executing')
        setActionError(undefined)
        setHasPendingAction(false)

        // Adicionar status de execução (não substitui, apenas adiciona)
        chat.setMessages(prev => {
            const newMsgs = [...prev]
            const lastIndex = newMsgs.length - 1
            if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                // Apenas adiciona ao final sem remover nada
                const currentContent = newMsgs[lastIndex].content
                // Evitar duplicar se já tiver o status
                if (!currentContent.includes('🔄 **Aplicando')) {
                    newMsgs[lastIndex] = {
                        ...newMsgs[lastIndex],
                        content: currentContent + '\n\n🔄 **Aplicando alterações...**'
                    }
                }
            }
            return newMsgs
        })

        // Set up streaming listener to detect action completion and update message in real-time
        let actionCompleted = false
        let streamBuffer = ''
        const { EventsOn, EventsOff } = await import("../wailsjs/runtime/runtime")

        const handleChunk = (chunk: string) => {
            streamBuffer += chunk

            // Detect when action execution completes
            if (!actionCompleted && streamBuffer.includes('✅ Ação executada com sucesso!')) {
                actionCompleted = true
                // Update message immediately
                chat.setMessages(prev => {
                    const newMsgs = [...prev]
                    const lastIndex = newMsgs.length - 1
                    if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                        const oldContent = newMsgs[lastIndex].content.replace(/🔄 \*\*Aplicando alterações\.\.\.\*\*/g, '').trim()
                        newMsgs[lastIndex] = {
                            ...newMsgs[lastIndex],
                            content: oldContent + '\n\n✅ **Alterações aplicadas com sucesso!**\n\n*IA continuando...*',
                            hasActions: true
                        }
                    }
                    return newMsgs
                })
            }
        }

        const cleanupListener = EventsOn("chat:chunk", handleChunk)

        try {
            // Import dynamically to avoid circular dependency
            const { ConfirmPendingAction } = await import("../wailsjs/go/app/App")

            // Call backend - this will execute the action
            const response = await ConfirmPendingAction()

            // Clean up listener
            cleanupListener()

            console.log('[DEBUG] ConfirmPendingAction response:', response?.substring(0, 100))

            if (response.startsWith("Error:")) {
                setActionState('error')
                setActionError(response.replace('Error: ', ''))
                // Update message with error
                chat.setMessages(prev => {
                    const newMsgs = [...prev]
                    const lastIndex = newMsgs.length - 1
                    if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                        const currentContent = newMsgs[lastIndex].content
                        newMsgs[lastIndex] = {
                            ...newMsgs[lastIndex],
                            content: currentContent + '\n\n❌ **Erro ao aplicar:** ' + response.replace('Error: ', '')
                        }
                    }
                    return newMsgs
                })
            } else {
                // Update last message to show success
                chat.setMessages(prev => {
                    const newMsgs = [...prev]
                    const lastIndex = newMsgs.length - 1
                    if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                        const currentContent = newMsgs[lastIndex].content
                        // Evitar duplicar se já tiver sucesso
                        if (!currentContent.includes('✅ **Alterações aplicadas')) {
                            newMsgs[lastIndex] = {
                                ...newMsgs[lastIndex],
                                content: currentContent + '\n\n✅ **Alterações aplicadas com sucesso!**',
                                hasActions: true
                            }
                        }
                    }
                    return newMsgs
                })

                // Atualizar preview automaticamente após modificações da IA
                if (excelUpload.sessionId) {
                    await excelUpload.refreshPreview()
                    toast.success('Preview atualizado com as modificações')
                }

                // Add AI continuation response to chat if there is one
                if (response && response.trim()) {
                    const { processAIResponse } = await import("@/services/aiProcessor")
                    const { displayContent, actionsExecuted } = await processAIResponse(response, {
                        askBeforeApply: true,
                        onPendingAction: (action) => chat.setPendingActions(prev => [...prev, action]),
                    })

                    // Only add new message if there's meaningful content beyond tool results
                    if (displayContent && !displayContent.includes('TOOL RESULTS')) {
                        chat.setMessages(prev => [...prev, {
                            role: 'assistant',
                            content: displayContent,
                            hasActions: actionsExecuted > 0
                        }])
                    }

                    // Check if there are new pending actions
                    const { HasPendingAction } = await import("../wailsjs/go/app/App")
                    const hasPending = await HasPendingAction()
                    if (hasPending) {
                        setHasPendingAction(true)
                        setActionState('pending')
                        return // Don't show completed, there's another action pending
                    }
                }

                // Set to completed state - will show Keep/Undo buttons
                setActionState('completed')
            }
        } catch (err) {
            console.error('Error confirming action:', err)
            setActionState('error')
            setActionError(err instanceof Error ? err.message : String(err))
        }
    }

    // Handle discarding pending actions
    const handleDiscardActions = async () => {
        // Update last message to show action was cancelled
        chat.setMessages(prev => {
            const newMsgs = [...prev]
            const lastIndex = newMsgs.length - 1
            if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                const currentContent = newMsgs[lastIndex].content
                newMsgs[lastIndex] = {
                    ...newMsgs[lastIndex],
                    content: currentContent + '\n\n🚫 **Ação descartada pelo usuário.**'
                }
            }
            return newMsgs
        })

        try {
            const { RejectPendingAction } = await import("../wailsjs/go/app/App")
            await RejectPendingAction()
        } catch (err) {
            console.error('Error rejecting action:', err)
        }
        chat.setPendingActions([])
        setActionState('pending')
        setHasPendingAction(false)
        toast.info('Ação descartada')
    }

    // Handle keeping changes (after completed state)
    const handleKeepChanges = async () => {
        try {
            // Aprovar ações no banco de dados (não podem mais ser desfeitas)
            const { ApproveUndoActions, GetCurrentConversationID } = await import("../wailsjs/go/app/App")
            const convID = await GetCurrentConversationID()
            if (convID) {
                await ApproveUndoActions(convID)
            }
        } catch (err) {
            console.error('Error approving undo actions:', err)
        }
        setActionState('pending')
        setHasPendingAction(false)
        chat.setPendingActions([])
        toast.success('Alterações confirmadas!')
    }

    // Handle undo (revert changes)
    const handleUndo = async () => {
        try {
            const { UndoByConversation, GetCurrentConversationID } = await import("../wailsjs/go/app/App")
            const convID = await GetCurrentConversationID()
            if (convID) {
                const undoneCount = await UndoByConversation(convID)
                toast.success(`${undoneCount} célula(s) restaurada(s)!`)
            } else {
                toast.error('Nenhuma conversa ativa')
            }
            // Clear completed state
            setActionState('pending')
            setHasPendingAction(false)
            chat.setPendingActions([])
            // Refresh preview
            if (excelUpload.sessionId) {
                await excelUpload.refreshPreview()
            }
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : String(err)
            toast.error(errorMsg || 'Nada para desfazer')
        }
    }

    // Handle new conversation
    const handleNewConversation = async () => {
        console.log('[DEBUG] Nova conversa clicada')
        try {
            await conversations.handleNewConversation()
            chat.setMessages([])
            // Reset reasoning content
            lastReasoningRef.current = ''
            console.log('[DEBUG] Nova conversa criada com sucesso')
        } catch (err) {
            console.error('[ERROR] Erro ao criar nova conversa:', err)
            toast.error('Erro ao criar nova conversa')
        }
    }

    // Handle loading a conversation
    const handleLoadConversation = async (convId: string) => {
        const messages = await conversations.handleLoadConversation(convId)
        if (messages.length > 0) {
            chat.setMessages(messages)
        }
    }

    // Handle file upload via chat input (+ button)
    const handleFileUpload = async (file: File) => {
        // Validar arquivo
        if (!file.name.endsWith('.xlsx') && !file.name.endsWith('.xls')) {
            toast.error('Por favor, selecione um arquivo Excel (.xlsx ou .xls)')
            return
        }

        // Converter para Uint8Array
        const arrayBuffer = await file.arrayBuffer()
        const data = new Uint8Array(arrayBuffer)

        try {
            // Fazer upload usando useExcelUpload
            await excelUpload.handleUpload(file.name, data)

            toast.success(`Arquivo "${file.name}" carregado com sucesso!`)
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao carregar arquivo: ' + errorMsg)
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
                onAskBeforeApplyChange={async (value) => {
                    console.log('[DEBUG] askBeforeApply changing to:', value)
                    setAskBeforeApply(value)
                    // Save to backend
                    try {
                        const { SetAskBeforeApply } = await import("../wailsjs/go/app/App")
                        await SetAskBeforeApply(value)
                    } catch (err) {
                        console.error('Error saving askBeforeApply:', err)
                    }
                }}
            />
        )
    }

    // Preview data from upload mode (Excelize only)
    // Criar previewData assim que arquivo é carregado, mesmo antes de clicar na aba
    console.log('[DEBUG App.tsx] Calculando previewData:')
    console.log('[DEBUG App.tsx] - excelUpload.previewData:', excelUpload.previewData)
    console.log('[DEBUG App.tsx] - excelUpload.sheetData:', excelUpload.sheetData)

    const previewData = excelUpload.previewData ? {
        fileName: excelUpload.previewData.fileName,
        sheets: excelUpload.previewData.sheets,
        activeSheet: excelUpload.activeSheet || excelUpload.previewData.activeSheet,
        headers: excelUpload.sheetData ? excelUpload.sheetData[0] || [] : [],
        rows: excelUpload.sheetData ? excelUpload.sheetData.slice(1) || [] : [],
        totalRows: excelUpload.sheetData ? excelUpload.sheetData.length || 0 : 0
    } : null

    console.log('[DEBUG App.tsx] - previewData calculado:', previewData)

    const activeSheet = excelUpload.activeSheet || excelUpload.previewData?.activeSheet || ''
    const selectedSheets = excelUpload.previewData?.sheets.map(s => s.name) || []

    // Workbooks to show in sidebar
    const workbooks = excelUpload.previewData
        ? [{ name: excelUpload.previewData.fileName, sheets: excelUpload.previewData.sheets.map(s => s.name) }]
        : []

    return (
        <div className="flex flex-col h-screen bg-background text-foreground">
            {/* Header */}
            <Header
                connected={!!excelUpload.sessionId}
                theme={theme}
                onNewConversation={handleNewConversation}
                onOpenSettings={() => setShowSettings(true)}
                onToggleTheme={toggleTheme}
                onTogglePreview={() => { setShowPreview(!showPreview); setShowChart(false); }}
                onToggleChart={() => { setShowChart(!showChart); setShowPreview(false); }}
                onUndo={handleUndo}
                onOpenExportDialog={() => setShowExportDialog(true)}
                onToggleSidebar={() => setIsSidebarOpen(!isSidebarOpen)}
            />

            {/* Main - Unified interface for both COM and Upload modes */}
            <main className="flex flex-1 overflow-hidden relative">
                {/* Collapsible Sidebar Wrapper */}
                <div
                    className={cn(
                        "shrink-0 transition-all duration-300 ease-in-out overflow-hidden bg-card border-r border-border",
                        isSidebarOpen ? "w-72 opacity-100" : "w-0 opacity-0 border-none"
                    )}
                >
                    <div className="w-72 h-full">
                        <Sidebar
                            workbooks={workbooks}
                            connected={!!excelUpload.sessionId}
                            selectedWorkbook={excelUpload.previewData?.fileName || ''}
                            selectedSheets={activeSheet ? [activeSheet] : []}
                            expandedWorkbook={excelUpload.previewData?.fileName || null}
                            contextLoaded={excelUpload.sessionId ? `Arquivo: ${excelUpload.previewData?.fileName || 'carregado'}` : ''}
                            onExpandWorkbook={() => { }}
                            onSelectSheet={async (_wbName, sheetName) => {
                                console.log('[DEBUG App.tsx] onSelectSheet chamado com sheet:', sheetName)
                                await excelUpload.loadSheetData(sheetName)
                            }}
                            conversations={conversations.conversations}
                            isLoadingConversations={conversations.isLoading}
                            onLoadConversations={conversations.loadConversations}
                            onLoadConversation={handleLoadConversation}
                            onDeleteConversation={conversations.handleDeleteConversation}
                            onCloseSession={excelUpload.closeSession}
                        />
                    </div>
                </div>

                {/* Chat Area - same for both modes */}
                <section className="flex-1 flex flex-col bg-linear-to-b from-background to-muted/20">
                    {/* Toolbar - show if there's preview data */}
                    {previewData && (
                        <Toolbar
                            showPreview={showPreview}
                            showChart={showChart}
                            onTogglePreview={() => { setShowPreview(!showPreview); setShowChart(false); }}
                            onToggleChart={() => { setShowChart(!showChart); setShowPreview(false); }}
                            selectedSheets={selectedSheets}
                            activeSheet={activeSheet}
                            onSwitchSheet={async (sheet) => { await excelUpload.loadSheetData(sheet); }}
                            onRefresh={excelUpload.refreshPreview}
                            onOpenNative={excelUpload.handleOpenFileNative}
                            onSaveNative={() => setShowSaveConfirm(true)}
                            isSaving={excelUpload.downloading}
                        />
                    )}

                    {/* Data Preview - Handsontable for Excel-like experience */}
                    {showPreview && excelUpload.sessionId && (
                        <div className="flex-1 overflow-hidden w-full relative">
                            <HandsontableViewer
                                sessionId={excelUpload.sessionId}
                                activeSheet={excelUpload.activeSheet || 'Planilha1'}
                                sheetData={excelUpload.sheetData}
                                loading={excelUpload.loadingSheet}
                            />
                        </div>
                    )}

                    {/* Chart View */}
                    {showChart && previewData && previewData.headers && previewData.headers.length > 0 && (
                        <div className="flex-1 flex flex-col overflow-hidden">
                            <ChartViewer previewData={previewData} />
                        </div>
                    )}

                    {/* Pending Actions Banner */}
                    <PendingActions
                        actions={chat.pendingActions}
                        state={actionState}
                        error={actionError}
                        hasPendingAction={hasPendingAction}
                        onApply={handleApplyActions}
                        onDiscard={handleDiscardActions}
                        onKeep={handleKeepChanges}
                        onUndo={handleUndo}
                    />

                    {/* Chat Messages */}
                    {!showPreview && !showChart && (
                        <div className="flex-1 overflow-y-auto p-6 space-y-6">
                            {chat.messages.length === 0 ? (
                                <EmptyState
                                    selectedSheets={selectedSheets}
                                    onOpenNative={excelUpload.handleOpenFileNative}
                                />
                            ) : (
                                <>
                                    {chat.messages
                                        .filter(msg => !msg.hidden && msg.role !== 'system' && !msg.content.startsWith('Resultados das ferramentas'))
                                        .filter(msg => msg.role !== 'assistant' || msg.content.trim().length > 0)
                                        .map((msg, idx) => (
                                            <MessageBubble
                                                key={idx}
                                                message={msg}
                                                index={idx}
                                                isLastAssistant={msg.role === 'assistant' && idx === chat.messages.filter(m => !m.hidden && m.role !== 'system' && !m.content.startsWith('Resultados das ferramentas')).length - 1}
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
                                        ))}
                                </>
                            )}
                            <div ref={chat.messagesEndRef} />
                        </div>
                    )}

                    {/* Chat Input */}
                    <ChatInput
                        inputMessage={chat.inputMessage}
                        isLoading={chat.isLoading || actionState === 'executing'}
                        inputRef={chat.inputRef}
                        onInputChange={chat.setInputMessage}
                        onSend={chat.handleSendMessage}
                        onCancel={chat.handleCancelChat}
                        onFileUpload={handleFileUpload}
                    />
                </section>
            </main>

            {/* Export Dialog */}
            <ExportDialog
                open={showExportDialog}
                onOpenChange={setShowExportDialog}
                messages={chat.messages}
                conversationTitle="Conversa"
            />

            {/* Save Confirmation Dialog */}
            <SaveConfirmationDialog
                open={showSaveConfirm}
                onOpenChange={setShowSaveConfirm}
                onConfirm={excelUpload.handleSaveNative}
                fileName={excelUpload.previewData?.fileName || 'arquivo'}
            />
        </div>
    )
}
