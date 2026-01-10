import React, { useState, useEffect } from 'react';
import { Sheet, ChevronDown, ChevronRight, Download, RefreshCw } from 'lucide-react';
import { Button } from '../ui/button';
import { Card } from '../ui/card';

interface SheetPreview {
  name: string;
  rows: number;
  cols: number;
}

interface PreviewData {
  sessionId: string;
  fileName: string;
  sheets: SheetPreview[];
  activeSheet: string;
}

interface ExcelViewerProps {
  previewData: PreviewData | null;
  sheetData: string[][] | null;
  loading?: boolean;
  onDownload: () => Promise<void>;
  onRefreshSheet?: (sheetName: string) => Promise<void>;
  onSheetChange?: (sheetName: string) => void;
  downloading?: boolean;
}

export const ExcelViewer: React.FC<ExcelViewerProps> = ({
  previewData,
  sheetData,
  loading = false,
  onDownload,
  onRefreshSheet,
  onSheetChange,
  downloading = false,
}) => {
  const [expandedSheets, setExpandedSheets] = useState<Set<string>>(new Set());
  const [currentSheet, setCurrentSheet] = useState<string>('');

  useEffect(() => {
    if (previewData?.activeSheet) {
      setCurrentSheet(previewData.activeSheet);
    }
  }, [previewData]);

  const toggleSheetExpand = (sheetName: string) => {
    setExpandedSheets((prev) => {
      const next = new Set(prev);
      if (next.has(sheetName)) {
        next.delete(sheetName);
      } else {
        next.add(sheetName);
      }
      return next;
    });
  };

  const handleSheetClick = (sheetName: string) => {
    setCurrentSheet(sheetName);
    if (onSheetChange) {
      onSheetChange(sheetName);
    }
    if (onRefreshSheet) {
      onRefreshSheet(sheetName);
    }
  };

  const handleDownload = async () => {
    await onDownload();
  };

  if (!previewData) {
    return (
      <Card className="p-6">
        <div className="flex flex-col items-center justify-center h-64 text-gray-500">
          <Sheet className="w-16 h-16 mb-4 text-gray-400" />
          <p className="text-lg font-medium">Nenhum arquivo carregado</p>
          <p className="text-sm">Carregue um arquivo Excel para começar</p>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <Card className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Sheet className="w-6 h-6 text-green-600" />
            <div>
              <h3 className="font-semibold text-lg">{previewData.fileName}</h3>
              <p className="text-sm text-gray-500">
                {previewData.sheets.length} planilha{previewData.sheets.length > 1 ? 's' : ''}
              </p>
            </div>
          </div>
          <Button
            onClick={handleDownload}
            disabled={downloading}
            className="flex items-center gap-2"
          >
            {downloading ? (
              <RefreshCw className="w-4 h-4 animate-spin" />
            ) : (
              <Download className="w-4 h-4" />
            )}
            {downloading ? 'Baixando...' : 'Baixar'}
          </Button>
        </div>
      </Card>

      {/* Sheets List */}
      <Card className="p-4">
        <div className="space-y-2">
          {previewData.sheets.map((sheet) => {
            const isExpanded = expandedSheets.has(sheet.name);
            const isActive = currentSheet === sheet.name;
            
            return (
              <div key={sheet.name}>
                <div
                  className={`flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-all ${
                    isActive
                      ? 'bg-blue-50 border-2 border-blue-300'
                      : 'hover:bg-gray-50 border-2 border-transparent'
                  }`}
                  onClick={() => handleSheetClick(sheet.name)}
                >
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleSheetExpand(sheet.name);
                    }}
                    className="shrink-0"
                  >
                    {isExpanded ? (
                      <ChevronDown className="w-4 h-4 text-gray-500" />
                    ) : (
                      <ChevronRight className="w-4 h-4 text-gray-500" />
                    )}
                  </button>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{sheet.name}</p>
                    <p className="text-xs text-gray-500">
                      {sheet.rows} linhas × {sheet.cols} colunas
                    </p>
                  </div>
                  {isActive && (
                    <div className="w-2 h-2 bg-blue-600 rounded-full shrink-0" />
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </Card>

      {/* Sheet Data Preview */}
      {currentSheet && (loading || sheetData) && (
        <Card className="p-4">
          <div className="mb-3">
            <h4 className="font-semibold text-lg">{currentSheet}</h4>
            <p className="text-sm text-gray-500">
              {loading ? 'Carregando dados...' : `${sheetData?.length || 0} linhas carregadas`}
            </p>
          </div>

          {loading ? (
            <div className="flex items-center justify-center h-64">
              <div className="flex items-center gap-3 text-blue-600">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
                <span className="font-medium">Carregando dados da planilha...</span>
              </div>
            </div>
          ) : sheetData && sheetData.length > 0 ? (
            <div className="overflow-x-auto border border-gray-200 rounded-lg">
              <table className="w-full text-sm">
                <thead className="bg-gray-50">
                  <tr>
                    {sheetData[0].map((header, index) => (
                      <th
                        key={index}
                        className="px-4 py-2 text-left font-semibold text-gray-700 border-b border-gray-200 whitespace-nowrap"
                      >
                        {header || `(Col ${index + 1})`}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {sheetData.slice(1, 101).map((row, rowIndex) => (
                    <tr
                      key={rowIndex}
                      className={rowIndex % 2 === 0 ? 'bg-white' : 'bg-gray-50'}
                    >
                      {row.map((cell, cellIndex) => (
                        <td
                          key={cellIndex}
                          className="px-4 py-2 text-left border-b border-gray-200 whitespace-nowrap text-gray-700"
                        >
                          {cell !== undefined && cell !== null ? cell : ''}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
              {sheetData.length > 101 && (
                <div className="p-3 text-center text-sm text-gray-500 bg-gray-50">
                  Mostrando primeiras 100 linhas de {sheetData.length} linhas totais
                </div>
              )}
            </div>
          ) : (
            <div className="flex items-center justify-center h-64 text-gray-500">
              <p className="text-lg">Planilha vazia</p>
            </div>
          )}
        </Card>
      )}
    </div>
  );
};
