package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var client openai.Client

func LLMClientInit() {

	client = openai.NewClient(
		option.WithBaseURL(CONFIG.BaseURL),
		option.WithAPIKey(CONFIG.APIKey),
	)
}

func AskLLM(
	ctx context.Context,
	tools []*mcp.Tool,
	messages []Message,
) (*LLMResponse, error) {

	var toolDescriptions strings.Builder

	for _, t := range tools {
		toolDescriptions.WriteString(
			fmt.Sprintf(
				"- %s : %s\n",
				t.Name,
				t.Description,
			),
		)
	}

	systemPrompt := fmt.Sprintf(`
You are an AI stock trading assistant.

You have access to MCP tools.

TOOLS:
%s

RESPONSE RULES:

You MUST ALWAYS respond with valid JSON.

Schema:

{
  "done": boolean,
  "final_answer": "string",
  "tool_call": {
    "name": "tool_name",
    "args": {}
  }
}

RULES:

1. If a tool is needed:
{
  "done": false,
  "tool_call": {
    "name": "...",
    "args": {}
  }
}

2. If task is complete:
{
  "done": true,
  "final_answer": "..."
}

3. Never output plain text outside JSON.

4. Never wrap JSON in markdown.

5. Never omit both done and tool_call.

6. Portfolio data is ALWAYS stale unless retrieved from tools.

7. You MUST call get_portfolio before answering portfolio questions.

8. Do NOT simulate tool results.

9. If external data is needed, tool_call is mandatory.

10. If previous response was invalid, fix formatting and retry.
`,
		toolDescriptions.String(),
	)

	var chatMessages []openai.ChatCompletionMessageParamUnion

	chatMessages = append(
		chatMessages,
		openai.SystemMessage(systemPrompt),
	)

	for _, m := range messages {

		switch m.Role {

		case "user":
			chatMessages = append(
				chatMessages,
				openai.UserMessage(m.Content),
			)

		case "assistant":
			chatMessages = append(
				chatMessages,
				openai.AssistantMessage(m.Content),
			)

		case "system":
			chatMessages = append(
				chatMessages,
				openai.SystemMessage(m.Content),
			)

		case "tool":
			chatMessages = append(
				chatMessages,
				openai.UserMessage(
					"Tool result:\n"+m.Content,
				),
			)
		}
	}

	resp, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model:       CONFIG.ModelID,
			Messages:    chatMessages,
			Temperature: openai.Float(0.1),
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
					Type: "json_schema",
					JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
						Name: "agent_response",
						Schema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"done":         map[string]any{"type": "boolean"},
								"final_answer": map[string]any{"type": "string"},
								"tool_call": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"name": map[string]any{"type": "string"},
										"args": map[string]any{"type": "object"},
									},
									"required":             []string{"name", "args"},
									"additionalProperties": false,
								},
							},
							"required":             []string{"done"},
							"additionalProperties": false,
						},
						// Schema: map[string]any{
						// 	"oneOf": []any{
						// 		map[string]any{
						// 			"type": "object",
						// 			"properties": map[string]any{
						// 				"done": map[string]any{
						// 					"const": true,
						// 				},
						// 				"final_answer": map[string]any{
						// 					"type": "string",
						// 				},
						// 			},
						// 			"required": []string{
						// 				"done",
						// 				"final_answer",
						// 			},
						// 			"additionalProperties": false,
						// 		},

						// 		map[string]any{
						// 			"type": "object",
						// 			"properties": map[string]any{
						// 				"done": map[string]any{
						// 					"const": false,
						// 				},
						// 				"tool_call": map[string]any{
						// 					"type": "object",
						// 					"properties": map[string]any{
						// 						"name": map[string]any{
						// 							"type": "string",
						// 						},
						// 						"args": map[string]any{
						// 							"type": "object",
						// 						},
						// 					},
						// 					"required": []string{
						// 						"name",
						// 						"args",
						// 					},
						// 					"additionalProperties": false,
						// 				},
						// 			},
						// 			"required": []string{
						// 				"done",
						// 				"tool_call",
						// 			},
						// 			"additionalProperties": false,
						// 		},
						// 	},
						// },
						Strict: openai.Bool(true),
					}},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(
		resp.Choices[0].Message.Content,
	)

	Debug("LLM Response:\n%s", content)

	var parsed LLMResponse

	if err := json.Unmarshal(
		[]byte(content),
		&parsed,
	); err != nil {

		Debug("invalid json: %v", err)

		return &LLMResponse{
				Done:        false,
				FinalAnswer: "",
				ToolCall:    nil,
			}, fmt.Errorf(
				"invalid llm json: %w",
				err,
			)
	}

	// validate state

	if parsed.Done && parsed.ToolCall != nil {
		return nil, fmt.Errorf(
			"invalid response: done=true with tool_call",
		)
	}

	if !parsed.Done && parsed.ToolCall == nil {
		return nil, fmt.Errorf(
			"invalid response: neither done nor tool_call",
		)
	}

	return &parsed, nil
}
