// Hook for managing Excel connection state and operations
import { useState, useEffect, useRef, useCallback } from 'react'
import { toast } from 'sonner'
import type { Workbook, PreviewDataType } from '@/types'

import {
    ConnectExcel,
    RefreshWorkbooks,
    SetExcelContext,
    GetPreviewData
} from "../../wailsjs/go/main/App"
import { EventsOn } from "../../wailsjs/runtime/runtime"

interface UseExcelConnectionReturn {
    connected: boolean
    workbooks: Workbook[]
    selectedWorkbook: string | null
    selectedSheets: string[]
    contextLoaded: string
    previewData: PreviewDataType | null
    expandedWorkbook: string | null
    setExpandedWorkbook: (name: string | null) => void
    handleConnect: () => Promise<void>
    handleSelectSheet: (wbName: string, sheetName: string) => Promise<void>
    refreshWorkbooks: () => Promise<void>
    setWorkbooks: (workbooks: Workbook[]) => void
    prepareChartData: (data: PreviewDataType) => any
}

export function useExcelConnection(): UseExcelConnectionReturn {
    const [connected, setConnected] = useState(false)
    const [workbooks, setWorkbooks] = useState<Workbook[]>([])
    const [selectedWorkbook, setSelectedWorkbook] = useState<string | null>(null)
    const [selectedSheets, setSelectedSheets] = useState<string[]>([])
    const [contextLoaded, setContextLoaded] = useState('')
    const [previewData, setPreviewData] = useState<PreviewDataType | null>(null)
    const [expandedWorkbook, setExpandedWorkbook] = useState<string | null>(null)

    // Ref to keep selectedWorkbook updated in callback without recreating listener
    const selectedWorkbookRef = useRef(selectedWorkbook)
    useEffect(() => {
        selectedWorkbookRef.current = selectedWorkbook
    }, [selectedWorkbook])

    const handleConnect = useCallback(async () => {
        if (typeof (window as any)?.go === 'undefined') {
            toast.error('Wails não detectado. Abra pelo app (wails dev/build).')
            return
        }
        try {
            const result = await ConnectExcel()
            setConnected(result.connected)
            if (result.connected && result.workbooks) {
                setWorkbooks(result.workbooks)
                toast.success('Conectado ao Excel!')
            } else if (result.error) {
                toast.error('Erro ao conectar: ' + result.error)
            }
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao conectar: ' + errorMessage)
        }
    }, [])

    const refreshWorkbooks = useCallback(async () => {
        try {
            const result = await RefreshWorkbooks()
            if (result.workbooks) {
                setWorkbooks(result.workbooks)
            }
        } catch (err) {
            console.error('Error refreshing workbooks:', err)
        }
    }, [])

    const handleSelectSheet = useCallback(async (wbName: string, sheetName: string) => {
        // Toggle sheet selection (multi-select)
        const isSelected = selectedWorkbook === wbName && selectedSheets.includes(sheetName)

        let newSheets: string[]
        let newWorkbook: string

        if (selectedWorkbook !== wbName) {
            // Changed workbook, reset selection
            newWorkbook = wbName
            newSheets = [sheetName]
            setSelectedWorkbook(wbName)
            setSelectedSheets([sheetName])
        } else if (isSelected) {
            // Deselect sheet
            newSheets = selectedSheets.filter(s => s !== sheetName)
            newWorkbook = wbName
            setSelectedSheets(newSheets)
            if (newSheets.length === 0) {
                setContextLoaded('')
                setPreviewData(null)
                return
            }
        } else {
            // Add sheet to selection
            newSheets = [...selectedSheets, sheetName]
            newWorkbook = wbName
            setSelectedSheets(newSheets)
        }

        const sheetsToLoad = isSelected
            ? selectedSheets.filter(s => s !== sheetName)
            : newSheets

        if (sheetsToLoad.length === 0) return

        setContextLoaded('')
        setPreviewData(null)

        try {
            // Load context for all selected sheets
            await SetExcelContext(newWorkbook, sheetsToLoad.join(','))

            // Get preview for first selected sheet
            const data = await GetPreviewData(newWorkbook, sheetsToLoad[0])
            if (data) {
                setPreviewData(data)
                const sheetNames = sheetsToLoad.join(', ')
                setContextLoaded(`${sheetNames} (${sheetsToLoad.length} aba${sheetsToLoad.length > 1 ? 's' : ''})`)
                toast.success(`Contexto carregado: ${sheetsToLoad.length} aba(s)`)
            }
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao carregar: ' + errorMessage)
            console.error('Erro ao carregar planilha:', err)
        }
    }, [selectedWorkbook, selectedSheets])

    const prepareChartData = useCallback((data: PreviewDataType) => {
        if (!data?.rows?.length) return null
        const labels = data.rows.slice(0, 10).map((r, i) => r[0] || `Item ${i + 1}`)
        const numericCol = data.rows[0]?.findIndex((_, i) => i > 0 && !isNaN(parseFloat(data.rows[0][i])))
        const values = data.rows.slice(0, 10).map(r => parseFloat(r[numericCol > 0 ? numericCol : 1]) || 0)

        return {
            labels,
            datasets: [{
                label: data.headers?.[numericCol > 0 ? numericCol : 1] || 'Valores',
                data: values,
                backgroundColor: ['#667eea', '#764ba2', '#f093fb', '#f5576c', '#4facfe', '#00f2fe', '#43e97b', '#38f9d7', '#fa709a', '#fee140'],
                borderColor: '#667eea',
                borderWidth: 1
            }]
        }
    }, [])

    // Listener for real-time workbook updates
    useEffect(() => {
        const cleanup = EventsOn("excel:workbooks-changed", (data: any) => {
            if (data?.workbooks !== undefined) {
                setWorkbooks(data.workbooks || [])
                setConnected(data.connected ?? true)

                // If selected workbook was closed, clear selection
                const currentSelected = selectedWorkbookRef.current
                if (currentSelected) {
                    const workbookStillExists = (data.workbooks || []).some(
                        (wb: any) => wb.name === currentSelected
                    )
                    if (!workbookStillExists) {
                        setSelectedWorkbook(null)
                        setSelectedSheets([])
                        setContextLoaded('')
                        setPreviewData(null)
                        setExpandedWorkbook(null)
                    }
                }
            }
        })
        return () => cleanup()
    }, [])

    return {
        connected,
        workbooks,
        selectedWorkbook,
        selectedSheets,
        contextLoaded,
        previewData,
        expandedWorkbook,
        setExpandedWorkbook,
        handleConnect,
        handleSelectSheet,
        refreshWorkbooks,
        setWorkbooks,
        prepareChartData
    }
}
