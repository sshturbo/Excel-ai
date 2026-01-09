// Hook for managing streaming message state with reasoning support
import { useState, useEffect, useRef } from 'react'
import { processStreamingContent } from '@/services/contentCleaner'
import { EventsOn } from "../../wailsjs/runtime/runtime"

interface UseStreamingMessageReturn {
    streamingContent: string
    reasoningContent: string
    resetStreamBuffer: () => void
}

export function useStreamingMessage(isLoading: boolean): UseStreamingMessageReturn {
    const [streamingContent, setStreamingContent] = useState('')
    const [reasoningContent, setReasoningContent] = useState('')
    const rawStreamBufferRef = useRef('')
    const reasoningBufferRef = useRef('')

    useEffect(() => {
        const cleanup = EventsOn("chat:chunk", (chunk: string) => {
            // Debug log
            if (chunk.includes('reasoning')) {
                console.log('[REASONING DEBUG] Chunk:', chunk.substring(0, 100))
            }

            // Check if chunk contains reasoning content
            if (chunk.startsWith(':::reasoning:::')) {
                // Extract reasoning content
                const reasoning = chunk.replace(':::reasoning:::', '')
                reasoningBufferRef.current += reasoning
                setReasoningContent(reasoningBufferRef.current)
                console.log('[REASONING] Updated:', reasoningBufferRef.current.substring(0, 100))
                return
            }

            // Regular content
            rawStreamBufferRef.current += chunk

            const { cleanContent, hasIncompleteAction, hasIncompleteQuery, hasIncompleteThinking } =
                processStreamingContent(rawStreamBufferRef.current)

            // If empty but there's technical activity, show status
            let displayContent = cleanContent
            if (!cleanContent && (hasIncompleteAction || hasIncompleteQuery || hasIncompleteThinking)) {
                if (hasIncompleteThinking) {
                    displayContent = '💭 Pensando...'
                } else {
                    displayContent = hasIncompleteQuery ? '🔍 Consultando...' : '⏳ Executando...'
                }
            }

            setStreamingContent(displayContent)
        })
        return () => cleanup()
    }, [])

    // Reset buffers when not loading
    useEffect(() => {
        if (!isLoading) {
            rawStreamBufferRef.current = ''
            reasoningBufferRef.current = ''
            // Clear states immediately when loading stops
            setStreamingContent('')
            setReasoningContent('')
        }
    }, [isLoading])

    const resetStreamBuffer = () => {
        rawStreamBufferRef.current = ''
        reasoningBufferRef.current = ''
        setStreamingContent('')
        setReasoningContent('')
    }

    return {
        streamingContent,
        reasoningContent,
        resetStreamBuffer
    }
}
