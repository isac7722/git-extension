package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

type Agent struct {
	client   anthropic.Client
	tools    []anthropic.BetaTool
	messages []anthropic.BetaMessageParam
	model    anthropic.Model
}

func New() (*Agent, error) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	client := anthropic.NewClient()
	tools, err := buildTools()
	if err != nil {
		return nil, fmt.Errorf("failed to build tools: %w", err)
	}

	return &Agent{
		client: client,
		tools:  tools,
		model:  anthropic.ModelClaudeSonnet4_5,
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

func (a *Agent) runInteractive(ctx context.Context, systemPrompt string) error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Git Agent started. Type your request (Ctrl+D to exit).")
	fmt.Println()

	for {
		fmt.Print("You: ")
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

		a.messages = append(a.messages, anthropic.NewBetaUserMessage(
			anthropic.NewBetaTextBlock(input),
		))

		if err := a.executeAndPrint(ctx, systemPrompt); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		fmt.Println()
	}

	return nil
}

func (a *Agent) executeAndPrint(ctx context.Context, systemPrompt string) error {
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

	fmt.Print("\nAgent: ")

	var lastMessage *anthropic.BetaMessage

	for eventsIter, err := range runner.AllStreaming(ctx) {
		if err != nil {
			return fmt.Errorf("streaming error: %w", err)
		}
		for event, err := range eventsIter {
			if err != nil {
				return fmt.Errorf("event error: %w", err)
			}
			switch ev := event.AsAny().(type) {
			case anthropic.BetaRawContentBlockStartEvent:
				switch cb := ev.ContentBlock.AsAny().(type) {
				case anthropic.BetaToolUseBlock:
					fmt.Printf("\n[tool: %s] ", cb.Name)
				}
			case anthropic.BetaRawContentBlockDeltaEvent:
				switch delta := ev.Delta.AsAny().(type) {
				case anthropic.BetaTextDelta:
					fmt.Print(delta.Text)
				case anthropic.BetaInputJSONDelta:
					// Show tool input being streamed
					fmt.Print(delta.PartialJSON)
				}
			}
		}
		lastMessage = runner.LastMessage()
	}

	fmt.Println()

	// Update message history with the full conversation
	if lastMessage != nil {
		a.messages = runner.Messages()
	}

	return nil
}
