// Toolbar component - Preview and Chart toggle buttons
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

interface ToolbarProps {
    showPreview: boolean
    showChart: boolean
    chartType: 'bar' | 'line' | 'pie'
    onTogglePreview: () => void
    onToggleChart: () => void
    onChartTypeChange: (type: 'bar' | 'line' | 'pie') => void
}

export function Toolbar({
    showPreview,
    showChart,
    chartType,
    onTogglePreview,
    onToggleChart,
    onChartTypeChange
}: ToolbarProps) {
    return (
        <div className="flex items-center gap-2 p-3 bg-card/60 border-b border-border">
            <Button
                variant={showPreview ? "default" : "outline"}
                size="sm"
                onClick={onTogglePreview}
            >
                📋 Preview
            </Button>
            <Button
                variant={showChart ? "default" : "outline"}
                size="sm"
                onClick={onToggleChart}
            >
                📊 Gráfico
            </Button>
            {showChart && (
                <Select value={chartType} onValueChange={(v) => onChartTypeChange(v as 'bar' | 'line' | 'pie')}>
                    <SelectTrigger className="w-32 ml-auto">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="bar">Barras</SelectItem>
                        <SelectItem value="line">Linha</SelectItem>
                        <SelectItem value="pie">Pizza</SelectItem>
                    </SelectContent>
                </Select>
            )}
        </div>
    )
}
