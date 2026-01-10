import { useExcelUpload } from '@/hooks/useExcelUpload'
import { UploadExcel } from '@/components/excel/UploadExcel'
import { ExcelViewer } from '@/components/excel/ExcelViewer'

export function ExcelUploadPage() {
  const {
    previewData,
    sheetData,
    uploading,
    downloading,
    loadingSheet,
    handleUpload,
    loadSheetData,
    handleDownload,
    closeSession,
    refreshPreview,
    setSheetData
  } = useExcelUpload()

  const handleSheetChange = (sheetName: string) => {
    setSheetData(null)
    loadSheetData(sheetName)
  }

  return (
    <div className="flex flex-col h-full p-6 space-y-6 bg-background">
      <div className="space-y-6 max-w-7xl mx-auto w-full">
        <div className="mb-6">
          <h1 className="text-2xl font-bold mb-2">Importar Arquivo Excel</h1>
          <p className="text-muted-foreground">
            Carregue um arquivo Excel para analisar e modificar com a IA
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-2">
          {/* Upload Component */}
          <UploadExcel
            onUpload={handleUpload}
            uploading={uploading}
          />

          {/* Excel Viewer Component */}
          <ExcelViewer
            previewData={previewData}
            sheetData={sheetData}
            loading={loadingSheet}
            onDownload={handleDownload}
            onRefreshSheet={loadSheetData}
            onSheetChange={handleSheetChange}
            downloading={downloading}
          />
        </div>

        {/* Botão para fechar sessão se houver uma ativa */}
        {previewData && (
          <div className="flex justify-center">
            <button
              onClick={closeSession}
              className="px-6 py-2 bg-destructive text-destructive-foreground rounded-md hover:opacity-90 transition-opacity"
            >
              Fechar Arquivo e Nova Sessão
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
