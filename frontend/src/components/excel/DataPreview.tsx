// DataPreview component - Professional spreadsheet-like table preview
import { useState, useMemo, useRef, useEffect } from 'react'
import type { PreviewDataType } from '@/types'
import { cn } from '@/lib/utils'

interface DataPreviewProps {
    previewData: PreviewDataType
    selectedColumns?: number[]
    onColumnSelect?: (columnIndex: number) => void
}

// Convert column index to Excel-style letter (0=A, 1=B, 26=AA, etc.)
function getColumnLetter(index: number): string {
    let letter = ''
    let temp = index
    while (temp >= 0) {
        letter = String.fromCharCode((temp % 26) + 65) + letter
        temp = Math.floor(temp / 26) - 1
    }
    return letter
}

export function DataPreview({ previewData, selectedColumns = [], onColumnSelect }: DataPreviewProps) {
    const containerRef = useRef<HTMLDivElement>(null)
    const [hoveredCell, setHoveredCell] = useState<{ row: number; col: number } | null>(null)

    const totalRows = previewData.rows?.length || 0
    const totalCols = previewData.headers?.length || 0
    const displayRows = Math.min(totalRows, 100) // Max 100 rows for performance

    // Determine column widths based on content
    const columnWidths = useMemo(() => {
        const widths: number[] = []
        const maxWidth = 200
        const minWidth = 60
        const charWidth = 8

        previewData.headers?.forEach((header, i) => {
            let maxLen = header?.length || 3
            // Sample first 20 rows for width calculation
            previewData.rows?.slice(0, 20).forEach(row => {
                const cellLen = String(row[i] || '').length
                if (cellLen > maxLen) maxLen = cellLen
            })
            widths[i] = Math.min(maxWidth, Math.max(minWidth, maxLen * charWidth + 16))
        })
        return widths
    }, [previewData])

    return (
        <div className="flex-1 flex flex-col bg-background overflow-hidden">
            {/* Status Bar Top */}
            <div className="flex items-center justify-between px-3 py-1.5 bg-muted/40 border-b border-border text-xs text-muted-foreground">
                <div className="flex items-center gap-4">
                    <span className="font-medium text-foreground">📊 Visualização de Dados</span>
                    <span>{totalCols} colunas × {totalRows} linhas</span>
                </div>
                <div className="flex items-center gap-3">
                    {selectedColumns.length > 0 && (
                        <span className="bg-primary/20 text-primary px-2 py-0.5 rounded-full">
                            {selectedColumns.length} col. selecionada{selectedColumns.length > 1 ? 's' : ''}
                        </span>
                    )}
                    {hoveredCell && (
                        <span className="font-mono bg-background px-2 py-0.5 rounded border border-border">
                            {getColumnLetter(hoveredCell.col)}{hoveredCell.row + 1}
                        </span>
                    )}
                </div>
            </div>

            {/* Spreadsheet Container */}
            <div ref={containerRef} className="flex-1 overflow-auto">
                <table className="border-collapse text-sm" style={{ minWidth: '100%' }}>
                    {/* Column Headers (A, B, C...) */}
                    <thead className="sticky top-0 z-20">
                        {/* Column letters row */}
                        <tr className="bg-muted/80 backdrop-blur-sm">
                            {/* Corner cell (row number header) */}
                            <th className="sticky left-0 z-30 w-12 min-w-12 bg-muted border border-border text-center text-muted-foreground font-normal">
                                <div className="w-full h-full flex items-center justify-center">
                                    <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                        <path d="M4 4l16 16M20 4l-16 16" strokeLinecap="round" />
                                    </svg>
                                </div>
                            </th>
                            {/* Column letter headers */}
                            {previewData.headers?.map((_, colIndex) => (
                                <th
                                    key={colIndex}
                                    className={cn(
                                        "border border-border text-center font-medium py-1 px-2 cursor-pointer transition-colors select-none",
                                        selectedColumns.includes(colIndex)
                                            ? "bg-primary/30 text-primary"
                                            : "bg-muted/80 text-muted-foreground hover:bg-primary/10"
                                    )}
                                    style={{ width: columnWidths[colIndex], minWidth: columnWidths[colIndex] }}
                                    onClick={() => onColumnSelect?.(colIndex)}
                                    title={`Clique para selecionar coluna ${getColumnLetter(colIndex)}`}
                                >
                                    {getColumnLetter(colIndex)}
                                </th>
                            ))}
                        </tr>
                        {/* Data headers row */}
                        <tr className="bg-muted/60 backdrop-blur-sm">
                            <th className="sticky left-0 z-30 w-12 min-w-12 bg-muted border border-border text-center text-xs text-muted-foreground font-normal py-2">
                                1
                            </th>
                            {previewData.headers?.map((header, colIndex) => (
                                <th
                                    key={colIndex}
                                    className={cn(
                                        "border border-border text-left font-semibold py-2 px-3 truncate",
                                        selectedColumns.includes(colIndex)
                                            ? "bg-primary/20"
                                            : "bg-muted/60"
                                    )}
                                    style={{ maxWidth: columnWidths[colIndex] }}
                                    title={header}
                                >
                                    {header || '—'}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {previewData.rows?.slice(0, displayRows).map((row, rowIndex) => (
                            <tr
                                key={rowIndex}
                                className={cn(
                                    "transition-colors",
                                    rowIndex % 2 === 0 ? "bg-background" : "bg-muted/20",
                                    "hover:bg-primary/5"
                                )}
                            >
                                {/* Row number */}
                                <td className="sticky left-0 z-10 w-12 min-w-12 bg-muted/40 border border-border text-center text-xs text-muted-foreground font-mono py-2">
                                    {rowIndex + 2}
                                </td>
                                {/* Data cells */}
                                {row.map((cell, colIndex) => (
                                    <td
                                        key={colIndex}
                                        className={cn(
                                            "border border-border py-2 px-3 truncate",
                                            selectedColumns.includes(colIndex) && "bg-primary/10",
                                            hoveredCell?.row === rowIndex && hoveredCell?.col === colIndex && "ring-2 ring-primary ring-inset"
                                        )}
                                        style={{ maxWidth: columnWidths[colIndex] }}
                                        onMouseEnter={() => setHoveredCell({ row: rowIndex + 1, col: colIndex })}
                                        onMouseLeave={() => setHoveredCell(null)}
                                        title={cell || ''}
                                    >
                                        {cell || ''}
                                    </td>
                                ))}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {/* Status Bar Bottom */}
            {totalRows > displayRows && (
                <div className="flex items-center justify-center px-3 py-2 bg-muted/40 border-t border-border text-xs text-muted-foreground">
                    <span>Mostrando {displayRows} de {totalRows} linhas</span>
                    <span className="mx-2">•</span>
                    <span className="text-primary cursor-pointer hover:underline">
                        Clique nas letras das colunas para selecioná-las para o gráfico
                    </span>
                </div>
            )}
        </div>
    )
}
