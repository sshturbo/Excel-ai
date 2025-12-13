import { useState, useEffect } from 'react'
import { toast } from 'sonner'
import {
    SetAPIKey,
    SetModel,
    GetSavedConfig,
    UpdateConfig
} from "../wailsjs/go/main/App"

// shadcn components
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Slider } from "@/components/ui/slider"
import { Switch } from "@/components/ui/switch"

interface SettingsProps {
    onClose: () => void
}

const models = [
    { value: 'openai/gpt-4o-mini', label: 'GPT-4o Mini', desc: 'Rápido e econômico' },
    { value: 'openai/gpt-4o', label: 'GPT-4o', desc: 'Avançado' },
    { value: 'anthropic/claude-3.5-sonnet', label: 'Claude 3.5 Sonnet', desc: 'Análise excelente' },
    { value: 'anthropic/claude-3-haiku', label: 'Claude 3 Haiku', desc: 'Ultra rápido' },
    { value: 'google/gemini-pro-1.5', label: 'Gemini Pro 1.5', desc: 'Contexto longo' },
    { value: 'deepseek/deepseek-chat', label: 'DeepSeek Chat', desc: 'Ótimo custo' },
]

export default function Settings({ onClose }: SettingsProps) {
    const [apiKey, setApiKey] = useState('')
    const [model, setModel] = useState('openai/gpt-4o-mini')
    const [maxRowsContext, setMaxRowsContext] = useState(50)
    const [maxRowsPreview, setMaxRowsPreview] = useState(100)
    const [detailLevel, setDetailLevel] = useState('normal')
    const [customPrompt, setCustomPrompt] = useState('')
    const [language, setLanguage] = useState('pt-BR')
    const [includeHeaders, setIncludeHeaders] = useState(true)
    const [isSaving, setIsSaving] = useState(false)

    useEffect(() => {
        const hasWailsRuntime = typeof (window as any)?.go !== 'undefined'
        if (hasWailsRuntime) {
            loadConfig()
        } else {
            console.warn('Wails runtime não detectado. Settings em modo somente UI (Vite puro).')
        }
    }, [])

    const loadConfig = async () => {
        try {
            const cfg = await GetSavedConfig()
            if (cfg) {
                if (cfg.apiKey) setApiKey(cfg.apiKey)
                if (cfg.model) setModel(cfg.model)
                if (cfg.maxRowsContext) setMaxRowsContext(cfg.maxRowsContext)
                if (cfg.maxRowsPreview) setMaxRowsPreview(cfg.maxRowsPreview)
                if (cfg.detailLevel) setDetailLevel(cfg.detailLevel)
                if (cfg.customPrompt) setCustomPrompt(cfg.customPrompt)
                if (cfg.language) setLanguage(cfg.language)
                setIncludeHeaders(cfg.includeHeaders !== false)
            }
        } catch (err) {
            toast.error('Erro ao carregar configurações')
        }
    }

    const handleSave = async () => {
        if (typeof (window as any)?.go === 'undefined') {
            toast.error('Wails não detectado. Não é possível salvar fora do app.')
            return
        }
        setIsSaving(true)
        try {
            await SetAPIKey(apiKey)
            await SetModel(model)
            await UpdateConfig(maxRowsContext, maxRowsPreview, includeHeaders, detailLevel, customPrompt, language)
            toast.success('✅ Configurações salvas!')
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('❌ Erro ao salvar: ' + errorMessage)
        } finally {
            setIsSaving(false)
        }
    }

    return (
        <div className="min-h-screen bg-background text-foreground">
            {/* Gradient Background */}
            <div className="fixed inset-0 pointer-events-none">
                <div className="absolute top-0 left-1/4 w-96 h-96 bg-primary/10 rounded-full blur-3xl" />
                <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-muted/40 rounded-full blur-3xl" />
            </div>

            <div className="relative max-w-4xl mx-auto py-8 px-4">
                {/* Header */}
                <header className="flex items-center justify-between mb-8 pb-6 border-b border-border">
                    <Button variant="ghost" onClick={onClose} className="gap-2">
                        ← Voltar
                    </Button>
                    <div className="flex items-center gap-3">
                        <span className="text-3xl">⚙️</span>
                        <h1 className="text-2xl font-semibold tracking-tight">
                            Configurações
                        </h1>
                    </div>
                    <Button onClick={handleSave} disabled={isSaving} className="gap-2">
                        {isSaving ? (
                            <>
                                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                Salvando
                            </>
                        ) : (
                            <>💾 Salvar</>
                        )}
                    </Button>
                </header>

                {/* Tabs */}
                <Tabs defaultValue="api" className="space-y-6">
                    <TabsList className="grid w-full grid-cols-4 bg-muted/40">
                        <TabsTrigger value="api" className="gap-2">🔑 API</TabsTrigger>
                        <TabsTrigger value="data" className="gap-2">📊 Dados</TabsTrigger>
                        <TabsTrigger value="custom" className="gap-2">✨ Pessoal</TabsTrigger>
                        <TabsTrigger value="about" className="gap-2">ℹ️ Sobre</TabsTrigger>
                    </TabsList>

                    {/* API Tab */}
                    <TabsContent value="api" className="space-y-6">
                        <Card className="bg-card/60">
                            <CardHeader>
                                <div className="w-12 h-12 bg-primary text-primary-foreground rounded-xl flex items-center justify-center text-2xl mb-2">
                                    🔑
                                </div>
                                <CardTitle>Chave de API</CardTitle>
                                <CardDescription>Configure sua chave do OpenRouter</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="space-y-2">
                                    <Label>API Key</Label>
                                    <Input
                                        type="password"
                                        value={apiKey}
                                        onChange={(e) => setApiKey(e.target.value)}
                                        placeholder="sk-or-v1-..."
                                    />
                                    <p className="text-sm text-muted-foreground">
                                        Obtenha em <a href="https://openrouter.ai/keys" target="_blank" className="text-primary hover:underline">openrouter.ai/keys</a>
                                    </p>
                                </div>
                            </CardContent>
                        </Card>

                        <Card className="bg-card/60">
                            <CardHeader>
                                <div className="w-12 h-12 bg-muted rounded-xl flex items-center justify-center text-2xl mb-2">
                                    🤖
                                </div>
                                <CardTitle>Modelo de IA</CardTitle>
                                <CardDescription>Escolha o modelo que melhor atende suas necessidades</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="grid grid-cols-2 gap-3">
                                    {models.map((m) => (
                                        <button
                                            key={m.value}
                                            onClick={() => setModel(m.value)}
                                            className={`p-3 rounded-lg text-left transition-all ${model === m.value
                                                ? 'bg-primary/10 border-2 border-primary/50'
                                                : 'bg-background/40 border border-border hover:bg-muted/40'
                                                }`}
                                        >
                                            <div className="font-medium text-sm">{m.label}</div>
                                            <div className="text-xs text-muted-foreground">{m.desc}</div>
                                        </button>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    </TabsContent>

                    {/* Data Tab */}
                    <TabsContent value="data" className="space-y-6">
                        <Card className="bg-card/60">
                            <CardHeader>
                                <div className="w-12 h-12 bg-muted rounded-xl flex items-center justify-center text-2xl mb-2">
                                    📈
                                </div>
                                <CardTitle>Quantidade de Dados</CardTitle>
                                <CardDescription>Configure quantas linhas serão enviadas para análise</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-6">
                                <div className="space-y-4">
                                    <div className="flex justify-between">
                                        <Label>Linhas para IA</Label>
                                        <span className="text-sm text-primary font-medium">{maxRowsContext}</span>
                                    </div>
                                    <Slider
                                        value={[maxRowsContext]}
                                        onValueChange={(v) => setMaxRowsContext(v[0])}
                                        min={10}
                                        max={200}
                                        step={10}
                                    />
                                    <div className="flex justify-between text-xs text-muted-foreground">
                                        <span>10 (Rápido)</span>
                                        <span>200 (Detalhado)</span>
                                    </div>
                                </div>

                                <div className="space-y-4">
                                    <div className="flex justify-between">
                                        <Label>Linhas no Preview</Label>
                                        <span className="text-sm text-blue-400 font-medium">{maxRowsPreview}</span>
                                    </div>
                                    <Slider
                                        value={[maxRowsPreview]}
                                        onValueChange={(v) => setMaxRowsPreview(v[0])}
                                        min={50}
                                        max={500}
                                        step={50}
                                    />
                                </div>

                                <div className="flex items-center justify-between p-4 bg-muted/30 border border-border rounded-lg">
                                    <Label>Incluir cabeçalhos no contexto</Label>
                                    <Switch checked={includeHeaders} onCheckedChange={setIncludeHeaders} />
                                </div>
                            </CardContent>
                        </Card>

                        <Card className="bg-card/60">
                            <CardHeader>
                                <div className="w-12 h-12 bg-muted rounded-xl flex items-center justify-center text-2xl mb-2">
                                    🎚️
                                </div>
                                <CardTitle>Nível de Detalhe</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="grid grid-cols-3 gap-3">
                                    {[
                                        { value: 'minimal', icon: '📝', title: 'Mínimo', desc: 'Só cabeçalhos' },
                                        { value: 'normal', icon: '📋', title: 'Normal', desc: 'Recomendado' },
                                        { value: 'detailed', icon: '📚', title: 'Detalhado', desc: 'Tudo' },
                                    ].map((opt) => (
                                        <button
                                            key={opt.value}
                                            onClick={() => setDetailLevel(opt.value)}
                                            className={`flex flex-col items-center gap-2 p-4 rounded-lg transition-all ${detailLevel === opt.value
                                                ? 'bg-primary/10 border-2 border-primary/50'
                                                : 'bg-background/40 border border-border hover:bg-muted/40'
                                                }`}
                                        >
                                            <span className="text-2xl">{opt.icon}</span>
                                            <span className="font-medium text-sm">{opt.title}</span>
                                            <span className="text-xs text-muted-foreground">{opt.desc}</span>
                                        </button>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    </TabsContent>

                    {/* Custom Tab */}
                    <TabsContent value="custom" className="space-y-6">
                        <Card className="bg-card/60">
                            <CardHeader>
                                <div className="w-12 h-12 bg-muted rounded-xl flex items-center justify-center text-2xl mb-2">
                                    💬
                                </div>
                                <CardTitle>Instrução Personalizada</CardTitle>
                                <CardDescription>Adicione instruções extras para todas as conversas</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <Textarea
                                    value={customPrompt}
                                    onChange={(e) => setCustomPrompt(e.target.value)}
                                    placeholder="Ex: Sempre responda em português. Formate números com 2 casas decimais..."
                                    className="min-h-32"
                                />
                                <p className="text-xs text-muted-foreground mt-2 text-right">
                                    {customPrompt.length} / 500 caracteres
                                </p>
                            </CardContent>
                        </Card>

                        <Card className="bg-card/60">
                            <CardHeader>
                                <div className="w-12 h-12 bg-muted rounded-xl flex items-center justify-center text-2xl mb-2">
                                    🌍
                                </div>
                                <CardTitle>Idioma</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="flex gap-3">
                                    {[
                                        { value: 'pt-BR', flag: '🇧🇷', name: 'Português' },
                                        { value: 'en', flag: '🇺🇸', name: 'English' },
                                        { value: 'es', flag: '🇪🇸', name: 'Español' },
                                    ].map((lang) => (
                                        <button
                                            key={lang.value}
                                            onClick={() => setLanguage(lang.value)}
                                            className={`flex items-center gap-3 flex-1 p-4 rounded-lg transition-all ${language === lang.value
                                                ? 'bg-primary/10 border-2 border-primary/50'
                                                : 'bg-background/40 border border-border hover:bg-muted/40'
                                                }`}
                                        >
                                            <span className="text-2xl">{lang.flag}</span>
                                            <span className="font-medium">{lang.name}</span>
                                        </button>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    </TabsContent>

                    {/* About Tab */}
                    <TabsContent value="about">
                        <Card className="bg-card/60 text-center py-8">
                            <CardContent className="space-y-4">
                                <div className="text-6xl animate-bounce">📊</div>
                                <h2 className="text-3xl font-bold">
                                    Excel-AI
                                </h2>
                                <p className="text-muted-foreground">Seu assistente inteligente para planilhas</p>
                                <span className="inline-block px-3 py-1 bg-muted rounded-full text-sm text-primary">
                                    v2.0.0
                                </span>
                                <div className="flex justify-center gap-2 pt-4">
                                    {['Go', 'React', 'Wails', 'Tailwind', 'shadcn/ui'].map((tech) => (
                                        <span key={tech} className="px-2 py-1 bg-muted/60 border border-border rounded text-xs text-muted-foreground">
                                            {tech}
                                        </span>
                                    ))}
                                </div>
                                <p className="text-xs text-muted-foreground pt-4">
                                    Dados em <code className="bg-muted px-1 rounded">~/.excel-ai/</code>
                                </p>
                            </CardContent>
                        </Card>
                    </TabsContent>
                </Tabs>
            </div>
        </div>
    )
}
