// DataPreview component - Professional spreadsheet-like table preview
import { useState, useMemo, useRef, useEffect } from 'react'
import { toast } from 'sonner'
import type { PreviewDataType } from '@/types'
import { cn } from '@/lib/utils'
import { UpdateCellValue } from '../../../wailsjs/go/app/App'

interface DataPreviewProps {
    previewData: PreviewDataType
    selectedColumns?: number[]
    onColumnSelect?: (columnIndex: number) => void
    activeSheet?: string
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

export function DataPreview({ previewData, selectedColumns = [], onColumnSelect, activeSheet = 'Planilha1' }: DataPreviewProps) {
    const containerRef = useRef<HTMLDivElement>(null)
    const [hoveredCell, setHoveredCell] = useState<{ row: number; col: number } | null>(null)
    const [selectedCell, setSelectedCell] = useState<{ row: number; col: number } | null>(null)

    // Editing state
    const [editingCell, setEditingCell] = useState<{ row: number; col: number } | null>(null)
    const [tempValue, setTempValue] = useState('')
    const [editedCells, setEditedCells] = useState<Map<string, string>>(new Map())
    const [savingCells, setSavingCells] = useState<Set<string>>(new Set())

    const totalRows = previewData.rows?.length || 0
    const totalCols = previewData.headers?.length || 0
    const displayRows = Math.min(totalRows, 1000) // Show up to 1000 rows

    // Start editing a cell
    const startEdit = (row: number, col: number) => {
        const currentValue = previewData.rows?.[row]?.[col] || ''
        setEditingCell({ row, col })
        setTempValue(currentValue)
        setSelectedCell({ row, col })
    }

    // Cancel editing
    const cancelEdit = () => {
        setEditingCell(null)
        setTempValue('')
    }

    // Confirm edit (mark as edited, don't save yet)
    const confirmEdit = () => {
        if (!editingCell) return

        const key = `${editingCell.row}-${editingCell.col}`
        const originalValue = previewData.rows?.[editingCell.row]?.[editingCell.col] || ''

        // Only mark as edited if value changed
        if (tempValue !== originalValue) {
            setEditedCells(prev => {
                const next = new Map(prev)
                next.set(key, tempValue)
                return next
            })
        }

        setEditingCell(null)
    }

    // Save a single cell to backend
    const handleSaveCell = async (row: number, col: number) => {
        const key = `${row}-${col}`
        const newValue = editedCells.get(key)

        if (!newValue) return

        setSavingCells(prev => new Set(prev).add(key))

        try {
            const cellRef = getColumnLetter(col) + (row + 2) // +2 because row 0 is headers (row 1 in Excel)

            await UpdateCellValue(activeSheet, cellRef, newValue)

            // Remove from edited list after successful save
            setEditedCells(prev => {
                const next = new Map(prev)
                next.delete(key)
                return next
            })

            // Update the preview data locally
            if (previewData.rows && previewData.rows[row]) {
                previewData.rows[row][col] = newValue
            }

            toast.success(`Célula ${cellRef} salva!`)
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao salvar célula: ' + errorMsg)
        } finally {
            setSavingCells(prev => {
                const next = new Set(prev)
                next.delete(key)
                return next
            })
        }
    }

    // Determine column widths based on content
    const columnWidths = useMemo(() => {
        const widths: number[] = []
        const maxWidth = 250
        const minWidth = 80
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
        <div className="flex-1 flex flex-col bg-background overflow-hidden border-t border-border">
            {/* Status Bar Top */}
            <div className="flex items-center justify-between px-4 py-2 bg-muted/50 border-b border-border text-xs text-muted-foreground">
                <div className="flex items-center gap-4">
                    <span className="font-semibold text-foreground">📊 Visualização de Dados</span>
                    <span className="bg-background px-2 py-0.5 rounded border border-border">
                        {totalCols} colunas × {totalRows} linhas
                    </span>
                </div>
                <div className="flex items-center gap-3">
                    {selectedColumns.length > 0 && (
                        <span className="bg-primary/20 text-primary px-2 py-0.5 rounded-full">
                            {selectedColumns.length} col. selecionada{selectedColumns.length > 1 ? 's' : ''}
                        </span>
                    )}
                    {selectedCell && (
                        <span className="font-mono bg-background px-2 py-0.5 rounded border border-border font-semibold">
                            {getColumnLetter(selectedCell.col)}{selectedCell.row + 1}
                        </span>
                    )}
                </div>
            </div>

            {/* Excel-style Spreadsheet Container */}
            <div ref={containerRef} className="flex-1 overflow-x-auto overflow-y-auto bg-white">
                <div className="inline-block min-w-full" style={{ minWidth: 'fit-content' }}>
                    <table
                        className="border-collapse text-sm"
                        style={{
                            tableLayout: 'fixed',
                            minWidth: `${Math.max(totalCols, 5) * 120}px`,
                        }}
                    >
                        {/* Column Headers (A, B, C...) */}
                        <thead className="sticky top-0 z-20">
                            {/* Column letters row */}
                            <tr>
                                {/* Corner cell (row number header) */}
                                <th
                                    className="sticky left-0 top-0 z-30 w-12 min-w-12 bg-gray-200 border border-gray-400 font-bold text-gray-700 text-center py-2 select-none"
                                    style={{
                                        boxShadow: '2px 0 4px rgba(0,0,0,0.1)'
                                    }}
                                >
                                    <svg className="w-3 h-3 mx-auto" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                        <path d="M4 4l16 16M20 4l-16 16" strokeLinecap="round" />
                                    </svg>
                                </th>
                                {/* Column letter headers */}
                                {previewData.headers?.map((_, colIndex) => (
                                    <th
                                        key={colIndex}
                                        className={cn(
                                            "border border-gray-400 font-bold text-center py-2 cursor-pointer transition-colors select-none",
                                            selectedColumns.includes(colIndex)
                                                ? "bg-blue-200 text-blue-700"
                                                : "bg-gray-200 text-gray-700 hover:bg-gray-300"
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
                            <tr>
                                <th className="sticky left-0 z-30 w-12 min-w-12 bg-gray-200 border border-gray-400 font-bold text-gray-700 text-center py-2 select-none text-xs">
                                    1
                                </th>
                                {previewData.headers?.map((header, colIndex) => (
                                    <th
                                        key={colIndex}
                                        className={cn(
                                            "border border-gray-400 text-left font-semibold py-2 px-3 truncate",
                                            selectedColumns.includes(colIndex)
                                                ? "bg-blue-100 text-blue-900"
                                                : "bg-gray-50 text-gray-700"
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
                                        rowIndex % 2 === 0 ? "bg-white" : "bg-gray-50/50",
                                        "hover:bg-blue-50/50"
                                    )}
                                >
                                    {/* Row number */}
                                    <td
                                        className="sticky left-0 z-10 w-12 min-w-12 bg-gray-200 border border-gray-400 font-bold text-gray-700 text-center py-2 select-none text-xs"
                                        style={{
                                            boxShadow: '2px 0 4px rgba(0,0,0,0.1)'
                                        }}
                                    >
                                        {rowIndex + 2}
                                    </td>
                                    {/* Data cells */}
                                    {row.map((cell, colIndex) => {
                                        const cellKey = `${rowIndex}-${colIndex}`
                                        const isEditing = editingCell?.row === rowIndex && editingCell?.col === colIndex
                                        const isEdited = editedCells.has(cellKey)
                                        const isSaving = savingCells.has(cellKey)

                                        return (
                                            <td
                                                key={cellKey}
                                                className={cn(
                                                    "border border-gray-300 py-2 px-3 transition-colors relative",
                                                    selectedColumns.includes(colIndex) && "bg-blue-50/50",
                                                    selectedCell?.row === rowIndex && selectedCell?.col === colIndex
                                                        ? "bg-blue-100 ring-2 ring-blue-500 ring-inset"
                                                        : "",
                                                    hoveredCell?.row === rowIndex && hoveredCell?.col === colIndex
                                                        ? "bg-blue-50/30"
                                                        : "",
                                                    isEdited && "bg-yellow-50"
                                                )}
                                                style={{ maxWidth: columnWidths[colIndex] }}
                                                onMouseEnter={() => setHoveredCell({ row: rowIndex, col: colIndex })}
                                                onMouseLeave={() => setHoveredCell(null)}
                                            >
                                                {isEditing ? (
                                                    <input
                                                        autoFocus
                                                        type="text"
                                                        value={tempValue}
                                                        onChange={(e) => setTempValue(e.target.value)}
                                                        onKeyDown={(e) => {
                                                            if (e.key === 'Enter') confirmEdit()
                                                            if (e.key === 'Escape') cancelEdit()
                                                        }}
                                                        onBlur={confirmEdit}
                                                        className="w-full px-1 py-0.5 border border-blue-500 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                                                    />
                                                ) : (
                                                    <div className="flex items-center gap-2 overflow-hidden">
                                                        <span
                                                            className="flex-1 overflow-hidden text-ellipsis whitespace-nowrap cursor-text"
                                                            onDoubleClick={() => startEdit(rowIndex, colIndex)}
                                                            onClick={() => setSelectedCell({ row: rowIndex, col: colIndex })}
                                                            title={cell || ''}
                                                        >
                                                            {isEdited ? editedCells.get(cellKey) : (cell !== undefined && cell !== null ? cell : '')}
                                                        </span>
                                                        {isEdited && (
                                                            <button
                                                                onClick={() => handleSaveCell(rowIndex, colIndex)}
                                                                disabled={isSaving}
                                                                className={cn(
                                                                    "shrink-0 w-6 h-6 rounded flex items-center justify-center transition-colors",
                                                                    isSaving
                                                                        ? "bg-gray-200 text-gray-400 cursor-not-allowed"
                                                                        : "bg-green-500 hover:bg-green-600 text-white cursor-pointer"
                                                                )}
                                                                title="Salvar célula"
                                                            >
                                                                {isSaving ? (
                                                                    <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                                                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                                                    </svg>
                                                                ) : (
                                                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                                                    </svg>
                                                                )}
                                                            </button>
                                                        )}
                                                    </div>
                                                )}
                                            </td>
                                        )
                                    })}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                {/* Row count indicator at bottom */}
                {totalRows > displayRows && (
                    <div className="sticky bottom-0 left-0 right-0 bg-gray-100 border-t-2 border-gray-400 px-4 py-2 text-sm text-gray-600 font-medium">
                        Mostrando {displayRows} de {totalRows} linhas totais
                    </div>
                )}
            </div>

            {/* Status Bar Bottom */}
            {totalRows > displayRows && (
                <div className="flex items-center justify-center px-4 py-2 bg-muted/40 border-t border-border text-xs text-muted-foreground">
                    <span className="text-primary cursor-pointer hover:underline">
                        Clique nas letras das colunas (A, B, C...) para selecioná-las para o gráfico
                    </span>
                    <span className="mx-2">•</span>
                    <span>Clique nas células para ver o endereço (ex: A1, B2)</span>
                </div>
            )}
        </div>
    )
}
