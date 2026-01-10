// Hook for managing settings state and operations - Z.AI Only
import { useState, useEffect, useMemo, useCallback } from 'react'
import { toast } from 'sonner'
import { dto } from "../../wailsjs/go/models"

import {
    SetAPIKey,
    SetModel,
    SetToolModel,
    GetSavedConfig,
    UpdateConfig,
    GetAvailableModels
} from "../../wailsjs/go/app/App"

// Z.AI Models
export const zaiModels = [
    { value: 'glm-4.7', label: 'GLM-4.7', desc: 'Flagship - melhor para coding', context: 128000 },
    { value: 'glm-4.6v', label: 'GLM-4.6V', desc: 'Multimodal com visão', context: 128000 },
    { value: 'glm-4.6', label: 'GLM-4.6', desc: 'Versátil e equilibrado', context: 128000 },
    { value: 'glm-4.5', label: 'GLM-4.5', desc: 'Econômico', context: 128000 },
    { value: 'glm-4.5-air', label: 'GLM-4.5 Air', desc: 'Ultraleve', context: 128000 },
]

interface UseSettingsProps {
    askBeforeApply: boolean
    onAskBeforeApplyChange: (value: boolean) => void
}

export function useSettings({ askBeforeApply, onAskBeforeApplyChange }: UseSettingsProps) {
    // API settings
    const [apiKey, setApiKey] = useState('')
    const [model, setModel] = useState('glm-4.7')
    const [toolModel, setToolModel] = useState('')
    const [baseUrl, setBaseUrl] = useState('https://api.z.ai/api/coding/paas/v4/')
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

    const loadConfig = useCallback(async () => {
        try {
            const cfg = await GetSavedConfig()
            if (cfg) {
                if (cfg.apiKey) {
                    setApiKey(cfg.apiKey)
                    loadModels()
                }
                if (cfg.baseUrl) setBaseUrl(cfg.baseUrl)
                if (cfg.model) {
                    setModel(cfg.model)
                    const isZaiModel = zaiModels.some(m => m.value === cfg.model)
                    if (!isZaiModel) {
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
        setAvailableModels([])

        try {
            const url = baseUrl || 'https://api.z.ai/api/coding/paas/v4/'
            console.log('[DEBUG] Carregando modelos Z.AI, URL:', url)
            const models = await GetAvailableModels(apiKey, url)
            if (models && models.length > 0) {
                setAvailableModels(models)
                toast.success(`${models.length} modelos carregados!`)
            } else {
                toast.warning('Nenhum modelo retornado pela API')
            }
        } catch (err) {
            console.error('Erro ao carregar modelos:', err)
            toast.error('Erro ao carregar modelos')
        } finally {
            setIsLoadingModels(false)
        }
    }, [apiKey, baseUrl])

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

            const effectiveToolModel = toolModel === 'same-as-chat' ? '' : toolModel
            if (effectiveToolModel) {
                await SetToolModel(effectiveToolModel)
            }

            const finalBaseUrl = baseUrl || 'https://api.z.ai/api/coding/paas/v4/'
            
            await UpdateConfig(
                maxRowsContext,
                maxContextChars,
                maxRowsPreview,
                includeHeaders,
                'normal',
                '',
                'pt-BR',
                'zai',
                effectiveToolModel,
                finalBaseUrl
            )
            toast.success('✅ Configurações salvas!')
            await loadModels()
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err)
            toast.error('❌ Erro ao salvar: ' + errorMessage)
        } finally {
            setIsSaving(false)
        }
    }, [apiKey, customModel, model, useCustomModel, toolModel, maxRowsContext, maxContextChars, maxRowsPreview, includeHeaders, baseUrl, loadModels])

    const handleProviderChange = useCallback(() => {
        // Não usado mais - Z.AI é o único provider
        toast.info('Z.AI é o único provider suportado')
    }, [])

    return {
        // API settings
        apiKey,
        setApiKey,
        model,
        setModel,
        toolModel,
        setToolModel,
        provider: 'zai',
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
