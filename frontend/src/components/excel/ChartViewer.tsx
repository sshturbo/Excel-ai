// ChartViewer component - Professional chart viewer with advanced customization
import { useState, useMemo, useEffect, useCallback } from 'react'
import { Bar, Line, Pie, Doughnut, Scatter } from 'react-chartjs-2'
import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    BarElement,
    LineElement,
    PointElement,
    Title,
    Tooltip,
    Legend,
    ArcElement,
    Filler
} from 'chart.js'
import ChartDataLabels from 'chartjs-plugin-datalabels'
import type { PreviewDataType } from '@/types'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

// Hook to detect dark mode
function useIsDarkMode() {
    const [isDark, setIsDark] = useState(false)

    useEffect(() => {
        // Check initial state
        const checkDark = () => {
            setIsDark(document.documentElement.classList.contains('dark'))
        }
        checkDark()

        // Watch for changes
        const observer = new MutationObserver(checkDark)
        observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

        return () => observer.disconnect()
    }, [])

    return isDark
}

// Get theme-aware colors
function getThemeColors(isDark: boolean) {
    return {
        foreground: isDark ? '#e2e8f0' : '#1e293b',
        mutedForeground: isDark ? '#94a3b8' : '#64748b',
        background: isDark ? '#0f172a' : '#ffffff',
        card: isDark ? '#1e293b' : '#ffffff',
        border: isDark ? '#334155' : '#e2e8f0',
        gridColor: isDark ? 'rgba(71, 85, 105, 0.3)' : 'rgba(203, 213, 225, 0.5)',
    }
}

// Register Chart.js components
ChartJS.register(
    CategoryScale,
    LinearScale,
    BarElement,
    LineElement,
    PointElement,
    Title,
    Tooltip,
    Legend,
    ArcElement,
    Filler,
    ChartDataLabels
)

// Chart color palette - more vibrant
const CHART_COLORS = [
    '#3b82f6', // Blue
    '#ef4444', // Red
    '#22c55e', // Green
    '#f59e0b', // Amber
    '#8b5cf6', // Purple
    '#06b6d4', // Cyan
    '#ec4899', // Pink
    '#84cc16', // Lime
    '#f97316', // Orange
    '#6366f1', // Indigo
]

// Convert column index to Excel-style letter
function getColumnLetter(index: number): string {
    let letter = ''
    let temp = index
    while (temp >= 0) {
        letter = String.fromCharCode((temp % 26) + 65) + letter
        temp = Math.floor(temp / 26) - 1
    }
    return letter
}

type ChartType = 'bar' | 'bar-stacked' | 'horizontal' | 'line' | 'pie' | 'doughnut' | 'area' | 'scatter' | 'bullet'

interface ChartViewerProps {
    previewData: PreviewDataType
}

