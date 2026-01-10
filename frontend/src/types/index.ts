// Types for the Excel-AI application

export interface Message {
    role: 'user' | 'assistant' | 'system'
    content: string
    hasActions?: boolean
    hidden?: boolean  // Se true, não aparece no chat UI
}

export interface Workbook {
    name: string
    path?: string
    sheets: string[]
}

export interface ConversationItem {
    id: string
    title: string
    updatedAt: string
    preview?: string  // Preview do conteúdo para busca
}

export interface PreviewDataType {
    headers: string[]
    rows: string[][]
    totalRows: number
}

// Excel action types
export interface ExcelAction {
    op: string
    workbook?: string
    sheet?: string
    cell?: string
    value?: string
    name?: string
    range?: string
    chartType?: string
    title?: string
    sourceSheet?: string
    sourceRange?: string
    destSheet?: string
    destCell?: string
    tableName?: string
    rowFields?: string[]
    valueFields?: (string | { field: string; function: string })[]
    bold?: boolean
    italic?: boolean
    fontSize?: number
    fontColor?: string
    bgColor?: string
    oldName?: string
    newName?: string
    row?: number
    count?: number
    style?: string
    width?: number
    height?: number
    column?: number
    ascending?: boolean
    source?: string
    dest?: string
}

export interface ExcelActionResult {
    success: boolean
    error?: string
}

export interface QueryResult {
    success: boolean
    data?: any
    error?: string
}

export interface AIProcessingResult {
    displayContent: string
    actionsExecuted: number
}
