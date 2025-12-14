package domain

// ModelPricing define os preços de um modelo
type ModelPricing struct {
	Prompt     string `json:"prompt"`     // Ex: "$0.01 / 1M tokens"
	Completion string `json:"completion"` // Ex: "$0.02 / 1M tokens"
}

// ModelInfo representa as capacidades e informações de um modelo de IA
type ModelInfo struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	ContextLength int          `json:"contextLength"`
	Pricing       ModelPricing `json:"pricing"`
}
