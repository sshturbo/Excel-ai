// DataPreview component - Table preview of spreadsheet data
import type { PreviewDataType } from '@/types'

interface DataPreviewProps {
    previewData: PreviewDataType
}

export function DataPreview({ previewData }: DataPreviewProps) {
    return (
        <div className="flex-1 overflow-auto p-4">
            <table className="w-full border-collapse text-sm">
                <thead>
                    <tr>
                        {previewData.headers?.map((h, i) => (
                            <th key={i} className="border border-border bg-muted/60 p-2 text-left sticky top-0 text-foreground">
                                {h}
                            </th>
                        ))}
                    </tr>
                </thead>
                <tbody>
                    {previewData.rows?.slice(0, 20).map((row, i) => (
                        <tr key={i} className="hover:bg-muted/40">
                            {row.map((cell, j) => (
                                <td key={j} className="border border-border p-2">{cell}</td>
                            ))}
                        </tr>
                    ))}
                </tbody>
            </table>
            {previewData.rows?.length > 20 && (
                <p className="text-center text-muted-foreground text-sm mt-3">
                    ... e mais {previewData.rows.length - 20} linhas
                </p>
            )}
        </div>
    )
}
