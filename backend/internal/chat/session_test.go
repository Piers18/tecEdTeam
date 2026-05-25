package chat

import (
	"testing"

	"github.com/openai/openai-go"
)

func TestSessionStore_AppendAndGet(t *testing.T) {
	s := NewSessionStore()
	s.Append("sess1", openai.UserMessage("hello"))
	s.Append("sess1", openai.AssistantMessage("hi there"))

	hist := s.Get("sess1")
	if len(hist) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(hist))
	}
}

func TestSessionStore_Delete(t *testing.T) {
	s := NewSessionStore()
	s.Append("sess2", openai.UserMessage("msg"))
	s.Delete("sess2")

	if s.Exists("sess2") {
		t.Fatal("session should not exist after delete")
	}
	if len(s.Get("sess2")) != 0 {
		t.Fatal("get on deleted session should return empty slice")
	}
}

func TestSessionStore_IsolatesSessionsFromEachOther(t *testing.T) {
	s := NewSessionStore()
	s.Append("a", openai.UserMessage("msg-a"))
	s.Append("b", openai.UserMessage("msg-b"))

	if len(s.Get("a")) != 1 {
		t.Fatal("session a should have 1 message")
	}
	if len(s.Get("b")) != 1 {
		t.Fatal("session b should have 1 message")
	}
}
