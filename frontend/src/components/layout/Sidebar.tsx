// Sidebar component - Workbooks list and conversation history with search
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
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
    isLoadingConversations?: boolean
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
    isLoadingConversations = false,
    onLoadConversations,
    onLoadConversation,
    onDeleteConversation
}: SidebarProps) {
    const [searchQuery, setSearchQuery] = useState('')

    // Filter conversations based on search query
    const filteredConversations = conversations.filter(conv => {
        if (!searchQuery.trim()) return true
        const query = searchQuery.toLowerCase()
        return (
            (conv.title && conv.title.toLowerCase().includes(query)) ||
            (conv.preview && conv.preview.toLowerCase().includes(query))
        )
    })

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
            <div className="p-4 flex-1 overflow-hidden flex flex-col">
                <button
                    onClick={onLoadConversations}
                    className="w-full flex items-center justify-between text-xs font-semibold uppercase text-muted-foreground mb-3 hover:text-foreground transition-colors"
                >
                    <span>💬 Histórico</span>
                    <span className="text-xs flex items-center gap-1">
                        {isLoadingConversations ? '⏳' : '🔄'}
                        {conversations.length > 0 && !isLoadingConversations && (
                            <span className="bg-primary/20 text-primary px-1.5 py-0.5 rounded text-[10px]">
                                {conversations.length}
                            </span>
                        )}
                    </span>
                </button>

                {/* Search Input */}
                <div className="mb-3">
                    <Input
                        placeholder="🔍 Buscar conversas..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="h-8 text-sm"
                    />
                </div>

                {/* Conversations List */}
                <div className="space-y-2 overflow-y-auto flex-1">
                    {isLoadingConversations ? (
                        // Skeleton loading state
                        <>
                            <div className="p-3 bg-muted/30 border border-border rounded">
                                <Skeleton className="h-4 w-3/4 mb-2" />
                                <Skeleton className="h-3 w-1/2" />
                            </div>
                            <div className="p-3 bg-muted/30 border border-border rounded">
                                <Skeleton className="h-4 w-2/3 mb-2" />
                                <Skeleton className="h-3 w-1/3" />
                            </div>
                            <div className="p-3 bg-muted/30 border border-border rounded">
                                <Skeleton className="h-4 w-4/5 mb-2" />
                                <Skeleton className="h-3 w-1/2" />
                            </div>
                        </>
                    ) : filteredConversations.length === 0 ? (
                        <p className="text-center text-muted-foreground text-sm py-4">
                            {searchQuery.trim()
                                ? 'Nenhuma conversa encontrada'
                                : conversations.length === 0
                                    ? 'Clique em "Atualizar" para carregar'
                                    : 'Nenhuma conversa'
                            }
                        </p>
                    ) : (
                        filteredConversations.map(conv => (
                            <div
                                key={conv.id}
                                onClick={() => onLoadConversation(conv.id)}
                                className="group p-2 bg-muted/30 border border-border rounded text-sm cursor-pointer hover:bg-muted/60 hover:border-primary/50 transition-all"
                            >
                                <div className="flex items-center justify-between gap-2">
                                    <div className="flex-1 min-w-0">
                                        {/* Highlight search terms */}
                                        <div className="truncate font-medium">
                                            {searchQuery && conv.title
                                                ? highlightText(conv.title, searchQuery)
                                                : conv.title || 'Sem título'
                                            }
                                        </div>
                                        <div className="text-xs text-muted-foreground truncate">
                                            {conv.preview
                                                ? searchQuery
                                                    ? highlightText(conv.preview, searchQuery)
                                                    : conv.preview
                                                : conv.updatedAt
                                            }
                                        </div>
                                    </div>
                                    <button
                                        onClick={(e) => onDeleteConversation(conv.id, e)}
                                        className="opacity-0 group-hover:opacity-100 p-1 hover:bg-destructive/20 rounded text-destructive transition-opacity shrink-0"
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

// Helper function to highlight search terms
function highlightText(text: string, query: string): React.ReactNode {
    if (!query || !text) return text

    const parts = text.split(new RegExp(`(${query})`, 'gi'))
    
    return parts.map((part, i) => 
        part.toLowerCase() === query.toLowerCase() ? (
            <mark key={i} className="bg-yellow-300/50 text-foreground rounded px-0.5">
                {part}
            </mark>
        ) : (
            <span key={i}>{part}</span>
        )
    )
}
