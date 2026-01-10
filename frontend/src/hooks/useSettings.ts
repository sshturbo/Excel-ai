// Hook for managing settings state and operations
import { useState, useEffect, useMemo, useCallback } from 'react'
import { toast } from 'sonner'
import { dto } from "../../wailsjs/go/models"

import {
    SetAPIKey,
    SetModel,
    GetSavedConfig,
    UpdateConfig,
    GetAvailableModels,
    SwitchProvider
} from "../../wailsjs/go/app/App"

// Popular models fallback
export const popularModels = [
    { value: 'openai/gpt-4o-mini', label: 'GPT-4o Mini', desc: 'Rápido e econômico' },
    { value: 'openai/gpt-4o', label: 'GPT-4o', desc: 'Avançado' },
    { value: 'anthropic/claude-3.5-sonnet', label: 'Claude 3.5 Sonnet', desc: 'Análise excelente' },
    { value: 'anthropic/claude-3-haiku', label: 'Claude 3 Haiku', desc: 'Ultra rápido' },
    { value: 'google/gemini-pro-1.5', label: 'Gemini Pro 1.5', desc: 'Contexto longo' },
    { value: 'deepseek/deepseek-chat', label: 'DeepSeek Chat', desc: 'Ótimo custo' },
    { value: 'glm-4.7', label: 'GLM-4.7', desc: 'Z.AI - Coding otimizado' },
    { value: 'glm-4.6', label: 'GLM-4.6', desc: 'Z.AI - Versátil' },
]

interface UseSettingsProps {
    askBeforeApply: boolean
    onAskBeforeApplyChange: (value: boolean) => void
}

