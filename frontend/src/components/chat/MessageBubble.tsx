// MessageBubble component - Professional chat message with reasoning support
import React, { ReactNode } from 'react'
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { ReasoningSection } from './ReasoningSection'
import type { Message } from '@/types'

interface MessageBubbleProps {
    message: Message
    index: number
    isLastAssistant: boolean
    isLoading: boolean
    isEditing: boolean
    editContent: string
    onEditContentChange: (value: string) => void
    onStartEdit: (index: number, content: string) => void
    onCancelEdit: () => void
    onSaveEdit: (index: number) => void
    onCopy: (text: string) => void
    onShare: (text: string) => void
    onRegenerate: () => void
    onUndo?: () => void
    renderContent: (content: string) => ReactNode
}

// Parse message to extract reasoning content if present, robustly handling streaming
function parseMessageContent(content: string): { mainContent: string; reasoning: string | null } {
    // Regex to find reasoning block:
    // Matches :::reasoning:::
    // Captures content (lazy) until...
    // Matches :::::/reasoning::: OR End of String ($)
    const match = content.match(/:::reasoning:::([\s\S]*?)(?:$|:::\/reasoning:::)/)

    if (match) {
        const reasoning = match[1].trim()

        // Remove the entire reasoning block from the content to get mainContent
        // We use the same regex logic but replace with empty string
        // We also trim start to remove any leading newlines left behind
        const mainContent = content.replace(/:::reasoning:::[\s\S]*?(?:$|:::\/reasoning:::)/, '').trim()

        return { mainContent, reasoning }
    }

    return { mainContent: content, reasoning: null }
}

export function MessageBubble({
    message,
    index,
    isLastAssistant,
    isLoading,
    isEditing,
    editContent,
    onEditContentChange,
    onStartEdit,
    onCancelEdit,
    onSaveEdit,
    onCopy,
    onShare,
    onRegenerate,
    onUndo,
    renderContent,
    isStreaming // Receive isStreaming prop
}: MessageBubbleProps & { isStreaming?: boolean }) {
    const { mainContent, reasoning } = parseMessageContent(message.content)
    const isUser = message.role === 'user'
    const isAssistant = message.role === 'assistant'

    // Auto-expand reasoning if it's currently streaming/loading
    const shouldExpandReasoning = isStreaming || (isLoading && !mainContent && !!reasoning)

    return (
        <div className={`group flex gap-4 ${isUser ? 'justify-end' : 'justify-start'} animate-in fade-in slide-in-from-bottom-2 duration-300`}>
            {/* Avatar - Assistant */}
            {isAssistant && (
                <div className="w-10 h-10 rounded-full bg-linear-to-br from-primary/20 to-primary/10 flex items-center justify-center border-2 border-primary/30 shrink-0 mt-1 shadow-sm">
                    <span className="text-xl">🤖</span>
                </div>
            )}

            {/* Message Column */}
            <div className={`flex flex-col gap-2 ${isUser ? 'max-w-[75%]' : 'max-w-[85%]'}`}>
                {/* Bubble */}
                <div
                    className={`rounded-2xl shadow-md transition-all duration-200 ${isEditing
                        ? 'bg-card border-2 border-primary w-full max-w-full p-4'
                        : isUser
                            ? 'bg-linear-to-br from-primary to-primary/90 text-primary-foreground rounded-br-md p-4 shadow-primary/20'
                            : 'bg-card border border-border/50 rounded-bl-md p-5 hover:shadow-lg hover:border-border'
                        }`}
                >
                    {isEditing ? (
                        <div className="space-y-3">
                            <Textarea
                                value={editContent}
                                onChange={(e) => onEditContentChange(e.target.value)}
                                className="min-h-[100px] bg-transparent text-foreground resize-none"
                                autoFocus
                            />
                            <div className="flex justify-end gap-2">
                                <Button size="sm" variant="ghost" onClick={onCancelEdit}>
                                    Cancelar
                                </Button>
                                <Button size="sm" onClick={() => onSaveEdit(index)}>
                                    Salvar e Enviar
                                </Button>
                            </div>
                        </div>
                    ) : (
                        <>
                            {/* Reasoning Section (Assistant only) */}
                            {isAssistant && reasoning && (
                                <ReasoningSection
                                    reasoning={reasoning}
                                    defaultExpanded={shouldExpandReasoning}
                                />
                            )}

                            {/* Main Content */}
                            {mainContent ? (
                                <div className={`prose prose-sm dark:prose-invert max-w-none wrap-break-word ${isUser ? 'text-primary-foreground' : ''}`}>
                                    {renderContent(mainContent)}
                                </div>
                            ) : (
                                // While loading and empty main content, preserve height or show nothing if reasoning handles it
                                isLoading && !reasoning ? <div className="h-4"></div> : null
                            )}

                            {/* Loading Indicator for Main Content */}
                            {isLoading && !mainContent && !reasoning && (
                                <div className="flex gap-1 h-6 items-center">
                                    <div className="w-2 h-2 bg-primary/50 rounded-full animate-bounce"></div>
                                    <div className="w-2 h-2 bg-primary/50 rounded-full animate-bounce [animation-delay:-.3s]"></div>
                                    <div className="w-2 h-2 bg-primary/50 rounded-full animate-bounce [animation-delay:-.5s]"></div>
                                </div>
                            )}

                            {/* Undo button for actions */}
                            {message.hasActions && onUndo && (
                                <div className="mt-4 pt-4 border-t border-border/50 flex justify-end">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={onUndo}
                                        className="text-xs h-8 gap-2"
                                    >
                                        <span>↩️</span>
                                        Desfazer Alteração
                                    </Button>
                                </div>
                            )}
                        </>
                    )}
                </div>

                {/* Message Actions - Show on hover */}
                {!isLoading && !isEditing && (
                    <div className="flex items-center gap-1 px-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                        {/* Edit (User only) */}
                        {isUser && (
                            <button
                                onClick={() => onStartEdit(index, message.content)}
                                className="p-2 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                                title="Editar mensagem"
                            >
                                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                </svg>
                            </button>
                        )}

                        {/* Copy */}
                        <button
                            onClick={() => onCopy(message.content)}
                            className="p-2 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                            title="Copiar"
                        >
                            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                            </svg>
                        </button>

                        {/* Share */}
                        <button
                            onClick={() => onShare(message.content)}
                            className="p-2 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                            title="Compartilhar"
                        >
                            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
                            </svg>
                        </button>

                        {/* Regenerate (Assistant only, last message) */}
                        {isAssistant && isLastAssistant && (
                            <button
                                onClick={onRegenerate}
                                className="p-2 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                                title="Regenerar resposta"
                            >
                                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                                </svg>
                            </button>
                        )}
                    </div>
                )}
            </div>

            {/* Avatar - User (Right side) */}
            {isUser && (
                <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center border-2 border-primary/10 shrink-0 mt-1">
                    <span className="text-xl">👤</span>
                </div>
            )}
        </div>
    )
}
