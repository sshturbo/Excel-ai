// MessageBubble component - Individual chat message with actions
import React, { ReactNode } from 'react'
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
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
    renderContent
}: MessageBubbleProps) {
    return (
        <div className={`group flex gap-3 ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            {message.role === 'assistant' && (
                <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center border border-primary/20 shrink-0 mt-1">
                    🤖
                </div>
            )}
            <div
                className={`max-w-[85%] p-4 rounded-2xl shadow-sm animate-in fade-in slide-in-from-bottom-2 ${isEditing
                    ? 'bg-card border border-border w-full max-w-full'
                    : (message.role === 'user'
                        ? 'bg-primary text-primary-foreground rounded-br-sm'
                        : 'bg-card border border-border rounded-bl-sm')
                    }`}
            >
                {isEditing ? (
                    <div className="space-y-3">
                        <Textarea
                            value={editContent}
                            onChange={(e) => onEditContentChange(e.target.value)}
                            className="min-h-25 bg-transparent text-foreground"
                        />
                        <div className="flex justify-end gap-2">
                            <Button size="sm" variant="ghost" onClick={onCancelEdit}>Cancelar</Button>
                            <Button size="sm" onClick={() => onSaveEdit(index)}>Salvar e Enviar</Button>
                        </div>
                    </div>
                ) : (
                    <>
                        {message.role === 'assistant' ? (
                            <div className="text-sm relative">
                                {message.content ? (
                                    renderContent(message.content)
                                ) : (
                                    <div className="flex items-center gap-1.5 h-6">
                                        <div className="w-2 h-2 bg-foreground/40 rounded-full animate-[bounce_1s_infinite_-0.3s]"></div>
                                        <div className="w-2 h-2 bg-foreground/40 rounded-full animate-[bounce_1s_infinite_-0.15s]"></div>
                                        <div className="w-2 h-2 bg-foreground/40 rounded-full animate-bounce"></div>
                                    </div>
                                )}
                                {message.hasActions && onUndo && (
                                    <div className="mt-3 pt-3 border-t border-border flex justify-end">
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            onClick={onUndo}
                                            className="text-xs h-7"
                                        >
                                            ↩️ Desfazer Alteração
                                        </Button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <div className="relative">
                                <div className="whitespace-pre-wrap text-sm">{message.content}</div>
                            </div>
                        )}

                        {/* Message Actions Footer */}
                        {!isLoading && (
                            <div className={`flex items-center justify-end gap-1 mt-2 pt-1 opacity-0 group-hover:opacity-100 transition-opacity ${message.role === 'user'
                                ? 'border-t border-primary-foreground/20'
                                : 'border-t border-border'
                                }`}>
                                {message.role === 'user' && (
                                    <button
                                        onClick={() => onStartEdit(index, message.content)}
                                        className="p-1.5 rounded hover:bg-primary-foreground/10 text-primary-foreground/80 hover:text-primary-foreground transition-colors"
                                        title="Editar"
                                    >
                                        ✏️
                                    </button>
                                )}

                                <button
                                    onClick={() => onCopy(message.content)}
                                    className={`p-1.5 rounded transition-colors ${message.role === 'user'
                                        ? 'hover:bg-primary-foreground/10 text-primary-foreground/80 hover:text-primary-foreground'
                                        : 'hover:bg-muted text-muted-foreground hover:text-foreground'
                                        }`}
                                    title="Copiar"
                                >
                                    📋
                                </button>

                                <button
                                    onClick={() => onShare(message.content)}
                                    className={`p-1.5 rounded transition-colors ${message.role === 'user'
                                        ? 'hover:bg-primary-foreground/10 text-primary-foreground/80 hover:text-primary-foreground'
                                        : 'hover:bg-muted text-muted-foreground hover:text-foreground'
                                        }`}
                                    title="Compartilhar"
                                >
                                    📤
                                </button>

                                {message.role === 'assistant' && isLastAssistant && (
                                    <button
                                        onClick={onRegenerate}
                                        className="p-1.5 rounded hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                                        title="Regenerar resposta"
                                    >
                                        🔄
                                    </button>
                                )}
                            </div>
                        )}
                    </>
                )}
            </div>
        </div>
    )
}
