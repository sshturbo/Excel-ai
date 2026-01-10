// ChatInput component - ChatGPT-style input with rounded pill design
import { Button } from "@/components/ui/button"
import React from "react"

interface ChatInputProps {
    inputMessage: string
    isLoading: boolean
    inputRef: React.RefObject<HTMLTextAreaElement>
    onInputChange: (value: string) => void
    onSend: () => void
    onCancel: () => void
    onFileUpload?: (file: File) => void
}

export function ChatInput({
    inputMessage,
    isLoading,
    inputRef,
    onInputChange,
    onSend,
    onCancel,
    onFileUpload
}: ChatInputProps) {
    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault()
            onSend()
        }
    }

    // Reset height when message is cleared
    React.useEffect(() => {
        if (inputRef.current && inputMessage === '') {
            inputRef.current.style.height = 'auto'
        }
    }, [inputMessage, inputRef])

    const hasText = inputMessage.trim().length > 0

    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0]
        if (file && onFileUpload) {
            onFileUpload(file)
        }
        // Reset input
        e.target.value = ''
    }

    return (
        <div className="p-4 bg-transparent">
            <div className="max-w-4xl mx-auto">
                <div className="relative flex items-center bg-background border border-border rounded-full shadow-sm hover:shadow-md transition-shadow">
                    {/* Plus button (left) */}
                    <div className="relative shrink-0">
                        <input
                            type="file"
                            accept=".xlsx,.xls"
                            onChange={handleFileUpload}
                            className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                            disabled={isLoading}
                        />
                        <button
                            className="p-3 pl-4 text-muted-foreground hover:text-foreground transition-colors pointer-events-none"
                            title="Anexar arquivo"
                        >
                            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                            </svg>
                        </button>
                    </div>

                    {/* Input textarea */}
                    <textarea
                        ref={inputRef}
                        value={inputMessage}
                        onChange={(e) => onInputChange(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder="Pergunte alguma coisa"
                        className="flex-1 bg-transparent border-none outline-none resize-none py-3 px-2 text-foreground placeholder:text-muted-foreground min-h-[44px] max-h-32"
                        disabled={isLoading}
                        rows={1}
                        style={{
                            overflow: 'hidden',
                            height: 'auto'
                        }}
                        onInput={(e) => {
                            const target = e.target as HTMLTextAreaElement
                            target.style.height = 'auto'
                            target.style.height = Math.min(target.scrollHeight, 128) + 'px'
                        }}
                    />

                    {/* Right side buttons */}
                    <div className="flex items-center gap-1 pr-2">
                        {/* Microphone button (decorative) */}
                        <button
                            className="p-2 text-muted-foreground hover:text-foreground transition-colors rounded-full"
                            title="Entrada por voz"
                        >
                            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
                            </svg>
                        </button>

                        {/* Send/Stop button */}
                        {isLoading ? (
                            <Button
                                onClick={onCancel}
                                size="icon"
                                className="rounded-full w-9 h-9 bg-destructive hover:bg-destructive/90"
                                title="Parar"
                            >
                                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                                    <rect x="6" y="6" width="12" height="12" rx="1" />
                                </svg>
                            </Button>
                        ) : (
                            <Button
                                onClick={onSend}
                                disabled={!hasText}
                                size="icon"
                                className={`rounded-full w-9 h-9 transition-all ${hasText
                                    ? 'bg-foreground hover:bg-foreground/90 text-background'
                                    : 'bg-muted text-muted-foreground cursor-not-allowed'
                                    }`}
                                title="Enviar"
                            >
                                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 10l7-7m0 0l7 7m-7-7v18" />
                                </svg>
                            </Button>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}
