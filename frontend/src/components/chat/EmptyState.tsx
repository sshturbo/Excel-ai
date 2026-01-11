// EmptyState component - Shown when no messages, with dynamic greeting
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useMemo } from "react"

interface EmptyStateProps {
    selectedSheets: string[]
    onOpenNative?: () => void
}

// Get greeting based on time of day
function getGreeting(): string {
    const hour = new Date().getHours()
    if (hour >= 5 && hour < 12) return "Bom dia"
    if (hour >= 12 && hour < 18) return "Boa tarde"
    return "Boa noite"
}

export function EmptyState({ selectedSheets, onOpenNative }: EmptyStateProps) {
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
                <CardContent className="space-y-6">
                    {!selectedSheets.length && (
                        <div className="flex justify-center pb-2">
                            <button
                                onClick={onOpenNative}
                                className="w-full h-16 bg-primary hover:scale-[1.02] transition-transform text-primary-foreground font-bold rounded-xl shadow-lg shadow-primary/20 flex items-center justify-center gap-3 text-lg"
                            >
                                <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                                    <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                                </svg>
                                Abrir Arquivo Excel
                            </button>
                        </div>
                    )}
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
