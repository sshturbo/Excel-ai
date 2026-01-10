import { useState, useCallback } from 'react'
import { toast } from 'sonner'

interface PreviewData {
  sessionId: string
  fileName: string
  sheets: SheetPreview[]
  activeSheet: string
}

interface SheetPreview {
  name: string
  rows: number
  cols: number
}

export function useExcelUpload() {
  const [sessionId, setSessionId] = useState<string>('')
  const [previewData, setPreviewData] = useState<PreviewData | null>(null)
  const [sheetData, setSheetData] = useState<string[][] | null>(null)
  const [activeSheet, setActiveSheet] = useState<string>('')
  const [uploading, setUploading] = useState(false)
  const [downloading, setDownloading] = useState(false)
  const [loadingSheet, setLoadingSheet] = useState(false)

  // Upload de arquivo
  const handleUpload = useCallback(async (filename: string, data: Uint8Array) => {
    setUploading(true)
    try {
      // Importar dinamicamente para evitar erros fora do Wails
      const { UploadExcel } = await import('../../wailsjs/go/app/App')

      // Converter Uint8Array para Array<number>
      const dataAsArray = Array.from(data)

      const newSessionId = await UploadExcel(filename, dataAsArray)
      setSessionId(newSessionId)

      toast.success(`Arquivo "${filename}" carregado com sucesso!`)

      // Carregar preview automaticamente
      await loadPreview(newSessionId)
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      toast.error('Erro ao carregar arquivo: ' + errorMsg)
      throw err
    } finally {
      setUploading(false)
    }
  }, [])

  // Carregar preview do arquivo
  const loadPreview = useCallback(async (sid: string) => {
    try {
      const { GetExcelPreview } = await import('../../wailsjs/go/app/App')
      const preview = await GetExcelPreview(sid)
      setPreviewData(preview)
      // Set initial active sheet from preview
      if (preview?.activeSheet) {
        setActiveSheet(preview.activeSheet)
      }
      return preview
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      toast.error('Erro ao carregar preview: ' + errorMsg)
      throw err
    }
  }, [])

  // Carregar dados de uma planilha
  const loadSheetData = useCallback(async (sheetName: string) => {
    console.log('[DEBUG useExcelUpload] loadSheetData chamado com:', sheetName)
    console.log('[DEBUG useExcelUpload] sessionId atual:', sessionId)

    if (!sessionId) {
      console.error('[DEBUG useExcelUpload] sessionId está vazio!')
      return
    }

    setLoadingSheet(true)
    try {
      console.log('[DEBUG useExcelUpload] Chamando GetSheetData com sessionId:', sessionId, 'sheetName:', sheetName)
      const { GetSheetData } = await import('../../wailsjs/go/app/App')
      const data = await GetSheetData(sessionId, sheetName)
      console.log('[DEBUG useExcelUpload] Dados recebidos:', data?.length, 'linhas')
      setSheetData(data)
      setActiveSheet(sheetName) // Update active sheet when data is loaded
      toast.success(`Dados da aba "${sheetName}" carregados!`)
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      console.error('[DEBUG useExcelUpload] Erro:', errorMsg)
      toast.error('Erro ao carregar dados da planilha: ' + errorMsg)
    } finally {
      setLoadingSheet(false)
    }
  }, [sessionId])

  // Download do arquivo modificado
  const handleDownload = useCallback(async () => {
    if (!sessionId) {
      toast.error('Nenhum arquivo carregado')
      return
    }

    setDownloading(true)
    try {
      const { DownloadExcel } = await import('../../wailsjs/go/app/App')
      const data = await DownloadExcel(sessionId)

      // Converter Array<number> para Uint8Array
      const dataAsUint8Array = new Uint8Array(data)

      // Criar blob e download
      const blob = new Blob([dataAsUint8Array], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = previewData?.fileName || 'arquivo_modificado.xlsx'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)

      toast.success('Arquivo baixado com sucesso!')
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      toast.error('Erro ao baixar arquivo: ' + errorMsg)
    } finally {
      setDownloading(false)
    }
  }, [sessionId, previewData])

  // Fechar sessão
  const closeSession = useCallback(async () => {
    if (!sessionId) return

    try {
      const { CloseSession } = await import('../../wailsjs/go/app/App')
      await CloseSession(sessionId)
      setSessionId('')
      setPreviewData(null)
      setSheetData(null)
      setActiveSheet('')
      toast.success('Sessão fechada')
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      toast.error('Erro ao fechar sessão: ' + errorMsg)
    }
  }, [sessionId])

  // Refresh do preview
  const refreshPreview = useCallback(async () => {
    if (!sessionId) return
    await loadPreview(sessionId)
  }, [sessionId, loadPreview])

  return {
    sessionId,
    previewData,
    sheetData,
    activeSheet,
    uploading,
    downloading,
    loadingSheet,
    handleUpload,
    loadSheetData,
    handleDownload,
    closeSession,
    refreshPreview,
    setSheetData
  }
}
