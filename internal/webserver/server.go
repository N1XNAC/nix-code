package webserver

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/n1xcode/n1x/internal/config"
	"github.com/n1xcode/n1x/internal/llm/models"
)

//go:embed config.html
var configHTML string

type Server struct {
	cfg    *config.Config
	server *http.Server
}

func New(cfg *config.Config) *Server {
	s := &Server{cfg: cfg}
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/config", s.handleGetConfig)
	mux.HandleFunc("/api/config/save", s.handleSaveConfig)
	mux.HandleFunc("/api/providers", s.handleGetProviders)
	mux.HandleFunc("/api/models", s.handleGetModels)
	mux.HandleFunc("/api/test", s.handleTestConnection)

	s.server = &http.Server{
		Handler: mux,
	}
	return s
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	fmt.Printf("\n  N1X Code Configuration\n")
	fmt.Printf("  ───────────────────────────\n")
	fmt.Printf("  Open your browser to:\n")
	fmt.Printf("  http://localhost:8080\n\n")
	fmt.Printf("  Close this page or press Ctrl+C to stop.\n\n")

	openBrowser("http://localhost:8080")

	go func() {
		<-ctx.Done()
		s.server.Shutdown(context.Background())
	}()

	return s.server.Serve(listener)
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(configHTML))
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.cfg)
}

func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var updated config.Config
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	for name, p := range updated.Providers {
		s.cfg.SetProvider(name, p)
	}

	if updated.DefaultMode != "" {
		s.cfg.DefaultMode = updated.DefaultMode
	}
	if updated.Theme != "" {
		s.cfg.Theme = updated.Theme
	}
	s.cfg.AutoCompact = updated.AutoCompact
	s.cfg.Permissions = updated.Permissions

	if err := s.cfg.Save(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleGetProviders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]string{"anthropic", "openai", "gemini", "openrouter", "groq", "nvidia-nim", "kimi", "glm", "deepseek", "mistral", "together", "fireworks", "perplexity", "xai", "ollama", "azure", "bedrock"})
}

func (s *Server) handleGetModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ModelRegistry())
}

func (s *Server) handleTestConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		Provider string `json:"provider"`
		APIKey   string `json:"apiKey"`
		Model    string `json:"model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if req.APIKey == "" {
		json.NewEncoder(w).Encode(map[string]any{"success": false, "error": "API key is required"})
		return
	}

	os.Setenv("NIX_TEST_KEY", req.APIKey)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true, "message": "API key saved. Test your connection with 'n1x run hello'"})
}
