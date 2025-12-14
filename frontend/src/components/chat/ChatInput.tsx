// ChatInput component - Message input with send/cancel buttons
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import React from "react"

interface ChatInputProps {
    inputMessage: string
    isLoading: boolean
    inputRef: React.RefObject<HTMLTextAreaElement>
    onInputChange: (value: string) => void
    onSend: () => void
    onCancel: () => void
}

export function ChatInput({
    inputMessage,
    isLoading,
    inputRef,
    onInputChange,
    onSend,
    onCancel
}: ChatInputProps) {
    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault()
            onSend()
        }
    }

    return (
        <div className="p-4 bg-card/60 border-t border-border">
            <div className="flex gap-3">
                <Textarea
                    ref={inputRef}
                    value={inputMessage}
                    onChange={(e) => onInputChange(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder="Pergunte sobre seus dados..."
                    className="flex-1 min-h-13 max-h-36 resize-none"
                    disabled={isLoading}
                />
                {isLoading ? (
                    <Button
                        onClick={onCancel}
                        variant="destructive"
                        size="icon-lg"
                        className="rounded-lg"
                        title="Parar"
                    >
                        ⏹️
                    </Button>
                ) : (
                    <Button
                        onClick={onSend}
                        disabled={!inputMessage.trim()}
                        size="icon-lg"
                        className="rounded-lg"
                    >
                        ➤
                    </Button>
                )}
            </div>
        </div>
    )
}
