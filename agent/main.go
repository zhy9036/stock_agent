package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var DEBUG bool = false

var CONFIG *Config

func main() {
	CONFIG, _ = LoadConfig("config.txt")
	LLMClientInit()
	ctx := context.Background()

	// Start MCP client
	mcpClient := mcp.NewClient(
		&mcp.Implementation{
			Name:    "stock-agent",
			Version: "1.0.0",
		},
		nil,
	)

	// transport := &mcp.CommandTransport{
	// 	Command: exec.Command("../mcp-server/mcp-server"),
	// }
	transport := &mcp.StreamableClientTransport{
		Endpoint: "http://localhost:8123",
	}

	session, err := mcpClient.Connect(
		ctx,
		transport,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer session.Close()

	// List available tools dynamically
	toolsResult, err := session.ListTools(
		ctx,
		&mcp.ListToolsParams{},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected MCP Tools:")
	for _, tool := range toolsResult.Tools {
		fmt.Printf("- %s : %s\n", tool.Name, tool.Description)
	}
	fmt.Println()

	state := &AgentState{}

	runInteractiveAgent(
		ctx,
		session,
		state,
		toolsResult.Tools,
	)
}
func runInteractiveAgent(
	ctx context.Context,
	session *mcp.ClientSession,
	state *AgentState,
	tools []*mcp.Tool,
) {

	fmt.Println("=== AI Stock Agent ===")
	fmt.Println("Examples:")
	fmt.Println("  buy 10 AAPL")
	fmt.Println("  sell 5 NVDA")
	fmt.Println("  should i buy tsla?")
	fmt.Println("  analyze my portfolio")
	fmt.Println("  quit or exit")
	fmt.Println()

	runtime := &AgentRuntime{
		Events:   make(chan Event, 100),
		State:    state,
		Session:  session,
		Tools:    tools,
		Messages: state.Messages,
	}

	ctx, cancel := context.WithCancel(ctx)
	runtime.CancelFunc = cancel

	// user input listener (always running)
	go runtime.listenUserInput()

	// main event loop (blocks forever)
	runtime.eventLoop(ctx)
}

func runAgentLoop_bk(
	ctx context.Context,
	session *mcp.ClientSession,
	state *AgentState,
	tools []*mcp.Tool,
	userInput string,
) {

	state.Messages = append(
		state.Messages,
		Message{
			Role:    "user",
			Content: userInput,
		},
	)

	invalidResponses := 0

	for step := 0; step < 10; step++ {

		Debug("_____________ steps %d", step)

		resp, err := AskLLM(
			ctx,
			tools,
			state.Messages,
		)

		if err != nil {

			log.Println("LLM parse error:", err)

			invalidResponses++

			if invalidResponses >= 3 {

				fmt.Println(
					"agent exceeded invalid response limit",
				)

				return
			}

			// repair prompt

			state.Messages = append(
				state.Messages,
				Message{
					Role: "user",
					Content: `
Your previous response was invalid.

You MUST respond with valid JSON only.

Valid formats:

Tool call:
{
  "done": false,
  "tool_call": {
    "name": "tool_name",
    "args": {}
  }
}

Final response:
{
  "done": true,
  "final_answer": "..."
}
`,
				},
			)

			continue
		}

		invalidResponses = 0

		// -------------------------
		// FINAL ANSWER
		// -------------------------

		if resp.Done {

			fmt.Println(resp.FinalAnswer)

			state.Messages = append(
				state.Messages,
				Message{
					Role:    "assistant",
					Content: resp.FinalAnswer,
				},
			)

			return
		}

		// -------------------------
		// TOOL CALL
		// -------------------------

		if resp.ToolCall == nil {

			log.Println(
				"invalid state: missing tool_call",
			)

			continue
		}

		Debug(
			"Tool Call: %s %+v",
			resp.ToolCall.Name,
			resp.ToolCall.Args,
		)

		result, err := session.CallTool(
			ctx,
			&mcp.CallToolParams{
				Name:      resp.ToolCall.Name,
				Arguments: resp.ToolCall.Args,
			},
		)

		if err != nil {

			log.Println("tool error:", err)

			state.Messages = append(
				state.Messages,
				Message{
					Role: "tool",
					Content: fmt.Sprintf(
						"Tool error: %v",
						err,
					),
				},
			)

			continue
		}

		var toolText strings.Builder

		for _, c := range result.Content {

			if text, ok := c.(*mcp.TextContent); ok {

				toolText.WriteString(text.Text)
				toolText.WriteString("\n")
			}
		}

		toolOutput := toolText.String()

		Debug(
			"Tool Result:\n%s",
			toolOutput,
		)

		// store assistant tool-call intent
		state.AgentMemory = map[string]any{
			"last_action": "tool_call",
			"last_tool":   resp.ToolCall.Name,
			"last_done":   false,
		}

		memoryBytes, _ := json.Marshal(state.AgentMemory)

		state.Messages = append(state.Messages, Message{
			Role: "system",
			Content: fmt.Sprintf(
				"Agent state:\n%s",
				string(memoryBytes),
			),
		})
		// state.Messages = append(
		// 	state.Messages,
		// 	Message{
		// 		Role: "assistant",
		// 		Content: fmt.Sprintf(
		// 			`Calling tool "%s"`,
		// 			resp.ToolCall.Name,
		// 		),
		// 	},
		// )

		// store tool output

		state.Messages = append(
			state.Messages,
			Message{
				Role:    "tool",
				Content: toolOutput,
			},
		)
	}

	fmt.Println("agent exceeded max loop iterations")
}
