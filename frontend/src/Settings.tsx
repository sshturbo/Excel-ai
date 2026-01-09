// Settings.tsx - Refactored settings page using modular components and hooks
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

// Custom hook
import { useSettings } from '@/hooks/useSettings'

// Components
import { SettingsHeader } from '@/components/settings/SettingsHeader'
import { ApiTab } from '@/components/settings/ApiTab'
import { DataTab } from '@/components/settings/DataTab'
import { AboutTab } from '@/components/settings/AboutTab'

interface SettingsProps {
    onClose: () => void
    askBeforeApply: boolean
    onAskBeforeApplyChange: (value: boolean) => void
}

export default function Settings({ onClose, askBeforeApply, onAskBeforeApplyChange }: SettingsProps) {
    const settings = useSettings({ askBeforeApply, onAskBeforeApplyChange })

    return (
        <div className="min-h-screen bg-background text-foreground">
            {/* Gradient Background */}
            <div className="fixed inset-0 pointer-events-none">
                <div className="absolute top-0 left-1/4 w-96 h-96 bg-primary/10 rounded-full blur-3xl" />
                <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-muted/40 rounded-full blur-3xl" />
            </div>

            <div className="relative max-w-4xl mx-auto py-8 px-4">
                {/* Header */}
                <SettingsHeader
                    onClose={onClose}
                    onSave={settings.handleSave}
                    isSaving={settings.isSaving}
                />

                {/* Tabs */}
                <Tabs defaultValue="api" className="space-y-6">
                    <TabsList className="grid w-full grid-cols-3 bg-muted/40">
                        <TabsTrigger value="api" className="gap-2">🔑 API</TabsTrigger>
                        <TabsTrigger value="data" className="gap-2">📊 Dados</TabsTrigger>
                        <TabsTrigger value="about" className="gap-2">ℹ️ Sobre</TabsTrigger>
                    </TabsList>

                    {/* API Tab */}
                    <TabsContent value="api">
                        <ApiTab
                            provider={settings.provider}
                            apiKey={settings.apiKey}
                            baseUrl={settings.baseUrl}
                            model={settings.model}
                            toolModel={settings.toolModel}
                            customModel={settings.customModel}
                            useCustomModel={settings.useCustomModel}
                            availableModels={settings.availableModels}
                            filteredModels={settings.filteredModels}
                            modelFilter={settings.modelFilter}
                            isLoadingModels={settings.isLoadingModels}
                            onProviderChange={settings.handleProviderChange}
                            onApiKeyChange={settings.setApiKey}
                            onBaseUrlChange={settings.setBaseUrl}
                            onModelChange={settings.setModel}
                            onToolModelChange={settings.setToolModel}
                            onCustomModelChange={settings.setCustomModel}
                            onUseCustomModelChange={settings.setUseCustomModel}
                            onModelFilterChange={settings.setModelFilter}
                            onLoadModels={settings.loadModels}
                        />
                    </TabsContent>

                    {/* Data Tab */}
                    <TabsContent value="data">
                        <DataTab
                            maxRowsContext={settings.maxRowsContext}
                            maxContextChars={settings.maxContextChars}
                            maxRowsPreview={settings.maxRowsPreview}
                            includeHeaders={settings.includeHeaders}
                            askBeforeApply={settings.askBeforeApply}
                            onMaxRowsContextChange={settings.setMaxRowsContext}
                            onMaxContextCharsChange={settings.setMaxContextChars}
                            onMaxRowsPreviewChange={settings.setMaxRowsPreview}
                            onIncludeHeadersChange={settings.setIncludeHeaders}
                            onAskBeforeApplyChange={settings.onAskBeforeApplyChange}
                        />
                    </TabsContent>

                    {/* About Tab */}
                    <TabsContent value="about">
                        <AboutTab />
                    </TabsContent>
                </Tabs>
            </div>
        </div>
    )
}
