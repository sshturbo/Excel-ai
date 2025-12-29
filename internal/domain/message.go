package domain

import "time"

// MessageRole define os papéis possíveis em uma mensagem
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
	RoleTool      MessageRole = "tool"
)

// Message representa uma mensagem independente de infraestrutura
type Message struct {
	Role       MessageRole `json:"role"`
	Content    string      `json:"content"`
	Timestamp  time.Time   `json:"timestamp"`
	Hidden     bool        `json:"hidden,omitempty"` // Se true, não aparece no chat UI
	ToolCalls  interface{} `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}
