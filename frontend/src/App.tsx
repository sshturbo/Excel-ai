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
    CreatePivotTable
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
            } else if (result.error) {
                toast.error('Erro ao conectar: ' + result.error)
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

    const [streamingContent, setStreamingContent] = useState('')

    useEffect(() => {
        const cleanup = EventsOn("chat:chunk", (chunk: string) => {
            setStreamingContent(prev => prev + chunk)
        })
        return () => cleanup()
    }, [])

    useEffect(() => {
        if (streamingContent && isLoading) {
            setMessages(prev => {
                const newMsgs = [...prev]
                const lastIndex = newMsgs.length - 1
                if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                    // Clean streaming content
                    let cleanContent = streamingContent.replace(/:::excel-action\s*([\s\S]*?)\s*:::/g, '')
                    // Also hide incomplete action block at the end to avoid flickering
                    cleanContent = cleanContent.replace(/:::excel-action[\s\S]*$/, '')
                    cleanContent = cleanContent.trim()

                    // Only update if content is different to avoid loops
                    if (newMsgs[lastIndex].content !== cleanContent) {
                        newMsgs[lastIndex] = { ...newMsgs[lastIndex], content: cleanContent }
                    }
                }
                return newMsgs
            })
        }
    }, [streamingContent, isLoading])

    const processMessage = async (text: string) => {
        setIsLoading(true)
        setStreamingContent('')

        // Add placeholder for assistant message
        setMessages(prev => [...prev, { role: 'assistant', content: '' }])

        try {
            const response = await SendMessage(text)
            
            // Check for Excel Actions (Global regex)
            const actionRegex = /:::excel-action\s*([\s\S]*?)\s*:::/g
            const matches = [...response.matchAll(actionRegex)]
            
            let actionsExecuted = 0
            
            if (matches.length > 0 && !askBeforeApply) {
                await StartUndoBatch()
            }

            for (const match of matches) {
                try {
                    const jsonStr = match[1]
                    // Sanitize potential markdown code blocks
                    const cleanJson = jsonStr.replace(/```json/g, '').replace(/```/g, '').trim()
                    const action = JSON.parse(cleanJson)
                    
                    if (action.op === 'write') {
                        if (askBeforeApply) {
                            setPendingActions(prev => [...prev, action])
                        } else {
                            await UpdateExcelCell(
                                action.workbook || '', 
                                action.sheet || '', 
                                action.cell, 
                                action.value
                            )
                            actionsExecuted++
                        }
                    } else if (action.op === 'create-workbook') {
                        if (askBeforeApply) {
                            setPendingActions(prev => [...prev, action])
                        } else {
                            const name = await CreateNewWorkbook()
                            toast.success(`Nova pasta de trabalho criada: ${name}`)
                            actionsExecuted++
                            // Refresh workbooks list
                            const result = await RefreshWorkbooks()
                            if (result.workbooks) setWorkbooks(result.workbooks)
                        }
                    } else if (action.op === 'create-sheet') {
                        if (askBeforeApply) {
                            setPendingActions(prev => [...prev, action])
                        } else {
                            await CreateNewSheet(action.name)
                            toast.success(`Nova aba criada: ${action.name}`)
                            actionsExecuted++
                            // Refresh workbooks list
                            const result = await RefreshWorkbooks()
                            if (result.workbooks) setWorkbooks(result.workbooks)
                        }
                    } else if (action.op === 'create-chart') {
                        if (askBeforeApply) {
                            setPendingActions(prev => [...prev, action])
                        } else {
                            await CreateChart(
                                action.sheet || '',
                                action.range,
                                action.chartType || 'column',
                                action.title || ''
                            )
                            toast.success('Gráfico criado!')
                            actionsExecuted++
                        }
                    } else if (action.op === 'create-pivot') {
                        if (askBeforeApply) {
                            setPendingActions(prev => [...prev, action])
                        } else {
                            await CreatePivotTable(
                                action.sourceSheet || '',
                                action.sourceRange,
                                action.destSheet || '',
                                action.destCell,
                                action.tableName || 'PivotTable1'
                            )
                            toast.success('Tabela dinâmica criada!')
                            actionsExecuted++
                        }
                    }
                } catch (e) {
                    console.error("Failed to execute Excel action", e)
                }
            }

            if (matches.length > 0 && !askBeforeApply) {
                await EndUndoBatch()
            }

            if (actionsExecuted > 0) {
                toast.success(`${actionsExecuted} alterações aplicadas!`)
                // Refresh context if we have a selected sheet
                if (selectedWorkbook && selectedSheet) {
                     const contextMsg = await SetExcelContext(selectedWorkbook, selectedSheet)
                     setContextLoaded(contextMsg)
                     const preview = await GetPreviewData(selectedWorkbook, selectedSheet)
                     setPreviewData(preview)
                }
            }

            // Remove all action blocks from the displayed message
            const displayContent = response.replace(actionRegex, '').trim()

            setMessages(prev => {
                const newMsgs = [...prev]
                const lastIndex = newMsgs.length - 1
                if (lastIndex >= 0 && newMsgs[lastIndex].role === 'assistant') {
                    newMsgs[lastIndex] = { 
                        ...newMsgs[lastIndex], 
                        content: displayContent,
                        hasActions: actionsExecuted > 0
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
                } else if (action.op === 'create-chart') {
                    await CreateChart(
                        action.sheet || '',
                        action.range,
                        action.chartType || 'column',
                        action.title || ''
                    )
                } else if (action.op === 'create-pivot') {
                    await CreatePivotTable(
                        action.sourceSheet || '',
                        action.sourceRange,
                        action.destSheet || '',
                        action.destCell,
                        action.tableName || 'PivotTable1'
                    )
                }
                executed++
            } catch (e) {
                console.error("Erro ao aplicar ação:", action, e)
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

            if (selectedWorkbook && selectedSheet) {
                const contextMsg = await SetExcelContext(selectedWorkbook, selectedSheet)
                setContextLoaded(contextMsg)
                const preview = await GetPreviewData(selectedWorkbook, selectedSheet)
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
            if (selectedWorkbook && selectedSheet) {
                const preview = await GetPreviewData(selectedWorkbook, selectedSheet)
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
            setSelectedSheet(null)
            
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
            if (list) setConversations(list)
        } catch (err) {
            console.error(err)
        }
    }

    const handleLoadConversation = async (convId: string) => {
        try {
            const messages = await LoadConversation(convId)
            if (messages && messages.length > 0) {
                const loadedMessages: Message[] = messages.map((m) => {
                    let content = m.content
                    // Clean Excel action blocks from assistant messages
                    if (m.role === 'assistant') {
                        content = content.replace(/:::excel-action\s*([\s\S]*?)\s*:::/g, '').trim()
                    }
                    return {
                        role: m.role as 'user' | 'assistant',
                        content: content
                    }
                })
                setMessages(loadedMessages)
                toast.success('Conversa carregada!')
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
                    <span className="text-xl font-semibold tracking-tight">
                        Excel-AI
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
                                                <span className="text-2xl">{selectedSheet ? '📊' : '🤖'}</span>
                                                <span>{selectedSheet ? `Planilha: ${selectedSheet}` : 'Excel-AI pronto'}</span>
                                            </CardTitle>
                                        </CardHeader>
                                        <CardContent>
                                            {selectedSheet ? (
                                                <div className="space-y-3">
                                                    <p className="text-sm text-muted-foreground">
                                                        ✅ Planilha <strong className="text-primary">{selectedSheet}</strong> carregada!
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
                                            className={`max-w-[85%] p-4 rounded-2xl shadow-sm animate-in fade-in slide-in-from-bottom-2 ${
                                                editingMessageIndex === idx
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
                                                                <ReactMarkdown 
                                                                    remarkPlugins={[remarkGfm]}
                                                                    components={markdownComponents}
                                                                >
                                                                    {msg.content}
                                                                </ReactMarkdown>
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
                                                        <div className={`flex items-center justify-end gap-1 mt-2 pt-1 opacity-0 group-hover:opacity-100 transition-opacity ${
                                                            msg.role === 'user' 
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
                                                                className={`p-1.5 rounded transition-colors ${
                                                                    msg.role === 'user'
                                                                        ? 'hover:bg-primary-foreground/10 text-primary-foreground/80 hover:text-primary-foreground'
                                                                        : 'hover:bg-muted text-muted-foreground hover:text-foreground'
                                                                }`}
                                                                title="Copiar"
                                                            >
                                                                📋
                                                            </button>

                                                            <button
                                                                onClick={() => handleShare(msg.content)}
                                                                className={`p-1.5 rounded transition-colors ${
                                                                    msg.role === 'user'
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
