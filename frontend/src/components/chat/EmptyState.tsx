// EmptyState component - Shown when no messages, with dynamic greeting
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useMemo } from "react"

interface EmptyStateProps {
    selectedSheets: string[]
}

// Get greeting based on time of day
function getGreeting(): string {
    const hour = new Date().getHours()
    if (hour >= 5 && hour < 12) return "Bom dia"
    if (hour >= 12 && hour < 18) return "Boa tarde"
    return "Boa noite"
}

export function EmptyState({ selectedSheets }: EmptyStateProps) {
    const greeting = useMemo(() => getGreeting(), [])

    return (
        <div className="flex flex-col items-center justify-center h-full">
            <Card className="w-full max-w-lg bg-card/60">
                <CardHeader className="text-center pb-2">
                    <div className="text-4xl mb-2">📊</div>
                    <CardTitle className="text-2xl font-bold bg-linear-to-r from-primary to-blue-500 bg-clip-text text-transparent">
                        {greeting}! Sou o HipoSystem
                    </CardTitle>
                    <p className="text-muted-foreground text-sm mt-1">
                        ✨ Inteligência Artificial ao alcance da sua planilha
                    </p>
                </CardHeader>
                <CardContent className="space-y-4">
                    {selectedSheets.length > 0 ? (
                        <div className="space-y-3">
                            <p className="text-sm text-center text-muted-foreground">
                                ✅ Abas carregadas: <strong className="text-primary">{selectedSheets.join(', ')}</strong>
                            </p>
                            <div className="bg-muted/30 rounded-lg p-4">
                                <p className="text-sm font-medium mb-2">💡 Posso ajudar você a:</p>
                                <ul className="text-sm text-muted-foreground space-y-1.5">
                                    <li className="flex items-start gap-2">
                                        <span>📈</span>
                                        <span>Analisar e resumir seus dados</span>
                                    </li>
                                    <li className="flex items-start gap-2">
                                        <span>🔢</span>
                                        <span>Criar fórmulas (SOMA, MÉDIA, PROCV...)</span>
                                    </li>
                                    <li className="flex items-start gap-2">
                                        <span>📊</span>
                                        <span>Gerar gráficos e tabelas dinâmicas</span>
                                    </li>
                                    <li className="flex items-start gap-2">
                                        <span>🎨</span>
                                        <span>Formatar e organizar sua planilha</span>
                                    </li>
                                </ul>
                            </div>
                            <p className="text-xs text-center text-muted-foreground">
                                Digite sua pergunta abaixo para começar!
                            </p>
                        </div>
                    ) : (
                        <div className="space-y-4">
                            <p className="text-sm text-center text-muted-foreground">
                                Sou seu assistente de planilhas com IA.
                            </p>
                            <div className="bg-muted/30 rounded-lg p-4">
                                <p className="text-sm font-medium mb-2">🚀 Para começar:</p>
                                <ol className="text-sm text-muted-foreground space-y-1.5 list-decimal list-inside">
                                    <li>Abra uma planilha no Excel</li>
                                    <li>Selecione uma aba na lateral esquerda</li>
                                    <li>Faça perguntas sobre seus dados!</li>
                                </ol>
                            </div>
                            <p className="text-xs text-center text-muted-foreground">
                                💡 Dica: quanto mais contexto você der, melhor posso ajudar
                            </p>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}
