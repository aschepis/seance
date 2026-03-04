package conversations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create a fake project directory
	projDir := filepath.Join(dir, "projects", "-Users-test-myproject")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a fake conversation JSONL file
	convFile := filepath.Join(projDir, "abc-123.jsonl")
	f, err := os.Create(convFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	ts := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	messages := []rawMessage{
		{
			Type:      "user",
			UUID:      "msg-1",
			SessionID: "abc-123",
			CWD:       "/Users/test/myproject",
			GitBranch: "main",
			Timestamp: ts,
			Message:   json.RawMessage(`{"role":"user","content":"Hello, can you help me?"}`),
		},
		{
			Type:       "assistant",
			UUID:       "msg-2",
			SessionID:  "abc-123",
			CWD:        "/Users/test/myproject",
			GitBranch:  "main",
			Timestamp:  ts.Add(5 * time.Second),
			Message:    json.RawMessage(`{"role":"assistant","content":[{"type":"text","text":"Sure, I can help you with that."}]}`),
		},
		{
			Type:      "user",
			UUID:      "msg-3",
			SessionID: "abc-123",
			CWD:       "/Users/test/myproject",
			Timestamp: ts.Add(10 * time.Second),
			Message:   json.RawMessage(`{"role":"user","content":"Write some code for me"}`),
		},
		{
			Type:      "assistant",
			UUID:      "msg-4",
			SessionID: "abc-123",
			CWD:       "/Users/test/myproject",
			Timestamp: ts.Add(15 * time.Second),
			Message:   json.RawMessage(`{"role":"assistant","content":[{"type":"thinking","thinking":"Let me think about this..."},{"type":"text","text":"Here is some code."},{"type":"tool_use","id":"tool-1","name":"Write","input":{}}]}`),
		},
	}

	enc := json.NewEncoder(f)
	for _, msg := range messages {
		if err := enc.Encode(msg); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestListProjects(t *testing.T) {
	dir := setupTestDir(t)
	p := NewParser(dir)

	groups, err := p.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("expected 1 project group, got %d", len(groups))
	}

	group := groups[0]
	if len(group.Conversations) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(group.Conversations))
	}

	conv := group.Conversations[0]
	if conv.ID != "abc-123" {
		t.Errorf("expected conversation ID abc-123, got %s", conv.ID)
	}
	if conv.Summary != "Hello, can you help me?" {
		t.Errorf("unexpected summary: %s", conv.Summary)
	}
	if conv.MessageCount != 4 {
		t.Errorf("expected 4 messages, got %d", conv.MessageCount)
	}
}

func TestGetConversation(t *testing.T) {
	dir := setupTestDir(t)
	p := NewParser(dir)

	conv, err := p.GetConversation("abc-123")
	if err != nil {
		t.Fatalf("GetConversation: %v", err)
	}

	if conv.ID != "abc-123" {
		t.Errorf("expected ID abc-123, got %s", conv.ID)
	}
	if conv.CWD != "/Users/test/myproject" {
		t.Errorf("expected CWD /Users/test/myproject, got %s", conv.CWD)
	}
	if conv.GitBranch != "main" {
		t.Errorf("expected branch main, got %s", conv.GitBranch)
	}

	if len(conv.Messages) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(conv.Messages))
	}

	// First message should be user
	if conv.Messages[0].Role != "user" {
		t.Errorf("expected first message role user, got %s", conv.Messages[0].Role)
	}

	// Fourth message should have text, thinking, and tool_use blocks
	msg4 := conv.Messages[3]
	if len(msg4.Content) < 3 {
		t.Fatalf("expected at least 3 content blocks in msg4, got %d", len(msg4.Content))
	}

	// Check thinking block
	found := false
	for _, block := range msg4.Content {
		if block.Type == "thinking" {
			found = true
			if block.Text != "Let me think about this..." {
				t.Errorf("unexpected thinking text: %s", block.Text)
			}
		}
	}
	if !found {
		t.Error("expected thinking block in msg4")
	}
}

func TestSearch(t *testing.T) {
	dir := setupTestDir(t)
	p := NewParser(dir)

	results, err := p.Search("help")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected search results for 'help'")
	}

	// Should find "Hello, can you help me?" and "Sure, I can help you"
	if len(results) < 2 {
		t.Errorf("expected at least 2 results, got %d", len(results))
	}
}

func TestSearchNoResults(t *testing.T) {
	dir := setupTestDir(t)
	p := NewParser(dir)

	results, err := p.Search("nonexistentquery12345")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestGetConversationNotFound(t *testing.T) {
	dir := setupTestDir(t)
	p := NewParser(dir)

	_, err := p.GetConversation("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent conversation")
	}
}

func TestDecodeProjectPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-Users-test-myproject", "/Users/test/myproject"},
		{"-Users-adam-src-github-com-user-repo", "/Users/adam/src/github/com/user/repo"},
	}

	for _, tc := range tests {
		got := decodeProjectPath(tc.input)
		if got != tc.expected {
			t.Errorf("decodeProjectPath(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
