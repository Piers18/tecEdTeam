package config_test

import (
	"os"
	"testing"

	"movie-tracker/internal/config"
)

func setEnv(kv map[string]string) func() {
	for k, v := range kv {
		os.Setenv(k, v)
	}
	return func() {
		for k := range kv {
			os.Unsetenv(k)
		}
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	restore := setEnv(map[string]string{
		"TMDB_API_KEY":   "key",
		"OPENAI_API_KEY": "sk-key",
	})
	defer restore()
	os.Unsetenv("DATABASE_URL")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_MissingTMDBKey(t *testing.T) {
	restore := setEnv(map[string]string{
		"DATABASE_URL":   "postgres://x:x@x/x",
		"OPENAI_API_KEY": "sk-key",
	})
	defer restore()
	os.Unsetenv("TMDB_API_KEY")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when TMDB_API_KEY is missing")
	}
}

func TestLoad_MissingOpenAIKey(t *testing.T) {
	restore := setEnv(map[string]string{
		"DATABASE_URL": "postgres://x:x@x/x",
		"TMDB_API_KEY": "key",
	})
	defer restore()
	os.Unsetenv("OPENAI_API_KEY")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when OPENAI_API_KEY is missing")
	}
}

func TestLoad_AllRequiredVarsPresent(t *testing.T) {
	restore := setEnv(map[string]string{
		"DATABASE_URL":   "postgres://u:p@host/db",
		"TMDB_API_KEY":   "tmdb123",
		"OPENAI_API_KEY": "sk-abc",
	})
	defer restore()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "postgres://u:p@host/db" {
		t.Errorf("wrong DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.TMDBAPIKey != "tmdb123" {
		t.Errorf("wrong TMDBAPIKey: %s", cfg.TMDBAPIKey)
	}
	if cfg.OpenAIKey != "sk-abc" {
		t.Errorf("wrong OpenAIKey: %s", cfg.OpenAIKey)
	}
}

func TestLoad_PortDefaultsTo8080(t *testing.T) {
	restore := setEnv(map[string]string{
		"DATABASE_URL":   "postgres://x:x@x/x",
		"TMDB_API_KEY":   "key",
		"OPENAI_API_KEY": "sk-key",
	})
	defer restore()
	os.Unsetenv("PORT")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
}

func TestLoad_PortOverride(t *testing.T) {
	restore := setEnv(map[string]string{
		"DATABASE_URL":   "postgres://x:x@x/x",
		"TMDB_API_KEY":   "key",
		"OPENAI_API_KEY": "sk-key",
		"PORT":           "9090",
	})
	defer restore()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
}

func TestLoad_TMDBBaseURLDefault(t *testing.T) {
	restore := setEnv(map[string]string{
		"DATABASE_URL":   "postgres://x:x@x/x",
		"TMDB_API_KEY":   "key",
		"OPENAI_API_KEY": "sk-key",
	})
	defer restore()
	os.Unsetenv("TMDB_BASE_URL")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TMDBBaseURL != "https://api.themoviedb.org/3" {
		t.Errorf("wrong default TMDB base URL: %s", cfg.TMDBBaseURL)
	}
}
