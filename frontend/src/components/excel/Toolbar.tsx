// Toolbar component - Professional preview and chart controls with sheet tabs
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

interface ToolbarProps {
    showPreview: boolean
    showChart: boolean
    onTogglePreview: () => void
    onToggleChart: () => void
    selectedSheets: string[]
    activeSheet: string | null
    onSwitchSheet: (sheetName: string) => void
    onRefresh?: () => void
    onDownload?: () => void // Legacy download (byte array)
    onOpenNative?: () => void
    onSaveNative?: () => void
    isSaving?: boolean
}

export function Toolbar({
    showPreview,
    showChart,
    onTogglePreview,
    onToggleChart,
    selectedSheets,
    activeSheet,
    onSwitchSheet,
    onRefresh,
    onDownload,
    onOpenNative,
    onSaveNative,
    isSaving
}: ToolbarProps) {
    return (
        <div className="flex items-center gap-2 px-4 py-2 bg-card/80 backdrop-blur-sm border-b border-border">
            {/* Left side - View toggles */}
            <div className="flex items-center gap-1 bg-muted/50 rounded-lg p-1">
                <Button
                    variant={showPreview ? "default" : "ghost"}
                    size="sm"
                    onClick={onTogglePreview}
                    className="h-8 gap-2"
                >
                    <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <rect x="3" y="3" width="18" height="18" rx="2" />
                        <path d="M3 9h18M9 21V9" />
                    </svg>
                    Dados
                </Button>
                <Button
                    variant={showChart ? "default" : "ghost"}
                    size="sm"
                    onClick={onToggleChart}
                    className="h-8 gap-2"
                >
                    <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M18 20V10M12 20V4M6 20v-6" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                    Gráfico
                </Button>
            </div>

            {/* Center - Sheet tabs */}
            {selectedSheets.length > 0 && (
                <div className="flex items-center gap-1 ml-4">
                    <svg className="w-4 h-4 text-muted-foreground mr-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M14.5 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V7.5L14.5 2z" />
                        <path d="M14 2v6h6" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>

                    {selectedSheets.length === 1 ? (
                        // Single sheet - just show name
                        <span className="text-sm font-medium text-foreground px-2">
                            {selectedSheets[0]}
                        </span>
                    ) : (
                        // Multiple sheets - show tabs
                        <div className="flex items-center gap-0.5 bg-muted/50 rounded-md p-0.5">
                            {selectedSheets.map((sheet) => (
                                <button
                                    key={sheet}
                                    onClick={() => onSwitchSheet(sheet)}
                                    className={cn(
                                        "px-3 py-1 text-sm rounded-sm transition-all",
                                        activeSheet === sheet
                                            ? "bg-background text-foreground font-medium shadow-sm"
                                            : "text-muted-foreground hover:text-foreground hover:bg-background/50"
                                    )}
                                >
                                    {sheet}
                                </button>
                            ))}
                        </div>
                    )}
                </div>
            )}

            {/* Right side - Actions */}
            <div className="flex items-center gap-2 ml-auto">
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={onOpenNative}
                    className="h-8 gap-2"
                    title="Abrir outro arquivo Excel"
                >
                    <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                    </svg>
                    Abrir
                </Button>

                {onRefresh && (
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={onRefresh}
                        className="h-8 gap-2"
                        title="Atualizar dados"
                    >
                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <path d="M21 12a9 9 0 11-9-9c2.52 0 4.93 1 6.74 2.74L21 8" strokeLinecap="round" strokeLinejoin="round" />
                            <path d="M21 3v5h-5" strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                        Atualizar
                    </Button>
                )}

                {/* Legacy download button kept for compatibility but hidden if save is priority */}
                {onDownload && !onSaveNative && (
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={onDownload}
                        className="h-8 gap-2"
                    >
                        Download
                    </Button>
                )}

                {onSaveNative && (
                    <Button
                        variant="default"
                        size="sm"
                        onClick={onSaveNative}
                        disabled={isSaving}
                        className="h-8 gap-2 bg-green-600 hover:bg-green-700 text-white"
                        title="Salvar alterações no arquivo original (Sobrescrever)"
                    >
                        {isSaving ? (
                            <svg className="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                <path d="M21 12a9 9 0 11-6.219-8.56" strokeLinecap="round" />
                            </svg>
                        ) : (
                            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                <path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z" />
                                <path d="M17 21v-8H7v8" />
                                <path d="M7 3v5h8" />
                            </svg>
                        )}
                        Salvar
                    </Button>
                )}

                {(showPreview || showChart) && (
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                            if (showPreview) onTogglePreview()
                            if (showChart) onToggleChart()
                        }}
                        className="h-8 gap-2"
                    >
                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <path d="M18 6L6 18M6 6l12 12" strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                        Fechar
                    </Button>
                )}
            </div>
        </div>
    )
}
