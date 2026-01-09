// API Tab content for Settings - Professional Version
import { useState } from "react"
import { dto } from "../../../wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Command, CommandEmpty, CommandGroup, CommandItem, CommandList } from "@/components/ui/command"
import { Check, ChevronsUpDown } from "lucide-react"
import { cn } from "@/lib/utils"
import { popularModels } from "@/hooks/useSettings"

// Provider information
interface ProviderInfo {
    id: string
    name: string
    icon: string
    description: string
    baseUrl?: string
    baseUrlPlaceholder: string
    apiKeyUrl?: string
    apiKeyHelp: string
    hasBaseUrl: boolean
    hasApiKey: boolean
    defaultUrl: string
}

// Popular models by provider
const providerModels: Record<string, Array<{ value: string; label: string; desc: string }>> = {
    openrouter: [
        { value: "openai/gpt-4o-mini", label: "GPT-4o Mini", desc: "Rápido e econômico" },
        { value: "openai/gpt-4o", label: "GPT-4o", desc: "Mais poderoso" },
        { value: "anthropic/claude-3.5-sonnet", label: "Claude 3.5 Sonnet", desc: "Melhor raciocínio" },
        { value: "google/gemini-flash-1.5", label: "Gemini Flash 1.5", desc: "Ultra rápido" },
    ],
    zai: [
        { value: "glm-4.7", label: "GLM-4.7", desc: "Flagship - melhor para coding" },
        { value: "glm-4.6v", label: "GLM-4.6V", desc: "Multimodal com visão" },
        { value: "glm-4.6", label: "GLM-4.6", desc: "Versátil e equilibrado" },
        { value: "glm-4.5", label: "GLM-4.5", desc: "Econômico" },
        { value: "glm-4.5-air", label: "GLM-4.5 Air", desc: "Ultraleve" },
    ],
    google: [
        { value: "gemini-2.0-flash-exp", label: "Gemini 2.0 Flash", desc: "Ultra rápido" },
        { value: "gemini-1.5-flash", label: "Gemini 1.5 Flash", desc: "Muito rápido" },
        { value: "gemini-1.5-pro", label: "Gemini 1.5 Pro", desc: "Mais poderoso" },
    ],
    groq: [
        { value: "llama-3.3-70b-versatile", label: "Llama 3.3 70B", desc: "Muito rápido" },
        { value: "llama-3.1-8b-instant", label: "Llama 3.1 8B", desc: "Instantâneo" },
        { value: "mixtral-8x7b-32768", label: "Mixtral 8x7B", desc: "Equilibrado" },
    ],
    ollama: [
        { value: "qwen2.5:7b", label: "Qwen 2.5 7B", desc: "Melhor suporte a tools" },
        { value: "llama3.1:8b", label: "Llama 3.1 8B", desc: "Bom suporte geral" },
        { value: "mistral:7b", label: "Mistral 7B", desc: "Bom suporte" },
    ],
};

