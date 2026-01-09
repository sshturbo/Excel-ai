// ReasoningSection - Collapsible reasoning/thinking section
import React, { useState } from 'react'
import { MarkdownRenderer } from '@/components/markdown/MarkdownRenderer'

interface ReasoningSectionProps {
    reasoning: string
    title?: string
    defaultExpanded?: boolean
}

export function ReasoningSection({
    reasoning,
    title = "Raciocínio",
    defaultExpanded = false
}: ReasoningSectionProps) {
    const [isExpanded, setIsExpanded] = useState(defaultExpanded)

    // Update expansion state if defaultExpanded changes (e.g. when loading starts/stops)
    React.useEffect(() => {
        setIsExpanded(defaultExpanded)
    }, [defaultExpanded])

    if (!reasoning) return null

    return (
        <div className="mb-4 border-l-4 border-primary/30 bg-muted/20 rounded-r-lg overflow-hidden">
            {/* Header - Clickable to expand/collapse */}
            <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="w-full flex items-center gap-3 p-3 hover:bg-muted/40 transition-colors text-left"
            >
                {/* Brain icon */}
                <div className="text-xl shrink-0">🧠</div>

                {/* Title */}
                <div className="flex-1 font-medium text-sm text-foreground/90">
                    {title}
                </div>

                {/* Expand/Collapse arrow */}
                <div className={`text-muted-foreground transition-transform ${isExpanded ? 'rotate-180' : ''}`}>
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                    </svg>
                </div>
            </button>

            {/* Content - Collapsible */}
            {isExpanded && (
                <div className="px-3 pb-3 pt-1 text-sm text-muted-foreground animate-in slide-in-from-top-2 fade-in">
                    <div className="prose prose-sm dark:prose-invert max-w-none">
                        <MarkdownRenderer content={reasoning} />
                    </div>
                </div>
            )}
        </div>
    )
}
