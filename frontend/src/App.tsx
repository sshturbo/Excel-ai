import { useState, useEffect, useRef, useMemo } from 'react'
import { toast } from 'sonner'
import ReactMarkdown, { Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
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
    ArcElement
} from 'chart.js'
import { Bar, Line, Pie } from 'react-chartjs-2'

// Wails bindings
import {
    ConnectExcel,
    RefreshWorkbooks,
    SetExcelContext,
    SetAPIKey,
    SetModel,
    SendMessage,
    ClearChat,
    GetPreviewData,
    ListConversations,
    LoadConversation,
    DeleteConversation,
    NewConversation,
    GetSavedConfig,
    WriteToExcel,
    ApplyFormula,
    UpdateExcelCell,
    UndoLastChange,
    DeleteLastMessages,
    StartUndoBatch,
    EndUndoBatch,
    CreateNewWorkbook,
    CreateNewSheet,
    CreateChart,
    CreatePivotTable,
    ConfigurePivotFields,
    QueryExcel,
    SendErrorFeedback,
    FormatRange,
    DeleteSheet,
    RenameSheet,
    ClearRange,
    AutoFitColumns,
    InsertRows,
    DeleteRows,
    MergeCells,
    UnmergeCells,
    SetBorders,
    SetColumnWidth,
    SetRowHeight,
    ApplyFilter,
    ClearFilters,
    SortRange,
    CopyRange,
    ListCharts,
    DeleteChartByName,
    CreateTable,
    ListTables,
    DeleteTable,
    CancelChat
} from "../wailsjs/go/main/App"
import { EventsOn } from "../wailsjs/runtime/runtime"

// shadcn components
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Slider } from "@/components/ui/slider"
import { Switch } from "@/components/ui/switch"

// Settings component
import Settings from './Settings'

ChartJS.register(
    CategoryScale,
    LinearScale,
    BarElement,
    LineElement,
    PointElement,
    Title,
    Tooltip,
    Legend,
    ArcElement
)

interface Message {
    role: 'user' | 'assistant'
    content: string
    hasActions?: boolean
}

interface Workbook {
    name: string
    path?: string
    sheets: string[]
}

interface ConversationItem {
    id: string
    title: string
    updatedAt: string
}

interface PreviewDataType {
    headers: string[]
    rows: string[][]
    totalRows: number
}

