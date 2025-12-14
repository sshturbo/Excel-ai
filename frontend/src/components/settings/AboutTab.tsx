// About Tab content for Settings
import { Card, CardContent } from "@/components/ui/card"

export function AboutTab() {
    return (
        <Card className="bg-card/60 text-center py-8">
            <CardContent className="space-y-4">
                <div className="text-6xl animate-bounce">📊</div>
                <h2 className="text-3xl font-bold bg-linear-to-r from-primary to-blue-500 bg-clip-text text-transparent">
                    HipoSystem
                </h2>
                <p className="text-muted-foreground">✨ Inteligência Artificial ao alcance da sua planilha</p>
                <span className="inline-block px-3 py-1 bg-muted rounded-full text-sm text-primary">
                    v2.0.0
                </span>
                <div className="pt-6 space-y-2">
                    <p className="text-sm text-muted-foreground">Desenvolvido por</p>
                    <p className="text-lg font-semibold text-primary">
                        Jefferson Hipolito de Oliveira
                    </p>
                    <p className="text-sm text-muted-foreground">HipoSystem</p>
                </div>
                <p className="text-xs text-muted-foreground pt-4">
                    Dados em <code className="bg-muted px-1 rounded">~/.excel-ai/</code>
                </p>
            </CardContent>
        </Card>
    )
}
