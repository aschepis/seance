package conversations

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// sessionsIndex is the structure of a sessions-index.json file written by newer Claude Code versions.
type sessionsIndex struct {
	Version int                 `json:"version"`
	Entries []sessionsIndexEntry `json:"entries"`
}

type sessionsIndexEntry struct {
	SessionID    string    `json:"sessionId"`
	FullPath     string    `json:"fullPath"`
	FirstPrompt  string    `json:"firstPrompt"`
	Summary      string    `json:"summary"`
	MessageCount int       `json:"messageCount"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
	GitBranch    string    `json:"gitBranch"`
	ProjectPath  string    `json:"projectPath"`
	IsSidechain  bool      `json:"isSidechain"`
}

// rawMessage is the JSON structure of each line in a JSONL conversation file.
type rawMessage struct {
	Type      string          `json:"type"`
	UUID      string          `json:"uuid"`
	ParentUUID *string        `json:"parentUuid"`
	SessionID string          `json:"sessionId"`
	CWD       string          `json:"cwd"`
	GitBranch string          `json:"gitBranch"`
	Timestamp time.Time       `json:"timestamp"`
	IsSidechain bool          `json:"isSidechain"`
	Message   json.RawMessage `json:"message"`
}

type rawChatMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type rawContentBlock struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Thinking  string `json:"thinking,omitempty"`
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	ToolUseID string `json:"tool_use_id,omitempty"`
	// Content can be a string or array for tool_result
	Content json.RawMessage `json:"content,omitempty"`
}

// Parser reads Claude Code conversation files from the filesystem.
type Parser struct {
	claudeDir string
}

// NewParser creates a parser that reads from the given Claude config directory.
func NewParser(claudeDir string) *Parser {
	return &Parser{claudeDir: claudeDir}
}

// DefaultClaudeDir returns the default Claude config directory path.
func DefaultClaudeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

// ListProjects discovers all projects with conversations.
func (p *Parser) ListProjects() ([]ProjectGroup, error) {
	projectsDir := filepath.Join(p.claudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("reading projects dir: %w", err)
	}

	groups := make(map[string]*ProjectGroup)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		projDir := filepath.Join(projectsDir, dirName)

		// Prefer the real path from sessions-index.json over the lossy decoded dir name.
		projectPath := decodeProjectPath(dirName)
		projectName := filepath.Base(projectPath)
		indexPath := filepath.Join(projDir, "sessions-index.json")
		idx, indexErr := readSessionsIndex(indexPath)
		if indexErr == nil {
			for _, e := range idx.Entries {
				if !e.IsSidechain && e.ProjectPath != "" {
					projectPath = e.ProjectPath
					projectName = filepath.Base(projectPath)
					break
				}
			}
		}

		// Track which session IDs we've already loaded from flat JSONL files.
		seenIDs := make(map[string]bool)

		files, err := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
		if err == nil {
			for _, file := range files {
				summary, err := p.parseConversationSummary(file, projectPath)
				if err != nil {
					continue
				}
				seenIDs[summary.ID] = true

				group, ok := groups[projectPath]
				if !ok {
					group = &ProjectGroup{
						Path: projectPath,
						Name: projectName,
					}
					groups[projectPath] = group
				}
				group.Conversations = append(group.Conversations, *summary)
			}
		}

		// Newer Claude Code versions write a sessions-index.json instead of flat JSONL files.
		// Read it to pick up any sessions not already found above.
		if indexErr == nil {
			for _, e := range idx.Entries {
				if e.IsSidechain || seenIDs[e.SessionID] {
					continue
				}
				summary := summaryFromIndexEntry(e, projectPath)

				group, ok := groups[projectPath]
				if !ok {
					group = &ProjectGroup{
						Path: projectPath,
						Name: projectName,
					}
					groups[projectPath] = group
				}
				group.Conversations = append(group.Conversations, summary)
			}
		}
	}

	// Convert map to sorted slice
	result := make([]ProjectGroup, 0, len(groups))
	for _, group := range groups {
		// Sort conversations by date, most recent first
		sort.Slice(group.Conversations, func(i, j int) bool {
			return group.Conversations[i].UpdatedAt.After(group.Conversations[j].UpdatedAt)
		})
		result = append(result, *group)
	}

	// Sort groups by most recent conversation
	sort.Slice(result, func(i, j int) bool {
		iTime := time.Time{}
		jTime := time.Time{}
		if len(result[i].Conversations) > 0 {
			iTime = result[i].Conversations[0].UpdatedAt
		}
		if len(result[j].Conversations) > 0 {
			jTime = result[j].Conversations[0].UpdatedAt
		}
		return iTime.After(jTime)
	})

	return result, nil
}

// GetConversation reads and parses a full conversation by ID.
func (p *Parser) GetConversation(id string) (*Conversation, error) {
	file, err := p.findConversationFile(id)
	if err == nil {
		return p.parseConversation(file)
	}

	// Fall back to sessions-index.json for conversations without a flat JSONL file.
	entry, indexErr := p.findIndexEntry(id)
	if indexErr != nil {
		return nil, fmt.Errorf("conversation not found: %s", id)
	}

	// If the fullPath from the index actually exists on disk, parse it.
	if _, statErr := os.Stat(entry.FullPath); statErr == nil {
		return p.parseConversation(entry.FullPath)
	}

	// Construct a minimal conversation from the index metadata since the JSONL is gone.
	conv := &Conversation{
		ID:        entry.SessionID,
		Project:   entry.ProjectPath,
		CWD:       entry.ProjectPath,
		GitBranch: entry.GitBranch,
		StartedAt: entry.Created,
		UpdatedAt: entry.Modified,
		Summary:   entry.FirstPrompt,
	}
	if conv.Summary == "" {
		conv.Summary = entry.Summary
	}
	return conv, nil
}

// Search searches across all conversations for the given query.
func (p *Parser) Search(query string) ([]SearchResult, error) {
	query = strings.ToLower(query)
	var results []SearchResult

	projectsDir := filepath.Join(p.claudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("reading projects dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		projectPath := decodeProjectPath(dirName)
		projDir := filepath.Join(projectsDir, dirName)

		files, err := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
		if err != nil {
			continue
		}

		for _, file := range files {
			fileResults, err := p.searchInFile(file, projectPath, query)
			if err != nil {
				continue
			}
			results = append(results, fileResults...)
		}
	}

	// Sort by timestamp, most recent first
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	// Limit results
	if len(results) > 100 {
		results = results[:100]
	}

	return results, nil
}

func (p *Parser) findConversationFile(id string) (string, error) {
	projectsDir := filepath.Join(p.claudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return "", fmt.Errorf("reading projects dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(projectsDir, entry.Name(), id+".jsonl")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("conversation not found: %s", id)
}

func (p *Parser) parseConversationSummary(file, projectPath string) (*ConversationSummary, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sessionID := strings.TrimSuffix(filepath.Base(file), ".jsonl")

	summary := &ConversationSummary{
		ID:      sessionID,
		Project: projectPath,
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB max line size

	messageCount := 0
	var firstUserMessage string
	var latestTimestamp time.Time

	for scanner.Scan() {
		var raw rawMessage
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		if raw.Type != "user" && raw.Type != "assistant" {
			continue
		}

		if raw.CWD != "" && summary.CWD == "" {
			summary.CWD = raw.CWD
		}
		if raw.GitBranch != "" && summary.GitBranch == "" {
			summary.GitBranch = raw.GitBranch
		}

		if !raw.Timestamp.IsZero() {
			if summary.StartedAt.IsZero() {
				summary.StartedAt = raw.Timestamp
			}
			if raw.Timestamp.After(latestTimestamp) {
				latestTimestamp = raw.Timestamp
			}
		}

		if raw.Type == "user" || raw.Type == "assistant" {
			messageCount++
		}

		// Extract first user message as summary
		if raw.Type == "user" && firstUserMessage == "" {
			var chatMsg rawChatMessage
			if err := json.Unmarshal(raw.Message, &chatMsg); err == nil {
				firstUserMessage = extractTextFromContent(chatMsg.Content)
			}
		}
	}

	summary.MessageCount = messageCount
	summary.UpdatedAt = latestTimestamp
	if firstUserMessage != "" {
		if len(firstUserMessage) > 200 {
			firstUserMessage = firstUserMessage[:200] + "..."
		}
		summary.Summary = firstUserMessage
	}

	return summary, nil
}

func (p *Parser) parseConversation(file string) (*Conversation, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sessionID := strings.TrimSuffix(filepath.Base(file), ".jsonl")

	conv := &Conversation{
		ID: sessionID,
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	// Track messages by UUID to deduplicate (assistant messages come in chunks)
	messageMap := make(map[string]*Message)
	var messageOrder []string

	// Track tool results to attach to the right message
	toolResults := make(map[string]string) // toolUseID -> output

	for scanner.Scan() {
		var raw rawMessage
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		if raw.CWD != "" && conv.CWD == "" {
			conv.CWD = raw.CWD
			conv.Project = raw.CWD
		}
		if raw.GitBranch != "" && conv.GitBranch == "" {
			conv.GitBranch = raw.GitBranch
		}

		if raw.Type != "user" && raw.Type != "assistant" {
			continue
		}

		if !raw.Timestamp.IsZero() {
			if conv.StartedAt.IsZero() {
				conv.StartedAt = raw.Timestamp
			}
			conv.UpdatedAt = raw.Timestamp
		}

		var chatMsg rawChatMessage
		if err := json.Unmarshal(raw.Message, &chatMsg); err != nil {
			continue
		}

		// Parse content blocks
		blocks := parseContentBlocks(chatMsg.Content)
		if len(blocks) == 0 {
			continue
		}

		// Check if this is a tool_result message (part of user role)
		if chatMsg.Role == "user" {
			for _, block := range blocks {
				if block.Type == "tool_result" && block.ToolUseID != "" {
					toolResults[block.ToolUseID] = block.Output
				}
			}
		}

		// For user messages, skip pure tool_result messages (no text content)
		if chatMsg.Role == "user" {
			hasText := false
			for _, block := range blocks {
				if block.Type == "text" && block.Text != "" {
					hasText = true
					break
				}
			}
			if !hasText {
				continue
			}
		}

		// Deduplicate assistant chunks by UUID — keep accumulating content
		if raw.UUID != "" {
			if existing, ok := messageMap[raw.UUID]; ok {
				// Merge new blocks
				existing.Content = mergeContentBlocks(existing.Content, blocks)
				if !raw.Timestamp.IsZero() {
					existing.Timestamp = raw.Timestamp
				}
				continue
			}
		}

		msg := Message{
			UUID:       raw.UUID,
			Role:       chatMsg.Role,
			Content:    blocks,
			Timestamp:  raw.Timestamp,
			IsSubAgent: raw.IsSidechain,
		}

		if raw.UUID != "" {
			messageMap[raw.UUID] = &msg
			messageOrder = append(messageOrder, raw.UUID)
		}
	}

	// Attach tool results to assistant messages
	for _, msg := range messageMap {
		if msg.Role != "assistant" {
			continue
		}
		for i, block := range msg.Content {
			if block.Type == "tool_use" && block.ToolID != "" {
				if output, ok := toolResults[block.ToolID]; ok {
					msg.Content[i].Output = output
					msg.ToolUses = append(msg.ToolUses, ToolUse{
						ID:     block.ToolID,
						Name:   block.ToolName,
						Output: output,
					})
				}
			}
		}
	}

	// Build ordered message list
	for _, uuid := range messageOrder {
		if msg, ok := messageMap[uuid]; ok {
			conv.Messages = append(conv.Messages, *msg)
		}
	}

	// Set summary from first user message
	for _, msg := range conv.Messages {
		if msg.Role == "user" {
			for _, block := range msg.Content {
				if block.Type == "text" && block.Text != "" {
					s := block.Text
					if len(s) > 200 {
						s = s[:200] + "..."
					}
					conv.Summary = s
					break
				}
			}
			break
		}
	}

	return conv, nil
}

func (p *Parser) searchInFile(file, projectPath, query string) ([]SearchResult, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sessionID := strings.TrimSuffix(filepath.Base(file), ".jsonl")
	var results []SearchResult

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		var raw rawMessage
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		if raw.Type != "user" && raw.Type != "assistant" {
			continue
		}

		var chatMsg rawChatMessage
		if err := json.Unmarshal(raw.Message, &chatMsg); err != nil {
			continue
		}

		text := extractTextFromContent(chatMsg.Content)
		if text == "" {
			continue
		}

		lowerText := strings.ToLower(text)
		if !strings.Contains(lowerText, query) {
			continue
		}

		// Extract a snippet around the match
		idx := strings.Index(lowerText, query)
		start := max(idx-50, 0)
		end := min(idx+len(query)+50, len(text))
		snippet := text[start:end]

		results = append(results, SearchResult{
			ConversationID: sessionID,
			Project:        projectPath,
			MessageUUID:    raw.UUID,
			Role:           chatMsg.Role,
			MatchText:      snippet,
			Timestamp:      raw.Timestamp,
		})
	}

	return results, nil
}

func (p *Parser) findIndexEntry(id string) (*sessionsIndexEntry, error) {
	projectsDir := filepath.Join(p.claudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		indexPath := filepath.Join(projectsDir, entry.Name(), "sessions-index.json")
		idx, err := readSessionsIndex(indexPath)
		if err != nil {
			continue
		}
		for i := range idx.Entries {
			if idx.Entries[i].SessionID == id {
				return &idx.Entries[i], nil
			}
		}
	}

	return nil, fmt.Errorf("not found in any sessions index: %s", id)
}

func readSessionsIndex(path string) (*sessionsIndex, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var idx sessionsIndex
	if err := json.NewDecoder(f).Decode(&idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

func summaryFromIndexEntry(e sessionsIndexEntry, projectPath string) ConversationSummary {
	text := e.FirstPrompt
	if text == "" {
		text = e.Summary
	}
	if len(text) > 200 {
		text = text[:200] + "..."
	}

	cwd := e.ProjectPath
	if cwd == "" {
		cwd = projectPath
	}

	// If the JSONL file doesn't exist on disk, the conversation has no accessible
	// messages regardless of what the index says, so report 0.
	messageCount := e.MessageCount
	if _, err := os.Stat(e.FullPath); err != nil {
		messageCount = 0
	}

	return ConversationSummary{
		ID:           e.SessionID,
		Project:      projectPath,
		CWD:          cwd,
		GitBranch:    e.GitBranch,
		StartedAt:    e.Created,
		UpdatedAt:    e.Modified,
		MessageCount: messageCount,
		Summary:      text,
	}
}

// decodeProjectPath converts a Claude projects directory name back to a path.
// The encoding replaces path separators with hyphens. The name starts with "-"
// which represents the root "/".
func decodeProjectPath(encoded string) string {
	// The format is like: -Users-alice-src-github-com-alice-myproject
	// Convert back to: /Users/alice/src/github.com/alice/myproject
	return strings.ReplaceAll(encoded, "-", "/")
}

func extractTextFromContent(content json.RawMessage) string {
	// Try as string first
	var s string
	if err := json.Unmarshal(content, &s); err == nil {
		return s
	}

	// Try as array of content blocks
	var blocks []rawContentBlock
	if err := json.Unmarshal(content, &blocks); err == nil {
		var parts []string
		for _, block := range blocks {
			switch block.Type {
			case "text":
				if block.Text != "" {
					parts = append(parts, block.Text)
				}
			case "thinking":
				if block.Thinking != "" {
					parts = append(parts, block.Thinking)
				}
			}
		}
		return strings.Join(parts, "\n")
	}

	return ""
}

func parseContentBlocks(content json.RawMessage) []ContentBlock {
	// Try as string first
	var s string
	if err := json.Unmarshal(content, &s); err == nil {
		if s != "" {
			return []ContentBlock{{Type: "text", Text: s}}
		}
		return nil
	}

	// Try as array of content blocks
	var rawBlocks []rawContentBlock
	if err := json.Unmarshal(content, &rawBlocks); err != nil {
		return nil
	}

	var blocks []ContentBlock
	for _, raw := range rawBlocks {
		switch raw.Type {
		case "text":
			if raw.Text != "" {
				blocks = append(blocks, ContentBlock{Type: "text", Text: raw.Text})
			}
		case "thinking":
			if raw.Thinking != "" {
				blocks = append(blocks, ContentBlock{Type: "thinking", Text: raw.Thinking})
			}
		case "tool_use":
			blocks = append(blocks, ContentBlock{
				Type:     "tool_use",
				ToolName: raw.Name,
				ToolID:   raw.ID,
			})
		case "tool_result":
			output := ""
			if raw.Content != nil {
				// tool_result content can be string or structured
				var s string
				if err := json.Unmarshal(raw.Content, &s); err == nil {
					output = s
				} else {
					output = string(raw.Content)
				}
			}
			blocks = append(blocks, ContentBlock{
				Type:      "tool_result",
				ToolUseID: raw.ToolUseID,
				Output:    output,
			})
		}
	}

	return blocks
}

func mergeContentBlocks(existing, newBlocks []ContentBlock) []ContentBlock {
	// For assistant streaming, later chunks may have additional content blocks
	existingTypes := make(map[string]bool)
	for _, b := range existing {
		key := b.Type + ":" + b.ToolID + ":" + b.Text
		existingTypes[key] = true
	}

	for _, b := range newBlocks {
		key := b.Type + ":" + b.ToolID + ":" + b.Text
		if !existingTypes[key] {
			existing = append(existing, b)
			existingTypes[key] = true
		}
	}

	return existing
}
