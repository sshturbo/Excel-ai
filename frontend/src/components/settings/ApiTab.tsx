// API Tab content for Settings - Z.AI Only
import { useState } from "react"
import { dto } from "../../../wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Check, ExternalLink } from "lucide-react"
import { cn } from "@/lib/utils"

// Z.AI Models
const zaiModels = [
    { value: "glm-4.7", label: "GLM-4.7", desc: "Flagship - melhor para coding", context: 128000 },
    { value: "glm-4.6v", label: "GLM-4.6V", desc: "Multimodal com visão", context: 128000 },
    { value: "glm-4.6", label: "GLM-4.6", desc: "Versátil e equilibrado", context: 128000 },
    { value: "glm-4.5", label: "GLM-4.5", desc: "Econômico", context: 128000 },
    { value: "glm-4.5-air", label: "GLM-4.5 Air", desc: "Ultraleve", context: 128000 },
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
    onToolModelChange: (value: string) => void
    onCustomModelChange: (value: string) => void
    onUseCustomModelChange: (value: boolean) => void
    onModelFilterChange: (value: string) => void
    onLoadModels: () => void
}

export function ApiTab({
    apiKey,
    baseUrl,
    model,
    toolModel,
    customModel,
    useCustomModel,
    availableModels,
    isLoadingModels,
    onApiKeyChange,
    onBaseUrlChange,
    onModelChange,
    onToolModelChange,
    onCustomModelChange,
    onUseCustomModelChange,
    onLoadModels,
}: ApiTabProps) {
    return (
        <div className="space-y-6">
            {/* Provider Card - Z.AI Only */}
            <Card className="bg-linear-to-br from-card/60 to-card/40 border-2 border-primary/20">
                <CardHeader className="space-y-2">
                    <div className="flex items-center gap-4">
                        <div className="w-16 h-16 bg-linear-to-br from-primary/20 to-primary/40 rounded-2xl flex items-center justify-center text-3xl border-2 border-primary/30 shadow-lg">
                            🤖
                        </div>
                        <div className="flex-1 space-y-1">
                            <CardTitle className="text-2xl font-bold">Z.AI (GLM Models)</CardTitle>
                            <CardDescription className="text-base">
                                Modelos GLM da Zhipu AI - 128K tokens de contexto
                            </CardDescription>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="space-y-6">
                    {/* Base URL */}
                    <div className="space-y-3">
                        <Label className="text-base font-semibold flex items-center gap-2">
                            <span className="text-primary">🔗</span> Base URL
                        </Label>
                        <div className="relative group">
                            <div className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors text-base">
                                🌐
                            </div>
                            <Input
                                type="text"
                                value={baseUrl}
                                onChange={(e) => onBaseUrlChange(e.target.value)}
                                placeholder="https://api.z.ai/api/coding/paas/v4/"
                                className="h-14 pl-12 text-base font-mono bg-background/50 border-2 border-primary/20 focus:border-primary/50 rounded-xl transition-all"
                            />
                        </div>
                        <p className="text-sm text-muted-foreground bg-primary/5 p-3 rounded-lg border border-primary/10 flex items-center gap-2">
                            <span>💡</span>
                            <span>URL padrão: <code className="bg-background px-1.5 py-0.5 rounded border">https://api.z.ai/api/coding/paas/v4/</code></span>
                        </p>
                    </div>

                    {/* API Key */}
                    <div className="space-y-3">
                        <div className="flex items-center justify-between">
                            <Label className="text-base font-semibold flex items-center gap-2">
                                <span className="text-primary">🔑</span> API Key
                            </Label>
                            <a
                                href="https://z.ai/model-api"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-xs font-medium text-primary hover:text-primary/80 hover:underline flex items-center gap-1 transition-colors bg-primary/10 px-3 py-1.5 rounded-full"
                            >
                                Obter Chave <ExternalLink className="w-3 h-3" />
                            </a>
                        </div>
                        <div className="relative group">
                            <div className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors text-base">
                                🔒
                            </div>
                            <Input
                                type="password"
                                value={apiKey}
                                onChange={(e) => onApiKeyChange(e.target.value)}
                                placeholder="Cole sua API Key do Z.AI aqui..."
                                className="h-14 pl-12 text-base font-mono bg-background/50 border-2 border-primary/20 focus:border-primary/50 rounded-xl transition-all tracking-wider"
                            />
                        </div>
                        <p className="text-sm text-muted-foreground pl-1">
                            Obtenha sua API Key no console do Z.AI (BigModel)
                        </p>
                    </div>
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
                                        placeholder="Digite o ID do modelo (ex: glm-4.7)"
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
                                    Digite o ID exato do modelo conforme a documentação do Z.AI.
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
                                        {availableModels.length > 0 ? (
                                            <>
                                                <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground border-b mb-1">
                                                    Disponíveis ({availableModels.length})
                                                </div>
                                                {availableModels.slice(0, 100).map(m => (
                                                    <SelectItem key={m.id} value={m.id} className="py-2">
                                                        <div className="flex items-center gap-2">
                                                            <Check
                                                                className={cn(
                                                                    "h-4 w-4",
                                                                    model === m.id ? "opacity-100" : "opacity-0"
                                                                )}
                                                            />
                                                            <span className="flex-1">{m.name || m.id}</span>
                                                            {m.contextLength > 0 && (
                                                                <span className="text-xs text-muted-foreground">
                                                                    ({(m.contextLength / 1000).toFixed(0)}k tokens)
                                                                </span>
                                                            )}
                                                        </div>
                                                    </SelectItem>
                                                ))}
                                            </>
                                        ) : (
                                            <>
                                                <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground border-b mb-1">
                                                    Modelos Populares
                                                </div>
                                                {zaiModels.map(m => (
                                                    <SelectItem key={m.value} value={m.value} className="py-2">
                                                        <div className="flex items-center gap-2">
                                                            <Check
                                                                className={cn(
                                                                    "h-4 w-4",
                                                                    model === m.value ? "opacity-100" : "opacity-0"
                                                                )}
                                                            />
                                                            <div className="flex-1">
                                                                <span className="font-medium">{m.label}</span>
                                                                <span className="ml-2 text-xs text-muted-foreground">- {m.desc}</span>
                                                            </div>
                                                        </div>
                                                    </SelectItem>
                                                ))}
                                            </>
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
                                    {zaiModels.map(m => (
                                        <SelectItem key={m.value} value={m.value}>
                                            {m.label} - {m.desc}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <p className="text-sm text-muted-foreground">
                                Se vazio ou "Padrão", usará o modelo principal do chat.
                            </p>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
