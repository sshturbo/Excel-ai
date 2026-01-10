// Keyboard Shortcuts Modal - Shows all available shortcuts
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { useEffect, useState } from "react"

interface Shortcut {
    key: string
    description: string
    category: string
}

const shortcuts: Shortcut[] = [
    // Chat
    { key: "Ctrl + Enter", description: "Enviar mensagem", category: "Chat" },
    { key: "Ctrl + K", description: "Abrir atalhos de teclado", category: "Chat" },
    { key: "Esc", description: "Cancelar envio de mensagem", category: "Chat" },

    // Actions
    { key: "Y", description: "Aceitar ação pendente", category: "Ações" },
    { key: "N", description: "Rejeitar ação pendente", category: "Ações" },
    { key: "Ctrl + Z", description: "Desfazer última alteração", category: "Ações" },

    // Navigation
    { key: "Ctrl + N", description: "Nova conversa", category: "Navegação" },
    { key: "Ctrl + Shift + C", description: "Abrir configurações", category: "Navegação" },

    // Excel
    { key: "Ctrl + E", description: "Conectar ao Excel", category: "Excel" },
    { key: "Ctrl + R", description: "Recarregar planilhas", category: "Excel" },

    // View
    { key: "Ctrl + P", description: "Mostrar/ocultar preview de dados", category: "Visualização" },
    { key: "Ctrl + G", description: "Mostrar/ocultar gráfico", category: "Visualização" },
    { key: "F11", description: "Modo tela cheia", category: "Visualização" },

    // Theme
    { key: "Ctrl + T", description: "Alternar tema (claro/escuro)", category: "Tema" },
]

interface KeyboardShortcutsProps {
    onNewConversation?: () => void
    onOpenSettings?: () => void
    onConnectExcel?: () => void
    onRefreshWorkbooks?: () => void
    onTogglePreview?: () => void
    onToggleChart?: () => void
    onToggleTheme?: () => void
    onUndo?: () => void
    onToggleFullscreen?: () => void
}

export function KeyboardShortcuts({
    onNewConversation,
    onOpenSettings,
    onConnectExcel,
    onRefreshWorkbooks,
    onTogglePreview,
    onToggleChart,
    onToggleTheme,
    onUndo,
    onToggleFullscreen
}: KeyboardShortcutsProps) {
    const [open, setOpen] = useState(false)

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            // Ignore if user is typing in an input or textarea
            if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
            
            // Ctrl + K to open shortcuts modal
            if (e.ctrlKey && e.key.toLowerCase() === 'k') {
                e.preventDefault()
                setOpen(true)
            }

            // F11 to toggle fullscreen
            if (e.key === 'F11') {
                e.preventDefault()
                onToggleFullscreen?.()
            }
        }

        window.addEventListener('keydown', handleKeyDown)
        return () => window.removeEventListener('keydown', handleKeyDown)
    }, [onToggleFullscreen])

    const categories = Array.from(new Set(shortcuts.map(s => s.category)))

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button
                    variant="ghost"
                    size="sm"
                    className="text-xs text-muted-foreground"
                    onClick={() => setOpen(true)}
                >
                    <span className="mr-1">⌨️</span>
                    Atalhos
                </Button>
            </DialogTrigger>

            <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <span className="text-2xl">⌨️</span>
                        Atalhos de Teclado
                    </DialogTitle>
                    <DialogDescription>
                        Use esses atalhos para navegar e usar o aplicativo mais rapidamente
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-6 mt-4">
                    {categories.map(category => (
                        <div key={category}>
                            <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3">
                                {category}
                            </h3>
                            <div className="space-y-2">
                                {shortcuts
                                    .filter(s => s.category === category)
                                    .map((shortcut, idx) => (
                                        <div
                                            key={idx}
                                            className="flex items-center justify-between p-3 rounded-lg bg-muted/30 hover:bg-muted/60 transition-colors"
                                        >
                                            <span className="text-sm">{shortcut.description}</span>
                                            <kbd className="px-3 py-1.5 text-xs font-mono bg-background border border-border rounded-md shadow-sm">
                                                {shortcut.key}
                                            </kbd>
                                        </div>
                                    ))}
                            </div>
                        </div>
                    ))}
                </div>

                <div className="mt-6 pt-4 border-t border-border">
                    <p className="text-xs text-muted-foreground text-center">
                        💡 Dica: Pressione <kbd className="px-1.5 py-0.5 bg-background border border-border rounded">Esc</kbd> para fechar este modal
                    </p>
                </div>
            </DialogContent>
        </Dialog>
    )
}
