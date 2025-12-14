package chat

import (
	"encoding/json"
	"regexp"
	"strings"
)

type ToolType string

const (
	ToolTypeQuery  ToolType = "query"
	ToolTypeAction ToolType = "action"
)

type ToolCommand struct {
	Type    ToolType
	Content string      // Raw JSON content
	Payload interface{} // Parsed map
}

// ParseToolCommands extracts commands from the AI response string
// Format: :::excel-query\n{...}\n::: or :::excel-action\n{...}\n:::
func (s *Service) ParseToolCommands(text string) []ToolCommand {
	var commands []ToolCommand

	// Regex to find blocks.
	// (?s) enables dot to match newlines
	// matches :::excel-(query|action)\s*({.*?})\s*:::
	re := regexp.MustCompile(`(?s):::excel-(query|action)\s*(.*?)\s*:::`)

	matches := re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		cmdTypeStr := match[1] // "query" or "action"
		content := match[2]

		var toolType ToolType
		if cmdTypeStr == "query" {
			toolType = ToolTypeQuery
		} else {
			toolType = ToolTypeAction
		}

		// Handle multiple JSON objects in one block (newline separated)
		// Simples split by newline isn't robust for pretty-printed JSON,
		// but our prompts usually output single-line JSONs per line or one big JSON.
		// For robustness, lets try to parse the whole block first.
		// If it fails, we try to split lines.

		// Strategy: The prompt instructions say:
		// :::excel-query
		// {"type": "list-sheets"}
		// {"type": "sheet-exists"...}
		// :::
		// So we might have multiple JSONs in the content string.

		decoder := json.NewDecoder(strings.NewReader(content))
		for decoder.More() {
			var payload interface{}
			err := decoder.Decode(&payload)
			if err != nil {
				// If we can't decode, skip this part or log it?
				// For now, let's just continue to try to decode the next token
				continue
			}

			commands = append(commands, ToolCommand{
				Type:    toolType,
				Content: content, // This might be inaccurate if multiple, but useful for debugging
				Payload: payload,
			})
		}
	}

	return commands
}