export function ChartViewer({ previewData }: ChartViewerProps) {
    // Theme detection
    const isDark = useIsDarkMode()
    const themeColors = useMemo(() => getThemeColors(isDark), [isDark])

    // Chart configuration
    const [chartType, setChartType] = useState<ChartType>('bar')
    const [chartTitle, setChartTitle] = useState('Gráfico de Dados')

    // Column selection
    const [labelColumn, setLabelColumn] = useState<number>(0)
    const [valueColumns, setValueColumns] = useState<number[]>([])
    const [goalColumn, setGoalColumn] = useState<number | null>(null)

    // Display options
    const [maxRows, setMaxRows] = useState(15)
    const [showLegend, setShowLegend] = useState(true)
    const [showValues, setShowValues] = useState(true)
    const [showPercentage, setShowPercentage] = useState(false)
    const [showGrid, setShowGrid] = useState(true)

    // Style options
    const [barWidth, setBarWidth] = useState<'thin' | 'normal' | 'thick'>('normal')
    const [colorScheme, setColorScheme] = useState<'default' | 'warm' | 'cool' | 'mono'>('default')

    // Get column options
    const columnOptions = useMemo(() => {
        return previewData.headers?.map((header, index) => ({
            index,
            letter: getColumnLetter(index),
            name: header || `Coluna ${getColumnLetter(index)}`,
            isNumeric: previewData.rows?.slice(0, 10).some(row => {
                const val = row[index]
                return val !== '' && val !== null && !isNaN(parseFloat(String(val)))
            }) ?? false
        })) || []
    }, [previewData])

    // Get colors based on scheme
    const getColors = useMemo(() => {
        switch (colorScheme) {
            case 'warm':
                return ['#ef4444', '#f97316', '#f59e0b', '#eab308', '#84cc16']
            case 'cool':
                return ['#3b82f6', '#06b6d4', '#14b8a6', '#22c55e', '#8b5cf6']
            case 'mono':
                return ['#1e293b', '#475569', '#64748b', '#94a3b8', '#cbd5e1']
            default:
                return CHART_COLORS
        }
    }, [colorScheme])

    // Bar width settings
    const barSettings = useMemo(() => {
        switch (barWidth) {
            case 'thin': return { barPercentage: 0.4, categoryPercentage: 0.7 }
            case 'thick': return { barPercentage: 0.95, categoryPercentage: 0.95 }
            default: return { barPercentage: 0.7, categoryPercentage: 0.85 }
        }
    }, [barWidth])

    // Auto-detect numeric columns on load
    useMemo(() => {
        if (valueColumns.length === 0 && columnOptions.length > 1) {
            const numericCols = columnOptions.filter((col, i) => i > 0 && col.isNumeric).slice(0, 2)
            if (numericCols.length > 0) {
                setValueColumns(numericCols.map(c => c.index))
            } else if (columnOptions.length > 1) {
                setValueColumns([1])
            }
        }
    }, [columnOptions])

    // Toggle value column
    const toggleValueColumn = (colIndex: number) => {
        if (colIndex === labelColumn) return
        setValueColumns(prev => {
            if (prev.includes(colIndex)) {
                return prev.filter(c => c !== colIndex)
            }
            return [...prev, colIndex]
        })
    }

    // Prepare chart data
    const chartData = useMemo(() => {
        if (!previewData.rows?.length || valueColumns.length === 0) return null

        const rows = previewData.rows.slice(0, maxRows)
        const labels = rows.map((row, i) => row[labelColumn] || `Item ${i + 1}`)

        // Bullet chart - goal comparison
        if (chartType === 'bullet' && goalColumn !== null) {
            const actualValues = rows.map(row => parseFloat(String(row[valueColumns[0]] || 0)) || 0)
            const goalValues = rows.map(row => parseFloat(String(row[goalColumn] || 0)) || 0)

            return {
                labels,
                datasets: [
                    {
                        label: previewData.headers?.[goalColumn] || 'Meta',
                        data: goalValues,
                        backgroundColor: '#e2e8f0',
                        borderColor: '#94a3b8',
                        borderWidth: 1,
                        ...barSettings,
                        barPercentage: 0.9,
                        order: 2
                    },
                    {
                        label: previewData.headers?.[valueColumns[0]] || 'Realizado',
                        data: actualValues,
                        backgroundColor: actualValues.map((val, i) => {
                            const goal = goalValues[i]
                            if (goal === 0) return getColors[0]
                            const percent = (val / goal) * 100
                            if (percent >= 100) return '#22c55e'
                            if (percent >= 70) return '#eab308'
                            return '#ef4444'
                        }),
                        borderWidth: 0,
                        ...barSettings,
                        barPercentage: 0.5,
                        order: 1
                    }
                ],
                _goalValues: goalValues
            }
        }

        // Standard charts - multiple series
        const datasets = valueColumns.map((colIndex, datasetIndex) => {
            const values = rows.map(row => parseFloat(String(row[colIndex] || 0)) || 0)
            const color = getColors[datasetIndex % getColors.length]
            const header = previewData.headers?.[colIndex] || `Série ${datasetIndex + 1}`

            const isPieType = chartType === 'pie' || chartType === 'doughnut'

            return {
                label: header,
                data: values,
                backgroundColor: isPieType
                    ? values.map((_, i) => getColors[i % getColors.length])
                    : chartType === 'line' || chartType === 'scatter'
                        ? `${color}40`
                        : color,
                borderColor: color,
                borderWidth: chartType === 'line' || chartType === 'area' ? 3 : 1,
                fill: chartType === 'area',
                tension: chartType === 'line' || chartType === 'area' ? 0.4 : 0,
                pointRadius: chartType === 'scatter' ? 8 : 4,
                pointHoverRadius: 10,
                pointBackgroundColor: color,
                ...barSettings
            }
        })

        return { labels, datasets }
    }, [previewData, labelColumn, valueColumns, goalColumn, maxRows, chartType, getColors, barSettings])

    // Chart options
    const chartOptions = useMemo(() => {
        const isHorizontal = chartType === 'horizontal' || chartType === 'bullet'
        const isPieType = chartType === 'pie' || chartType === 'doughnut'
        const isStacked = chartType === 'bar-stacked'

        return {
            responsive: true,
            maintainAspectRatio: false,
            indexAxis: isHorizontal ? 'y' as const : 'x' as const,
            plugins: {
                legend: {
                    display: showLegend,
                    position: 'top' as const,
                    labels: {
                        color: themeColors.foreground,
                        font: { size: 13, weight: 500 },
                        usePointStyle: true,
                        padding: 20
                    }
                },
                title: {
                    display: !!chartTitle,
                    text: chartTitle,
                    color: themeColors.foreground,
                    font: { size: 18, weight: 'bold' as const },
                    padding: { bottom: 20 }
                },
                tooltip: {
                    enabled: true,
                    backgroundColor: themeColors.card,
                    titleColor: themeColors.foreground,
                    bodyColor: themeColors.mutedForeground,
                    borderColor: themeColors.border,
                    borderWidth: 1,
                    padding: 12,
                    displayColors: true,
                    callbacks: chartType === 'bullet' ? {
                        afterLabel: (context: any) => {
                            const goalValues = (chartData as any)?._goalValues
                            if (context.datasetIndex === 1 && goalValues) {
                                const goal = goalValues[context.dataIndex]
                                if (goal > 0) {
                                    return `${((context.raw / goal) * 100).toFixed(1)}% da meta`
                                }
                            }
                            return ''
                        }
                    } : undefined
                },
                datalabels: {
                    display: showValues || showPercentage,
                    color: (context: any) => {
                        if (chartType === 'bullet') return context.datasetIndex === 0 ? themeColors.mutedForeground : '#fff'
                        if (isPieType) return '#fff'
                        return themeColors.foreground
                    },
                    font: { weight: 'bold' as const, size: isPieType ? 13 : 11 },
                    anchor: isPieType ? 'center' as const : 'end' as const,
                    align: isPieType ? 'center' as const : isHorizontal ? 'end' as const : 'top' as const,
                    offset: 4,
                    formatter: (value: number, context: any) => {
                        if (value === 0) return ''

                        if (showPercentage) {
                            if (chartType === 'bullet') {
                                const goalValues = (chartData as any)?._goalValues
                                if (context.datasetIndex === 1 && goalValues?.[context.dataIndex]) {
                                    return `${((value / goalValues[context.dataIndex]) * 100).toFixed(0)}%`
                                }
                                return ''
                            }
                            if (isPieType) {
                                const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
                                return `${((value / total) * 100).toFixed(1)}%`
                            }
                        }

                        if (showValues) {
                            if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
                            if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
                            return value.toLocaleString('pt-BR')
                        }
                        return ''
                    }
                }
            },
            scales: !isPieType ? {
                x: {
                    stacked: isStacked || chartType === 'bullet',
                    display: true,
                    ticks: {
                        color: themeColors.mutedForeground,
                        font: { size: 11 },
                        maxRotation: isHorizontal ? 0 : 45
                    },
                    grid: {
                        display: showGrid,
                        color: themeColors.gridColor
                    }
                },
                y: {
                    stacked: isStacked || chartType === 'bullet',
                    display: true,
                    beginAtZero: true,
                    ticks: {
                        color: themeColors.mutedForeground,
                        font: { size: 11 }
                    },
                    grid: {
                        display: showGrid,
                        color: themeColors.gridColor
                    }
                }
            } : undefined
        }
    }, [chartTitle, showLegend, showValues, showPercentage, showGrid, chartType, chartData, themeColors])

    const renderChart = () => {
        if (!chartData) {
            return (
                <div className="flex-1 flex items-center justify-center text-muted-foreground">
                    <div className="text-center space-y-3">
                        <p className="text-2xl">📊</p>
                        <p className="text-lg font-medium">Selecione as colunas de valores</p>
                        <p className="text-sm max-w-md">
                            Escolha uma coluna para identificadores (ex: Nome, Mês)
                            e uma ou mais colunas de valores numéricos (ex: Vendas, Meta)
                        </p>
                    </div>
                </div>
            )
        }

        if (chartType === 'bar' || chartType === 'bar-stacked' || chartType === 'horizontal' || chartType === 'bullet') {
            return <Bar data={chartData} options={chartOptions} />
        }
        if (chartType === 'line' || chartType === 'area') {
            return <Line data={chartData} options={chartOptions} />
        }
        if (chartType === 'pie') return <Pie data={chartData} options={chartOptions} />
        if (chartType === 'doughnut') return <Doughnut data={chartData} options={chartOptions} />
        if (chartType === 'scatter') return <Scatter data={chartData} options={chartOptions} />

        return <Bar data={chartData} options={chartOptions} />
    }

    return (
        <div className="flex-1 flex overflow-hidden">
            {/* Configuration Panel */}
            <div className="w-80 shrink-0 bg-card border-r border-border overflow-y-auto">
                <Tabs defaultValue="data" className="w-full">
                    <TabsList className="w-full grid grid-cols-3 p-1 m-2">
                        <TabsTrigger value="data" className="text-xs">📊 Dados</TabsTrigger>
                        <TabsTrigger value="style" className="text-xs">🎨 Estilo</TabsTrigger>
                        <TabsTrigger value="options" className="text-xs">⚙️ Opções</TabsTrigger>
                    </TabsList>

                    {/* Data Tab */}
                    <TabsContent value="data" className="p-4 space-y-5 mt-0">
                        {/* Chart Type */}
                        <div className="space-y-2">
                            <Label className="text-sm font-semibold">Tipo de Gráfico</Label>
                            <div className="grid grid-cols-3 gap-1.5">
                                {[
                                    { type: 'bar', icon: '📊', label: 'Barras' },
                                    { type: 'bar-stacked', icon: '📚', label: 'Empilhado' },
                                    { type: 'horizontal', icon: '📶', label: 'Horizontal' },
                                    { type: 'bullet', icon: '🎯', label: 'Metas' },
                                    { type: 'line', icon: '📈', label: 'Linha' },
                                    { type: 'area', icon: '🏔️', label: 'Área' },
                                    { type: 'pie', icon: '🥧', label: 'Pizza' },
                                    { type: 'doughnut', icon: '🍩', label: 'Rosca' },
                                    { type: 'scatter', icon: '⭐', label: 'Pontos' }
                                ].map(({ type, icon, label }) => (
                                    <Button
                                        key={type}
                                        variant={chartType === type ? 'default' : 'outline'}
                                        size="sm"
                                        onClick={() => setChartType(type as ChartType)}
                                        className={cn(
                                            "h-auto py-1.5 flex-col gap-0.5",
                                            chartType === type && "ring-2 ring-primary ring-offset-1"
                                        )}
                                    >
                                        <span className="text-sm">{icon}</span>
                                        <span className="text-[10px]">{label}</span>
                                    </Button>
                                ))}
                            </div>
                        </div>

                        {/* Label Column */}
                        <div className="space-y-2">
                            <Label className="text-sm font-semibold">
                                📋 Identificadores (Eixo X)
                            </Label>
                            <Select value={String(labelColumn)} onValueChange={(v) => setLabelColumn(parseInt(v))}>
                                <SelectTrigger className="h-10">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {columnOptions.map((col) => (
                                        <SelectItem key={col.index} value={String(col.index)}>
                                            <div className="flex items-center gap-2">
                                                <span className="font-mono text-xs bg-primary/20 text-primary px-1.5 py-0.5 rounded">{col.letter}</span>
                                                <span className="truncate">{col.name}</span>
                                            </div>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <p className="text-xs text-muted-foreground">
                                Ex: Nome do produto, Mês, Categoria
                            </p>
                        </div>

                        {/* Value Columns */}
                        {chartType === 'bullet' ? (
                            <>
                                <div className="space-y-2">
                                    <Label className="text-sm font-semibold">📊 Valor Realizado</Label>
                                    <Select
                                        value={valueColumns[0]?.toString() || ''}
                                        onValueChange={(v) => setValueColumns([parseInt(v)])}
                                    >
                                        <SelectTrigger className="h-10">
                                            <SelectValue placeholder="Selecione a coluna..." />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {columnOptions.filter(c => c.index !== labelColumn && c.isNumeric).map((col) => (
                                                <SelectItem key={col.index} value={String(col.index)}>
                                                    <div className="flex items-center gap-2">
                                                        <span className="font-mono text-xs bg-primary/20 text-primary px-1.5 py-0.5 rounded">{col.letter}</span>
                                                        <span>{col.name}</span>
                                                    </div>
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="space-y-2">
                                    <Label className="text-sm font-semibold">🎯 Meta</Label>
                                    <Select
                                        value={goalColumn?.toString() || ''}
                                        onValueChange={(v) => setGoalColumn(parseInt(v))}
                                    >
                                        <SelectTrigger className="h-10">
                                            <SelectValue placeholder="Selecione a meta..." />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {columnOptions.filter(c => c.index !== labelColumn && c.index !== valueColumns[0] && c.isNumeric).map((col) => (
                                                <SelectItem key={col.index} value={String(col.index)}>
                                                    <div className="flex items-center gap-2">
                                                        <span className="font-mono text-xs bg-amber-500/20 text-amber-600 px-1.5 py-0.5 rounded">{col.letter}</span>
                                                        <span>{col.name}</span>
                                                    </div>
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="p-3 bg-muted/50 rounded-lg space-y-2">
                                    <p className="text-xs font-medium">Legenda de Cores:</p>
                                    <div className="grid gap-1 text-xs">
                                        <div className="flex items-center gap-2"><div className="w-3 h-3 rounded bg-green-500" /> ≥100% atingido</div>
                                        <div className="flex items-center gap-2"><div className="w-3 h-3 rounded bg-yellow-500" /> 70-99% próximo</div>
                                        <div className="flex items-center gap-2"><div className="w-3 h-3 rounded bg-red-500" /> &lt;70% atrasado</div>
                                    </div>
                                </div>
                            </>
                        ) : (
                            <div className="space-y-2">
                                <div className="flex items-center justify-between">
                                    <Label className="text-sm font-semibold">📈 Valores (Séries)</Label>
                                    {valueColumns.length > 0 && (
                                        <span className="text-xs bg-primary/20 text-primary px-2 py-0.5 rounded-full">
                                            {valueColumns.length} selecionada{valueColumns.length > 1 ? 's' : ''}
                                        </span>
                                    )}
                                </div>
                                <div className="border border-border rounded-lg p-2 max-h-48 overflow-y-auto space-y-1">
                                    {columnOptions.filter(col => col.index !== labelColumn).map((col) => (
                                        <button
                                            key={col.index}
                                            onClick={() => toggleValueColumn(col.index)}
                                            className={cn(
                                                "w-full flex items-center gap-2 px-2 py-2 rounded-md text-sm transition-colors text-left",
                                                valueColumns.includes(col.index)
                                                    ? "bg-primary text-primary-foreground"
                                                    : "hover:bg-muted",
                                                !col.isNumeric && "opacity-50"
                                            )}
                                        >
                                            <span className={cn(
                                                "font-mono text-xs px-1.5 py-0.5 rounded",
                                                valueColumns.includes(col.index)
                                                    ? "bg-primary-foreground/20"
                                                    : "bg-muted"
                                            )}>
                                                {col.letter}
                                            </span>
                                            <span className="flex-1 truncate">{col.name}</span>
                                            {!col.isNumeric && <span title="Não numérico">📝</span>}
                                            {valueColumns.includes(col.index) && <span>✓</span>}
                                        </button>
                                    ))}
                                </div>
                                <p className="text-xs text-muted-foreground">
                                    Clique para selecionar múltiplas colunas de valores
                                </p>
                            </div>
                        )}

                        {/* Max Rows */}
                        <div className="space-y-2">
                            <Label className="text-sm font-semibold">Quantidade de Itens</Label>
                            <Select value={String(maxRows)} onValueChange={(v) => setMaxRows(parseInt(v))}>
                                <SelectTrigger className="h-10">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {[5, 10, 15, 20, 30, 50].map(n => (
                                        <SelectItem key={n} value={String(n)}>{n} itens</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                    </TabsContent>

                    {/* Style Tab */}
                    <TabsContent value="style" className="p-4 space-y-5 mt-0">
                        {/* Title */}
                        <div className="space-y-2">
                            <Label className="text-sm font-semibold">Título do Gráfico</Label>
                            <Input
                                value={chartTitle}
                                onChange={(e) => setChartTitle(e.target.value)}
                                placeholder="Digite o título..."
                                className="h-10"
                            />
                        </div>

                        {/* Color Scheme */}
                        <div className="space-y-2">
                            <Label className="text-sm font-semibold">Esquema de Cores</Label>
                            <div className="grid grid-cols-2 gap-2">
                                {[
                                    { id: 'default', name: 'Padrão', colors: CHART_COLORS.slice(0, 4) },
                                    { id: 'warm', name: 'Quentes', colors: ['#ef4444', '#f97316', '#f59e0b', '#eab308'] },
                                    { id: 'cool', name: 'Frias', colors: ['#3b82f6', '#06b6d4', '#14b8a6', '#22c55e'] },
                                    { id: 'mono', name: 'Mono', colors: ['#1e293b', '#475569', '#64748b', '#94a3b8'] }
                                ].map(scheme => (
                                    <button
                                        key={scheme.id}
                                        onClick={() => setColorScheme(scheme.id as any)}
                                        className={cn(
                                            "p-3 rounded-lg border-2 transition-all",
                                            colorScheme === scheme.id
                                                ? "border-primary bg-primary/5"
                                                : "border-border hover:border-primary/50"
                                        )}
                                    >
                                        <div className="flex gap-1 mb-2">
                                            {scheme.colors.map((c, i) => (
                                                <div key={i} className="w-4 h-4 rounded" style={{ backgroundColor: c }} />
                                            ))}
                                        </div>
                                        <span className="text-xs font-medium">{scheme.name}</span>
                                    </button>
                                ))}
                            </div>
                        </div>

                        {/* Bar Width */}
                        <div className="space-y-2">
                            <Label className="text-sm font-semibold">Largura das Barras</Label>
                            <div className="grid grid-cols-3 gap-2">
                                {[
                                    { id: 'thin', name: 'Finas' },
                                    { id: 'normal', name: 'Normal' },
                                    { id: 'thick', name: 'Grossas' }
                                ].map(opt => (
                                    <Button
                                        key={opt.id}
                                        variant={barWidth === opt.id ? 'default' : 'outline'}
                                        size="sm"
                                        onClick={() => setBarWidth(opt.id as any)}
                                    >
                                        {opt.name}
                                    </Button>
                                ))}
                            </div>
                        </div>
                    </TabsContent>

                    {/* Options Tab */}
                    <TabsContent value="options" className="p-4 space-y-3 mt-0">
                        <Label className="text-sm font-semibold">Exibição</Label>

                        <div className="space-y-2">
                            {[
                                { key: 'showLegend', label: 'Mostrar Legenda', value: showLegend, setter: setShowLegend },
                                { key: 'showGrid', label: 'Mostrar Grade', value: showGrid, setter: setShowGrid },
                                { key: 'showValues', label: 'Mostrar Valores', value: showValues, setter: setShowValues },
                                { key: 'showPercentage', label: 'Mostrar Percentual', value: showPercentage, setter: setShowPercentage }
                            ].map(opt => (
                                <button
                                    key={opt.key}
                                    onClick={() => opt.setter(!opt.value)}
                                    className={cn(
                                        "w-full flex items-center justify-between px-3 py-2.5 rounded-lg border transition-all",
                                        opt.value
                                            ? "bg-primary/10 border-primary/30 text-primary"
                                            : "border-border hover:bg-muted"
                                    )}
                                >
                                    <span className="text-sm">{opt.label}</span>
                                    <span className="text-base">{opt.value ? '✅' : '⬜'}</span>
                                </button>
                            ))}
                        </div>
                    </TabsContent>
                </Tabs>
            </div>

            {/* Chart Display */}
            <div className="flex-1 flex flex-col bg-background">
                {/* Header */}
                <div className="flex items-center justify-between px-4 py-3 bg-muted/40 border-b border-border">
                    <div className="flex items-center gap-3">
                        <span className="text-lg">📊</span>
                        <span className="font-medium">{chartTitle || 'Gráfico'}</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        {valueColumns.length > 0 && (
                            <span className="bg-primary/10 text-primary px-2 py-1 rounded">
                                {valueColumns.length} série{valueColumns.length > 1 ? 's' : ''}
                            </span>
                        )}
                        <span>{Math.min(maxRows, previewData.rows?.length || 0)} itens</span>
                    </div>
                </div>

                {/* Chart */}
                <div className="flex-1 p-6 min-h-0">
                    <div className="w-full h-full">
                        {renderChart()}
                    </div>
                </div>
            </div>
        </div>
    )
}
