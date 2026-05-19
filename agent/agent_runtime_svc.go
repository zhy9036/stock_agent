package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AgentRuntime struct {
	Events chan Event

	State *AgentState

	CancelFunc context.CancelFunc

	Session *mcp.ClientSession
	Tools   []*mcp.Tool

	Messages []Message
}

func (r *AgentRuntime) runLLM(ctx context.Context) {

	streamCtx, cancel := context.WithCancel(ctx)
	r.CancelFunc = cancel

	go func() {

		resp, err := AskLLMStream(
			streamCtx,
			r.Tools,
			r.Messages,
			func(token string) {
				r.Events <- Event{
					Type: EventToken,
					Data: token,
				}
			},
		)

		if err != nil {
			fmt.Println("error: ", err.Error())
			r.Events <- Event{
				Type: EventDone,
				Data: err,
			}
			return
		}

		if resp.ToolCall != nil {
			r.Events <- Event{
				Type: EventToolCall,
				Data: resp,
			}
		}

	}()
}

func (r *AgentRuntime) eventLoop(ctx context.Context) {

	// r.runLLM(ctx)

	var buffer strings.Builder

	// done := make(chan int)
	// uiOn := false

	for evt := range r.Events {

		switch evt.Type {

		// -------------------------
		// STREAM TOKENS
		// -------------------------
		case EventToken:
			token := evt.Data.(string)
			buffer.WriteString(token)

			// only print when we have a complete valid JSON
			if json.Valid([]byte(buffer.String())) {
				// done <- 1
				var resp LLMResponse

				err := json.Unmarshal([]byte(buffer.String()), &resp)
				if err != nil {
					// optional: log malformed JSON
					return
				}
				if resp.Done {
					fmt.Println(resp.FinalAnswer)
					fmt.Print("> ")
				}
				buffer.Reset()
				// } else {
				// 	if !uiOn {
				// 		uiOn = true
				// 		go func() {
				// 			dots := []string{"", ".", "..", "..."}
				// 			i := 0
				// 			for {
				// 				select {
				// 				case <-done:
				// 					fmt.Print("\r            \r")
				// 					return
				// 				default:
				// 					fmt.Printf("\rthinking%s", dots[i%len(dots)])
				// 					time.Sleep(300 * time.Millisecond)
				// 					i++
				// 				}
				// 			}
				// 		}()
				// 	}
			}

		// -------------------------
		// USER INTERRUPT
		// -------------------------
		case EventUserInput:
			input := evt.Data.(string)

			fmt.Println("\n\n🧑 User interrupt:", input)

			// cancel current LLM
			// if r.CancelFunc != nil {
			// 	r.CancelFunc()
			// }

			// inject new user message
			r.Messages = append(r.Messages, Message{
				Role:    "user",
				Content: input,
			})

			r.runLLM(ctx)

		// -------------------------
		// TOOL CALL OR FINAL OUTPUT
		// -------------------------
		case EventToolCall:

			resp := evt.Data.(*LLMResponse)

			// FINAL ANSWER
			if resp.Done {

				fmt.Println("\n\n✅ FINAL:", resp.FinalAnswer)
				fmt.Println("\n\n✅ tool call:", resp.ToolCall)

				r.Messages = append(r.Messages, Message{
					Role:    "assistant",
					Content: resp.FinalAnswer,
				})

				r.Events <- Event{Type: EventDone}
			}

			// TOOL CALL
			if resp.ToolCall != nil {

				fmt.Println("\n\n🔧 Tool:", resp.ToolCall.Name)

				result, err := r.Session.CallTool(
					ctx,
					&mcp.CallToolParams{
						Name:      resp.ToolCall.Name,
						Arguments: resp.ToolCall.Args,
					},
				)

				if err != nil {
					fmt.Println("tool error:", err)
					return
				}

				var toolText strings.Builder
				for _, c := range result.Content {
					if t, ok := c.(*mcp.TextContent); ok {
						toolText.WriteString(t.Text)
					}
				}

				r.Messages = append(r.Messages, Message{
					Role:    "tool",
					Content: toolText.String(),
				})

				// restart reasoning
				go r.runLLM(ctx)
			}

		// -------------------------
		// DONE
		// -------------------------
		case EventDone:
			fmt.Println("\n\nAgent finished.")
			// return
		case EventShutdown:
			fmt.Println("\n Bye!")
			return
		}
	}
}
func (r *AgentRuntime) listenUserInput() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ") // 👈 add prompt here

		if !scanner.Scan() {
			return
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		if text == "exit" || text == "quit" {
			r.Events <- Event{Type: EventShutdown}
			return
		}

		r.Events <- Event{
			Type: EventUserInput,
			Data: text,
		}
	}
}
