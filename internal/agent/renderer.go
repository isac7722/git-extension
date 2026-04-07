package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/isac7722/git-extension/internal/tui"
)

type AgentRenderer struct {
	toolName    string
	toolInput   strings.Builder
	inToolBlock bool
	mu          sync.Mutex
	needNewline bool // tracks whether we need a newline before the next text block
}

func NewAgentRenderer() *AgentRenderer {
	return &AgentRenderer{}
}

func (r *AgentRenderer) RenderWelcome(model string, tools []string, version, latestVersion string) {
	header := tui.Bold.Render("── Git Agent ─────────────────────────────")

	versionLine := "  Version: " + version
	if latestVersion != "" {
		versionLine += " " + tui.Yellow.Render(fmt.Sprintf("(update available: v%s)", latestVersion))
	}

	info := tui.Dim.Render(fmt.Sprintf("  Model: %s | Tools: %s", model, strings.Join(tools, ", ")))
	hint := tui.Dim.Render("  Type your request. Press Ctrl+D to exit.")
	footer := tui.Bold.Render("──────────────────────────────────────────")

	fmt.Fprintln(os.Stderr, header)
	fmt.Fprintln(os.Stderr, versionLine)
	fmt.Fprintln(os.Stderr, info)
	fmt.Fprintln(os.Stderr, hint)
	fmt.Fprintln(os.Stderr, footer)
	fmt.Fprintln(os.Stderr)
}

func (r *AgentRenderer) RenderPrompt() {
	fmt.Fprintf(os.Stderr, "%s ", tui.Cyan.Render("❯"))
}

func (r *AgentRenderer) RenderAgentPrefix() {
	fmt.Fprintf(os.Stderr, "\n%s ", tui.Green.Render("◆"))
}

func (r *AgentRenderer) RenderText(s string) {
	if r.needNewline {
		r.needNewline = false
		fmt.Fprintln(os.Stderr)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Fprint(os.Stderr, s)
}

func (r *AgentRenderer) StartToolBlock(name string) {
	r.toolName = name
	r.toolInput.Reset()
	r.inToolBlock = true
	r.needNewline = true
}

func (r *AgentRenderer) AccumulateToolInput(partialJSON string) {
	r.toolInput.WriteString(partialJSON)
}

func (r *AgentRenderer) FinishToolBlock() {
	input := r.toolInput.String()
	display := r.formatToolCall(r.toolName, input)

	r.mu.Lock()
	fmt.Fprintf(os.Stderr, "\n  %s %s",
		tui.Dim.Render("▸"),
		tui.Dim.Render(display))
	r.mu.Unlock()

	r.inToolBlock = false
}

func (r *AgentRenderer) formatToolCall(name, inputJSON string) string {
	switch name {
	case "git":
		var g struct {
			Command string `json:"command"`
		}
		if json.Unmarshal([]byte(inputJSON), &g) == nil && g.Command != "" {
			return "$ git " + g.Command
		}
	case "gh":
		var g struct {
			Command string `json:"command"`
		}
		if json.Unmarshal([]byte(inputJSON), &g) == nil && g.Command != "" {
			return "$ gh " + g.Command
		}
	case "read_file":
		var f struct {
			Path string `json:"path"`
		}
		if json.Unmarshal([]byte(inputJSON), &f) == nil && f.Path != "" {
			return "reading " + f.Path
		}
	case "shell":
		var s struct {
			Command string `json:"command"`
		}
		if json.Unmarshal([]byte(inputJSON), &s) == nil && s.Command != "" {
			return "$ " + s.Command
		}
	}
	return name + " " + inputJSON
}

func (r *AgentRenderer) RenderError(err error) {
	fmt.Fprintf(os.Stderr, "\n%s %s\n",
		tui.Red.Render("✗"),
		tui.Red.Render(fmt.Sprintf("Error: %v", err)))
}

func (r *AgentRenderer) RenderNewline() {
	fmt.Fprintln(os.Stderr)
}
