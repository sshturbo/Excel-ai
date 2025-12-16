// PendingActions component - Banner for pending Excel actions
import { Button } from "@/components/ui/button"

interface PendingActionsProps {
    count: number
    validationMode?: boolean
    onApply: () => void
    onDiscard: () => void
    onKeep?: () => void
    onUndo?: () => void
}

export function PendingActions({ count, validationMode, onApply, onDiscard, onKeep, onUndo }: PendingActionsProps) {
    if (count === 0 && !validationMode) return null

    if (validationMode) {
        return (
            <div className="px-6 py-2 bg-green-500/10 border-b border-green-500/20 flex items-center justify-between animate-in slide-in-from-top-2">
                <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
                    <span>✅</span>
                    <span>Alterações aplicadas! Deseja manter?</span>
                </div>
                <div className="flex items-center gap-2">
                    <Button size="sm" variant="ghost" onClick={onUndo} className="text-muted-foreground hover:text-destructive">
                        Desfazer
                    </Button>
                    <Button size="sm" onClick={onKeep} className="bg-green-500 hover:bg-green-600 text-white">
                        Manter Alterações
                    </Button>
                </div>
            </div>
        )
    }

    return (
        <div className="px-6 py-2 bg-yellow-500/10 border-b border-yellow-500/20 flex items-center justify-between animate-in slide-in-from-top-2">
            <div className="flex items-center gap-2 text-sm text-yellow-600 dark:text-yellow-400">
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
