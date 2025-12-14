// Data Tab content for Settings
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Slider } from "@/components/ui/slider"
import { Switch } from "@/components/ui/switch"

interface DataTabProps {
    maxRowsContext: number
    maxRowsPreview: number
    includeHeaders: boolean
    askBeforeApply: boolean
    onMaxRowsContextChange: (value: number) => void
    onMaxRowsPreviewChange: (value: number) => void
    onIncludeHeadersChange: (value: boolean) => void
    onAskBeforeApplyChange: (value: boolean) => void
}

export function DataTab({
    maxRowsContext,
    maxRowsPreview,
    includeHeaders,
    askBeforeApply,
    onMaxRowsContextChange,
    onMaxRowsPreviewChange,
    onIncludeHeadersChange,
    onAskBeforeApplyChange
}: DataTabProps) {
    return (
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
                        onValueChange={(v) => onMaxRowsContextChange(v[0])}
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
                        onValueChange={(v) => onMaxRowsPreviewChange(v[0])}
                        min={50}
                        max={500}
                        step={50}
                    />
                </div>

                <div className="flex items-center justify-between p-4 bg-muted/30 border border-border rounded-lg">
                    <Label>Incluir cabeçalhos no contexto</Label>
                    <Switch checked={includeHeaders} onCheckedChange={onIncludeHeadersChange} />
                </div>

                <div className="flex items-center justify-between p-4 bg-muted/30 border border-border rounded-lg">
                    <div className="space-y-1">
                        <Label>Perguntar antes de aplicar</Label>
                        <p className="text-xs text-muted-foreground">
                            A IA pedirá confirmação antes de modificar a planilha
                        </p>
                    </div>
                    <Switch checked={askBeforeApply} onCheckedChange={onAskBeforeApplyChange} />
                </div>
            </CardContent>
        </Card>
    )
}
