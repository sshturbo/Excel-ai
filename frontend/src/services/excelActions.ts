// Excel action execution service
// Extracted from App.tsx executeExcelAction function

import { toast } from 'sonner'
import type { ExcelAction, ExcelActionResult, Workbook } from '@/types'

import {
    UpdateExcelCell,
    CreateNewWorkbook,
    CreateNewSheet,
    CreateChart,
    CreatePivotTable,
    ConfigurePivotFields,
    FormatRange,
    DeleteSheet,
    RenameSheet,
    ClearRange,
    AutoFitColumns,
    InsertRows,
    DeleteRows,
    MergeCells,
    UnmergeCells,
    SetBorders,
    SetColumnWidth,
    SetRowHeight,
    ApplyFilter,
    ClearFilters,
    SortRange,
    CopyRange,
    ListCharts,
    DeleteChartByName,
    CreateTable,
    DeleteTable,
    RefreshWorkbooks
} from "../../wailsjs/go/app/App"

/**
 * Executes an Excel action and returns the result
 * @param action The Excel action to execute
 * @param onWorkbooksUpdate Optional callback to update workbooks state
 */
export async function executeExcelAction(
    action: ExcelAction,
    onWorkbooksUpdate?: (workbooks: Workbook[]) => void
): Promise<ExcelActionResult> {
    try {
        if (action.op === 'write') {
            await UpdateExcelCell(
                action.workbook || '',
                action.sheet || '',
                action.cell || '',
                action.value || ''
            )
        } else if (action.op === 'create-workbook') {
            const name = await CreateNewWorkbook()
            toast.success(`Nova pasta de trabalho criada: ${name}`)
            if (onWorkbooksUpdate) {
                const result = await RefreshWorkbooks()
                if (result.workbooks) onWorkbooksUpdate(result.workbooks)
            }
        } else if (action.op === 'create-sheet') {
            await CreateNewSheet(action.name || '')
            toast.success(`Nova aba criada: ${action.name}`)
            if (onWorkbooksUpdate) {
                const result = await RefreshWorkbooks()
                if (result.workbooks) onWorkbooksUpdate(result.workbooks)
            }
            // Small delay to ensure Excel processed the sheet creation
            await new Promise(resolve => setTimeout(resolve, 300))
        } else if (action.op === 'create-chart') {
            await CreateChart(
                action.sheet || '',
                action.range || '',
                action.chartType || 'column',
                action.title || ''
            )
            toast.success('Gráfico criado!')
        } else if (action.op === 'create-pivot') {
            console.log('[DEBUG] create-pivot:', action)
            const tableName = action.tableName || 'PivotTable1'
            await CreatePivotTable(
                action.sourceSheet || '',
                action.sourceRange || '',
                action.destSheet || '',
                action.destCell || '',
                tableName
            )

            // Configure fields if specified
            if (action.rowFields || action.valueFields) {
                // Small delay to ensure the table was created
                await new Promise(resolve => setTimeout(resolve, 500))

                const rowFields = action.rowFields || []
                const valueFields = (action.valueFields || []).map((vf) => {
                    if (typeof vf === 'string') {
                        return { field: vf, function: 'sum' }
                    }
                    return vf
                })

                await ConfigurePivotFields(
                    action.destSheet || '',
                    tableName,
                    rowFields,
                    valueFields
                )
            }

            toast.success('Tabela dinâmica criada!')
        } else if (action.op === 'format-range') {
            await FormatRange(
                action.sheet || '',
                action.range || '',
                action.bold || false,
                action.italic || false,
                action.fontSize || 0,
                action.fontColor || '',
                action.bgColor || ''
            )
            toast.success('Formatação aplicada!')
        } else if (action.op === 'delete-sheet') {
            await DeleteSheet(action.name || '')
            toast.success(`Aba "${action.name}" excluída!`)
            if (onWorkbooksUpdate) {
                const result = await RefreshWorkbooks()
                if (result.workbooks) onWorkbooksUpdate(result.workbooks)
            }
        } else if (action.op === 'rename-sheet') {
            await RenameSheet(action.oldName || '', action.newName || '')
            toast.success(`Aba renomeada: ${action.oldName} → ${action.newName}`)
            if (onWorkbooksUpdate) {
                const result = await RefreshWorkbooks()
                if (result.workbooks) onWorkbooksUpdate(result.workbooks)
            }
        } else if (action.op === 'clear-range') {
            await ClearRange(action.sheet || '', action.range || '')
            toast.success('Conteúdo limpo!')
        } else if (action.op === 'autofit') {
            await AutoFitColumns(action.sheet || '', action.range || '')
            toast.success('Colunas ajustadas!')
        } else if (action.op === 'insert-rows') {
            await InsertRows(action.sheet || '', action.row || 0, action.count || 1)
            toast.success(`${action.count || 1} linha(s) inserida(s)!`)
        } else if (action.op === 'delete-rows') {
            await DeleteRows(action.sheet || '', action.row || 0, action.count || 1)
            toast.success(`${action.count || 1} linha(s) excluída(s)!`)
        } else if (action.op === 'merge-cells') {
            await MergeCells(action.sheet || '', action.range || '')
            toast.success('Células mescladas!')
        } else if (action.op === 'unmerge-cells') {
            await UnmergeCells(action.sheet || '', action.range || '')
            toast.success('Células desmescladas!')
        } else if (action.op === 'set-borders') {
            await SetBorders(action.sheet || '', action.range || '', action.style || 'thin')
            toast.success('Bordas aplicadas!')
        } else if (action.op === 'set-column-width') {
            await SetColumnWidth(action.sheet || '', action.range || '', action.width || 15)
            toast.success('Largura definida!')
        } else if (action.op === 'set-row-height') {
            await SetRowHeight(action.sheet || '', action.range || '', action.height || 20)
            toast.success('Altura definida!')
        } else if (action.op === 'apply-filter') {
            await ApplyFilter(action.sheet || '', action.range || '')
            toast.success('Filtro aplicado!')
        } else if (action.op === 'clear-filters') {
            await ClearFilters(action.sheet || '')
            toast.success('Filtros limpos!')
        } else if (action.op === 'sort') {
            await SortRange(action.sheet || '', action.range || '', action.column || 1, action.ascending !== false)
            toast.success('Dados ordenados!')
        } else if (action.op === 'copy-range') {
            await CopyRange(action.sheet || '', action.source || '', action.dest || '')
            toast.success('Range copiado!')
        } else if (action.op === 'list-charts') {
            const charts = await ListCharts(action.sheet || '')
            toast.info(`Gráficos encontrados: ${charts.join(', ') || 'nenhum'}`)
        } else if (action.op === 'delete-chart') {
            await DeleteChartByName(action.sheet || '', action.name || '')
            toast.success(`Gráfico "${action.name}" excluído!`)
        } else if (action.op === 'create-table') {
            await CreateTable(action.sheet || '', action.range || '', action.name || '', action.style || '')
            toast.success(`Tabela "${action.name || 'Tabela'}" criada!`)
        } else if (action.op === 'delete-table') {
            await DeleteTable(action.sheet || '', action.name || '')
            toast.success(`Tabela "${action.name}" removida!`)
        }
        return { success: true }
    } catch (e: unknown) {
        const errorMsg = e instanceof Error ? e.message : String(e)
        console.error("Erro na ação Excel:", errorMsg)
        return { success: false, error: errorMsg }
    }
}
