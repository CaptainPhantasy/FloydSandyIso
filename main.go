package main

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/cmd"
)

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			cmd.Execute()
			return
		}
	}

	// Load global config from ~/.floyd/.env.local first
	if homeDir, err := os.UserHomeDir(); err == nil {
		globalEnv := filepath.Join(homeDir, ".floyd", ".env.local")
		_ = godotenv.Load(globalEnv)
	}

	// Then load local .env.local (overrides global)
	_ = godotenv.Load(".env.local")

	if os.Getenv("FLOYD_PROFILE") != "" {
		go func() {
			slog.Info("Serving pprof at localhost:6060")
			if httpErr := http.ListenAndServe("localhost:6060", nil); httpErr != nil {
				slog.Error("Failed to pprof listen", "error", httpErr)
			}
		}()
	}

	cmd.Execute()
}
