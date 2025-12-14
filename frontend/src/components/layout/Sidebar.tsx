// Sidebar component - Workbooks list and conversation history
import type { Workbook, ConversationItem } from '@/types'

interface SidebarProps {
    // Workbooks
    workbooks: Workbook[]
    connected: boolean
    selectedWorkbook: string | null
    selectedSheets: string[]
    expandedWorkbook: string | null
    contextLoaded: string
    onExpandWorkbook: (name: string | null) => void
    onSelectSheet: (wbName: string, sheetName: string) => void
    // Conversations
    conversations: ConversationItem[]
    onLoadConversations: () => void
    onLoadConversation: (convId: string) => void
    onDeleteConversation: (convId: string, e: React.MouseEvent) => void
}

export function Sidebar({
    workbooks,
    connected,
    selectedWorkbook,
    selectedSheets,
    expandedWorkbook,
    contextLoaded,
    onExpandWorkbook,
    onSelectSheet,
    conversations,
    onLoadConversations,
    onLoadConversation,
    onDeleteConversation
}: SidebarProps) {
    return (
        <aside className="w-72 bg-card border-r border-border flex flex-col overflow-hidden">
            {/* Workbooks */}
            <div className="p-4 border-b border-border">
                <h3 className="text-xs font-semibold uppercase text-muted-foreground mb-3">📗 Planilhas</h3>
                <div className="space-y-2 max-h-60 overflow-y-auto">
                    {workbooks.length > 0 ? workbooks.map(wb => (
                        <div key={wb.name} className="rounded-lg overflow-hidden border border-border bg-muted/30">
                            <button
                                onClick={() => onExpandWorkbook(expandedWorkbook === wb.name ? null : wb.name)}
                                className="w-full flex items-center gap-2 p-3 hover:bg-muted/60 transition-colors"
                            >
                                <span>📓</span>
                                <span className="flex-1 text-left text-sm truncate">{wb.name}</span>
                                <span className="text-xs text-muted-foreground bg-background/50 px-2 py-0.5 rounded-full border border-border">
                                    {wb.sheets?.length || 0}
                                </span>
                            </button>
                            {expandedWorkbook === wb.name && (
                                <div className="bg-background/40 border-t border-border max-h-40 overflow-y-auto">
                                    <div className="px-4 py-1 text-xs text-muted-foreground border-b border-border">
                                        💡 Clique para selecionar múltiplas abas
                                    </div>
                                    {wb.sheets?.map((sheet: string) => {
                                        const isSelected = selectedWorkbook === wb.name && selectedSheets.includes(sheet)
                                        return (
                                            <button
                                                key={sheet}
                                                onClick={() => onSelectSheet(wb.name, sheet)}
                                                className={`w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted/60 transition-colors border-l-2 ${isSelected
                                                    ? 'border-l-primary bg-muted/60'
                                                    : 'border-transparent'
                                                    }`}
                                            >
                                                <span className="opacity-70">{isSelected ? '☑️' : '📄'}</span>
                                                <span className="flex-1 text-left">{sheet}</span>
                                                {isSelected && (
                                                    <span className="text-emerald-500">✓</span>
                                                )}
                                            </button>
                                        )
                                    })}
                                </div>
                            )}
                        </div>
                    )) : (
                        <p className="text-center text-muted-foreground text-sm py-4">
                            {connected ? 'Nenhuma planilha' : 'Conecte ao Excel'}
                        </p>
                    )}
                </div>
                {contextLoaded && (
                    <div className="mt-3 p-2 bg-primary/10 border border-primary/30 rounded text-xs text-primary">
                        ✅ {contextLoaded}
                    </div>
                )}
            </div>

            {/* History */}
            <div className="p-4 flex-1 overflow-hidden">
                <button
                    onClick={onLoadConversations}
                    className="w-full flex items-center justify-between text-xs font-semibold uppercase text-muted-foreground mb-3 hover:text-foreground"
                >
                    <span>💬 Histórico</span>
                    <span className="text-xs">🔄 Atualizar</span>
                </button>
                <div className="space-y-2 overflow-y-auto max-h-48">
                    {conversations.length === 0 ? (
                        <p className="text-center text-muted-foreground text-sm py-4">
                            Clique em "Atualizar" para carregar
                        </p>
                    ) : (
                        conversations.slice(0, 10).map(conv => (
                            <div
                                key={conv.id}
                                onClick={() => onLoadConversation(conv.id)}
                                className="group p-2 bg-muted/30 border border-border rounded text-sm cursor-pointer hover:bg-muted/60 hover:border-primary/50 transition-all"
                            >
                                <div className="flex items-center justify-between gap-2">
                                    <div className="flex-1 min-w-0">
                                        <div className="truncate font-medium">{conv.title || 'Sem título'}</div>
                                        <div className="text-xs text-muted-foreground">{conv.updatedAt}</div>
                                    </div>
                                    <button
                                        onClick={(e) => onDeleteConversation(conv.id, e)}
                                        className="opacity-0 group-hover:opacity-100 p-1 hover:bg-destructive/20 rounded text-destructive transition-opacity"
                                        title="Excluir conversa"
                                    >
                                        🗑️
                                    </button>
                                </div>
                            </div>
                        ))
                    )}
                </div>
            </div>
        </aside>
    )
}
