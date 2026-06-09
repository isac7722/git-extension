package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/isac7722/git-extension/internal/config"
	"github.com/isac7722/git-extension/internal/update"
	"github.com/isac7722/git-extension/internal/version"
)

type Agent struct {
	client   anthropic.Client
	tools    []anthropic.BetaTool
	messages []anthropic.BetaMessageParam
	model    anthropic.Model
	renderer *AgentRenderer
}

// ErrNoAPIKey is returned when no Anthropic API key is found.
type ErrNoAPIKey struct{}

func (e *ErrNoAPIKey) Error() string {
	return "no Anthropic API key found"
}

func New() (*Agent, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	model := anthropic.ModelClaudeSonnet4_6

	cfg, err := config.LoadAgentConfig()
	if err == nil {
		if cfg.AnthropicAPIKey != "" && apiKey == "" {
			apiKey = cfg.AnthropicAPIKey
		}
		if cfg.Model != "" {
			model = anthropic.Model(cfg.Model)
		}
	}

	if apiKey == "" {
		return nil, &ErrNoAPIKey{}
	}

	// Set env var so anthropic-sdk-go picks it up
	if err := os.Setenv("ANTHROPIC_API_KEY", apiKey); err != nil {
		return nil, err
	}

	client := anthropic.NewClient()
	tools, err := buildTools()
	if err != nil {
		return nil, fmt.Errorf("failed to build tools: %w", err)
	}

	return &Agent{
		client:   client,
		tools:    tools,
		model:    model,
		renderer: NewAgentRenderer(),
	}, nil
}

func (a *Agent) Run(ctx context.Context, oneshot string) error {
	systemPrompt := buildSystemPrompt()

	if oneshot != "" {
		return a.runOnce(ctx, systemPrompt, oneshot)
	}
	return a.runInteractive(ctx, systemPrompt)
}

func (a *Agent) runOnce(ctx context.Context, systemPrompt, prompt string) error {
	a.messages = append(a.messages, anthropic.NewBetaUserMessage(
		anthropic.NewBetaTextBlock(prompt),
	))

	return a.executeAndPrint(ctx, systemPrompt)
}

func (a *Agent) toolNames() []string {
	names := make([]string, len(a.tools))
	for i, t := range a.tools {
		names[i] = t.Name()
	}
	return names
}

func (a *Agent) runInteractive(ctx context.Context, systemPrompt string) error {
	ver := version.Version
	latestVersion := ""
	if latest, hasUpdate := update.CheckCached(ver); hasUpdate {
		latestVersion = latest
	}
	a.renderer.RenderWelcome(string(a.model), a.toolNames(), ver, latestVersion)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		a.renderer.RenderPrompt()
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}
		if input == "/model" {
			model, err := selectModel(a.model)
			if err != nil {
				a.renderer.RenderError(err)
			} else {
				a.model = model
			}
			continue
		}

		a.messages = append(a.messages, anthropic.NewBetaUserMessage(
			anthropic.NewBetaTextBlock(input),
		))

		if err := a.executeAndPrint(ctx, systemPrompt); err != nil {
			a.renderer.RenderError(err)
			continue
		}
		a.renderer.RenderNewline()
	}

	return nil
}

func (a *Agent) executeAndPrint(ctx context.Context, systemPrompt string) error {
	r := a.renderer

	runner := a.client.Beta.Messages.NewToolRunnerStreaming(
		a.tools,
		anthropic.BetaToolRunnerParams{
			BetaMessageNewParams: anthropic.BetaMessageNewParams{
				Model:     a.model,
				MaxTokens: 8192,
				System: []anthropic.BetaTextBlockParam{
					{Text: systemPrompt},
				},
				Messages: a.messages,
			},
			MaxIterations: 20,
		},
	)

	r.RenderAgentPrefix()

	var lastMessage *anthropic.BetaMessage

	for eventsIter, err := range runner.AllStreaming(ctx) {
		if err != nil {
			r.RenderError(err)
			return fmt.Errorf("streaming error: %w", err)
		}

		for event, err := range eventsIter {
			if err != nil {
				r.RenderError(err)
				return fmt.Errorf("event error: %w", err)
			}
			switch ev := event.AsAny().(type) {
			case anthropic.BetaRawContentBlockStartEvent:
				switch cb := ev.ContentBlock.AsAny().(type) {
				case anthropic.BetaToolUseBlock:
					r.StartToolBlock(cb.Name)
				}
			case anthropic.BetaRawContentBlockDeltaEvent:
				switch delta := ev.Delta.AsAny().(type) {
				case anthropic.BetaTextDelta:
					r.RenderText(delta.Text)
				case anthropic.BetaInputJSONDelta:
					r.AccumulateToolInput(delta.PartialJSON)
				}
			case anthropic.BetaRawContentBlockStopEvent:
				if r.inToolBlock {
					r.FinishToolBlock()
				}
			}
		}
		lastMessage = runner.LastMessage()
	}

	r.RenderNewline()

	// Update message history with the full conversation
	if lastMessage != nil {
		a.messages = runner.Messages()
	}

	return nil
}
