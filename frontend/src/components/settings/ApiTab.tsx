// API Tab content for Settings
import { dto } from "../../../wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { popularModels } from "@/hooks/useSettings"

interface ApiTabProps {
    provider: string
    apiKey: string
    baseUrl: string
    model: string
    customModel: string
    useCustomModel: boolean
    availableModels: dto.ModelInfo[]
    filteredModels: dto.ModelInfo[]
    modelFilter: string
    isLoadingModels: boolean
    onProviderChange: (value: string) => void
    onApiKeyChange: (value: string) => void
    onBaseUrlChange: (value: string) => void
    onModelChange: (value: string) => void
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
    onCustomModelChange,
    onUseCustomModelChange,
    onModelFilterChange,
    onLoadModels
}: ApiTabProps) {
    return (
        <div className="space-y-6">
            {/* Provider Card */}
            <Card className="bg-card/60">
                <CardHeader>
                    <div className="w-12 h-12 bg-primary text-primary-foreground rounded-xl flex items-center justify-center text-2xl mb-2">
                        🔑
                    </div>
                    <CardTitle>Provedor e Chave de API</CardTitle>
                    <CardDescription>Configure o provedor de IA e sua chave de acesso</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label>Provedor</Label>
                        <Select value={provider} onValueChange={onProviderChange}>
                            <SelectTrigger>
                                <SelectValue placeholder="Selecione o provedor" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="openrouter">OpenRouter (Recomendado)</SelectItem>
                                <SelectItem value="groq">Groq (Rápido e Gratuito)</SelectItem>
                                <SelectItem value="custom">Personalizado (OpenAI Compatible)</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                    {provider === 'custom' && (
                        <div className="space-y-2">
                            <Label>Base URL</Label>
                            <Input
                                type="text"
                                value={baseUrl}
                                onChange={(e) => onBaseUrlChange(e.target.value)}
                                placeholder="https://api.exemplo.com/v1"
                            />
                        </div>
                    )}

                    <div className="space-y-2">
                        <Label>API Key</Label>
                        <Input
                            type="password"
                            value={apiKey}
                            onChange={(e) => onApiKeyChange(e.target.value)}
                            placeholder="sk-or-v1-..."
                        />
                        <p className="text-sm text-muted-foreground">
                            {provider === 'groq' ? (
                                <>Obtenha em <a href="https://console.groq.com/keys" target="_blank" className="text-primary hover:underline">console.groq.com/keys</a></>
                            ) : (
                                <>Obtenha em <a href="https://openrouter.ai/keys" target="_blank" className="text-primary hover:underline">openrouter.ai/keys</a></>
                            )}
                        </p>
                    </div>
                </CardContent>
            </Card>

            {/* Model Card */}
            <Card className="bg-card/60">
                <CardHeader>
                    <div className="w-12 h-12 bg-muted rounded-xl flex items-center justify-center text-2xl mb-2">
                        🤖
                    </div>
                    <CardTitle>Modelo de IA</CardTitle>
                    <CardDescription>Escolha o modelo que melhor atende suas necessidades</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    {/* Toggle for custom model */}
                    <div className="flex items-center justify-between p-3 bg-muted/30 rounded-lg">
                        <div>
                            <Label className="font-medium">Usar modelo personalizado</Label>
                            <p className="text-xs text-muted-foreground">Digite o ID do modelo manualmente</p>
                        </div>
                        <Switch
                            checked={useCustomModel}
                            onCheckedChange={onUseCustomModelChange}
                        />
                    </div>

                    {useCustomModel ? (
                        <div className="space-y-2">
                            <Label>ID do Modelo</Label>
                            <Input
                                value={customModel}
                                onChange={(e) => onCustomModelChange(e.target.value)}
                                placeholder="ex: anthropic/claude-3.5-sonnet"
                            />
                            <p className="text-xs text-muted-foreground">
                                Veja modelos disponíveis em <a href="https://openrouter.ai/models" target="_blank" className="text-primary hover:underline">openrouter.ai/models</a>
                            </p>
                        </div>
                    ) : (
                        <>
                            {/* Load models button */}
                            <div className="flex gap-2">
                                <Button
                                    variant="outline"
                                    onClick={onLoadModels}
                                    disabled={isLoadingModels || !apiKey}
                                    className="flex-1"
                                >
                                    {isLoadingModels ? (
                                        <>
                                            <div className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin mr-2" />
                                            Carregando...
                                        </>
                                    ) : (
                                        <>🔄 Carregar Modelos {provider === 'groq' ? 'da Groq' : provider === 'openrouter' ? 'da OpenRouter' : 'da API'}</>
                                    )}
                                </Button>
                            </div>

                            {!apiKey && (
                                <p className="text-xs text-amber-500">⚠️ Configure a API Key primeiro para carregar os modelos</p>
                            )}

                            {/* Model filter */}
                            {availableModels.length > 0 && (
                                <div className="space-y-2">
                                    <Label>Buscar modelo</Label>
                                    <Input
                                        value={modelFilter}
                                        onChange={(e) => onModelFilterChange(e.target.value)}
                                        placeholder="Digite para filtrar... (ex: gpt, claude, gemini)"
                                    />
                                </div>
                            )}

                            {/* Model list */}
                            {availableModels.length > 0 ? (
                                <div className="space-y-2">
                                    <Label className="text-muted-foreground">
                                        {filteredModels.length} de {availableModels.length} modelos
                                    </Label>
                                    <div className="grid grid-cols-1 gap-2 max-h-64 overflow-y-auto pr-2">
                                        {filteredModels.map((m) => (
                                            <button
                                                key={m.id}
                                                onClick={() => onModelChange(m.id)}
                                                className={`p-3 rounded-lg text-left transition-all ${model === m.id
                                                    ? 'bg-primary/10 border-2 border-primary/50'
                                                    : 'bg-background/40 border border-border hover:bg-muted/40'
                                                    }`}
                                            >
                                                <div className="flex items-center justify-between">
                                                    <div className="font-medium text-sm truncate flex-1">{m.name || m.id}</div>
                                                    {m.contextLength > 0 && (
                                                        <span className="text-xs text-muted-foreground ml-2">
                                                            {(m.contextLength / 1000).toFixed(0)}K ctx
                                                        </span>
                                                    )}
                                                </div>
                                                <div className="text-xs text-muted-foreground truncate">{m.id}</div>
                                            </button>
                                        ))}
                                    </div>
                                </div>
                            ) : (
                                <div className="space-y-2">
                                    <Label className="text-muted-foreground">Modelos populares</Label>
                                    <div className="grid grid-cols-2 gap-3">
                                        {popularModels.map((m) => (
                                            <button
                                                key={m.value}
                                                onClick={() => onModelChange(m.value)}
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
                                </div>
                            )}
                        </>
                    )}

                    {/* Selected model display */}
                    <div className="p-3 bg-primary/5 border border-primary/20 rounded-lg">
                        <Label className="text-xs text-muted-foreground">Modelo selecionado:</Label>
                        <div className="font-mono text-sm text-primary mt-1">
                            {useCustomModel ? customModel || 'Nenhum' : model}
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
    )
}
