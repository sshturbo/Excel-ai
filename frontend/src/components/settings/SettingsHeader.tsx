// SettingsHeader component
import { Button } from "@/components/ui/button"

interface SettingsHeaderProps {
    onClose: () => void
    onSave: () => void
    isSaving: boolean
}

export function SettingsHeader({ onClose, onSave, isSaving }: SettingsHeaderProps) {
    return (
        <header className="flex items-center justify-between mb-8 pb-6 border-b border-border">
            <Button variant="ghost" onClick={onClose} className="gap-2">
                ← Voltar
            </Button>
            <div className="flex items-center gap-3">
                <span className="text-3xl">⚙️</span>
                <h1 className="text-2xl font-semibold tracking-tight">
                    Configurações
                </h1>
            </div>
            <Button onClick={onSave} disabled={isSaving} className="gap-2">
                {isSaving ? (
                    <>
                        <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                        Salvando
                    </>
                ) : (
                    <>💾 Salvar</>
                )}
            </Button>
        </header>
    )
}
