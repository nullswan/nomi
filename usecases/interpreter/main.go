package interpreter

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/code"
	"github.com/nullswan/nomi/internal/tools"
)

const executionErrorLimit = 3

// TODO(nullswan): storage: store code prompt
func OnStart(
	ctx context.Context,
	selector tools.Selector,
	logger tools.Logger,
	textToJSON tools.TextToJSONBackend,
	inputHandler tools.InputHandler,
	conversation chat.Conversation,
) error {
	logger.Info("Starting console usecase")

	systemPrompt, err := getConsoleInstruction(
		runtime.GOOS,
	)
	if err != nil {
		return fmt.Errorf("failed to get console instruction: %w", err)
	}

	conversation.AddMessage(
		chat.NewMessage(
			chat.RoleSystem,
			systemPrompt,
		),
	)

	req, err := inputHandler.Read(ctx, ">>> ")
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	conversation.AddMessage(
		chat.NewMessage(
			chat.RoleUser,
			req,
		),
	)

	errorRetries := 0
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done: %w", ctx.Err())
		default:
			// Handle too many errors
			if errorRetries > executionErrorLimit {
				fmt.Println("Too many errors, how can I help you?")
				resp, err := inputHandler.Read(ctx, ">>> ")
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}

				conversation.AddMessage(
					chat.NewMessage(
						chat.RoleUser,
						resp,
					),
				)

				errorRetries = 0
				continue
			}

			resp, err := textToJSON.Do(ctx, conversation)
			if err != nil {
				return fmt.Errorf("error generating completion: %w", err)
			}

			conversation.AddMessage(
				chat.NewMessage(
					chat.RoleAssistant,
					resp,
				),
			)

			var consoleResp consoleResponse
			if err := json.Unmarshal([]byte(resp), &consoleResp); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}

			logger.Debug(
				"Received console response: " + string(consoleResp.Action),
			)

			switch consoleResp.Action {
			case consoleActionCode:
				// Sanitize code, add code block if necessary
				if consoleResp.Language != "" && consoleResp.Code != "" &&
					!strings.HasPrefix(consoleResp.Code, "```") {
					consoleResp.Code = "```" + consoleResp.Language + "\n" + consoleResp.Code + "\n```"
				}

				result := code.InterpretCodeBlocks(consoleResp.Code)

				if len(result) == 0 {
					logger.Info("No code blocks found")
					continue
				}

				containsError := true
				for _, r := range result {
					// TODO(nullswan): save the block

					fmt.Printf(
						"Received (%d): %s\n%s\n",
						r.ExitCode,
						r.Stdout,
						r.Stderr,
					)

					if r.ExitCode == 0 {
						containsError = false
						break
					}
				}

				formattedResult := code.FormatExecutionResultForLLM(result)
				conversation.AddMessage(
					chat.NewMessage(
						chat.RoleAssistant,
						formattedResult,
					),
				)

				if containsError {
					logger.Info("Code execution failed")
					errorRetries++
					continue
				} else {
					logger.Info("Code execution succeeded")
					errorRetries = 0
					if !selector.SelectBool(
						"Do you want to continue?",
						false,
					) {
						return nil
					}

					req, err := inputHandler.Read(ctx, consoleResp.Question)
					if err != nil {
						return fmt.Errorf("failed to read input: %w", err)
					}

					conversation.AddMessage(
						chat.NewMessage(
							chat.RoleUser,
							req,
						),
					)
				}
			case consoleActionAsk:
				fmt.Println(consoleResp.Question)
				req, err := inputHandler.Read(
					ctx,
					">>> ",
				)
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}

				conversation.AddMessage(
					chat.NewMessage(
						chat.RoleUser,
						req,
					),
				)

				// ask memory
				continue
			}
		}
	}
}

type consoleResponse struct {
	Action   consoleAction `json:"action"`
	Question string        `json:"question"`
	Language string        `json:"language"`
	Code     string        `json:"code"`
}

type consoleAction string

const (
	consoleActionCode consoleAction = "code"
	consoleActionAsk  consoleAction = "ask"
)
