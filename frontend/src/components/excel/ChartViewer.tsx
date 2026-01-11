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

// Helper to safely parse localized numbers (BR/US)
function parseValue(val: any): number {
    if (typeof val === 'number') return val
    if (!val) return 0

    let str = String(val).trim()

    // Remove everything that is NOT a number, dot, comma, or minus
    // This handles R$, $, USD, spaces, etc automatically
    str = str.replace(/[^0-9,.-]/g, '')

    // Handle empty after strip
    if (!str) return 0

    // Check for BR format: 1.234,56 or 1234,56
    // Logic: has comma, and either no dots OR last dot comes before last comma
    const lastComma = str.lastIndexOf(',')
    const lastDot = str.lastIndexOf('.')

    if (lastComma > -1 && (lastDot === -1 || lastDot < lastComma)) {
        // BR detected: remove thousands separator (.) and replace decimal (,) with (.)
        str = str.replace(/\./g, '').replace(',', '.')
    } else {
        // US/Standard detected: 1,234.56 or 1234.56
        // Remove thousands separator (,)
        str = str.replace(/,/g, '')
    }

    const parsed = parseFloat(str)
    return isNaN(parsed) ? 0 : parsed
}

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
    const [colorScheme, setColorScheme] = useState('blue')

    // Get column options
    const columnOptions = useMemo(() => {
        return previewData.headers?.map((header, index) => ({
            index,
            letter: getColumnLetter(index),
            name: header || `Coluna ${getColumnLetter(index)}`,
            isNumeric: previewData.rows?.slice(0, 10).some(row => {
                const val = row[index]
                // Use robust parsing to check if it's a number
                return val !== '' && val !== null && !isNaN(parseValue(val)) && parseValue(val) !== 0
            }) ?? false
        })) || []
    }, [previewData])

    // Get colors based on scheme
    const getColors = useMemo(() => {
        switch (colorScheme) {
            case 'green':
                return ['#22c55e', '#16a34a', '#15803d', '#166534', '#14532d']
            case 'purple':
                return ['#a855f7', '#9333ea', '#7e22ce', '#6b21a8', '#581c87']
            case 'orange':
                return ['#f97316', '#ea580c', '#c2410c', '#9a3412', '#7c2d12']
            case 'gray':
                return ['#64748b', '#475569', '#334155', '#1e293b', '#0f172a']
            case 'mixed':
                return ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6']
            case 'blue':
            default:
                return ['#3b82f6', '#2563eb', '#1d4ed8', '#1e40af', '#1e3a8a']
        }
    }, [colorScheme])


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

    // State for filtering and scrolling
    const [filterText, setFilterText] = useState('')
    const [enableScroll, setEnableScroll] = useState(false)
    const [barSettings, setBarSettings] = useState({ barPercentage: 0.6 })

    // Prepare chart data
    const chartData = useMemo(() => {
        if (!previewData.rows?.length || valueColumns.length === 0) return null

        // Filter rows first (search in whole dataset), then slice
        const allMatchingRows = filterText.trim()
            ? previewData.rows.filter((row: any) => {
                const label = row[labelColumn] || ''
                return String(label).toLowerCase().includes(filterText.toLowerCase())
            })
            : previewData.rows

        // Apply limit after filtering
        const filteredRows = allMatchingRows.slice(0, maxRows)

        const labels = filteredRows.map((row: any, i: number) => row[labelColumn] || `Item ${i + 1}`)

        // Helper for parsing (reuse existing if possible or define here)
        // Since we are replacing the whole block, let's redefine parseValue or reuse if in scope.
        // It relies on parseValue defined outside? No, previous code had it inside or used global?
        // I will define it inside to be safe as previously it was missing in the small view.

        const parseValue = (val: any) => {
            if (val === null || val === undefined || val === '') return 0
            if (typeof val === 'number') return val
            const strVal = String(val).trim()
            const cleanStr = strVal.replace(/[^0-9,.-]/g, '')
            if (cleanStr.includes(',') && cleanStr.includes('.')) {
                if (cleanStr.indexOf(',') < cleanStr.indexOf('.')) return parseFloat(cleanStr.replace(/,/g, ''))
                else return parseFloat(cleanStr.replace(/\./g, '').replace(',', '.'))
            } else if (cleanStr.includes(',')) return parseFloat(cleanStr.replace(',', '.'))
            return parseFloat(cleanStr)
        }

        // Bullet chart - goal comparison
        if (chartType === 'bullet' && goalColumn !== null) {
            const actualValues = filteredRows.map((row: any) => parseValue(row[valueColumns[0]]))
            const goalValues = filteredRows.map((row: any) => parseValue(row[goalColumn]))

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
                        label: previewData.headers?.[valueColumns[0]] || 'Valor',
                        data: actualValues,
                        backgroundColor: (context: any) => {
                            const value = context.raw
                            const goal = goalValues[context.dataIndex]
                            if (goal === 0) return getColors[0] // Default color if goal is zero
                            if (value >= goal) return '#22c55e' // Green
                            if (value >= goal * 0.8) return '#eab308' // Yellow
                            return '#ef4444' // Red
                        },
                        borderColor: 'transparent',
                        borderWidth: 0,
                        barPercentage: 0.5,
                        order: 1
                    }
                ],
                _goalValues: goalValues // Store for tooltip
            }
        }

        // Standard charts - multiple series
        const datasets = valueColumns.map((colIndex, datasetIndex) => {
            const values = filteredRows.map(row => parseValue(row[colIndex]))
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
    }, [previewData, labelColumn, valueColumns, goalColumn, maxRows, chartType, getColors, barSettings, filterText])

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
                    align: 'start' as const,
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
                    align: 'start' as const,
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
                    callbacks: {
                        label: (context: any) => {
                            let label = context.dataset.label || ''
                            if (label) label += ': '
                            if (context.parsed.y !== null) {
                                label += new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(context.parsed.y)
                            }
                            return label
                        },
                        afterLabel: chartType === 'bullet' ? (context: any) => {
                            const goalValues = (chartData as any)?._goalValues
                            if (context.datasetIndex === 1 && goalValues) {
                                const goal = goalValues[context.dataIndex]
                                if (goal > 0) {
                                    return `${((context.raw / goal) * 100).toFixed(1)}% da meta`
                                }
                            }
                            return ''
                        } : undefined
                    }
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
                            if (value >= 1000000) return `R$ ${(value / 1000000).toFixed(1)}M`
                            if (value >= 1000) return `R$ ${(value / 1000).toFixed(1)}K`
                            return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value)
                        }
                        return ''
                    }
                },
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

        // Use pixels instead of % to avoid canvas crash on large datasets (browser limit ~32k px)
        // 60px per bar/point is a good density for readable labels
        // Added 32000px cap to prevent canvas rendering failure on extremely large datasets
        const rawSize = chartData.labels ? chartData.labels.length * 60 : 0
        const dynamicSize = enableScroll && chartData.labels ? Math.min(32000, Math.max(400, rawSize)) + 'px' : '100%'

        const isHorizontal = chartType === 'horizontal' || chartType === 'bullet'

        const containerStyle = enableScroll
            ? isHorizontal
                ? { width: '100%', height: dynamicSize, minHeight: '100%' } // Scroll vertical
                : { width: dynamicSize, minWidth: '100%', height: '100%' } // Scroll horizontal
            : { width: '100%', height: '100%' }

        return (
            <div className="w-full h-full overflow-hidden">
                <div
                    className={cn(
                        "h-full relative transition-all",
                        enableScroll ? "overflow-auto" : "overflow-hidden"
                    )}
                >
                    <div style={containerStyle} className="relative">
                        {chartType === 'bar' || chartType === 'bar-stacked' || chartType === 'horizontal' || chartType === 'bullet' ? (
                            <Bar data={chartData} options={chartOptions} />
                        ) : null}
                        {chartType === 'line' || chartType === 'area' ? (
                            <Line data={chartData} options={chartOptions} />
                        ) : null}
                        {chartType === 'pie' ? <Pie data={chartData} options={chartOptions} /> : null}
                        {chartType === 'doughnut' ? <Doughnut data={chartData} options={chartOptions} /> : null}
                        {chartType === 'scatter' ? <Scatter data={chartData} options={chartOptions} /> : null}
                    </div>
                </div>
            </div>
        )
    }

    // Panel state
    const [isPanelOpen, setIsPanelOpen] = useState(true)

    return (
        <div className="w-full h-full relative overflow-hidden bg-background">
            {/* Background Chart Area - fills space but respects sidebar via padding */}
            <div
                className="absolute inset-0 flex flex-col transition-all duration-300 ease-in-out overflow-hidden"
                style={{ paddingLeft: isPanelOpen ? '320px' : '0' }}
            >
                {/* Header */}
                <div className="flex items-center justify-between px-4 py-3 bg-muted/40 border-b border-border shrink-0">
                    <div className="flex items-center gap-3">
                        <span className="text-lg">📊</span>
                        <span className="font-medium">{chartTitle || 'Gráfico'}</span>
                    </div>
                    {/* Centered Filter Input */}
                    <div className="flex-1 max-w-sm mx-4">
                        <div className="relative">
                            <span className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground text-xs">🔍</span>
                            <Input
                                value={filterText}
                                onChange={(e) => setFilterText(e.target.value)}
                                placeholder="Filtrar itens do gráfico..."
                                className="h-8 pl-8 text-xs bg-background"
                            />
                        </div>
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

                {/* Chart Canvas Container */}
                <div className="flex-1 p-6 min-h-0 relative">
                    {renderChart()}
                </div>
            </div>

            {/* Sidebar Configuration Panel - Absolutely positioned on left */}
            <div
                className={cn(
                    "absolute left-0 top-0 bottom-0 bg-card border-r border-border z-20 transition-all duration-300 ease-in-out flex flex-col shadow-xl",
                    isPanelOpen ? "w-80 translate-x-0" : "w-80 -translate-x-full"
                )}
            >
                {/* Scrollable Content inside Sidebar */}
                <div className="flex-1 overflow-y-auto w-full">
                    <Tabs defaultValue="data" className="w-full">
                        <TabsList className="w-full grid grid-cols-3 p-1 m-2">
                            <TabsTrigger value="data" className="text-xs">📊 Dados</TabsTrigger>
                            <TabsTrigger value="style" className="text-xs">🎨 Estilo</TabsTrigger>
                            <TabsTrigger value="options" className="text-xs">⚙️ Opções</TabsTrigger>
                        </TabsList>

                        {/* Data Tab */}
                        <TabsContent value="data" className="p-4 pt-0 space-y-4 data-[state=inactive]:hidden">
                            {/* Chart Type Config */}
                            <div className="space-y-2">
                                <Label className="text-sm font-semibold">Tipo de Gráfico</Label>
                                <div className="grid grid-cols-3 gap-2">
                                    {[
                                        { type: 'bar', label: 'Barras', icon: '📊' },
                                        { type: 'bar-stacked', label: 'Empilhado', icon: '🏗️' },
                                        { type: 'horizontal', label: 'Horizontal', icon: '📶' },
                                        { type: 'bullet', label: 'Metas', icon: '🎯' },
                                        { type: 'line', label: 'Linha', icon: '📈' },
                                        { type: 'area', label: 'Área', icon: '🏔️' },
                                        { type: 'pie', label: 'Pizza', icon: '🥧' },
                                        { type: 'doughnut', label: 'Rosca', icon: '🍩' },
                                        { type: 'scatter', label: 'Pontos', icon: '⭐' },
                                    ].map(({ type, label, icon }) => (
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
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {columnOptions.map((col) => (
                                                    <SelectItem key={col.index} value={String(col.index)}>
                                                        {col.name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                    </div>
                                    <div className="space-y-2">
                                        <Label className="text-sm font-semibold">🎯 Meta (Goal)</Label>
                                        <Select
                                            value={goalColumn?.toString() || ''}
                                            onValueChange={(v) => setGoalColumn(parseInt(v))}
                                        >
                                            <SelectTrigger className="h-10">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {columnOptions.map((col) => (
                                                    <SelectItem key={col.index} value={String(col.index)}>
                                                        {col.name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                    </div>
                                </>
                            ) : (
                                <div className="space-y-2">
                                    <div className="flex items-center justify-between">
                                        <Label className="text-sm font-semibold">📈 Valores (Séries)</Label>
                                        <span className="text-xs text-muted-foreground">
                                            {valueColumns.length} selecionadas
                                        </span>
                                    </div>
                                    <div className="border rounded-md max-h-48 overflow-y-auto p-1 bg-background">
                                        {columnOptions.map((col) => {
                                            const isSelected = valueColumns.includes(col.index)
                                            return (
                                                <div
                                                    key={col.index}
                                                    className={cn(
                                                        "flex items-center gap-2 px-2 py-1.5 cursor-pointer rounded text-sm transition-colors",
                                                        isSelected ? "bg-primary/10 text-primary font-medium" : "hover:bg-muted"
                                                    )}
                                                    onClick={() => toggleValueColumn(col.index)}
                                                >
                                                    <div className={cn(
                                                        "w-4 h-4 border rounded flex items-center justify-center transition-colors",
                                                        isSelected ? "bg-primary border-primary" : "border-muted-foreground/30"
                                                    )}>
                                                        {isSelected && <span className="text-primary-foreground-foreground text-[10px]">✓</span>}
                                                    </div>
                                                    <div className="flex items-center gap-2 flex-1 min-w-0">
                                                        <span className="font-mono text-xs text-muted-foreground w-4">{col.letter}</span>
                                                        <span className="truncate">{col.name}</span>
                                                    </div>
                                                </div>
                                            )
                                        })}
                                    </div>
                                    <p className="text-xs text-muted-foreground">
                                        Clique para selecionar múltiplas colunas de valores
                                    </p>
                                </div>
                            )}
                        </TabsContent>

                        {/* Style Tab */}
                        <TabsContent value="style" className="p-4 space-y-4 mt-0">
                            <div className="space-y-2">
                                <Label className="text-sm font-semibold">🎨 Esquema de Cores</Label>
                                <Select value={colorScheme} onValueChange={(v: any) => setColorScheme(v)}>
                                    <SelectTrigger className="h-10">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="blue">🔵 Azul Corporativo</SelectItem>
                                        <SelectItem value="green">🟢 Verde Natureza</SelectItem>
                                        <SelectItem value="purple">🟣 Roxo Criativo</SelectItem>
                                        <SelectItem value="orange">🟠 Laranja Quente</SelectItem>
                                        <SelectItem value="gray">⚫ Neutro / Cinza</SelectItem>
                                        <SelectItem value="mixed">🌈 Misto / Vibrante</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            <div className="space-y-4 pt-2">
                                <Label className="text-sm font-semibold">Espessura das Barras</Label>
                                <div className="space-y-3">
                                    <div className="flex items-center justify-between">
                                        <span className="text-xs text-muted-foreground">Mais finas</span>
                                        <span className="text-xs text-muted-foreground">Mais grossas</span>
                                    </div>
                                    <input
                                        type="range"
                                        min="0.1"
                                        max="1.0"
                                        step="0.1"
                                        value={barSettings.barPercentage}
                                        onChange={(e) => setBarSettings(prev => ({ ...prev, barPercentage: parseFloat(e.target.value) }))}
                                        className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer"
                                    />
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

                            {/* Filter and Scroll */}
                            <div className="space-y-4 pt-2 border-t border-border">
                                <div className="flex items-center justify-between">
                                    <Label className="text-sm font-semibold cursor-pointer" htmlFor="scroll-mode">
                                        ↔️ Rolagem Horizontal
                                    </Label>
                                    <div className="flex items-center gap-2">
                                        <input
                                            type="checkbox"
                                            id="scroll-mode"
                                            checked={enableScroll}
                                            onChange={(e) => setEnableScroll(e.target.checked)}
                                            className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
                                        />
                                        <span className="text-xs text-muted-foreground">{enableScroll ? 'On' : 'Off'}</span>
                                    </div>
                                </div>
                            </div>

                            {/* Max Rows */}
                            <div className="space-y-2">
                                <Label className="text-sm font-semibold">Quantidade de Itens</Label>
                                <Select value={String(maxRows)} onValueChange={(v) => setMaxRows(parseInt(v))}>
                                    <SelectTrigger className="h-10">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {[5, 10, 15, 20, 30, 50, 100].map(n => (
                                            <SelectItem key={n} value={String(n)}>{n} itens</SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                        </TabsContent>
                    </Tabs>
                </div>
            </div>

            {/* Toggle Button - Now Fixed overlay on top-left of chart area */}
            <Button
                variant="secondary"
                size="sm"
                className="absolute z-30 w-6 h-12 rounded-r-lg rounded-l-none border-y border-r border-border p-0 flex items-center justify-center shadow-md bg-background hover:bg-muted transition-transform duration-300"
                style={{
                    left: isPanelOpen ? '320px' : '0px',
                    top: '50%',
                    transform: 'translateY(-50%)',
                    transition: 'left 300ms ease-in-out'
                }}
                onClick={() => setIsPanelOpen(!isPanelOpen)}
                title={isPanelOpen ? "Recolher menu" : "Expandir menu"}
            >
                {isPanelOpen ? '◀' : '▶'}
            </Button>
        </div>
    )
}
