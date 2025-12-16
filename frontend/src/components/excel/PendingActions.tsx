// PendingActions component - Banner for pending Excel actions with preview
// Inspired by gemini-cli permission model

import { Button } from "@/components/ui/button"
import type { ExcelAction } from "@/types"

// Action execution states
export type ActionState = 'pending' | 'executing' | 'completed' | 'error'

interface PendingActionsProps {
    actions: ExcelAction[]
    state: ActionState
    error?: string
    hasPendingAction?: boolean  // From backend HasPendingAction()
    onApply: () => void
    onDiscard: () => void
    onKeep?: () => void
    onUndo?: () => void
}

// Helper to describe an action in Portuguese
function describeAction(action: ExcelAction): string {
    const ops: Record<string, string> = {
        'write': 'Escrever dados',
        'create-sheet': 'Criar planilha',
        'create-workbook': 'Criar pasta de trabalho',
        'create-chart': 'Criar gráfico',
        'create-pivot': 'Criar tabela dinâmica',
        'format-range': 'Formatar células',
        'delete-sheet': 'Excluir planilha',
        'rename-sheet': 'Renomear planilha',
        'clear-range': 'Limpar conteúdo',
        'autofit': 'Ajustar colunas',
        'insert-rows': 'Inserir linhas',
        'delete-rows': 'Excluir linhas',
        'merge-cells': 'Mesclar células',
        'unmerge-cells': 'Desmesclar células',
        'set-borders': 'Adicionar bordas',
        'apply-filter': 'Aplicar filtro',
        'sort': 'Ordenar dados',
        'copy-range': 'Copiar intervalo',
        'macro': 'Macro (múltiplas ações)',
    }

    const opName = ops[action.op] || action.op
    const target = action.sheet ? ` em "${action.sheet}"` : ''
    const cell = action.cell ? ` (${action.cell})` : ''

    return `${opName}${target}${cell}`
}

// Helper to count sub-actions in a macro
function countActions(actions: ExcelAction[]): number {
    let count = 0
    for (const action of actions) {
        if (action.op === 'macro' && (action as any).actions) {
            count += (action as any).actions.length
        } else {
            count++
        }
    }
    return count
}

export function PendingActions({
    actions,
    state,
    error,
    hasPendingAction = false,
    onApply,
    onDiscard,
    onKeep,
    onUndo
}: PendingActionsProps) {
    // Show if: we have actions in array OR backend says there's a pending action OR state is not pending
    const shouldShow = actions.length > 0 || hasPendingAction || state !== 'pending'
    if (!shouldShow) return null

    const totalActions = countActions(actions)

    // Estado: Executando
    if (state === 'executing') {
        return (
            <div className="px-6 py-3 bg-blue-500/10 border-b border-blue-500/20 animate-in slide-in-from-top-2">
                <div className="flex items-center gap-3 text-sm text-blue-600 dark:text-blue-400">
                    <div className="animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full" />
                    <span>Executando {totalActions} ação(ões)...</span>
                </div>
            </div>
        )
    }

    // Estado: Concluído
    if (state === 'completed') {
        return (
            <div className="px-6 py-3 bg-green-500/10 border-b border-green-500/20 flex items-center justify-between animate-in slide-in-from-top-2">
                <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
                    <span>✅</span>
                    <span>Ações executadas! Deseja manter as alterações?</span>
                </div>
                <div className="flex items-center gap-2">
                    {onUndo && (
                        <Button
                            size="sm"
                            variant="ghost"
                            onClick={onUndo}
                            className="text-muted-foreground hover:text-destructive"
                        >
                            Desfazer
                        </Button>
                    )}
                    {onKeep && (
                        <Button
                            size="sm"
                            onClick={onKeep}
                            className="bg-green-500 hover:bg-green-600 text-white"
                        >
                            Manter Alterações
                        </Button>
                    )}
                </div>
            </div>
        )
    }

    // Estado: Erro
    if (state === 'error') {
        return (
            <div className="px-6 py-3 bg-red-500/10 border-b border-red-500/20 flex items-center justify-between animate-in slide-in-from-top-2">
                <div className="flex items-center gap-2 text-sm text-red-600 dark:text-red-400">
                    <span>❌</span>
                    <span>Erro: {error || 'Falha ao executar ação'}</span>
                </div>
                <Button
                    size="sm"
                    variant="ghost"
                    onClick={onDiscard}
                    className="text-muted-foreground"
                >
                    Fechar
                </Button>
            </div>
        )
    }

    // Estado: Pendente (aguardando aprovação)
    return (
        <div className="border-b border-yellow-500/20 animate-in slide-in-from-top-2">
            {/* Header */}
            <div className="px-6 py-3 bg-yellow-500/10 flex items-center justify-between">
                <div className="flex items-center gap-2 text-sm text-yellow-600 dark:text-yellow-400">
                    <span>⚠️</span>
                    <span>
                        <strong>Confirmação necessária:</strong> {totalActions > 0
                            ? `A IA quer executar ${totalActions} ação(ões)`
                            : 'A IA propôs uma alteração'}
                    </span>
                </div>
                <div className="flex items-center gap-2">
                    <Button
                        size="sm"
                        variant="ghost"
                        onClick={onDiscard}
                        className="text-muted-foreground hover:text-destructive"
                    >
                        Descartar (n)
                    </Button>
                    <Button
                        size="sm"
                        onClick={onApply}
                        className="bg-yellow-500 hover:bg-yellow-600 text-black font-medium"
                    >
                        Aplicar (Y)
                    </Button>
                </div>
            </div>

            {/* Preview das ações */}
            <div className="px-6 py-2 bg-muted/30 text-xs space-y-1 max-h-32 overflow-y-auto">
                {actions.map((action, idx) => {
                    // Se for macro, listar sub-ações
                    if (action.op === 'macro' && (action as any).actions) {
                        return (
                            <div key={idx} className="space-y-1">
                                {((action as any).actions as ExcelAction[]).map((subAction, subIdx) => (
                                    <div key={`${idx}-${subIdx}`} className="flex items-center gap-2 text-muted-foreground">
                                        <span className="text-yellow-500">→</span>
                                        <span>{describeAction(subAction)}</span>
                                    </div>
                                ))}
                            </div>
                        )
                    }
                    return (
                        <div key={idx} className="flex items-center gap-2 text-muted-foreground">
                            <span className="text-yellow-500">→</span>
                            <span>{describeAction(action)}</span>
                        </div>
                    )
                })}
            </div>
        </div>
    )
}
