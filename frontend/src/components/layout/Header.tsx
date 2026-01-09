// Header component - Top navigation bar with theme toggle
import { Button } from "@/components/ui/button"

interface HeaderProps {
    connected: boolean
    theme: 'light' | 'dark'
    onNewConversation: () => void
    onOpenSettings: () => void
    onConnect: () => void
    onToggleTheme: () => void
}

export function Header({
    connected,
    theme,
    onNewConversation,
    onOpenSettings,
    onConnect,
    onToggleTheme
}: HeaderProps) {
    return (
        <header className="flex items-center justify-between px-6 py-4 border-b border-border bg-card/60 backdrop-blur supports-backdrop-filter:bg-card/40">
            <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-primary text-primary-foreground flex items-center justify-center text-xl">
                    📊
                </div>
                <span className="text-xl font-semibold tracking-tight bg-linear-to-r from-primary to-blue-500 bg-clip-text text-transparent">
                    HipoSystem
                </span>
            </div>

            <div className="flex items-center gap-3">
                {/* Theme toggle */}
                <button
                    onClick={onToggleTheme}
                    className="p-2 rounded-lg bg-muted/50 hover:bg-muted transition-colors"
                    title={theme === 'dark' ? 'Tema claro' : 'Tema escuro'}
                >
                    {theme === 'dark' ? '☀️' : '🌙'}
                </button>

                <Button
                    onClick={() => {
                        console.log('[HEADER] Botão Nova Conversa clicado')
                        onNewConversation()
                    }}
                    variant="default"
                >
                    ➕ Nova Conversa
                </Button>
                <Button onClick={onOpenSettings} variant="secondary">
                    ⚙️ Configurações
                </Button>
                <div className="flex items-center gap-2 px-3 py-1.5 bg-muted/50 border border-border rounded-full text-sm">
                    <span className={`w-2 h-2 rounded-full ${connected ? 'bg-emerald-500' : 'bg-rose-500'}`}></span>
                    <span>{connected ? 'Conectado' : 'Desconectado'}</span>
                    <button onClick={onConnect} className="hover:text-foreground transition-colors">🔄</button>
                </div>
            </div>
        </header>
    )
}