export function useSettings({ askBeforeApply, onAskBeforeApplyChange }: UseSettingsProps) {
    // API settings
    const [apiKey, setApiKey] = useState('')
    const [model, setModel] = useState('openai/gpt-4o-mini')
    const [toolModel, setToolModel] = useState('') // New secondary model state
    const [provider, setProvider] = useState('openrouter')
    const [baseUrl, setBaseUrl] = useState('')
    const [customModel, setCustomModel] = useState('')
    const [useCustomModel, setUseCustomModel] = useState(false)

    // Data settings
    const [maxRowsContext, setMaxRowsContext] = useState(50)
    const [maxContextChars, setMaxContextChars] = useState(6000)
    const [maxRowsPreview, setMaxRowsPreview] = useState(100)
    const [includeHeaders, setIncludeHeaders] = useState(true)

    // UI state
    const [isSaving, setIsSaving] = useState(false)
    const [availableModels, setAvailableModels] = useState<dto.ModelInfo[]>([])
    const [isLoadingModels, setIsLoadingModels] = useState(false)
    const [modelFilter, setModelFilter] = useState('')

    useEffect(() => {
        const hasWailsRuntime = typeof (window as any)?.go !== 'undefined'
        if (hasWailsRuntime) {
            loadConfig()
        } else {
            console.warn('Wails runtime não detectado. Settings em modo somente UI (Vite puro).')
        }
    }, [])

    // Clear models when changing provider
    useEffect(() => {
        setAvailableModels([])
        setModelFilter('')
    }, [provider])

    const loadConfig = useCallback(async () => {
        try {
            const cfg = await GetSavedConfig()
            if (cfg) {
                if (cfg.apiKey) {
                    setApiKey(cfg.apiKey)
                    loadModels()
                }
                if (cfg.provider) setProvider(cfg.provider)
                if (cfg.baseUrl) setBaseUrl(cfg.baseUrl)
                if (cfg.model) {
                    setModel(cfg.model)
                    const isPopular = popularModels.some(m => m.value === cfg.model)
                    if (!isPopular) {
                        setCustomModel(cfg.model)
                        setUseCustomModel(true)
                    }
                }
                if (cfg.toolModel) {
                    setToolModel(cfg.toolModel)
                } else {
                    setToolModel('')
                }
                if (cfg.maxRowsContext) setMaxRowsContext(cfg.maxRowsContext)
                if (cfg.maxContextChars) setMaxContextChars(cfg.maxContextChars)
                if (cfg.maxRowsPreview) setMaxRowsPreview(cfg.maxRowsPreview)
                setIncludeHeaders(cfg.includeHeaders !== false)
            }
        } catch (err) {
            toast.error('Erro ao carregar configurações')
        }
    }, [])

    const loadModels = useCallback(async () => {
        if (typeof (window as any)?.go === 'undefined') return
        if (!apiKey) {
            toast.warning('Configure a API Key primeiro')
            return
        }

        setIsLoadingModels(true)
        setAvailableModels([]) // Limpar modelos anteriores

        try {
            // Determinar URL baseado no provider atual
            let url = ''
            if (provider === 'groq') {
                url = 'https://api.groq.com/openai/v1'
            } else if (provider === 'google') {
                url = 'https://generativelanguage.googleapis.com/v1beta'
            } else if (provider === 'zai') {
                url = 'https://api.z.ai/api/paas/v4'
            } else if (provider === 'openrouter') {
                url = 'https://openrouter.ai/api/v1'
            } else {
                url = baseUrl || ''
            }

            console.log('[DEBUG] Carregando modelos para provider:', provider, 'URL:', url)
            const models = await GetAvailableModels(apiKey, url)
            if (models && models.length > 0) {
                setAvailableModels(models)
                toast.success(`${models.length} modelos carregados!`)
            } else {
                toast.warning('Nenhum modelo retornado pela API')
            }
        } catch (err) {
            console.error('Erro ao carregar modelos:', err)
            toast.error('Erro ao carregar modelos. Usando lista padrão.')
        } finally {
            setIsLoadingModels(false)
        }
    }, [apiKey, provider, baseUrl])

    // Filter models
    const filteredModels = useMemo(() => {
        if (!modelFilter.trim()) return availableModels.slice(0, 50)
        const search = modelFilter.toLowerCase()
        return availableModels.filter(m =>
            m.id.toLowerCase().includes(search) ||
            m.name.toLowerCase().includes(search)
        ).slice(0, 50)
    }, [availableModels, modelFilter])

    const handleSave = useCallback(async () => {
        if (typeof (window as any)?.go === 'undefined') {
            toast.error('Wails não detectado. Não é possível salvar fora do app.')
            return
        }
        setIsSaving(true)
        try {
            await SetAPIKey(apiKey)
            const selectedModel = useCustomModel ? customModel : model
            await SetModel(selectedModel)

            // Determinar URL correta baseado no provider
            let correctBaseUrl = ''
            if (provider === 'groq') {
                correctBaseUrl = 'https://api.groq.com/openai/v1'
            } else if (provider === 'google') {
                correctBaseUrl = 'https://generativelanguage.googleapis.com/v1beta'
            } else if (provider === 'zai') {
                correctBaseUrl = 'https://api.z.ai/api/paas/v4'
            } else if (provider === 'openrouter') {
                correctBaseUrl = 'https://openrouter.ai/api/v1'
            } else {
                correctBaseUrl = baseUrl
            }

            // Parâmetros: maxRowsContext, maxContextChars, maxRowsPreview, includeHeaders, detailLevel, customPrompt, language, provider, toolModel, baseUrl
            const effectiveToolModel = toolModel === 'same-as-chat' ? '' : toolModel
            await UpdateConfig(maxRowsContext, maxContextChars, maxRowsPreview, includeHeaders, 'normal', '', 'pt-BR', provider, effectiveToolModel, correctBaseUrl)
            toast.success('✅ Configurações salvas!')
            await loadModels()
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('❌ Erro ao salvar: ' + errorMessage)
        } finally {
            setIsSaving(false)
        }
    }, [apiKey, customModel, model, useCustomModel, toolModel, maxRowsContext, maxContextChars, maxRowsPreview, includeHeaders, provider, baseUrl, loadModels])

    const handleProviderChange = useCallback(async (val: string) => {
        setProvider(val)
        setAvailableModels([])

        try {
            const cfg = await SwitchProvider(val)
            if (cfg) {
                if (cfg.apiKey) {
                    setApiKey(cfg.apiKey)
                } else {
                    setApiKey('')
                }
                if (cfg.model) {
                    setModel(cfg.model)
                } else {
                    setModel('')
                }
                if (cfg.toolModel) {
                    setToolModel(cfg.toolModel)
                } else {
                    setToolModel('')
                }
                if (cfg.baseUrl) {
                    setBaseUrl(cfg.baseUrl)
                } else {
                    if (val === 'groq') {
                        setBaseUrl('https://api.groq.com/openai/v1')
                    } else if (val === 'zai') {
                        setBaseUrl('https://api.z.ai/api/paas/v4')
                    } else if (val === 'openrouter') {
                        setBaseUrl('https://openrouter.ai/api/v1')
                    } else if (val === 'google') {
                        setBaseUrl('https://generativelanguage.googleapis.com/v1beta')
                    } else {
                        setBaseUrl('')
                    }
                }

                if (cfg.apiKey) {
                    toast.success(`Configurações do ${val} carregadas!`)
                } else {
                    toast.info(`Configure a API Key para ${val}`)
                }
            }
        } catch (err) {
            console.error('Erro ao trocar provider:', err)
            if (val === 'groq') {
                setBaseUrl('https://api.groq.com/openai/v1')
            } else if (val === 'zai') {
                setBaseUrl('https://api.z.ai/api/paas/v4')
            } else if (val === 'openrouter') {
                setBaseUrl('https://openrouter.ai/api/v1')
            } else if (val === 'google') {
                setBaseUrl('https://generativelanguage.googleapis.com/v1beta')
            }
            setApiKey('')
        }
    }, [])

    return {
        // API settings
        apiKey,
        setApiKey,
        model,
        setModel,
        toolModel,
        setToolModel,
        provider,
        baseUrl,
        setBaseUrl,
        customModel,
        setCustomModel,
        useCustomModel,
        setUseCustomModel,

        // Data settings
        maxRowsContext,
        setMaxRowsContext,
        maxContextChars,
        setMaxContextChars,
        maxRowsPreview,
        setMaxRowsPreview,
        includeHeaders,
        setIncludeHeaders,
        askBeforeApply,
        onAskBeforeApplyChange,

        // UI state
        isSaving,
        availableModels,
        isLoadingModels,
        modelFilter,
        setModelFilter,
        filteredModels,

        // Actions
        loadModels,
        handleSave,
        handleProviderChange
    }
}
