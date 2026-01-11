import { useState, useCallback, useEffect } from 'react'
import { toast } from 'sonner'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

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

  // Listen for auto-loaded files from backend (e.g. when switching conversations)
  useEffect(() => {
    const handleNativeFileLoaded = (data: { sessionId: string, fileName: string }) => {
      console.log('[DEBUG useExcelUpload] Evento native:file-loaded recebido:', data)
      setSessionId(data.sessionId)
      loadPreview(data.sessionId)
    }

    EventsOn('native:file-loaded', handleNativeFileLoaded)
    return () => {
      EventsOff('native:file-loaded')
    }
  }, [loadPreview])

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

  // Abrir arquivo nativo (System Dialog)
  const handleOpenFileNative = useCallback(async () => {
    setUploading(true)
    try {
      const { OpenFileNative } = await import('../../wailsjs/go/app/App')
      const newSessionId = await OpenFileNative()

      if (newSessionId) {
        setSessionId(newSessionId)
        toast.success(`Arquivo aberto com sucesso!`)
        await loadPreview(newSessionId)
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      toast.error('Erro ao abrir arquivo: ' + errorMsg)
    } finally {
      setUploading(false)
    }
  }, [loadPreview])

  // Salvar arquivo nativo (Sobrescrever original)
  const handleSaveNative = useCallback(async () => {
    if (!sessionId) return

    setDownloading(true)
    try {
      const { SaveFileNative } = await import('../../wailsjs/go/app/App')
      await SaveFileNative()
      toast.success('Alterações salvas no arquivo original!')
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      toast.error('Erro ao salvar arquivo: ' + errorMsg)
    } finally {
      setDownloading(false)
    }
  }, [sessionId])

  // Refresh do preview
  const refreshPreview = useCallback(async () => {
    if (!sessionId) return

    // Recarregar meta dados
    const newPreview = await loadPreview(sessionId)

    // Usar o estado mais recente da aba ativa
    // Se não tivermos no estado, tentamos pegar do novo preview
    setActiveSheet(currentActive => {
      const targetSheet = currentActive || newPreview?.activeSheet
      if (targetSheet) {
        // Disparar o carregamento dos dados da aba de forma assíncrona
        // mas fora do ciclo de atualização do estado para evitar avisos de renderização
        setTimeout(async () => {
          try {
            setLoadingSheet(true)
            const { GetSheetData } = await import('../../wailsjs/go/app/App')
            const data = await GetSheetData(sessionId, targetSheet)
            setSheetData(data)
          } catch (err) {
            console.error('Erro ao recarregar dados da aba ativa:', err)
          } finally {
            setLoadingSheet(false)
          }
        }, 0)
      }
      return currentActive
    })
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
    handleOpenFileNative,
    handleSaveNative,
    closeSession,
    refreshPreview,
    setSheetData
  }
}