// Provider information
const providersInfo: ProviderInfo[] = [
    {
        id: "openrouter",
        name: "OpenRouter",
        icon: "🌐",
        description: "Acesse múltiplos modelos de IA de um só lugar",
        baseUrlPlaceholder: "https://openrouter.ai/api/v1",
        apiKeyUrl: "https://openrouter.ai/keys",
        apiKeyHelp: "Obtenha sua API Key no painel do OpenRouter",
        hasBaseUrl: true,
        hasApiKey: true,
        defaultUrl: "https://openrouter.ai/api/v1",
    },
    {
        id: "zai",
        name: "Z.AI (GLM Models)",
        icon: "🤖",
        description: "Modelos GLM da Zhipu AI - 128K tokens de contexto",
        baseUrlPlaceholder: "https://api.z.ai/api/paas/v4",
        apiKeyUrl: "https://open.bigmodel.cn/developercenter/balance",
        apiKeyHelp: "Obtenha sua API Key no console do Z.AI (BigModel)",
        hasBaseUrl: false,
        hasApiKey: true,
        defaultUrl: "https://api.z.ai/api/paas/v4",
    },
    {
        id: "google",
        name: "Google AI (Gemini)",
        icon: "🔷",
        description: "Modelos Gemini do Google - raciocínio avançado",
        baseUrlPlaceholder: "https://generativelanguage.googleapis.com/v1beta",
        apiKeyUrl: "https://aistudio.google.com/apikey",
        apiKeyHelp: "Obtenha sua API Key no AI Studio do Google",
        hasBaseUrl: false,
        hasApiKey: true,
        defaultUrl: "",
    },
    {
        id: "groq",
        name: "Groq",
        icon: "⚡",
        description: "Inferência ultra-rápida com Llama 3",
        baseUrlPlaceholder: "https://api.groq.com/openai/v1",
        apiKeyUrl: "https://console.groq.com/keys",
        apiKeyHelp: "Obtenha sua API Key no console do Groq",
        hasBaseUrl: false,
        hasApiKey: true,
        defaultUrl: "https://api.groq.com/openai/v1",
    },
    {
        id: "ollama",
        name: "Ollama",
        icon: "🦙",
        description: "Execute modelos localmente - privacidade total",
        baseUrlPlaceholder: "http://localhost:11434",
        apiKeyHelp: "Ollama local não requer API Key. Deixe qualquer valor.",
        hasBaseUrl: true,
        hasApiKey: false,
        defaultUrl: "http://localhost:11434",
    },
    {
        id: "custom",
        name: "Personalizado",
        icon: "⚙️",
        description: "Qualquer API compatível com OpenAI",
        baseUrlPlaceholder: "https://api.exemplo.com/v1",
        apiKeyHelp: "Configure a URL base e API Key do seu provedor",
        hasBaseUrl: true,
        hasApiKey: true,
        defaultUrl: "",
    },
];

interface ApiTabProps {
    provider: string
    apiKey: string
    baseUrl: string
    model: string
    customModel: string
    toolModel: string
    useCustomModel: boolean
    availableModels: dto.ModelInfo[]
    filteredModels: dto.ModelInfo[]
    modelFilter: string
    isLoadingModels: boolean
    onProviderChange: (value: string) => void
    onApiKeyChange: (value: string) => void
    onBaseUrlChange: (value: string) => void
    onModelChange: (value: string) => void
    onToolModelChange: (value: string) => void // New Handler
    onCustomModelChange: (value: string) => void
    onUseCustomModelChange: (value: boolean) => void
    onModelFilterChange: (value: string) => void
    onLoadModels: () => void
}

