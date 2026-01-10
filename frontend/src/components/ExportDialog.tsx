// Export Dialog - Modal to export conversations in various formats
import { useState } from 'react'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Checkbox } from "@/components/ui/checkbox"
import type { Message } from '@/types'
import { exportToMarkdown, exportToJSON, exportToText, downloadFile, EXPORT_FORMATS } from '@/services/conversationExporter'

interface ExportDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    messages: Message[]
    conversationTitle?: string
}

export function ExportDialog({
    open,
    onOpenChange,
    messages,
    conversationTitle = 'Conversation'
}: ExportDialogProps) {
    const [format, setFormat] = useState<'markdown' | 'json' | 'txt'>('markdown')
    const [includeActions, setIncludeActions] = useState(true)
    const [includeReasoning, setIncludeReasoning] = useState(false)

    const handleExport = () => {
        let exportData

        switch (format) {
            case 'markdown':
                exportData = exportToMarkdown(messages, conversationTitle)
                break
            case 'json':
                exportData = exportToJSON(messages, conversationTitle)
                break
            case 'txt':
                exportData = exportToText(messages, conversationTitle)
                break
            default:
                exportData = exportToMarkdown(messages, conversationTitle)
        }

        downloadFile(exportData)
        onOpenChange(false)
        
        // Show success toast
        import('sonner').then(({ toast }) => {
            toast.success(`Conversa exportada em ${format.toUpperCase()}!`)
        })
    }

    const visibleMessages = messages.filter(m => !m.hidden && m.role !== 'system')

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <span className="text-2xl">📤</span>
                        Exportar Conversa
                    </DialogTitle>
                    <DialogDescription>
                        Escolha o formato e opções de exportação
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-6 py-4">
                    {/* Format Selection */}
                    <div className="space-y-3">
                        <Label className="text-sm font-semibold">Formato de Exportação</Label>
                        <RadioGroup value={format} onValueChange={(v: 'markdown' | 'json' | 'txt') => setFormat(v)}>
                            {EXPORT_FORMATS.map((fmt) => (
                                <div key={fmt.value} className="flex items-center space-x-2 p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors">
                                    <RadioGroupItem value={fmt.value} id={fmt.value} />
                                    <Label htmlFor={fmt.value} className="flex-1 cursor-pointer">
                                        <span className="flex items-center gap-2">
                                            <span>{fmt.icon}</span>
                                            <span className="font-medium">{fmt.label}</span>
                                        </span>
                                        <div className="text-xs text-muted-foreground mt-1">
                                            {format === 'markdown' && 'Ideal para documentação, Markdown suporta formatação rica'}
                                            {format === 'json' && 'Formato estruturado, útil para processamento programático'}
                                            {format === 'txt' && 'Texto simples, compatível com qualquer editor'}
                                        </div>
                                    </Label>
                                </div>
                            ))}
                        </RadioGroup>
                    </div>

                    {/* Export Options */}
                    <div className="space-y-3">
                        <Label className="text-sm font-semibold">Opções</Label>
                        <div className="space-y-2">
                            <div className="flex items-center space-x-2 p-3 rounded-lg border border-border">
                                <Checkbox
                                    id="include-actions"
                                    checked={includeActions}
                                    onCheckedChange={(checked: boolean) => setIncludeActions(checked)}
                                />
                                <Label htmlFor="include-actions" className="cursor-pointer">
                                    Incluir marcadores de ações executadas
                                </Label>
                            </div>
                            <div className="flex items-center space-x-2 p-3 rounded-lg border border-border">
                                <Checkbox
                                    id="include-reasoning"
                                    checked={includeReasoning}
                                    onCheckedChange={(checked: boolean) => setIncludeReasoning(checked)}
                                />
                                <Label htmlFor="include-reasoning" className="cursor-pointer">
                                    Incluir reasoning do modelo
                                </Label>
                            </div>
                        </div>
                    </div>

                    {/* Preview Info */}
                    <div className="bg-muted/50 rounded-lg p-3 space-y-2">
                        <div className="flex justify-between text-sm">
                            <span className="text-muted-foreground">Mensagens:</span>
                            <span className="font-medium">{visibleMessages.length}</span>
                        </div>
                        <div className="flex justify-between text-sm">
                            <span className="text-muted-foreground">Com ações:</span>
                            <span className="font-medium">{visibleMessages.filter(m => m.hasActions).length}</span>
                        </div>
                    </div>
                </div>

                {/* Actions */}
                <div className="flex gap-2 justify-end">
                    <Button variant="outline" onClick={() => onOpenChange(false)}>
                        Cancelar
                    </Button>
                    <Button onClick={handleExport}>
                        <span className="mr-2">📥</span>
                        Exportar
                    </Button>
                </div>
            </DialogContent>
        </Dialog>
    )
}
