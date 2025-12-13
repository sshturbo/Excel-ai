import { useState, useEffect, useRef } from 'react'
import { toast } from 'sonner'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
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
    ApplyFormula
} from "../wailsjs/go/main/App"

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
    const [selectedSheet, setSelectedSheet] = useState<string | null>(null)
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

    const messagesEndRef = useRef<HTMLDivElement>(null)
    const inputRef = useRef<HTMLTextAreaElement>(null)

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, [messages])

    useEffect(() => {
        const hasWailsRuntime = typeof (window as any)?.go !== 'undefined'
        if (hasWailsRuntime) {
            handleConnect()
            loadConfig()
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
            }
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro ao conectar: ' + errorMessage)
        }
    }

    const handleSelectSheet = async (wbName: string, sheetName: string) => {
        setSelectedWorkbook(wbName)
        setSelectedSheet(sheetName)
        setContextLoaded('')
        setPreviewData(null)

        try {
            await SetExcelContext(wbName, sheetName)
            const data = await GetPreviewData(wbName, sheetName)
            if (data) {
                setPreviewData(data)
                setContextLoaded(`${sheetName} (${data.totalRows} linhas)`)
                toast.success(`Contexto carregado: ${sheetName}`)
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

    const handleSendMessage = async () => {
        if (!inputMessage.trim() || isLoading) return

        const userMessage = inputMessage.trim()
        setInputMessage('')
        setMessages(prev => [...prev, { role: 'user', content: userMessage }])
        setIsLoading(true)

        try {
            const response = await SendMessage(userMessage)
            setMessages(prev => [...prev, { role: 'assistant', content: response }])
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('Erro: ' + errorMessage)
        } finally {
            setIsLoading(false)
            inputRef.current?.focus()
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
            toast.success('Nova conversa criada')
        } catch (err) {
            console.error(err)
        }
    }

    const loadConversations = async () => {
        try {
            const list = await ListConversations()
            if (list) setConversations(list)
        } catch (err) {
            console.error(err)
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
        return <Settings onClose={() => setShowSettings(false)} />
    }

    return (
        <div className="flex flex-col h-screen bg-background text-foreground">
            {/* Header */}
            <header className="flex items-center justify-between px-6 py-4 border-b border-border bg-card/60 backdrop-blur supports-backdrop-filter:bg-card/40">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-primary text-primary-foreground flex items-center justify-center text-xl">
                        📊
                    </div>
                    <span className="text-xl font-semibold tracking-tight">
                        Excel-AI
                    </span>
                </div>

                <div className="flex items-center gap-3">
                    <Button onClick={handleNewConversation} variant="default">
                        ➕ Nova Conversa
                    </Button>
                    <Button onClick={() => setShowSettings(true)} variant="secondary">
                        ⚙️ Config
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
                                            {wb.sheets?.map((sheet: string) => (
                                                <button
                                                    key={sheet}
                                                    onClick={() => handleSelectSheet(wb.name, sheet)}
                                                    className={`w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted/60 transition-colors border-l-2 ${selectedSheet === sheet && selectedWorkbook === wb.name
                                                        ? 'border-l-primary bg-muted/60'
                                                        : 'border-transparent'
                                                        }`}
                                                >
                                                    <span className="opacity-70">📄</span>
                                                    <span className="flex-1 text-left">{sheet}</span>
                                                    {selectedSheet === sheet && selectedWorkbook === wb.name && (
                                                        <span className="text-emerald-500">✓</span>
                                                    )}
                                                </button>
                                            ))}
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
                            <span>▶</span>
                        </button>
                        <div className="space-y-2 overflow-y-auto max-h-40">
                            {conversations.slice(0, 5).map(conv => (
                                <div key={conv.id} className="p-2 bg-muted/30 border border-border rounded text-sm cursor-pointer hover:bg-muted/60">
                                    <div className="truncate">{conv.title || 'Sem título'}</div>
                                    <div className="text-xs text-muted-foreground">{conv.updatedAt}</div>
                                </div>
                            ))}
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

                    {/* Chat Messages */}
                    {!showPreview && !showChart && (
                        <div className="flex-1 overflow-y-auto p-6 space-y-4">
                            {messages.length === 0 ? (
                                <div className="flex flex-col items-center justify-center h-full">
                                    <Card className="w-full max-w-md bg-card/60">
                                        <CardHeader>
                                            <CardTitle className="flex items-center gap-2">
                                                <span className="text-2xl">🤖</span>
                                                <span>Excel-AI pronto</span>
                                            </CardTitle>
                                        </CardHeader>
                                        <CardContent>
                                            <p className="text-sm text-muted-foreground">
                                                Selecione uma planilha na lateral e faça perguntas sobre seus dados.
                                            </p>
                                        </CardContent>
                                    </Card>
                                </div>
                            ) : (
                                messages.map((msg, idx) => (
                                    <div
                                        key={idx}
                                        className={`max-w-[85%] p-4 rounded-2xl shadow-sm animate-in fade-in slide-in-from-bottom-2 ${msg.role === 'user'
                                            ? 'ml-auto bg-primary text-primary-foreground rounded-br-sm'
                                            : 'bg-card border border-border rounded-bl-sm'
                                            }`}
                                    >
                                        {msg.role === 'assistant' ? (
                                            <div className="prose prose-invert prose-sm max-w-none">
                                                <ReactMarkdown remarkPlugins={[remarkGfm]}>{msg.content}</ReactMarkdown>
                                            </div>
                                        ) : (
                                            <div className="whitespace-pre-wrap">{msg.content}</div>
                                        )}
                                    </div>
                                ))
                            )}
                            {isLoading && (
                                <div className="flex items-center gap-2 text-muted-foreground">
                                    <div className="w-5 h-5 border-2 border-muted border-t-primary rounded-full animate-spin"></div>
                                    <span>Pensando...</span>
                                </div>
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
                            />
                            <Button
                                onClick={handleSendMessage}
                                disabled={isLoading || !inputMessage.trim()}
                                size="icon-lg"
                                className="rounded-lg"
                            >
                                ➤
                            </Button>
                        </div>
                    </div>
                </section>
            </main>
        </div>
    )
}
