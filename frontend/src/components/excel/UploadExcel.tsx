import React, { useState, useCallback } from 'react';
import { Upload, X, FileSpreadsheet, AlertCircle, CheckCircle } from 'lucide-react';
import { read, utils } from 'xlsx';
import { Button } from '../ui/button';
import { Card } from '../ui/card';

interface UploadExcelProps {
  onUpload: (filename: string, data: Uint8Array) => Promise<void>;
  uploading?: boolean;
}

export const UploadExcel: React.FC<UploadExcelProps> = ({ onUpload, uploading = false }) => {
  const [dragActive, setDragActive] = useState(false);
  const [error, setError] = useState<string>('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  }, []);

  const handleDrop = useCallback(async (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      await processFile(e.dataTransfer.files[0]);
    }
  }, []);

  const handleChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    e.preventDefault();
    if (e.target.files && e.target.files[0]) {
      await processFile(e.target.files[0]);
    }
  }, []);

  const processFile = async (file: File) => {
    setError('');

    // Validar extensão
    if (!file.name.endsWith('.xlsx') && !file.name.endsWith('.xls')) {
      setError('Por favor, selecione um arquivo Excel (.xlsx ou .xls)');
      return;
    }

    // Validar tamanho (max 10MB)
    if (file.size > 10 * 1024 * 1024) {
      setError('O arquivo é muito grande. Tamanho máximo: 10MB');
      return;
    }

    try {
      // Ler o arquivo
      const arrayBuffer = await file.arrayBuffer();
      const data = new Uint8Array(arrayBuffer);

      // Tentar ler com xlsx para validar
      try {
        const workbook = read(data, { type: 'array' });
        if (workbook.SheetNames.length === 0) {
          setError('O arquivo não contém nenhuma planilha');
          return;
        }
      } catch (readError) {
        setError('Não foi possível ler o arquivo Excel. Verifique se o arquivo está válido.');
        return;
      }

      setSelectedFile(file);
      await onUpload(file.name, data);
      setSelectedFile(null);
    } catch (err) {
      setError('Erro ao processar o arquivo: ' + (err as Error).message);
    }
  };

  const clearSelectedFile = () => {
    setSelectedFile(null);
    setError('');
  };

  return (
    <Card className="p-6">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">Carregar Arquivo Excel</h3>
          <FileSpreadsheet className="w-6 h-6 text-green-600" />
        </div>

        {error && (
          <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700">
            <AlertCircle className="w-5 h-5" />
            <span className="text-sm">{error}</span>
          </div>
        )}

        {!uploading && !selectedFile ? (
          <div
            className={`relative border-2 border-dashed rounded-lg p-8 text-center transition-all ${
              dragActive
                ? 'border-blue-500 bg-blue-50'
                : 'border-gray-300 hover:border-gray-400'
            }`}
            onDragEnter={handleDrag}
            onDragLeave={handleDrag}
            onDragOver={handleDrag}
            onDrop={handleDrop}
          >
            <input
              type="file"
              accept=".xlsx,.xls"
              onChange={handleChange}
              className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
              disabled={uploading}
            />
            <Upload className="w-12 h-12 mx-auto mb-4 text-gray-400" />
            <p className="text-gray-700 font-medium mb-2">
              Arraste e solte seu arquivo Excel aqui
            </p>
            <p className="text-gray-500 text-sm mb-4">ou clique para selecionar</p>
            <div className="text-xs text-gray-400">
              <p>Formatos aceitos: .xlsx, .xls</p>
              <p>Tamanho máximo: 10MB</p>
            </div>
          </div>
        ) : selectedFile && !uploading ? (
          <div className="flex items-center gap-4 p-4 bg-green-50 border border-green-200 rounded-lg">
            <CheckCircle className="w-6 h-6 text-green-600 shrink-0" />
            <div className="flex-1 min-w-0">
              <p className="font-medium text-green-800 truncate">{selectedFile.name}</p>
              <p className="text-sm text-green-600">
                {(selectedFile.size / 1024).toFixed(2)} KB
              </p>
            </div>
            <Button
              variant="ghost"
              size="icon"
              onClick={clearSelectedFile}
              className="shrink-0"
            >
              <X className="w-4 h-4" />
            </Button>
          </div>
        ) : (
          <div className="flex items-center justify-center p-8">
            <div className="flex items-center gap-3 text-blue-600">
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600"></div>
              <span className="font-medium">Processando arquivo...</span>
            </div>
          </div>
        )}

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p className="text-sm text-blue-800">
            <strong>Nota:</strong> O arquivo será processado localmente. Suas informações não são enviadas para nenhum servidor.
          </p>
        </div>
      </div>
    </Card>
  );
};
