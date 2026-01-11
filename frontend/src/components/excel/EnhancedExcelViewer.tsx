import React, { useState, useRef, useEffect } from 'react';
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

interface EnhancedExcelViewerProps {
  previewData: PreviewData | null;
  sheetData: string[][] | null;
  loading?: boolean;
  onDownload: () => Promise<void>;
  onRefreshSheet?: (sheetName: string) => Promise<void>;
  onSheetChange?: (sheetName: string) => void;
  downloading?: boolean;
}

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

export const EnhancedExcelViewer: React.FC<EnhancedExcelViewerProps> = ({
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
  const [selectedCell, setSelectedCell] = useState<{ row: number; col: number } | null>(null);
  const gridRef = useRef<HTMLDivElement>(null);

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

  const totalRows = sheetData?.length || 0;
  const totalCols = sheetData?.[0]?.length || 0;
  const displayRows = Math.min(totalRows, 1000); // Limit to 1000 rows for performance

  return (
    <div className="flex flex-col h-full space-y-4">
      {/* Header */}
      <Card className="p-4 shrink-0">
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
      <Card className="p-4 shrink-0">
        <div className="space-y-2">
          {previewData.sheets.map((sheet) => {
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
                    {expandedSheets.has(sheet.name) ? (
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

      {/* Excel-style Grid */}
      <Card className="flex-1 flex flex-col overflow-hidden">
        {currentSheet && (loading || sheetData) ? (
          <>
            {/* Sheet Header */}
            <div className="p-3 border-b border-gray-200 bg-gray-50 shrink-0">
              <h4 className="font-semibold">{currentSheet}</h4>
              <p className="text-xs text-gray-500 mt-0.5">
                {loading 
                  ? 'Carregando dados...' 
                  : `${totalRows} linhas × ${totalCols} colunas`
                }
                {selectedCell && (
                  <span className="ml-3 font-mono bg-white px-2 py-0.5 rounded border border-gray-300">
                    {getColumnLetter(selectedCell.col)}{selectedCell.row + 1}
                  </span>
                )}
              </p>
            </div>

            {loading ? (
              <div className="flex-1 flex items-center justify-center">
                <div className="flex items-center gap-3 text-blue-600">
                  <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
                  <span className="font-medium">Carregando dados da planilha...</span>
                </div>
              </div>
            ) : sheetData && sheetData.length > 0 ? (
              <div className="flex-1 overflow-x-auto overflow-y-auto relative bg-white">
                <div ref={gridRef} className="inline-block min-w-full" style={{ minWidth: 'fit-content' }}>
                  <table 
                    className="border-collapse text-sm"
                    style={{ 
                      tableLayout: 'fixed',
                      minWidth: `${Math.max(totalCols, 5) * 100}px`,
                    }}
                  >
                    {/* Column Header Row */}
                    <thead className="sticky top-0 z-20">
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
                        {/* Column letter headers (A, B, C...) */}
                        {sheetData[0].map((_, colIndex) => (
                          <th
                            key={colIndex}
                            className="bg-gray-200 border border-gray-400 font-bold text-gray-700 text-center py-2 select-none hover:bg-gray-300 transition-colors"
                            style={{ width: '100px' }}
                          >
                            {getColumnLetter(colIndex)}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {sheetData.slice(0, displayRows).map((row, rowIndex) => (
                        <tr key={rowIndex} className="hover:bg-blue-50/30">
                          {/* Row number */}
                          <td 
                            className="sticky left-0 z-10 w-12 min-w-12 bg-gray-200 border border-gray-400 font-bold text-gray-700 text-center py-2 select-none"
                            style={{
                              boxShadow: '2px 0 4px rgba(0,0,0,0.1)'
                            }}
                          >
                            {rowIndex + 1}
                          </td>
                          {/* Data cells */}
                          {row.map((cell, colIndex) => (
                            <td
                              key={`${rowIndex}-${colIndex}`}
                              className={`border border-gray-300 py-1.5 px-2 overflow-hidden text-ellipsis whitespace-nowrap cursor-text transition-colors
                                ${selectedCell?.row === rowIndex && selectedCell?.col === colIndex 
                                  ? 'bg-blue-100 ring-2 ring-blue-500 ring-inset' 
                                  : ''
                                }
                              `}
                              style={{ width: '100px' }}
                              onClick={() => setSelectedCell({ row: rowIndex, col: colIndex })}
                              title={cell || ''}
                            >
                              {cell !== undefined && cell !== null ? cell : ''}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                {totalRows > displayRows && (
                  <div className="fixed bottom-0 left-0 right-0 bg-gray-100 border-t border-gray-300 px-4 py-2 text-sm text-gray-600">
                    Mostrando {displayRows} de {totalRows} linhas totais
                  </div>
                )}
              </div>
            ) : (
              <div className="flex-1 flex items-center justify-center text-gray-500">
                <p className="text-lg">Planilha vazia</p>
              </div>
            )}
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-gray-500">
            <p className="text-lg">Selecione uma planilha para visualizar</p>
          </div>
        )}
      </Card>
    </div>
  );
};
