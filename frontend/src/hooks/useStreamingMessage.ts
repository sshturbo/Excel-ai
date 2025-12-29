// Hook for managing streaming message state
import { useState, useEffect, useRef } from 'react'
import { processStreamingContent } from '@/services/contentCleaner'
import { EventsOn } from "../../wailsjs/runtime/runtime"

interface UseStreamingMessageReturn {
    streamingContent: string
    resetStreamBuffer: () => void
}

export function useStreamingMessage(isLoading: boolean): UseStreamingMessageReturn {
    const [streamingContent, setStreamingContent] = useState('')
    const rawStreamBufferRef = useRef('')

    useEffect(() => {
        const cleanup = EventsOn("chat:chunk", (chunk: string) => {
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

    // Reset buffer when not loading
    useEffect(() => {
        if (!isLoading) {
            rawStreamBufferRef.current = ''
        }
    }, [isLoading])

    const resetStreamBuffer = () => {
        rawStreamBufferRef.current = ''
        setStreamingContent('')
    }

    return {
        streamingContent,
        resetStreamBuffer
    }
}
