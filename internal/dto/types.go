package dto

import "excel-ai/pkg/excel"

// UndoAction representa uma ação que pode ser desfeita
type UndoAction struct {
	Workbook string `json:"workbook"`
	Sheet    string `json:"sheet"`
	Cell     string `json:"cell"`
	OldValue string `json:"oldValue"`
	BatchID  int64  `json:"batchId"`
}

// ExcelStatus status da conexão com Excel
type ExcelStatus struct {
	Connected bool             `json:"connected"`
	Workbooks []excel.Workbook `json:"workbooks"`
	Error     string           `json:"error,omitempty"`
}

// ChatMessage mensagem do chat para frontend
type ChatMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
}

// ConversationInfo informações de conversa para frontend
type ConversationInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Preview   string `json:"preview"`
	UpdatedAt string `json:"updatedAt"`
}

// PreviewData dados para preview no frontend
type PreviewData struct {
	Headers   []string   `json:"headers"`
	Rows      [][]string `json:"rows"`
	TotalRows int        `json:"totalRows"`
	TotalCols int        `json:"totalCols"`
	Workbook  string     `json:"workbook"`
	Sheet     string     `json:"sheet"`
}

// WriteRequest requisição de escrita no Excel
type WriteRequest struct {
	Row     int         `json:"row"`
	Col     int         `json:"col"`
	Value   interface{} `json:"value"`
	Formula string      `json:"formula,omitempty"`
}

// ModelInfo informações do modelo
type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ContextLength int    `json:"contextLength"`
	PricePrompt   string `json:"pricePrompt"`
	PriceComplete string `json:"priceComplete"`
}