export function ApiTab({
    provider,
    apiKey,
    baseUrl,
    model,
    toolModel,
    customModel,
    useCustomModel,
    availableModels,
    filteredModels,
    modelFilter,
    isLoadingModels,
    onProviderChange,
    onApiKeyChange,
    onBaseUrlChange,
    onModelChange,
    onToolModelChange,
    onCustomModelChange,
    onUseCustomModelChange,
    onModelFilterChange,
    onLoadModels,
}: ApiTabProps) {
    const [providerOpen, setProviderOpen] = useState(false)
    const currentProvider = providersInfo.find(p => p.id === provider);

    // Get popular models for current provider
    const currentPopularModels = providerModels[provider] || [];

    return (
        <div className="space-y-6">
            {/* Provider Selection Card */}
            <Card className="bg-linear-to-br from-card/60 to-card/40 border-2 border-primary/20">
                <CardHeader className="space-y-2">
                    <div className="flex items-center gap-4">
                        <div className="w-16 h-16 bg-linear-to-br from-primary/20 to-primary/40 rounded-2xl flex items-center justify-center text-3xl border-2 border-primary/30 shadow-lg">
                            {currentProvider?.icon || "🔑"}
                        </div>
                        <div className="flex-1 space-y-1">
                            <CardTitle className="text-2xl font-bold">Configuração de Provedor</CardTitle>
                            <CardDescription className="text-base">
                                {currentProvider?.description}
                            </CardDescription>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="space-y-6">
                    {/* Provider Selector - Using Combobox */}
                    <div className="space-y-3">
                        <Label className="text-base font-semibold flex items-center gap-2">
                            <span className="text-primary">🌐</span> Selecione o Provedor
                        </Label>
                        <Popover open={providerOpen} onOpenChange={setProviderOpen}>
                            <PopoverTrigger asChild>
                                <Button
                                    variant="outline"
                                    role="combobox"
                                    aria-expanded={providerOpen}
                                    className="w-full h-14 justify-between text-base bg-background/50 border-2 border-primary/20 hover:border-primary/50 hover:bg-background/70 transition-all rounded-xl px-4"
                                >
                                    {currentProvider ? (
                                        <div className="flex items-center gap-3">
                                            <div className="flex items-center justify-center w-8 h-8 bg-muted/50 rounded-lg text-lg shrink-0">
                                                {currentProvider.icon}
                                            </div>
                                            <span className="font-semibold text-foreground/90">{currentProvider.name}</span>
                                            <span className="text-xs bg-primary/20 text-primary px-2 py-0.5 rounded-full">Ativo</span>
                                        </div>
                                    ) : (
                                        <span className="text-muted-foreground">Selecione um provedor...</span>
                                    )}
                                    <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                                </Button>
                            </PopoverTrigger>
                            <PopoverContent className="w-[--radix-popover-trigger-width] p-0" align="start">
                                <Command>
                                    <CommandList>
                                        <CommandEmpty>Nenhum provedor encontrado.</CommandEmpty>
                                        <CommandGroup>
                                            {providersInfo.map((p) => (
                                                <CommandItem
                                                    key={p.id}
                                                    value={p.id}
                                                    onSelect={(value) => {
                                                        onProviderChange(value)
                                                        setProviderOpen(false)
                                                    }}
                                                    className="cursor-pointer py-3"
                                                >
                                                    <div className="flex items-center gap-3 flex-1">
                                                        <div className="p-1.5 bg-muted rounded-lg text-base shrink-0">
                                                            {p.icon}
                                                        </div>
                                                        <div className="flex-1 min-w-0">
                                                            <div className="font-bold">{p.name}</div>
                                                            <div className="text-xs text-muted-foreground font-medium truncate">{p.description}</div>
                                                        </div>
                                                    </div>
                                                    <Check
                                                        className={cn(
                                                            "h-4 w-4 shrink-0",
                                                            provider === p.id ? "opacity-100" : "opacity-0"
                                                        )}
                                                    />
                                                </CommandItem>
                                            ))}
                                        </CommandGroup>
                                    </CommandList>
                                </Command>
                            </PopoverContent>
                        </Popover>
                    </div>

                    {/* Base URL - Only for custom, ollama */}
                    {currentProvider?.hasBaseUrl && (
                        <div className="space-y-3">
                            <Label className="text-base font-semibold flex items-center gap-2">
                                <span className="text-primary">🔗</span> Base URL
                                {currentProvider?.id === 'ollama' && <span className="text-xs font-normal text-muted-foreground">(Opcional para Localhost)</span>}
                            </Label>
                            <div className="relative group">
                                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors text-base">
                                    🌐
                                </div>
                                <Input
                                    type="text"
                                    value={baseUrl}
                                    onChange={(e) => onBaseUrlChange(e.target.value)}
                                    placeholder={currentProvider?.baseUrlPlaceholder}
                                    className="h-14 pl-12 text-base font-mono bg-background/50 border-2 border-primary/20 focus:border-primary/50 rounded-xl transition-all"
                                />
                            </div>
                            {currentProvider?.id === 'ollama' && (
                                <p className="text-sm text-muted-foreground bg-primary/5 p-3 rounded-lg border border-primary/10 flex items-center gap-2">
                                    <span>💡</span>
                                    <span>Para uso local, deixe em branco para usar <code className="bg-background px-1.5 py-0.5 rounded border">http://localhost:11434</code></span>
                                </p>
                            )}
                        </div>
                    )}

                    {/* API Key - Only for providers that need it */}
                    {currentProvider?.hasApiKey && (
                        <div className="space-y-3">
                            <div className="flex items-center justify-between">
                                <Label className="text-base font-semibold flex items-center gap-2">
                                    <span className="text-primary">🔑</span> API Key
                                </Label>
                                {currentProvider?.apiKeyUrl && (
                                    <a
                                        href={currentProvider.apiKeyUrl}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="text-xs font-medium text-primary hover:text-primary/80 hover:underline flex items-center gap-1 transition-colors bg-primary/10 px-3 py-1.5 rounded-full"
                                    >
                                        Obter Chave ↗
                                    </a>
                                )}
                            </div>
                            <div className="relative group">
                                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors text-base">
                                    🔒
                                </div>
                                <Input
                                    type="password"
                                    value={apiKey}
                                    onChange={(e) => onApiKeyChange(e.target.value)}
                                    placeholder={`Cole sua API Key do ${currentProvider?.name} aqui...`}
                                    className="h-14 pl-12 text-base font-mono bg-background/50 border-2 border-primary/20 focus:border-primary/50 rounded-xl transition-all tracking-wider"
                                />
                            </div>
                            <p className="text-sm text-muted-foreground pl-1">
                                {currentProvider?.apiKeyHelp}
                            </p>
                        </div>
                    )}


                </CardContent>
            </Card>

            {/* Model Selection Card */}
            <Card className="bg-linear-to-br from-muted/50 to-muted/30 border-2 border-primary/20">
                <CardHeader className="space-y-2">
                    <div className="flex items-center gap-4">
                        <div className="w-16 h-16 bg-linear-to-br from-purple-500/20 to-pink-500/20 rounded-2xl flex items-center justify-center text-3xl border-2 border-purple-500/30 shadow-lg">
                            🤖
                        </div>
                        <div className="flex-1 space-y-1">
                            <CardTitle className="text-2xl font-bold">Seleção de Modelo</CardTitle>
                            <CardDescription className="text-base">
                                Escolha o modelo de IA que melhor atende suas necessidades
                            </CardDescription>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="space-y-6">
                    <div className="space-y-4">
                        <div className="flex items-center justify-between">
                            <Label className="text-base font-semibold">
                                Modelo Principal
                                <span className={`ml-2 text-xs px-2 py-0.5 rounded-full font-mono ${useCustomModel ? 'bg-amber-500/20 text-amber-600' : 'bg-primary/20 text-primary'}`}>
                                    {useCustomModel ? 'Personalizado' : 'Lista'}
                                </span>
                            </Label>
                            <Button
                                variant="ghost"
                                size="sm"
                                onClick={onLoadModels}
                                disabled={isLoadingModels || !apiKey}
                                className="h-8 text-xs"
                            >
                                {isLoadingModels ? (
                                    <span className="animate-spin mr-1">🔄</span>
                                ) : (
                                    <span className="mr-1">🔄</span>
                                )}
                                {isLoadingModels ? "Carregando..." : "Atualizar Lista"}
                            </Button>
                        </div>

                        {useCustomModel ? (
                            <div className="space-y-2">
                                <div className="flex gap-2">
                                    <Input
                                        value={customModel}
                                        onChange={(e) => onCustomModelChange(e.target.value)}
                                        placeholder="Digite o ID do modelo (ex: gpt-4)"
                                        className="h-12 text-base font-mono bg-background"
                                        autoFocus
                                    />
                                    <Button
                                        variant="outline"
                                        onClick={() => onUseCustomModelChange(false)}
                                        className="h-12 px-4"
                                        title="Voltar para a lista"
                                    >
                                        📋 Lista
                                    </Button>
                                </div>
                                <p className="text-sm text-muted-foreground">
                                    Digite o ID exato do modelo conforme a documentação do provedor.
                                </p>
                            </div>
                        ) : (
                            <div className="space-y-2">
                                <Select
                                    value={model}
                                    onValueChange={(val) => {
                                        if (val === 'custom-option') {
                                            onUseCustomModelChange(true);
                                        } else {
                                            onModelChange(val);
                                        }
                                    }}
                                >
                                    <SelectTrigger className="h-12 text-base bg-background">
                                        <SelectValue placeholder="Selecione um modelo..." />
                                    </SelectTrigger>
                                    <SelectContent className="max-h-80">
                                        <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground border-b mb-1">
                                            {availableModels.length > 0 ? `Disponíveis (${availableModels.length})` : 'Sugestões Populares'}
                                        </div>

                                        {availableModels.length > 0 ? (
                                            availableModels.slice(0, 100).map(m => (
                                                <SelectItem key={m.id} value={m.id} className="py-2">
                                                    <span>{m.name || m.id}</span>
                                                    {m.contextLength > 0 && (
                                                        <span className="ml-2 text-xs text-muted-foreground">
                                                            ({(m.contextLength / 1000).toFixed(0)}k)
                                                        </span>
                                                    )}
                                                </SelectItem>
                                            ))
                                        ) : (
                                            currentPopularModels.map(m => (
                                                <SelectItem key={m.value} value={m.value} className="py-2">
                                                    <span className="font-medium">{m.label}</span>
                                                    <span className="ml-2 text-xs text-muted-foreground">- {m.desc}</span>
                                                </SelectItem>
                                            ))
                                        )}

                                        <div className="my-1 border-t" />
                                        <SelectItem value="custom-option" className="py-2 text-amber-600 dark:text-amber-400 font-semibold cursor-pointer">
                                            ✏️ Digitar Modelo Manualmente...
                                        </SelectItem>
                                    </SelectContent>
                                </Select>
                                {availableModels.length === 0 && (
                                    <p className="text-xs text-muted-foreground">
                                        Use o botão "Atualizar Lista" acima para ver todos os modelos da sua conta.
                                    </p>
                                )}
                            </div>
                        )}
                    </div>
                </CardContent>
            </Card>

            {/* Secondary Model Card (For Tools) */}
            <Card className="bg-linear-to-br from-card/60 to-card/40 border-2 border-primary/20">
                <CardHeader>
                    <CardTitle className="text-xl font-bold flex items-center gap-2">
                        <span>🛠️</span> Modelo Secundário (Ferramentas)
                    </CardTitle>
                    <CardDescription>
                        Escolha um modelo mais rápido/barato para executar ações e ferramentas. (Opcional)
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="space-y-4">
                        <div className="space-y-2">
                            <Label className="text-base font-semibold">
                                Modelo para Ferramentas
                                {toolModel && <span className="ml-2 text-xs bg-primary/20 px-2 py-0.5 rounded-full text-primary font-mono">{toolModel}</span>}
                            </Label>
                            <Select value={toolModel} onValueChange={onToolModelChange}>
                                <SelectTrigger className="h-12 text-base">
                                    <SelectValue placeholder="Usar mesmo modelo do chat (Padrão)" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="same-as-chat">Usar mesmo modelo do chat (Padrão)</SelectItem>
                                    {availableModels.length > 0 ? (
                                        availableModels.slice(0, 50).map(m => (
                                            <SelectItem key={m.id} value={m.id}>
                                                {m.name || m.id} {(m.contextLength ? `(${(m.contextLength / 1000).toFixed(0)}k)` : '')}
                                            </SelectItem>
                                        ))
                                    ) : (
                                        // Fallback popular models if API not loaded
                                        currentPopularModels.map(m => (
                                            <SelectItem key={m.value} value={m.value}>
                                                {m.label} - {m.desc}
                                            </SelectItem>
                                        ))
                                    )}
                                </SelectContent>
                            </Select>
                            <p className="text-sm text-muted-foreground">
                                Se vazio ou "Padrão", usará o modelo principal do chat. Recomenda-se modelos "Flash" ou "Haiku" para maior velocidade.
                            </p>
                        </div>
                    </div>
                </CardContent>
            </Card>


        </div>
    );
}