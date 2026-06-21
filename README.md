# N1X Code

```
███╗  ██╗  ██╗ ██╗  ██╗
████╗ ██║  ══╝ ╚██╗██╔╝
██╔██╗██║  ██╗  ╚███╔╝ 
██║╚████║  ██║  ██╔██╗ 
██║ ╚███║  ██║ ██╔╝ ██╗
╚═╝  ╚══╝  ╚═╝ ╚═╝  ╚═╝
 ██████╗ ██████╗ ██████╗ ███████╗
██╔════╝██╔═══██╗██╔══██╗██╔════╝
██║     ██║   ██║██║  ██║█████╗
██║     ██║   ██║██║  ██║██╔══╝
╚██████╗╚██████╔╝██████╔╝███████╗
 ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝

       Term1na1 A1 Cod1ng Agent
```

**N1X Code** is a terminal-based AI coding agent like Claude Code / opencode. Connect your own API keys and code with AI assistance directly in your terminal.

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/N1XNAC/nix-code/main/install.sh | bash
```

Windows (PowerShell):
```powershell
powershell -c "irm https://raw.githubusercontent.com/N1XNAC/nix-code/main/install.ps1 | iex"
```

## Commands

| Command | Description |
|---------|-------------|
| `n1x` | Launch interactive TUI chat |
| `n1x run "prompt"` | Run AI agent non-interactively |
| `n1x config` | Open browser-based config UI at localhost:8080 |
| `n1x version` | Show version info |

## Features

- **Interactive TUI** — Full terminal chat interface with streaming responses
- **Think / Code modes** — Press Tab to switch between analysis and development modes
- **Tool system** — AI can read, write, edit files, run bash, search code, and manage tasks
- **Multiple providers** — Anthropic Claude, OpenAI GPT, Google Gemini
- **Web config UI** — Browser-based settings at localhost:8080 (`n1x config`)
- **Todo tracking** — AI auto-manages task lists during complex operations

## Configuration

```bash
n1x config
```

Opens `http://localhost:8080` in your browser. Add your API keys, set default models, and configure permissions.

## Building from Source

```bash
git clone https://github.com/N1XNAC/nix-code.git
cd n1x
go build -o n1x ./cmd/n1x/
```

## Architecture

```
cmd/n1x/          — CLI entry point (Cobra)
internal/
  app/            — Core application lifecycle
  config/         — JSON config (global + per-project)
  tui/            — Bubble Tea TUI (chat, panels, streaming)
  llm/
    models/       — Model definitions
    provider/     — Anthropic, OpenAI, Gemini providers
    agent/        — Agent loop with Think/Code modes
    tools/        — read, write, edit, bash, grep, glob, todowrite
  webserver/      — Browser config UI
  permission/     — Tool permission system (Allow/Ask/Deny)
  pubsub/         — Event bus
```

## License

MIT
