// PendingActions component - Banner for pending Excel actions
import { Button } from "@/components/ui/button"

interface PendingActionsProps {
    count: number
    onApply: () => void
    onDiscard: () => void
}

export function PendingActions({ count, onApply, onDiscard }: PendingActionsProps) {
    if (count === 0) return null

    return (
        <div className="px-6 py-2 bg-yellow-500/10 border-b border-yellow-500/20 flex items-center justify-between animate-in slide-in-from-top-2">
            <div className="flex items-center gap-2 text-sm text-yellow-500">
                <span>⚠️</span>
                <span>A IA sugeriu <strong>{count}</strong> alterações na planilha.</span>
            </div>
            <div className="flex items-center gap-2">
                <Button size="sm" variant="ghost" onClick={onDiscard} className="text-muted-foreground hover:text-destructive">
                    Descartar
                </Button>
                <Button size="sm" onClick={onApply} className="bg-yellow-500 hover:bg-yellow-600 text-black">
                    Aplicar Alterações
                </Button>
            </div>
        </div>
    )
}
