package chat

import (
	"sync"
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

func TestSessionStore_GetReturnsCopy(t *testing.T) {
	s := NewSessionStore()
	s.Append("sess", openai.UserMessage("original"))

	hist := s.Get("sess")
	// Mutate the returned slice — should not affect the stored state.
	hist = append(hist, openai.AssistantMessage("injected"))

	stored := s.Get("sess")
	if len(stored) != 1 {
		t.Errorf("mutation of returned slice should not affect stored messages, got %d", len(stored))
	}
}

func TestSessionStore_ExistsReturnsFalseForUnknownSession(t *testing.T) {
	s := NewSessionStore()
	if s.Exists("nonexistent") {
		t.Fatal("Exists should return false for unknown session")
	}
}

func TestSessionStore_DeleteNonexistentIsNoOp(t *testing.T) {
	s := NewSessionStore()
	// Should not panic.
	s.Delete("ghost")
}

func TestSessionStore_ConcurrentAccess(t *testing.T) {
	s := NewSessionStore()
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Concurrent writers.
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			id := "session"
			s.Append(id, openai.UserMessage("msg"))
		}(i)
	}
	// Concurrent readers.
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			s.Get("session")
		}()
	}
	wg.Wait()
}

func TestSessionStore_AppendMultipleAtOnce(t *testing.T) {
	s := NewSessionStore()
	s.Append("sess",
		openai.UserMessage("a"),
		openai.AssistantMessage("b"),
		openai.UserMessage("c"),
	)
	if n := len(s.Get("sess")); n != 3 {
		t.Errorf("expected 3 messages, got %d", n)
	}
}