export default function App() {
    // States
    const [connected, setConnected] = useState(false)
    const [workbooks, setWorkbooks] = useState<Workbook[]>([])
    const [selectedWorkbook, setSelectedWorkbook] = useState<string | null>(null)
    const [selectedSheets, setSelectedSheets] = useState<string[]>([])
    const [messages, setMessages] = useState<Message[]>([])
    const [inputMessage, setInputMessage] = useState('')
    const [isLoading, setIsLoading] = useState(false)
    const [apiKey, setApiKey] = useState('')
    const [model, setModel] = useState('openai/gpt-4o-mini')
    const [contextLoaded, setContextLoaded] = useState('')
    const [previewData, setPreviewData] = useState<PreviewDataType | null>(null)
    const [showPreview, setShowPreview] = useState(false)
    const [conversations, setConversations] = useState<ConversationItem[]>([])
    const [showChart, setShowChart] = useState(false)
    const [chartType, setChartType] = useState<'bar' | 'line' | 'pie'>('bar')
    const [chartData, setChartData] = useState<any>(null)
    const [expandedWorkbook, setExpandedWorkbook] = useState<string | null>(null)
    const [showSettings, setShowSettings] = useState(false)
    const [askBeforeApply, setAskBeforeApply] = useState(true)
    const [pendingActions, setPendingActions] = useState<any[]>([])
    const [editingMessageIndex, setEditingMessageIndex] = useState<number | null>(null)
    const [editContent, setEditContent] = useState('')

    const messagesEndRef = useRef<HTMLDivElement>(null)
    const inputRef = useRef<HTMLTextAreaElement>(null)

    // Componentes customizados para Markdown
    const markdownComponents: Components = useMemo(() => ({
        // Código inline
        code({ node, className, children, ...props }) {
            const match = /language-(\w+)/.exec(className || '')
            const isInline = !match && !className

            if (isInline) {
                return (
                    <code className="px-1.5 py-0.5 rounded bg-muted text-primary font-mono text-sm" {...props}>
                        {children}
                    </code>
                )
            }

            // Bloco de código com syntax highlighting
            return (
                <div className="relative group my-3">
                    <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                            onClick={() => {
                                navigator.clipboard.writeText(String(children).replace(/\n$/, ''))
                                toast.success('Código copiado!')
                            }}
                            className="px-2 py-1 text-xs bg-muted hover:bg-muted/80 rounded border border-border"
                        >
                            📋 Copiar
                        </button>
                    </div>
                    {match && (
                        <div className="text-xs text-muted-foreground px-3 py-1 bg-muted/50 border-b border-border rounded-t">
                            {match[1]}
                        </div>
                    )}
                    <SyntaxHighlighter
                        style={oneDark}
                        language={match?.[1] || 'text'}
                        PreTag="div"
                        customStyle={{
                            margin: 0,
                            borderRadius: match ? '0 0 0.5rem 0.5rem' : '0.5rem',
                            fontSize: '0.85rem',
                        }}
                    >
                        {String(children).replace(/\n$/, '')}
                    </SyntaxHighlighter>
                </div>
            )
        },
        // Tabelas
        table({ children }) {
            return (
                <div className="overflow-x-auto my-3">
                    <table className="min-w-full border border-border rounded-lg overflow-hidden">
                        {children}
                    </table>
                </div>
            )
        },
        thead({ children }) {
            return <thead className="bg-muted/50">{children}</thead>
        },
        th({ children }) {
            return <th className="px-3 py-2 text-left text-sm font-semibold border-b border-border">{children}</th>
        },
        td({ children }) {
            return <td className="px-3 py-2 text-sm border-b border-border/50">{children}</td>
        },
        // Listas
        ul({ children }) {
            return <ul className="list-disc list-inside space-y-1 my-2 ml-2">{children}</ul>
        },
        ol({ children }) {
            return <ol className="list-decimal list-inside space-y-1 my-2 ml-2">{children}</ol>
        },
        li({ children }) {
            return <li className="text-sm">{children}</li>
        },
        // Títulos
        h1({ children }) {
            return <h1 className="text-xl font-bold mt-4 mb-2 text-primary">{children}</h1>
        },
        h2({ children }) {
            return <h2 className="text-lg font-bold mt-3 mb-2 text-primary">{children}</h2>
        },
        h3({ children }) {
            return <h3 className="text-base font-semibold mt-2 mb-1">{children}</h3>
        },
        // Parágrafos
        p({ children }) {
            return <p className="my-2 leading-relaxed">{children}</p>
        },
        // Links
        a({ href, children }) {
            return (
                <a
                    href={href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                >
                    {children}
                </a>
            )
        },
        // Blockquote
        blockquote({ children }) {
            return (
                <blockquote className="border-l-4 border-primary/50 pl-4 my-3 italic text-muted-foreground bg-muted/20 py-2 rounded-r">
                    {children}
                </blockquote>
            )
        },
        // Horizontal rule
        hr() {
            return <hr className="my-4 border-border" />
        },
        // Strong/Bold
        strong({ children }) {
            return <strong className="font-semibold text-primary">{children}</strong>
        },
        // Emphasis/Italic
        em({ children }) {
            return <em className="italic">{children}</em>
        },
    }), [])

    // Função para limpar conteúdo técnico da resposta
    const cleanTechnicalBlocks = (content: string): string => {
        // Remove blocos :::excel-action (completamente)
        let cleaned = content.replace(/:::excel-action\s*[\s\S]*?\s*:::/g, '')
        // Remove blocos :::excel-query (completamente)
        cleaned = cleaned.replace(/:::excel-query\s*[\s\S]*?\s*:::/g, '')
        // Remove linhas vazias múltiplas
        cleaned = cleaned.replace(/\n{3,}/g, '\n\n')
        return cleaned.trim()
    }

    // Função para renderizar conteúdo com suporte a :::thinking blocks
    const renderMessageContent = (content: string) => {
        // Primeiro limpa os blocos técnicos
        const cleanedContent = cleanTechnicalBlocks(content)

        if (!cleanedContent) {
            return null
        }

        const thinkingRegex = /:::thinking\s*([\s\S]*?)\s*:::/g
        const parts: JSX.Element[] = []
        let lastIndex = 0
        let match
        let key = 0

        while ((match = thinkingRegex.exec(cleanedContent)) !== null) {
            // Adiciona texto antes do bloco thinking
            if (match.index > lastIndex) {
                const textBefore = cleanedContent.slice(lastIndex, match.index)
                if (textBefore.trim()) {
                    parts.push(
                        <ReactMarkdown key={key++} remarkPlugins={[remarkGfm]} components={markdownComponents}>
                            {textBefore}
                        </ReactMarkdown>
                    )
                }
            }

            // Adiciona o bloco thinking com design melhorado
            const thinkingContent = match[1].trim()
            const lines = thinkingContent.split('\n').filter(line => line.trim())

            parts.push(
                <div key={key++} className="my-3 overflow-hidden rounded-lg border border-blue-500/20 bg-blue-500/5">
                    <div className="flex items-center gap-2 px-3 py-2 bg-blue-500/10 border-b border-blue-500/20">
                        <span className="text-blue-400">💭</span>
                        <span className="text-xs font-medium text-blue-400">Raciocínio</span>
                    </div>
                    <div className="p-3 space-y-1.5">
                        {lines.map((line, i) => (
                            <div key={i} className="flex items-start gap-2 text-xs text-muted-foreground/80">
                                <span className="text-blue-400/60 mt-0.5">→</span>
                                <span>{line.trim().replace(/^\d+\.\s*/, '').replace(/^[-•]\s*/, '')}</span>
                            </div>
                        ))}
                    </div>
                </div>
            )

            lastIndex = match.index + match[0].length
        }

        // Adiciona texto restante após o último bloco thinking
        if (lastIndex < cleanedContent.length) {
            const textAfter = cleanedContent.slice(lastIndex)
            if (textAfter.trim()) {
                parts.push(
                    <ReactMarkdown key={key++} remarkPlugins={[remarkGfm]} components={markdownComponents}>
                        {textAfter}
                    </ReactMarkdown>
                )
            }
        }

        // Se não houver partes válidas após limpeza
        if (parts.length === 0) {
            // Se houver conteúdo limpo, renderiza
            if (cleanedContent.trim()) {
                return (
                    <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                        {cleanedContent}
                    </ReactMarkdown>
                )
            }
            return null
        }

        return <>{parts}</>
    }

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, [messages])

    useEffect(() => {
        const hasWailsRuntime = typeof (window as any)?.go !== 'undefined'
        if (hasWailsRuntime) {
            handleConnect()
            loadConfig()
            loadConversations()
        } else {
            console.warn('Wails runtime não detectado. Rodando fora do app (Vite puro).')
        }
    }, [])

    const loadConfig = async () => {
        try {
            const config = await GetSavedConfig()
            if (config) {
                if (config.apiKey) setApiKey(config.apiKey)
                if (config.model) setModel(config.model)
            }
        } catch (err) {
            console.error('Error loading config:', err)
        }
    }

    const handleConnect = async () => {
        if (typeof (window as any)?.go === 'undefined') {
            toast.error('Wails não detectado. Abra pelo app (wails dev/build).')
            return
        }
        try {
            const result = await ConnectExcel()
            setConnected(result.connected)
            if (result.connected && result.workbooks) {
                setWorkbooks(result.workbooks)
                toast.success('Conectado ao Excel!')
            } else if (result.error) {
                toast.error('Erro ao conectar: ' + result.error)
            }
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao conectar: ' + errorMessage)
        }
    }

    const handleSelectSheet = async (wbName: string, sheetName: string) => {
        // Toggle sheet selection (multi-select)
        const isSelected = selectedWorkbook === wbName && selectedSheets.includes(sheetName)

        if (selectedWorkbook !== wbName) {
            // Mudou de workbook, reset selection
            setSelectedWorkbook(wbName)
            setSelectedSheets([sheetName])
        } else if (isSelected) {
            // Deselect sheet
            const newSheets = selectedSheets.filter(s => s !== sheetName)
            setSelectedSheets(newSheets)
            if (newSheets.length === 0) {
                setContextLoaded('')
                setPreviewData(null)
                return
            }
        } else {
            // Add sheet to selection
            setSelectedSheets([...selectedSheets, sheetName])
        }

        const sheetsToLoad = isSelected
            ? selectedSheets.filter(s => s !== sheetName)
            : selectedWorkbook === wbName
                ? [...selectedSheets, sheetName]
                : [sheetName]

        if (sheetsToLoad.length === 0) return

        setContextLoaded('')
        setPreviewData(null)

        try {
            // Load context for all selected sheets
            await SetExcelContext(wbName, sheetsToLoad.join(','))

            // Get preview for first selected sheet
            const data = await GetPreviewData(wbName, sheetsToLoad[0])
            if (data) {
                setPreviewData(data)
                const sheetNames = sheetsToLoad.join(', ')
                setContextLoaded(`${sheetNames} (${sheetsToLoad.length} aba${sheetsToLoad.length > 1 ? 's' : ''})`)
                toast.success(`Contexto carregado: ${sheetsToLoad.length} aba(s)`)
                prepareChartData(data)
            }
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao carregar: ' + errorMessage)
            console.error('Erro ao carregar planilha:', err)
        }
    }

    const prepareChartData = (data: PreviewDataType) => {
        if (!data?.rows?.length) return
        const labels = data.rows.slice(0, 10).map((r, i) => r[0] || `Item ${i + 1}`)
        const numericCol = data.rows[0]?.findIndex((_, i) => i > 0 && !isNaN(parseFloat(data.rows[0][i])))
        const values = data.rows.slice(0, 10).map(r => parseFloat(r[numericCol > 0 ? numericCol : 1]) || 0)

        setChartData({
            labels,
            datasets: [{
                label: data.headers?.[numericCol > 0 ? numericCol : 1] || 'Valores',
                data: values,
                backgroundColor: ['#667eea', '#764ba2', '#f093fb', '#f5576c', '#4facfe', '#00f2fe', '#43e97b', '#38f9d7', '#fa709a', '#fee140'],
                borderColor: '#667eea',
                borderWidth: 1
            }]
        })
    }

    const [streamingContent, setStreamingContent] = useState('')
    const rawStreamBufferRef = useRef('')

    useEffect(() => {
        const cleanup = EventsOn("chat:chunk", (chunk: string) => {
            rawStreamBufferRef.current += chunk

            // Processar conteúdo imediatamente
            let cleanContent = rawStreamBufferRef.current

            // Remove blocos excel-action completos
            cleanContent = cleanContent.replace(/:::excel-action\s*[\s\S]*?\s*:::/g, '')
            // Remove blocos excel-query completos
            cleanContent = cleanContent.replace(/:::excel-query\s*[\s\S]*?\s*:::/g, '')

            // Verificar se há bloco técnico incompleto no final (NÃO thinking)
            // Só ocultar se for excel-action ou excel-query incompleto
            const incompleteActionMatch = cleanContent.match(/:::excel-action[\s\S]*$/)
            const incompleteQueryMatch = cleanContent.match(/:::excel-query[\s\S]*$/)

            if (incompleteActionMatch) {
                cleanContent = cleanContent.replace(/:::excel-action[\s\S]*$/, '')
            }
            if (incompleteQueryMatch) {
                cleanContent = cleanContent.replace(/:::excel-query[\s\S]*$/, '')
            }

            cleanContent = cleanContent.replace(/\n{3,}/g, '\n\n').trim()

            // Se está vazio mas há atividade técnica, mostrar status
            const hasIncompleteAction = /:::excel-action(?![\s\S]*:::)/.test(rawStreamBufferRef.current)
            const hasIncompleteQuery = /:::excel-query(?![\s\S]*:::)/.test(rawStreamBufferRef.current)
            if (!cleanContent && (hasIncompleteAction || hasIncompleteQuery)) {
                cleanContent = hasIncompleteQuery ? '🔍 Consultando...' : '⏳ Executando...'
            }

            setStreamingContent(cleanContent)
        })
        return () => cleanup()
    }, [])

    // Ref para manter selectedWorkbook atualizado no callback sem recriar listener
    const selectedWorkbookRef = useRef(selectedWorkbook)
    useEffect(() => {
        selectedWorkbookRef.current = selectedWorkbook
    }, [selectedWorkbook])

    // Listener para atualizações em tempo real das planilhas
    useEffect(() => {
        const cleanup = EventsOn("excel:workbooks-changed", (data: any) => {
            if (data?.workbooks !== undefined) {
                setWorkbooks(data.workbooks || [])
                setConnected(data.connected ?? true)

                // Se o workbook selecionado foi fechado, limpar seleção
                const currentSelected = selectedWorkbookRef.current
                if (currentSelected) {
                    const workbookStillExists = (data.workbooks || []).some(
                        (wb: any) => wb.name === currentSelected
                    )
                    if (!workbookStillExists) {
                        setSelectedWorkbook(null)
                        setSelectedSheets([])
                        setContextLoaded('')
                        setPreviewData(null)
                        setExpandedWorkbook(null)
                    }
                }
            }
        })
        return () => cleanup()
    }, [])

    // Resetar buffer quando não está carregando
    useEffect(() => {
        if (!isLoading) {
            rawStreamBufferRef.current = ''
        }
    }, [isLoading])

    // Atualizar mensagem apenas quando streamingContent muda significativamente
    const lastContentRef = useRef('')
    useEffect(() => {
        if (!isLoading || !streamingContent) return
        if (streamingContent === lastContentRef.current) return

        lastContentRef.current = streamingContent
        setMessages(prev => {
            const newMsgs = [...prev]
            const lastIndex = newMsgs.length - 1
            if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                if (newMsgs[lastIndex].content !== streamingContent) {
                    newMsgs[lastIndex] = { ...newMsgs[lastIndex], content: streamingContent }
                    return newMsgs
                }
            }
            return prev // Não criar novo array se não mudou
        })
    }, [streamingContent, isLoading])

    // Função auxiliar para executar ações do Excel
    const executeExcelAction = async (action: any): Promise<{ success: boolean; error?: string }> => {
        try {
            if (action.op === 'write') {
                await UpdateExcelCell(
                    action.workbook || '',
                    action.sheet || '',
                    action.cell,
                    action.value
                )
            } else if (action.op === 'create-workbook') {
                const name = await CreateNewWorkbook()
                toast.success(`Nova pasta de trabalho criada: ${name}`)
                const result = await RefreshWorkbooks()
                if (result.workbooks) setWorkbooks(result.workbooks)
            } else if (action.op === 'create-sheet') {
                await CreateNewSheet(action.name)
                toast.success(`Nova aba criada: ${action.name}`)
                const result = await RefreshWorkbooks()
                if (result.workbooks) setWorkbooks(result.workbooks)
                // Pequeno delay para garantir que o Excel processou a criação da aba
                await new Promise(resolve => setTimeout(resolve, 300))
            } else if (action.op === 'create-chart') {
                await CreateChart(
                    action.sheet || '',
                    action.range,
                    action.chartType || 'column',
                    action.title || ''
                )
                toast.success('Gráfico criado!')
            } else if (action.op === 'create-pivot') {
                console.log('[DEBUG] create-pivot:', action)
                const tableName = action.tableName || 'PivotTable1'
                await CreatePivotTable(
                    action.sourceSheet || '',
                    action.sourceRange,
                    action.destSheet || '',
                    action.destCell,
                    tableName
                )

                // Configurar campos se especificados
                if (action.rowFields || action.valueFields) {
                    // Pequeno delay para garantir que a tabela foi criada
                    await new Promise(resolve => setTimeout(resolve, 500))

                    const rowFields = action.rowFields || []
                    const valueFields = (action.valueFields || []).map((vf: any) => {
                        if (typeof vf === 'string') {
                            return { field: vf, function: 'sum' }
                        }
                        return vf
                    })

                    await ConfigurePivotFields(
                        action.destSheet || '',
                        tableName,
                        rowFields,
                        valueFields
                    )
                }

                toast.success('Tabela dinâmica criada!')
            } else if (action.op === 'format-range') {
                await FormatRange(
                    action.sheet || '',
                    action.range,
                    action.bold || false,
                    action.italic || false,
                    action.fontSize || 0,
                    action.fontColor || '',
                    action.bgColor || ''
                )
                toast.success('Formatação aplicada!')
            } else if (action.op === 'delete-sheet') {
                await DeleteSheet(action.name)
                toast.success(`Aba "${action.name}" excluída!`)
                const result = await RefreshWorkbooks()
                if (result.workbooks) setWorkbooks(result.workbooks)
            } else if (action.op === 'rename-sheet') {
                await RenameSheet(action.oldName, action.newName)
                toast.success(`Aba renomeada: ${action.oldName} → ${action.newName}`)
                const result = await RefreshWorkbooks()
                if (result.workbooks) setWorkbooks(result.workbooks)
            } else if (action.op === 'clear-range') {
                await ClearRange(action.sheet || '', action.range)
                toast.success('Conteúdo limpo!')
            } else if (action.op === 'autofit') {
                await AutoFitColumns(action.sheet || '', action.range)
                toast.success('Colunas ajustadas!')
            } else if (action.op === 'insert-rows') {
                await InsertRows(action.sheet || '', action.row, action.count || 1)
                toast.success(`${action.count || 1} linha(s) inserida(s)!`)
            } else if (action.op === 'delete-rows') {
                await DeleteRows(action.sheet || '', action.row, action.count || 1)
                toast.success(`${action.count || 1} linha(s) excluída(s)!`)
            } else if (action.op === 'merge-cells') {
                await MergeCells(action.sheet || '', action.range)
                toast.success('Células mescladas!')
            } else if (action.op === 'unmerge-cells') {
                await UnmergeCells(action.sheet || '', action.range)
                toast.success('Células desmescladas!')
            } else if (action.op === 'set-borders') {
                await SetBorders(action.sheet || '', action.range, action.style || 'thin')
                toast.success('Bordas aplicadas!')
            } else if (action.op === 'set-column-width') {
                await SetColumnWidth(action.sheet || '', action.range, action.width || 15)
                toast.success('Largura definida!')
            } else if (action.op === 'set-row-height') {
                await SetRowHeight(action.sheet || '', action.range, action.height || 20)
                toast.success('Altura definida!')
            } else if (action.op === 'apply-filter') {
                await ApplyFilter(action.sheet || '', action.range)
                toast.success('Filtro aplicado!')
            } else if (action.op === 'clear-filters') {
                await ClearFilters(action.sheet || '')
                toast.success('Filtros limpos!')
            } else if (action.op === 'sort') {
                await SortRange(action.sheet || '', action.range, action.column || 1, action.ascending !== false)
                toast.success('Dados ordenados!')
            } else if (action.op === 'copy-range') {
                await CopyRange(action.sheet || '', action.source, action.dest)
                toast.success('Range copiado!')
            } else if (action.op === 'list-charts') {
                const charts = await ListCharts(action.sheet || '')
                toast.info(`Gráficos encontrados: ${charts.join(', ') || 'nenhum'}`)
            } else if (action.op === 'delete-chart') {
                await DeleteChartByName(action.sheet || '', action.name)
                toast.success(`Gráfico "${action.name}" excluído!`)
            } else if (action.op === 'create-table') {
                await CreateTable(action.sheet || '', action.range, action.name || '', action.style || '')
                toast.success(`Tabela "${action.name || 'Tabela'}" criada!`)
            } else if (action.op === 'delete-table') {
                await DeleteTable(action.sheet || '', action.name)
                toast.success(`Tabela "${action.name}" removida!`)
            }
            return { success: true }
        } catch (e: any) {
            const errorMsg = e?.message || String(e)
            console.error("Erro na ação Excel:", errorMsg)
            return { success: false, error: errorMsg }
        }
    }

    // Função para processar resposta da IA e executar ações
    const processAIResponse = async (response: string, maxRetries: number = 2): Promise<{ displayContent: string; actionsExecuted: number }> => {
        // Processar queries primeiro (:::excel-query)
        const queryRegex = /:::excel-query\s*([\s\S]*?)\s*:::/g
        const queryMatches = [...response.matchAll(queryRegex)]

        // Resultados de queries para enviar de volta à IA
        const queryResults: string[] = []

        for (const match of queryMatches) {
            try {
                const jsonStr = match[1]
                // Suportar múltiplas queries no mesmo bloco
                const lines = jsonStr.split('\n').filter(l => l.trim().startsWith('{'))

                for (const line of lines) {
                    const query = JSON.parse(line.trim())
                    console.log('[QUERY]', query)

                    const result = await QueryExcel(query.type, {
                        name: query.name || '',
                        sheet: query.sheet || '',
                        range: query.range || ''
                    })

                    if (result.success) {
                        queryResults.push(`Query "${query.type}": ${JSON.stringify(result.data)}`)
                    } else {
                        queryResults.push(`Query "${query.type}" falhou: ${result.error}`)
                    }
                }
            } catch (err) {
                console.error('Erro ao processar query:', err)
            }
        }

        // Se temos resultados de query, podemos exibi-los no chat
        if (queryResults.length > 0) {
            console.log('[QUERY RESULTS]', queryResults)
        }

        const actionRegex = /:::excel-action\s*([\s\S]*?)\s*:::/g
        let matches = [...response.matchAll(actionRegex)]
        let actionsExecuted = 0
        let currentResponse = response
        let retryCount = 0
        let undoBatchStarted = false

        while (matches.length > 0 && retryCount <= maxRetries) {
            if (!askBeforeApply && retryCount === 0) {
                await StartUndoBatch()
                undoBatchStarted = true
            }

            let hasError = false
            let errorMessage = ''

            for (const match of matches) {
                try {
                    const jsonStr = match[1]
                    const cleanJson = jsonStr.replace(/```json/g, '').replace(/```/g, '').trim()
                    const action = JSON.parse(cleanJson)

                    if (askBeforeApply) {
                        setPendingActions(prev => [...prev, action])
                    } else {
                        const result = await executeExcelAction(action)
                        if (result.success) {
                            actionsExecuted++
                        } else {
                            hasError = true
                            errorMessage = result.error || 'Erro desconhecido'
                            toast.warning(`Erro: ${errorMessage}. Solicitando correção...`)
                            break // Para no primeiro erro para pedir correção
                        }
                    }
                } catch (e: any) {
                    console.error("Erro ao parsear ação", e)
                    hasError = true
                    errorMessage = `Erro ao processar comando: ${e?.message || e}`
                    break
                }
            }

            // Se houve erro, envia feedback para a IA
            if (hasError && retryCount < maxRetries) {
                retryCount++
                console.log(`[DEBUG] Enviando erro para IA (tentativa ${retryCount}):`, errorMessage)

                setStreamingContent('')
                toast.info(`Solicitando correção à IA (tentativa ${retryCount})...`)

                try {
                    const correctedResponse = await SendErrorFeedback(errorMessage)
                    currentResponse = correctedResponse
                    matches = [...correctedResponse.matchAll(actionRegex)]

                    // Atualiza mensagem com a nova resposta
                    const newDisplayContent = correctedResponse.replace(actionRegex, '').trim()
                    setMessages(prev => {
                        const newMsgs = [...prev]
                        const lastIndex = newMsgs.length - 1
                        if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                            newMsgs[lastIndex] = {
                                ...newMsgs[lastIndex],
                                content: newDisplayContent
                            }
                        }
                        return newMsgs
                    })
                } catch (feedbackErr) {
                    console.error("Erro ao enviar feedback:", feedbackErr)
                    const msg = feedbackErr instanceof Error ? feedbackErr.message : String(feedbackErr)
                    toast.error('Erro ao solicitar correção à IA: ' + msg)
                    break
                }
            } else {
                break
            }
        }

        // Sempre finalize o lote se tiver sido iniciado, mesmo que nenhuma ação tenha sido aplicada.
        if (!askBeforeApply && undoBatchStarted) {
            try {
                await EndUndoBatch()
            } catch (e) {
                console.warn('Falha ao finalizar lote de Undo:', e)
            }
        }

        const displayContent = currentResponse.replace(actionRegex, '').trim()
        return { displayContent, actionsExecuted }
    }

    const processMessage = async (text: string) => {
        setIsLoading(true)
        setStreamingContent('')

        // Add placeholder for assistant message
        setMessages(prev => [...prev, { role: 'assistant', content: '' }])

        const MAX_AGENT_ITERATIONS = 5 // Limite de iterações para evitar loops infinitos
        let currentMessage = text
        let totalActionsExecuted = 0
        let finalDisplayContent = ''
        let iteration = 0
        let continueAgent = true
        const previousQueries: string[] = [] // Para detectar loops repetitivos

        try {
            while (continueAgent) {
                // Loop principal do agente
                const startIteration = iteration

                while (iteration < startIteration + MAX_AGENT_ITERATIONS) {
                    iteration++
                    console.log(`[AGENT] Iteração ${iteration}`)

                    // Atualizar status visual durante o loop
                    if (iteration > 1) {
                        setStreamingContent(`🔄 Agente processando... (iteração ${iteration})`)
                    }

                    const response = await SendMessage(currentMessage)

                    // Processar queries primeiro
                    const queryRegex = /:::excel-query\s*([\s\S]*?)\s*:::/g
                    const queryMatches = [...response.matchAll(queryRegex)]
                    const queryResults: string[] = []

                    for (const match of queryMatches) {
                        try {
                            const jsonStr = match[1]
                            const lines = jsonStr.split('\n').filter(l => l.trim().startsWith('{'))

                            for (const line of lines) {
                                const query = JSON.parse(line.trim())
                                console.log('[AGENT QUERY]', query)

                                const result = await QueryExcel(query.type, {
                                    name: query.name || '',
                                    sheet: query.sheet || '',
                                    range: query.range || ''
                                })

                                if (result.success) {
                                    queryResults.push(`${query.type}: ${JSON.stringify(result.data)}`)
                                } else {
                                    queryResults.push(`${query.type} erro: ${result.error}`)
                                }
                            }
                        } catch (err) {
                            console.error('Erro ao processar query:', err)
                        }
                    }

                    // Processar ações
                    const { displayContent, actionsExecuted } = await processAIResponse(response)
                    totalActionsExecuted += actionsExecuted

                    // Se houve queries, enviamos os resultados de volta para a IA continuar
                    if (queryResults.length > 0) {
                        // Detectar loop repetitivo (mesma query enviada 2x seguidas)
                        const querySignature = queryResults.join('|')
                        if (previousQueries.length > 0 && previousQueries[previousQueries.length - 1] === querySignature) {
                            console.warn('[AGENT] Loop repetitivo detectado! Mesmas queries enviadas consecutivamente.')
                            toast.warning('Loop detectado - mesmas consultas repetidas. Agente interrompido.')
                            finalDisplayContent = displayContent + '\n\n⚠️ Agente interrompido: loop repetitivo detectado.'
                            continueAgent = false
                            break
                        }
                        previousQueries.push(querySignature)

                        console.log('[AGENT] Enviando resultados de volta:', queryResults)
                        currentMessage = `[Query Results]\n${queryResults.join('\n')}\n\nContinue with the task based on these results.`
                        // Continuar o loop
                    } else {
                        // Sem queries, terminamos o loop
                        finalDisplayContent = displayContent
                        continueAgent = false
                        break
                    }
                }

                // Verificar se atingiu o limite da rodada
                if (continueAgent && iteration >= startIteration + MAX_AGENT_ITERATIONS) {
                    console.warn('[AGENT] Limite de iterações atingido, perguntando ao usuário')

                    // Perguntar se quer continuar
                    const wantContinue = window.confirm(
                        `⚠️ Limite de ${MAX_AGENT_ITERATIONS} iterações atingido (total: ${iteration}).\n\n` +
                        `O agente ainda está processando a tarefa.\n` +
                        `Deseja continuar por mais ${MAX_AGENT_ITERATIONS} iterações?`
                    )

                    if (!wantContinue) {
                        finalDisplayContent += '\n\n⚠️ Agente interrompido pelo usuário.'
                        continueAgent = false
                    }
                }
            }

            if (totalActionsExecuted > 0) {
                toast.success(`${totalActionsExecuted} alterações aplicadas!`)
                // Refresh context if we have selected sheets
                if (selectedWorkbook && selectedSheets.length > 0) {
                    const contextMsg = await SetExcelContext(selectedWorkbook, selectedSheets.join(','))
                    setContextLoaded(contextMsg)
                    const preview = await GetPreviewData(selectedWorkbook, selectedSheets[0])
                    setPreviewData(preview)
                }
            }

            setMessages(prev => {
                const newMsgs = [...prev]
                const lastIndex = newMsgs.length - 1
                if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                    // Se não há conteúdo de texto mas houve ações, mostrar mensagem de sucesso
                    let finalContent = finalDisplayContent
                    if (!finalContent && totalActionsExecuted > 0) {
                        finalContent = `✅ ${totalActionsExecuted === 1 ? 'Ação executada' : `${totalActionsExecuted} ações executadas`} com sucesso!`
                    }

                    newMsgs[lastIndex] = {
                        ...newMsgs[lastIndex],
                        content: finalContent,
                        hasActions: totalActionsExecuted > 0
                    }
                }
                return newMsgs
            })
        } catch (err) {
            setMessages(prev => prev.slice(0, -1)) // Remove failed message
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro: ' + errorMessage)
        } finally {
            setIsLoading(false)
            inputRef.current?.focus()
        }
    }

    const handleSendMessage = async () => {
        if (!inputMessage.trim() || isLoading) return

        const userMessage = inputMessage.trim()
        setInputMessage('')
        setMessages(prev => [...prev, { role: 'user', content: userMessage }])

        await processMessage(userMessage)
    }

    const handleCancelChat = async () => {
        try {
            await CancelChat()
            setIsLoading(false)
            rawStreamBufferRef.current = ''
            setStreamingContent('')
            toast.info('⏹️ Chat interrompido')
        } catch (err) {
            console.error('Erro ao cancelar:', err)
        }
    }

    const handleRegenerate = async () => {
        if (isLoading || messages.length === 0) return

        let lastUserMsgIndex = -1
        for (let i = messages.length - 1; i >= 0; i--) {
            if (messages[i].role === 'user') {
                lastUserMsgIndex = i
                break
            }
        }
        if (lastUserMsgIndex === -1) return

        const text = messages[lastUserMsgIndex].content
        const countToRemove = messages.length - lastUserMsgIndex

        try {
            await DeleteLastMessages(countToRemove)
            setMessages(prev => prev.slice(0, lastUserMsgIndex))
            setMessages(prev => [...prev, { role: 'user', content: text }])
            await processMessage(text)
        } catch (err) {
            console.error(err)
            toast.error('Erro ao regenerar')
        }
    }

    const handleCopy = (text: string) => {
        navigator.clipboard.writeText(text)
        toast.success('Copiado para a área de transferência')
    }

    const handleShare = (text: string) => {
        navigator.clipboard.writeText(text)
        toast.success('Pronto para compartilhar!')
    }

    const handleEditMessage = (index: number, content: string) => {
        setEditingMessageIndex(index)
        setEditContent(content)
    }

    const handleSaveEdit = async (index: number) => {
        if (!editContent.trim()) return

        const countToRemove = messages.length - index

        try {
            await DeleteLastMessages(countToRemove)
            setMessages(prev => prev.slice(0, index))
            setMessages(prev => [...prev, { role: 'user', content: editContent }])
            setEditingMessageIndex(null)
            setEditContent('')
            await processMessage(editContent)
        } catch (err) {
            console.error(err)
            toast.error('Erro ao editar')
        }
    }

    const handleCancelEdit = () => {
        setEditingMessageIndex(null)
        setEditContent('')
    }

    const handleApplyActions = async () => {
        let executed = 0
        let errors = 0

        const toastId = toast.loading('Aplicando alterações...')

        if (pendingActions.length > 0) {
            await StartUndoBatch()
        }

        for (const action of pendingActions) {
            try {
                if (action.op === 'write') {
                    await UpdateExcelCell(
                        action.workbook || '',
                        action.sheet || '',
                        action.cell,
                        action.value
                    )
                } else if (action.op === 'create-workbook') {
                    await CreateNewWorkbook()
                } else if (action.op === 'create-sheet') {
                    await CreateNewSheet(action.name)
                    // Pequeno delay para garantir que o Excel processou a criação da aba
                    await new Promise(resolve => setTimeout(resolve, 300))
                } else if (action.op === 'create-chart') {
                    await CreateChart(
                        action.sheet || '',
                        action.range,
                        action.chartType || 'column',
                        action.title || ''
                    )
                } else if (action.op === 'create-pivot') {
                    console.log('[DEBUG] Criando pivot table:', {
                        sourceSheet: action.sourceSheet,
                        sourceRange: action.sourceRange,
                        destSheet: action.destSheet,
                        destCell: action.destCell,
                        tableName: action.tableName
                    })
                    await CreatePivotTable(
                        action.sourceSheet || '',
                        action.sourceRange,
                        action.destSheet || '',
                        action.destCell,
                        action.tableName || 'PivotTable1'
                    )
                }
                executed++
            } catch (e: any) {
                console.error("Erro ao aplicar ação:", action)
                console.error("Detalhes do erro:", e?.message || e)
                toast.error(`Erro: ${e?.message || 'Falha na operação'}`)
                errors++
            }
        }

        if (pendingActions.length > 0) {
            await EndUndoBatch()
        }

        toast.dismiss(toastId)

        if (executed > 0) {
            toast.success(`${executed} alterações aplicadas!`)
            if (errors > 0) {
                toast.warning(`${errors} falharam. Verifique o console.`)
            }
            setPendingActions([])

            // Update the last assistant message to show the Undo button
            setMessages(prev => {
                const newMsgs = [...prev]
                let lastIndex = -1
                for (let i = newMsgs.length - 1; i >= 0; i--) {
                    if (newMsgs[i].role === 'assistant') {
                        lastIndex = i
                        break
                    }
                }
                if (lastIndex >= 0) {
                    newMsgs[lastIndex] = { ...newMsgs[lastIndex], hasActions: true }
                }
                return newMsgs
            })

            if (selectedWorkbook && selectedSheets.length > 0) {
                const contextMsg = await SetExcelContext(selectedWorkbook, selectedSheets.join(','))
                setContextLoaded(contextMsg)
                const preview = await GetPreviewData(selectedWorkbook, selectedSheets[0])
                setPreviewData(preview)
            }
        } else if (errors > 0) {
            toast.error(`Falha ao aplicar alterações. ${errors} erros encontrados.`)
        }
    }

    const handleDiscardActions = () => {
        setPendingActions([])
        toast.info('Alterações descartadas')
    }

    const handleUndo = async () => {
        try {
            await UndoLastChange()
            toast.success('Alteração desfeita!')
            if (selectedWorkbook && selectedSheets.length > 0) {
                const preview = await GetPreviewData(selectedWorkbook, selectedSheets[0])
                setPreviewData(preview)
            }
        } catch (err) {
            toast.error('Nada para desfazer')
        }
    }

    const handleClearChat = async () => {
        try {
            await ClearChat()
            setMessages([])
            toast.success('Chat limpo')
        } catch (err) {
            toast.error('Erro ao limpar')
        }
    }

    const handleNewConversation = async () => {
        try {
            await NewConversation()
            setMessages([])
            setContextLoaded('')
            setPreviewData(null)
            setSelectedWorkbook(null)
            setSelectedSheets([])

            // Recarregar histórico de conversas
            const list = await ListConversations()
            if (list) setConversations(list)

            // Recarregar planilhas do Excel
            if (connected) {
                const result = await RefreshWorkbooks()
                if (result.workbooks) {
                    setWorkbooks(result.workbooks)
                }
            }

            toast.success('Nova conversa criada')
        } catch (err) {
            console.error(err)
        }
    }

    const loadConversations = async () => {
        try {
            const list = await ListConversations()
            console.log('[DEBUG] Conversas carregadas:', list)
            if (list && list.length > 0) {
                setConversations(list)
            } else {
                setConversations([])
            }
        } catch (err) {
            console.error('Erro ao carregar conversas:', err)
        }
    }

    const handleLoadConversation = async (convId: string) => {
        try {
            const messages = await LoadConversation(convId)
            if (messages && messages.length > 0) {
                const loadedMessages: Message[] = []

                for (const m of messages) {
                    // Ignorar mensagens do sistema (system prompt)
                    if (m.role === 'system') continue

                    // Ignorar mensagens internas do usuário (resultados de queries, feedback)
                    if (m.role === 'user') {
                        // Pular mensagens que são resultados de queries
                        if (m.content.startsWith('Resultados das queries:')) continue
                        if (m.content.startsWith('[ERRO na ação')) continue
                        if (m.content.startsWith('A ação anterior falhou')) continue
                        if (m.content.includes('Contexto do Excel:') && m.content.includes('Pergunta do usuário:')) {
                            // Extrair apenas a pergunta real do usuário
                            const match = m.content.match(/Pergunta do usuário:\s*([\s\S]+)$/)
                            if (match) {
                                loadedMessages.push({
                                    role: 'user',
                                    content: match[1].trim()
                                })
                            }
                            continue
                        }
                    }

                    // Limpar conteúdo de mensagens do assistente
                    let content = m.content
                    if (m.role === 'assistant') {
                        // Remove blocos técnicos
                        content = content.replace(/:::excel-action\s*[\s\S]*?\s*:::/g, '')
                        content = content.replace(/:::excel-query\s*[\s\S]*?\s*:::/g, '')
                        content = content.replace(/\n{3,}/g, '\n\n').trim()

                        // Se ficou vazio após limpeza, pular
                        if (!content) continue
                    }

                    loadedMessages.push({
                        role: m.role as 'user' | 'assistant',
                        content: content
                    })
                }

                if (loadedMessages.length > 0) {
                    setMessages(loadedMessages)
                    toast.success('Conversa carregada!')
                } else {
                    toast.info('Conversa vazia')
                }
            } else {
                toast.info('Conversa vazia')
            }
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao carregar conversa: ' + errorMessage)
        }
    }

    const handleDeleteConversation = async (convId: string, e: React.MouseEvent) => {
        e.stopPropagation() // Evita disparar o onClick do item pai
        try {
            await DeleteConversation(convId)
            setConversations(prev => prev.filter(c => c.id !== convId))
            toast.success('Conversa excluída!')
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao excluir: ' + errorMessage)
        }
    }

    const chartOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: { labels: { color: '#e6edf3' } },
            title: { display: true, text: 'Visualização', color: '#e6edf3' }
        },
        scales: chartType !== 'pie' ? {
            x: { ticks: { color: '#8b949e' }, grid: { color: '#21262d' } },
            y: { ticks: { color: '#8b949e' }, grid: { color: '#21262d' } }
        } : undefined
    }

    if (showSettings) {
        return (
            <Settings
                onClose={() => setShowSettings(false)}
                askBeforeApply={askBeforeApply}
                onAskBeforeApplyChange={setAskBeforeApply}
            />
        )
    }

    return (
        <div className="flex flex-col h-screen bg-background text-foreground">
            {/* Header */}
            <header className="flex items-center justify-between px-6 py-4 border-b border-border bg-card/60 backdrop-blur supports-backdrop-filter:bg-card/40">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-primary text-primary-foreground flex items-center justify-center text-xl">
                        📊
                    </div>
                    <span className="text-xl font-semibold tracking-tight bg-linear-to-r from-primary to-blue-500 bg-clip-text text-transparent">
                        HipoSystem
                    </span>
                </div>

                <div className="flex items-center gap-3">
                    <Button onClick={handleNewConversation} variant="default">
                        ➕ Nova Conversa
                    </Button>
                    <Button onClick={() => setShowSettings(true)} variant="secondary">
                        ⚙️ Configurações
                    </Button>
                    <div className="flex items-center gap-2 px-3 py-1.5 bg-muted/50 border border-border rounded-full text-sm">
                        <span className={`w-2 h-2 rounded-full ${connected ? 'bg-emerald-500' : 'bg-rose-500'}`}></span>
                        <span>{connected ? 'Conectado' : 'Desconectado'}</span>
                        <button onClick={handleConnect} className="hover:text-foreground transition-colors">🔄</button>
                    </div>
                </div>
            </header>

            {/* Main */}
            <main className="flex flex-1 overflow-hidden">
                {/* Sidebar */}
                <aside className="w-72 bg-card border-r border-border flex flex-col overflow-hidden">
                    {/* Workbooks */}
                    <div className="p-4 border-b border-border">
                        <h3 className="text-xs font-semibold uppercase text-muted-foreground mb-3">📗 Planilhas</h3>
                        <div className="space-y-2 max-h-60 overflow-y-auto">
                            {workbooks.length > 0 ? workbooks.map(wb => (
                                <div key={wb.name} className="rounded-lg overflow-hidden border border-border bg-muted/30">
                                    <button
                                        onClick={() => setExpandedWorkbook(expandedWorkbook === wb.name ? null : wb.name)}
                                        className="w-full flex items-center gap-2 p-3 hover:bg-muted/60 transition-colors"
                                    >
                                        <span>📓</span>
                                        <span className="flex-1 text-left text-sm truncate">{wb.name}</span>
                                        <span className="text-xs text-muted-foreground bg-background/50 px-2 py-0.5 rounded-full border border-border">
                                            {wb.sheets?.length || 0}
                                        </span>
                                    </button>
                                    {expandedWorkbook === wb.name && (
                                        <div className="bg-background/40 border-t border-border max-h-40 overflow-y-auto">
                                            <div className="px-4 py-1 text-xs text-muted-foreground border-b border-border">
                                                💡 Clique para selecionar múltiplas abas
                                            </div>
                                            {wb.sheets?.map((sheet: string) => {
                                                const isSelected = selectedWorkbook === wb.name && selectedSheets.includes(sheet)
                                                return (
                                                    <button
                                                        key={sheet}
                                                        onClick={() => handleSelectSheet(wb.name, sheet)}
                                                        className={`w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted/60 transition-colors border-l-2 ${isSelected
                                                            ? 'border-l-primary bg-muted/60'
                                                            : 'border-transparent'
                                                            }`}
                                                    >
                                                        <span className="opacity-70">{isSelected ? '☑️' : '📄'}</span>
                                                        <span className="flex-1 text-left">{sheet}</span>
                                                        {isSelected && (
                                                            <span className="text-emerald-500">✓</span>
                                                        )}
                                                    </button>
                                                )
                                            })}
                                        </div>
                                    )}
                                </div>
                            )) : (
                                <p className="text-center text-muted-foreground text-sm py-4">
                                    {connected ? 'Nenhuma planilha' : 'Conecte ao Excel'}
                                </p>
                            )}
                        </div>
                        {contextLoaded && (
                            <div className="mt-3 p-2 bg-primary/10 border border-primary/30 rounded text-xs text-primary">
                                ✅ {contextLoaded}
                            </div>
                        )}
                    </div>

                    {/* History */}
                    <div className="p-4 flex-1 overflow-hidden">
                        <button
                            onClick={loadConversations}
                            className="w-full flex items-center justify-between text-xs font-semibold uppercase text-muted-foreground mb-3 hover:text-foreground"
                        >
                            <span>💬 Histórico</span>
                            <span className="text-xs">🔄 Atualizar</span>
                        </button>
                        <div className="space-y-2 overflow-y-auto max-h-48">
                            {conversations.length === 0 ? (
                                <p className="text-center text-muted-foreground text-sm py-4">
                                    Clique em "Atualizar" para carregar
                                </p>
                            ) : (
                                conversations.slice(0, 10).map(conv => (
                                    <div
                                        key={conv.id}
                                        onClick={() => handleLoadConversation(conv.id)}
                                        className="group p-2 bg-muted/30 border border-border rounded text-sm cursor-pointer hover:bg-muted/60 hover:border-primary/50 transition-all"
                                    >
                                        <div className="flex items-center justify-between gap-2">
                                            <div className="flex-1 min-w-0">
                                                <div className="truncate font-medium">{conv.title || 'Sem título'}</div>
                                                <div className="text-xs text-muted-foreground">{conv.updatedAt}</div>
                                            </div>
                                            <button
                                                onClick={(e) => handleDeleteConversation(conv.id, e)}
                                                className="opacity-0 group-hover:opacity-100 p-1 hover:bg-destructive/20 rounded text-destructive transition-opacity"
                                                title="Excluir conversa"
                                            >
                                                🗑️
                                            </button>
                                        </div>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>
                </aside>

                {/* Chat Area */}
                <section className="flex-1 flex flex-col bg-linear-to-b from-background to-muted/20">
                    {/* Toolbar */}
                    {previewData && (
                        <div className="flex items-center gap-2 p-3 bg-card/60 border-b border-border">
                            <Button
                                variant={showPreview ? "default" : "outline"}
                                size="sm"
                                onClick={() => { setShowPreview(!showPreview); setShowChart(false); }}
                            >
                                📋 Preview
                            </Button>
                            <Button
                                variant={showChart ? "default" : "outline"}
                                size="sm"
                                onClick={() => { setShowChart(!showChart); setShowPreview(false); }}
                            >
                                📊 Gráfico
                            </Button>
                            {showChart && (
                                <Select value={chartType} onValueChange={(v) => setChartType(v as 'bar' | 'line' | 'pie')}>
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
                    )}

                    {/* Preview */}
                    {showPreview && previewData && (
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
                    )}

                    {/* Chart */}
                    {showChart && chartData && (
                        <div className="flex-1 flex items-center justify-center p-6">
                            <div className="w-full max-h-96">
                                {chartType === 'bar' && <Bar data={chartData} options={chartOptions} />}
                                {chartType === 'line' && <Line data={chartData} options={chartOptions} />}
                                {chartType === 'pie' && <Pie data={chartData} options={chartOptions} />}
                            </div>
                        </div>
                    )}

                    {/* Pending Actions */}
                    {pendingActions.length > 0 && (
                        <div className="px-6 py-2 bg-yellow-500/10 border-b border-yellow-500/20 flex items-center justify-between animate-in slide-in-from-top-2">
                            <div className="flex items-center gap-2 text-sm text-yellow-500">
                                <span>⚠️</span>
                                <span>A IA sugeriu <strong>{pendingActions.length}</strong> alterações na planilha.</span>
                            </div>
                            <div className="flex items-center gap-2">
                                <Button size="sm" variant="ghost" onClick={handleDiscardActions} className="text-muted-foreground hover:text-destructive">
                                    Descartar
                                </Button>
                                <Button size="sm" onClick={handleApplyActions} className="bg-yellow-500 hover:bg-yellow-600 text-black">
                                    Aplicar Alterações
                                </Button>
                            </div>
                        </div>
                    )}

                    {/* Chat Messages */}
                    {!showPreview && !showChart && (
                        <div className="flex-1 overflow-y-auto p-6 space-y-4">
                            {messages.length === 0 ? (
                                <div className="flex flex-col items-center justify-center h-full">
                                    <Card className="w-full max-w-md bg-card/60">
                                        <CardHeader>
                                            <CardTitle className="flex items-center gap-2">
                                                <span className="text-2xl">{selectedSheets.length > 0 ? '📊' : '🤖'}</span>
                                                <span>{selectedSheets.length > 0 ? `${selectedSheets.length} aba(s) selecionada(s)` : 'HipoSystem pronto'}</span>
                                            </CardTitle>
                                        </CardHeader>
                                        <CardContent>
                                            {selectedSheets.length > 0 ? (
                                                <div className="space-y-3">
                                                    <p className="text-sm text-muted-foreground">
                                                        ✅ Abas carregadas: <strong className="text-primary">{selectedSheets.join(', ')}</strong>
                                                    </p>
                                                    <p className="text-sm text-muted-foreground">
                                                        Faça perguntas como:
                                                    </p>
                                                    <ul className="text-sm text-muted-foreground space-y-1 ml-4">
                                                        <li>• "Qual a soma da coluna B?"</li>
                                                        <li>• "Mostre os 5 maiores valores"</li>
                                                        <li>• "Crie uma fórmula para calcular média"</li>
                                                        <li>• "Analise os dados dessa planilha"</li>
                                                    </ul>
                                                </div>
                                            ) : (
                                                <p className="text-sm text-muted-foreground">
                                                    Selecione uma planilha na lateral e faça perguntas sobre seus dados.
                                                </p>
                                            )}
                                        </CardContent>
                                    </Card>
                                </div>
                            ) : (
                                messages.map((msg, idx) => (
                                    <div key={idx} className={`group flex gap-3 ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                                        {msg.role === 'assistant' && (
                                            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center border border-primary/20 shrink-0 mt-1">
                                                🤖
                                            </div>
                                        )}
                                        <div
                                            className={`max-w-[85%] p-4 rounded-2xl shadow-sm animate-in fade-in slide-in-from-bottom-2 ${editingMessageIndex === idx
                                                ? 'bg-card border border-border w-full max-w-full'
                                                : (msg.role === 'user'
                                                    ? 'bg-primary text-primary-foreground rounded-br-sm'
                                                    : 'bg-card border border-border rounded-bl-sm')
                                                }`}
                                        >
                                            {editingMessageIndex === idx ? (
                                                <div className="space-y-3">
                                                    <Textarea
                                                        value={editContent}
                                                        onChange={(e) => setEditContent(e.target.value)}
                                                        className="min-h-25 bg-transparent text-foreground"
                                                    />
                                                    <div className="flex justify-end gap-2">
                                                        <Button size="sm" variant="ghost" onClick={handleCancelEdit}>Cancelar</Button>
                                                        <Button size="sm" onClick={() => handleSaveEdit(idx)}>Salvar e Enviar</Button>
                                                    </div>
                                                </div>
                                            ) : (
                                                <>
                                                    {msg.role === 'assistant' ? (
                                                        <div className="text-sm relative">
                                                            {msg.content ? (
                                                                renderMessageContent(msg.content)
                                                            ) : (
                                                                <div className="flex items-center gap-1.5 h-6">
                                                                    <div className="w-2 h-2 bg-foreground/40 rounded-full animate-[bounce_1s_infinite_-0.3s]"></div>
                                                                    <div className="w-2 h-2 bg-foreground/40 rounded-full animate-[bounce_1s_infinite_-0.15s]"></div>
                                                                    <div className="w-2 h-2 bg-foreground/40 rounded-full animate-bounce"></div>
                                                                </div>
                                                            )}
                                                            {msg.hasActions && (
                                                                <div className="mt-3 pt-3 border-t border-border flex justify-end">
                                                                    <Button
                                                                        variant="outline"
                                                                        size="sm"
                                                                        onClick={handleUndo}
                                                                        className="text-xs h-7"
                                                                    >
                                                                        ↩️ Desfazer Alteração
                                                                    </Button>
                                                                </div>
                                                            )}
                                                        </div>
                                                    ) : (
                                                        <div className="relative">
                                                            <div className="whitespace-pre-wrap text-sm">{msg.content}</div>
                                                        </div>
                                                    )}

                                                    {/* Message Actions Footer */}
                                                    {!isLoading && (
                                                        <div className={`flex items-center justify-end gap-1 mt-2 pt-1 opacity-0 group-hover:opacity-100 transition-opacity ${msg.role === 'user'
                                                            ? 'border-t border-primary-foreground/20'
                                                            : 'border-t border-border'
                                                            }`}>
                                                            {msg.role === 'user' && (
                                                                <button
                                                                    onClick={() => handleEditMessage(idx, msg.content)}
                                                                    className="p-1.5 rounded hover:bg-primary-foreground/10 text-primary-foreground/80 hover:text-primary-foreground transition-colors"
                                                                    title="Editar"
                                                                >
                                                                    ✏️
                                                                </button>
                                                            )}

                                                            <button
                                                                onClick={() => handleCopy(msg.content)}
                                                                className={`p-1.5 rounded transition-colors ${msg.role === 'user'
                                                                    ? 'hover:bg-primary-foreground/10 text-primary-foreground/80 hover:text-primary-foreground'
                                                                    : 'hover:bg-muted text-muted-foreground hover:text-foreground'
                                                                    }`}
                                                                title="Copiar"
                                                            >
                                                                📋
                                                            </button>

                                                            <button
                                                                onClick={() => handleShare(msg.content)}
                                                                className={`p-1.5 rounded transition-colors ${msg.role === 'user'
                                                                    ? 'hover:bg-primary-foreground/10 text-primary-foreground/80 hover:text-primary-foreground'
                                                                    : 'hover:bg-muted text-muted-foreground hover:text-foreground'
                                                                    }`}
                                                                title="Compartilhar"
                                                            >
                                                                📤
                                                            </button>

                                                            {msg.role === 'assistant' && idx === messages.length - 1 && (
                                                                <button
                                                                    onClick={handleRegenerate}
                                                                    className="p-1.5 rounded hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                                                                    title="Regenerar resposta"
                                                                >
                                                                    🔄
                                                                </button>
                                                            )}
                                                        </div>
                                                    )}
                                                </>
                                            )}
                                        </div>
                                    </div>
                                ))
                            )}
                            <div ref={messagesEndRef} />
                        </div>
                    )}

                    {/* Input */}
                    <div className="p-4 bg-card/60 border-t border-border">
                        <div className="flex gap-3">
                            <Textarea
                                ref={inputRef}
                                value={inputMessage}
                                onChange={(e) => setInputMessage(e.target.value)}
                                onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && (e.preventDefault(), handleSendMessage())}
                                placeholder="Pergunte sobre seus dados..."
                                className="flex-1 min-h-13 max-h-36 resize-none"
                                disabled={isLoading}
                            />
                            {isLoading ? (
                                <Button
                                    onClick={handleCancelChat}
                                    variant="destructive"
                                    size="icon-lg"
                                    className="rounded-lg"
                                    title="Parar"
                                >
                                    ⏹️
                                </Button>
                            ) : (
                                <Button
                                    onClick={handleSendMessage}
                                    disabled={!inputMessage.trim()}
                                    size="icon-lg"
                                    className="rounded-lg"
                                >
                                    ➤
                                </Button>
                            )}
                        </div>
                    </div>
                </section>
            </main>
        </div>
    )
}
