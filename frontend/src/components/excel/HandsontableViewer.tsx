import { useRef, useState, useEffect, useCallback, useMemo } from 'react'
import { AgGridReact } from 'ag-grid-react'
import { ModuleRegistry, AllCommunityModule, themeQuartz } from 'ag-grid-community'
import './aggrid-override.css'
import { toast } from 'sonner'
import { UpdateCellValue } from '../../../wailsjs/go/app/App'

// Register AG Grid modules
ModuleRegistry.registerModules([AllCommunityModule])

interface AGGridViewerProps {
    sessionId: string
    activeSheet: string
    sheetData: string[][] | null
    loading?: boolean
}

// Convert column index to Excel letter (0=A, 1=B, etc.)
function columnToLetter(col: number): string {
    let letter = ''
    let temp = col
    while (temp >= 0) {
        letter = String.fromCharCode((temp % 26) + 65) + letter
        temp = Math.floor(temp / 26) - 1
    }
    return letter
}

export function HandsontableViewer({
    sessionId,
    activeSheet,
    sheetData,
    loading = false
}: AGGridViewerProps) {
    const gridRef = useRef<AgGridReact>(null)
    const [isSaving, setIsSaving] = useState(false)

    // Detect dark mode and create appropriate theme
    const [isDark, setIsDark] = useState(() =>
        document.documentElement.classList.contains('dark')
    )

    useEffect(() => {
        const observer = new MutationObserver(() => {
            setIsDark(document.documentElement.classList.contains('dark'))
        })
        observer.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class']
        })
        return () => observer.disconnect()
    }, [])

    // Create theme based on current mode with correct OKLCH colors from index.css
    const customTheme = useMemo(() => {
        if (isDark) {
            // Dark mode - use exact colors from index.css .dark
            return themeQuartz.withParams({
                backgroundColor: 'oklch(0.141 0.005 285.823)', // --background dark
                headerBackgroundColor: 'oklch(0.141 0.005 285.823)',
                oddRowBackgroundColor: 'oklch(0.141 0.005 285.823)',
                chromeBackgroundColor: 'oklch(0.141 0.005 285.823)',
                textColor: 'oklch(0.985 0 0)', // --foreground dark
                headerTextColor: 'oklch(0.985 0 0)',
                borderColor: 'oklch(1 0 0 / 10%)', // --border dark
                rowHoverColor: 'oklch(0.274 0.006 286.033)', // --muted dark
                selectedRowBackgroundColor: 'oklch(0.274 0.006 286.033)', // --accent dark
                fontSize: 14,
                headerFontSize: 14,
                headerFontWeight: 600,
                spacing: 8,
            })
        } else {
            // Light mode - use exact colors from index.css :root
            return themeQuartz.withParams({
                backgroundColor: 'oklch(0.97 0.005 250)', // --background light
                headerBackgroundColor: 'oklch(0.97 0.005 250)',
                oddRowBackgroundColor: 'oklch(0.97 0.005 250)',
                chromeBackgroundColor: 'oklch(0.97 0.005 250)',
                textColor: 'oklch(0.20 0.02 260)', // --foreground light
                headerTextColor: 'oklch(0.20 0.02 260)',
                borderColor: 'oklch(0.88 0.01 250)', // --border light
                rowHoverColor: 'oklch(0.93 0.01 250)', // --muted light
                selectedRowBackgroundColor: 'oklch(0.90 0.02 250)', // --accent light
                fontSize: 14,
                headerFontSize: 14,
                headerFontWeight: 600,
                spacing: 8,
            })
        }
    }, [isDark])

    // Convert sheetData to AG Grid format
    const { columnDefs, rowData } = useMemo(() => {
        if (!sheetData || sheetData.length === 0) {
            return { columnDefs: [], rowData: [] }
        }

        // First row is headers
        const headers = sheetData[0]

        // Create column definitions
        const cols = headers.map((header, index) => ({
            field: `col${index}`,
            headerName: header || columnToLetter(index),
            editable: true,
            resizable: true,
            sortable: true,
            filter: true,
            minWidth: 100,
        }))

        // Create row data (skip first row which is headers)
        const rows = sheetData.slice(1).map((row, rowIndex) => {
            const rowObj: any = { _rowIndex: rowIndex }
            row.forEach((cell, colIndex) => {
                rowObj[`col${colIndex}`] = cell
            })
            return rowObj
        })

        return { columnDefs: cols, rowData: rows }
    }, [sheetData])

    // Handle cell value changes
    const onCellValueChanged = useCallback(async (event: any) => {
        const { data, colDef, newValue, oldValue } = event

        if (newValue === oldValue) return

        setIsSaving(true)

        try {
            // Get column index from field name (col0, col1, etc.)
            const colIndex = parseInt(colDef.field.replace('col', ''))
            const rowIndex = data._rowIndex

            // +2 because: +1 for 0-indexed to 1-indexed, +1 for header row
            const cellRef = columnToLetter(colIndex) + (rowIndex + 2)

            await UpdateCellValue(activeSheet, cellRef, String(newValue || ''))
            toast.success(`Célula ${cellRef} salva!`)
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao salvar: ' + errorMsg)
            // Revert the change on error
            event.node.setDataValue(event.colDef.field, oldValue)
        } finally {
            setIsSaving(false)
        }
    }, [activeSheet])

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full">
                <div className="flex items-center gap-3 text-blue-600">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                    <span className="font-medium">Carregando planilha...</span>
                </div>
            </div>
        )
    }

    if (!sessionId || !sheetData || sheetData.length === 0) {
        return (
            <div className="flex items-center justify-center h-full text-gray-500">
                <p className="text-lg">Nenhum dado disponível</p>
            </div>
        )
    }

    return (
        <div className="w-full h-full relative" style={{ borderRadius: 0 }}>
            {isSaving && (
                <div className="absolute top-2 right-2 z-50 bg-blue-500 text-white px-3 py-1 rounded-full text-sm flex items-center gap-2 shadow-lg">
                    <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-white"></div>
                    Salvando...
                </div>
            )}


            <AgGridReact
                ref={gridRef}
                theme={customTheme}
                columnDefs={columnDefs}
                rowData={rowData}
                onCellValueChanged={onCellValueChanged}
                defaultColDef={{
                    editable: true,
                    resizable: true,
                    sortable: true,
                    filter: true,
                }}
                undoRedoCellEditing={true}
                undoRedoCellEditingLimit={20}
                animateRows={true}
                enableCellTextSelection={true}
                ensureDomOrder={true}
            />
        </div>
    )
}
