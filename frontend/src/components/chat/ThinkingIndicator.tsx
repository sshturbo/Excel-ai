// ThinkingIndicator - Animated thinking indicator with reasoning content display
import React, { useState } from 'react'

interface ThinkingIndicatorProps {
    isThinking: boolean
    thinkingText?: string
    reasoningContent?: string
}

export function ThinkingIndicator({
    isThinking,
    thinkingText = "Pensando...",
    reasoningContent = ""
}: ThinkingIndicatorProps) {
    const [isExpanded, setIsExpanded] = useState(true)

    if (!isThinking) return null

    const hasReasoning = reasoningContent && reasoningContent.trim().length > 0

    return (
        <div className="flex items-start gap-4 p-4 bg-muted/30 rounded-lg border border-border/50 animate-in fade-in slide-in-from-bottom-2">
            {/* Animated brain icon */}
            <div className="relative shrink-0 mt-0.5">
                <div className="text-2xl animate-pulse">🧠</div>
                <div className="absolute inset-0 bg-primary/20 rounded-full blur-md animate-ping"></div>
            </div>

            <div className="flex-1 min-w-0">
                {/* Thinking text with animated dots */}
                <div className="flex items-center gap-2 mb-2">
                    <span className="text-sm font-medium text-foreground">{thinkingText}</span>
                    <div className="flex gap-1">
                        <div className="w-1.5 h-1.5 bg-primary rounded-full animate-[bounce_1s_infinite_-0.3s]"></div>
                        <div className="w-1.5 h-1.5 bg-primary rounded-full animate-[bounce_1s_infinite_-0.15s]"></div>
                        <div className="w-1.5 h-1.5 bg-primary rounded-full animate-bounce"></div>
                    </div>
                </div>

                {/* Reasoning content if available */}
                {hasReasoning && (
                    <div className="mt-3 pt-3 border-t border-border/50">
                        <button
                            onClick={() => setIsExpanded(!isExpanded)}
                            className="flex items-center gap-2 text-xs text-muted-foreground hover:text-foreground transition-colors mb-2"
                        >
                            <svg
                                className={`w-3 h-3 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                strokeWidth={2}
                            >
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                            </svg>
                            <span className="font-medium">Ver raciocínio</span>
                        </button>

                        {isExpanded && (
                            <div className="text-xs text-muted-foreground whitespace-pre-wrap font-mono bg-muted/20 p-3 rounded border-l-2 border-primary/30 animate-in slide-in-from-top-2 fade-in">
                                {reasoningContent}
                            </div>
                        )}
                    </div>
                )}

                {/* Debug info when no reasoning */}
                {!hasReasoning && (
                    <div className="text-xs text-muted-foreground/60 italic">
                        Aguardando resposta do modelo...
                    </div>
                )}
            </div>
        </div>
    )
}
