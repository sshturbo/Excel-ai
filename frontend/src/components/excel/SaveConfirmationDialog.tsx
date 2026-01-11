import React from 'react';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { AlertTriangle } from 'lucide-react';

interface SaveConfirmationDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: () => void;
    fileName: string;
}

export function SaveConfirmationDialog({
    open,
    onOpenChange,
    onConfirm,
    fileName
}: SaveConfirmationDialogProps) {
    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <div className="flex items-center gap-2 text-destructive mb-2">
                        <AlertTriangle className="h-6 w-6" />
                        <DialogTitle>Aviso de Substituição</DialogTitle>
                    </div>
                    <DialogDescription className="text-base">
                        Você está prestes a salvar as alterações diretamente no arquivo original:
                        <br />
                        <strong className="text-foreground mt-2 block">{fileName}</strong>
                    </DialogDescription>
                </DialogHeader>

                <div className="bg-destructive/10 p-4 rounded-lg border border-destructive/20 my-4">
                    <p className="text-sm text-destructive font-medium flex gap-2">
                        <span>⚠️</span>
                        Esta ação irá substituir permanentemente o conteúdo original e não poderá ser desfeita.
                    </p>
                </div>

                <DialogFooter className="gap-2 sm:gap-0">
                    <Button
                        variant="outline"
                        onClick={() => onOpenChange(false)}
                    >
                        Cancelar
                    </Button>
                    <Button
                        variant="destructive"
                        onClick={() => {
                            onConfirm();
                            onOpenChange(false);
                        }}
                    >
                        Substituir e Salvar
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
