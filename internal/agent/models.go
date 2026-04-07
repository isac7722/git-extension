package agent

import (
	"fmt"
	"os"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/isac7722/git-extension/internal/config"
	"github.com/isac7722/git-extension/internal/tui"
)

type modelOption struct {
	label string
	model anthropic.Model
}

var availableModels = []modelOption{
	{"Opus 4.6", anthropic.ModelClaudeOpus4_6},
	{"Sonnet 4.6", anthropic.ModelClaudeSonnet4_6},
	{"Sonnet 4.5", anthropic.ModelClaudeSonnet4_5},
	{"Haiku 4.5", anthropic.ModelClaudeHaiku4_5},
	{"Sonnet 3.5 v2", anthropic.Model("claude-3-5-sonnet-20241022")},
	{"Haiku 3.5", anthropic.Model("claude-3-5-haiku-20241022")},
}

func selectModel(current anthropic.Model) (anthropic.Model, error) {
	items := make([]tui.SelectorItem, len(availableModels))
	for i, m := range availableModels {
		items[i] = tui.SelectorItem{
			Label:    m.label,
			Hint:     string(m.model),
			Selected: m.model == current,
		}
	}

	idx, err := tui.RunSelector(items, "Select model:")
	if err != nil {
		return current, err
	}
	if idx < 0 {
		return current, nil
	}

	chosen := availableModels[idx]

	// Save to config
	cfg, _ := config.LoadAgentConfig()
	if cfg == nil {
		cfg = &config.AgentConfig{}
	}
	cfg.Model = string(chosen.model)
	if err := config.SaveAgentConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to save model config: %v\n",
			tui.Red.Render("✗"), err)
	}

	fmt.Fprintf(os.Stderr, "%s Model changed to %s\n",
		tui.Green.Render("✔"),
		tui.Bold.Render(string(chosen.model)))
	return chosen.model, nil
}
