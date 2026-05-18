package main

type Config struct {
	BaseURL string
	APIKey  string
	ModelID string
}
type Message struct {
	Role    string
	Content string
}

type AgentState struct {
	Messages    []Message
	AgentMemory map[string]any
}

type LLMToolCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type LLMResponse struct {
	Done        bool         `json:"done"`
	FinalAnswer string       `json:"final_answer,omitempty"`
	ToolCall    *LLMToolCall `json:"tool_call,omitempty"`
}
type EventType string

const (
	EventToken     EventType = "token"
	EventUserInput EventType = "user_input"
	EventToolCall  EventType = "tool_call"
	EventToolDone  EventType = "tool_done"
	EventCancel    EventType = "cancel"
	EventDone      EventType = "done"
)

type Event struct {
	Type EventType
	Data any
}
