package conversations

import "time"

// Project represents a grouping of conversations by working directory.
type Project struct {
	Path          string          `json:"path"`
	Name          string          `json:"name"`
	Conversations []*Conversation `json:"conversations"`
}

// Conversation represents a single Claude Code session.
type Conversation struct {
	ID        string    `json:"id"`
	Project   string    `json:"project"`
	CWD       string    `json:"cwd"`
	GitBranch string    `json:"gitBranch,omitempty"`
	StartedAt time.Time `json:"startedAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Messages  []Message `json:"messages,omitempty"`
	Summary   string    `json:"summary,omitempty"`
}

// Message represents a single user prompt or assistant response chunk.
type Message struct {
	UUID      string         `json:"uuid"`
	Role      string         `json:"role"` // "user" or "assistant"
	Content   []ContentBlock `json:"content"`
	Timestamp time.Time      `json:"timestamp"`
	ToolUses  []ToolUse      `json:"toolUses,omitempty"`
	IsSubAgent bool          `json:"isSubAgent"`
}

// ContentBlock represents a piece of message content.
type ContentBlock struct {
	Type string `json:"type"` // "text", "thinking", "tool_use", "tool_result"
	Text string `json:"text,omitempty"`
	// For tool_use blocks
	ToolName string `json:"toolName,omitempty"`
	ToolID   string `json:"toolId,omitempty"`
	// For tool_result blocks
	ToolUseID string `json:"toolUseId,omitempty"`
	Output    string `json:"output,omitempty"`
}

// ToolUse represents a tool invocation within a message.
type ToolUse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Input  string `json:"input,omitempty"`
	Output string `json:"output,omitempty"`
}

// ConversationSummary is a lightweight version without full messages.
type ConversationSummary struct {
	ID           string    `json:"id"`
	Project      string    `json:"project"`
	CWD          string    `json:"cwd"`
	GitBranch    string    `json:"gitBranch,omitempty"`
	StartedAt    time.Time `json:"startedAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	MessageCount int       `json:"messageCount"`
	Summary      string    `json:"summary,omitempty"`
}

// SearchResult represents a search match within a conversation.
type SearchResult struct {
	ConversationID string    `json:"conversationId"`
	Project        string    `json:"project"`
	MessageUUID    string    `json:"messageUuid"`
	Role           string    `json:"role"`
	MatchText      string    `json:"matchText"`
	Timestamp      time.Time `json:"timestamp"`
}

// ProjectGroup groups conversation summaries by project path.
type ProjectGroup struct {
	Path          string                `json:"path"`
	Name          string                `json:"name"`
	Conversations []ConversationSummary `json:"conversations"`
}
