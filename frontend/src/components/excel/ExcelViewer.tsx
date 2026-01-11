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
};

// Convert column index to Excel-style letter (0=A, 1=B, 26=AA, etc.)
function getColumnLetter(index: number): string {
  let letter = '';
  let temp = index;
  while (temp >= 0) {
    letter = String.fromCharCode((temp % 26) + 65) + letter;
    temp = Math.floor(temp / 26) - 1;
  }
  return letter;
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
            <div className="overflow-x-auto overflow-y-auto border border-gray-300 rounded-lg bg-white" style={{ maxHeight: '500px' }}>
              <div className="inline-block min-w-full" style={{ minWidth: 'fit-content' }}>
                <table className="w-full text-sm border-collapse" style={{ tableLayout: 'fixed', minWidth: `${Math.max(sheetData[0].length, 5) * 120}px` }}>
                <thead className="sticky top-0 z-20">
                  {/* Column letters row */}
                  <tr>
                    <th className="sticky left-0 z-30 w-12 min-w-12 bg-gray-200 border border-gray-400 font-bold text-gray-700 text-center py-2 select-none">
                      <svg className="w-3 h-3 mx-auto" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M4 4l16 16M20 4l-16 16" strokeLinecap="round" />
                      </svg>
                    </th>
                    {sheetData[0].map((_, index) => (
                      <th
                        key={index}
                        className="bg-gray-200 border border-gray-400 font-bold text-gray-700 text-center py-2 select-none hover:bg-gray-300"
                        style={{ width: '120px' }}
                      >
                        {getColumnLetter(index)}
                      </th>
                    ))}
                  </tr>
                  {/* Data headers row */}
                  <tr>
                    <th className="sticky left-0 z-30 w-12 min-w-12 bg-gray-200 border border-gray-400 font-bold text-gray-700 text-center py-2 select-none text-xs">
                      1
                    </th>
                    {sheetData[0].map((header, index) => (
                      <th
                        key={index}
                        className="bg-gray-50 border border-gray-300 text-left font-semibold py-2 px-3 truncate"
                        style={{ width: '120px' }}
                        title={header || ''}
                      >
                        {header || `Col ${index + 1}`}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {sheetData.slice(1, 101).map((row, rowIndex) => (
                    <tr key={rowIndex} className={rowIndex % 2 === 0 ? 'bg-white' : 'bg-gray-50/50 hover:bg-blue-50/30'}>
                      <td className="sticky left-0 z-10 w-12 min-w-12 bg-gray-200 border border-gray-400 font-bold text-gray-700 text-center py-2 select-none text-xs">
                        {rowIndex + 2}
                      </td>
                      {row.map((cell, cellIndex) => (
                        <td
                          key={cellIndex}
                          className="border border-gray-300 py-2 px-3 overflow-hidden text-ellipsis whitespace-nowrap cursor-text"
                          style={{ width: '120px' }}
                          title={cell !== undefined && cell !== null ? String(cell) : ''}
                        >
                          {cell !== undefined && cell !== null ? cell : ''}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
              {sheetData.length > 101 && (
                <div className="sticky bottom-0 left-0 right-0 bg-gray-100 border-t-2 border-gray-400 px-4 py-2 text-sm text-gray-600 font-medium">
                  Mostrando primeiras 100 linhas de {sheetData.length} linhas totais
                </div>
              )}
              </div>
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
