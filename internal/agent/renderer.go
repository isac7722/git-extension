package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/isac7722/git-extension/internal/tui"
	"github.com/mattn/go-isatty"
)

type AgentRenderer struct {
	isTTY          bool
	toolName       string
	toolInput      strings.Builder
	inToolBlock    bool
	spinnerStop    chan struct{}
	spinnerActive  bool // true after spinner has written at least one frame
	mu             sync.Mutex
	needNewline    bool // tracks whether we need a newline before the next text block
}

func NewAgentRenderer() *AgentRenderer {
	return &AgentRenderer{
		isTTY: isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd()),
	}
}

func (r *AgentRenderer) RenderWelcome(model string, tools []string) {
	header := tui.Bold.Render("── Git Agent ─────────────────────────────")
	info := tui.Dim.Render(fmt.Sprintf("  Model: %s | Tools: %s", model, strings.Join(tools, ", ")))
	hint := tui.Dim.Render("  Type your request. Press Ctrl+D to exit.")
	footer := tui.Bold.Render("──────────────────────────────────────────")

	fmt.Fprintln(os.Stderr, header)
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
	// Stop spinner if a tool just finished and text is now streaming
	r.StopSpinner(true)

	if r.needNewline {
		r.needNewline = false
		fmt.Fprintln(os.Stderr)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Fprint(os.Stderr, s)
}

func (r *AgentRenderer) StartToolBlock(name string) {
	// Stop spinner from a previous tool block if still running
	r.StopSpinner(true)

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
	r.startSpinner()
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

func (r *AgentRenderer) startSpinner() {
	if !r.isTTY {
		return
	}
	r.spinnerActive = false
	r.spinnerStop = make(chan struct{})
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-r.spinnerStop:
				return
			case <-ticker.C:
				r.mu.Lock()
				if r.spinnerActive {
					// Move cursor left 1 column, clear to end of line, write new frame
					fmt.Fprintf(os.Stderr, "\033[1D\033[K%s", frames[i%len(frames)])
				} else {
					// First frame: write leading space + frame
					fmt.Fprintf(os.Stderr, " %s", frames[i%len(frames)])
					r.spinnerActive = true
				}
				r.mu.Unlock()
				i++
			}
		}
	}()
}

func (r *AgentRenderer) StopSpinner(success bool) {
	if r.spinnerStop != nil {
		close(r.spinnerStop)
		r.spinnerStop = nil

		// Small delay to let spinner goroutine exit
		time.Sleep(10 * time.Millisecond)

		r.mu.Lock()
		if r.isTTY && r.spinnerActive {
			// Move cursor left 2 columns (space + frame), clear to end of line
			fmt.Fprint(os.Stderr, "\033[2D\033[K")
		}
		r.spinnerActive = false
		if success {
			fmt.Fprint(os.Stderr, " "+tui.Green.Render("✔"))
		} else {
			fmt.Fprint(os.Stderr, " "+tui.Red.Render("✗"))
		}
		r.mu.Unlock()
	}
}

func (r *AgentRenderer) RenderError(err error) {
	fmt.Fprintf(os.Stderr, "\n%s %s\n",
		tui.Red.Render("✗"),
		tui.Red.Render(fmt.Sprintf("Error: %v", err)))
}

func (r *AgentRenderer) RenderNewline() {
	fmt.Fprintln(os.Stderr)
}
