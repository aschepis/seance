package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aschepis/seance/internal/conversations"
)

func setupTestAPI(t *testing.T) (*Handler, string) {
	t.Helper()
	dir := t.TempDir()

	projDir := filepath.Join(dir, "projects", "-Users-test-myproject")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	convFile := filepath.Join(projDir, "test-conv-1.jsonl")
	f, err := os.Create(convFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	ts := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	lines := []string{
		`{"type":"user","uuid":"m1","sessionId":"test-conv-1","cwd":"/Users/test/myproject","gitBranch":"main","timestamp":"` + ts.Format(time.RFC3339) + `","message":{"role":"user","content":"Hello world"}}`,
		`{"type":"assistant","uuid":"m2","sessionId":"test-conv-1","cwd":"/Users/test/myproject","timestamp":"` + ts.Add(5*time.Second).Format(time.RFC3339) + `","message":{"role":"assistant","content":[{"type":"text","text":"Hi there!"}]}}`,
	}

	for _, line := range lines {
		if _, err := f.WriteString(line + "\n"); err != nil {
			t.Fatal(err)
		}
	}

	parser := conversations.NewParser(dir)
	handler := NewHandler(parser)
	return handler, dir
}

func TestListConversationsEndpoint(t *testing.T) {
	handler, _ := setupTestAPI(t)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/conversations", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var groups []conversations.ProjectGroup
	if err := json.Unmarshal(w.Body.Bytes(), &groups); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Conversations) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(groups[0].Conversations))
	}
}

func TestGetConversationEndpoint(t *testing.T) {
	handler, _ := setupTestAPI(t)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/conversations/test-conv-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var conv conversations.Conversation
	if err := json.Unmarshal(w.Body.Bytes(), &conv); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if conv.ID != "test-conv-1" {
		t.Errorf("expected ID test-conv-1, got %s", conv.ID)
	}
}

func TestGetConversationNotFoundEndpoint(t *testing.T) {
	handler, _ := setupTestAPI(t)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/conversations/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestSearchEndpoint(t *testing.T) {
	handler, _ := setupTestAPI(t)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/search?q=hello", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var results []conversations.SearchResult
	if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected search results for 'hello'")
	}
}

func TestSearchMissingQuery(t *testing.T) {
	handler, _ := setupTestAPI(t)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/search", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
