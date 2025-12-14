// EmptyState component - Shown when no messages
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

interface EmptyStateProps {
    selectedSheets: string[]
}

export function EmptyState({ selectedSheets }: EmptyStateProps) {
    return (
        <div className="flex flex-col items-center justify-center h-full">
            <Card className="w-full max-w-md bg-card/60">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <span className="text-2xl">{selectedSheets.length > 0 ? '📊' : '🤖'}</span>
                        <span>{selectedSheets.length > 0 ? `${selectedSheets.length} aba(s) selecionada(s)` : 'HipoSystem pronto'}</span>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    {selectedSheets.length > 0 ? (
                        <div className="space-y-3">
                            <p className="text-sm text-muted-foreground">
                                ✅ Abas carregadas: <strong className="text-primary">{selectedSheets.join(', ')}</strong>
                            </p>
                            <p className="text-sm text-muted-foreground">
                                Faça perguntas como:
                            </p>
                            <ul className="text-sm text-muted-foreground space-y-1 ml-4">
                                <li>• "Qual a soma da coluna B?"</li>
                                <li>• "Mostre os 5 maiores valores"</li>
                                <li>• "Crie uma fórmula para calcular média"</li>
                                <li>• "Analise os dados dessa planilha"</li>
                            </ul>
                        </div>
                    ) : (
                        <p className="text-sm text-muted-foreground">
                            Selecione uma planilha na lateral e faça perguntas sobre seus dados.
                        </p>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}
