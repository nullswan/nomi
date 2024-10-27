package browser

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"encoding/json"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
	playwright "github.com/playwright-community/playwright-go"
)

const browserPrompt = `
Extract and define a precise list of steps in JSON to achieve the user's goal using Playwright. Request user clarification if the goal is ambiguous.

Ensure all steps and interactions are accurately represented using specified action types and requirements. Do not return more than three steps at a time.

# Steps

- **Identify User Goal**: Determine the goal based on user input. If unclear, formulate a question using a question action to seek clarification.
- **Plan Steps**: Consider possible steps:
  - Use page_request to retrieve current page content if needed.
  - Use navigate to reach specific URLs.
  - Use click, fill, extract, etc., for different page interactions.
  - Use scroll, wait, screenshot, etc., for navigation and testing stability.
- **Execute Steps**: Plan up to three steps in advance, being prepared to adapt based on execution outcomes.
- **User Feedback**: Allow user input for altering or confirming the approach after each batch of steps.
- **Completion**: Return done: true when the sequence completes.

# Output Format

- **JSON**: Provide all steps in JSON format, structured with actionType for clarity.
- **Action Examples**: Ensure clear examples of each step type with comments or additional context as necessary.
- **Step Limit**: Do not include more than three steps in the response.
- **Done Property**: Include the "done" property set to true when the sequence is complete.
- **Page Request Reset**: When using a page request action or asking for user input, do not include further steps in the same response.

# Examples

**Example 1: Question for Clarification**
{
  "steps": [
    {
      "actionType": "question",
      "question": {
        "question": "Could you please provide more details about your request?"
      }
    }
  ],
  "done": false
}

**Example 2: Retrieve Page Content and Navigate**
{
  "steps": [
    {
      "actionType": "page_request",
      "page_request": {}
    },
    {
      "actionType": "navigate",
      "navigate": {
        "url": "http://example.com"
      }
    }
  ],
  "done": false
}

**Example 3: Fill a Form and Click a Button**
{
  "steps": [
    {
      "actionType": "fill",
      "fill": {
        "selector": "#search-bar",
        "fill_value": "sample search"
      }
    },
    {
      "actionType": "click",
      "click": {
        "selector": "#submit-button"
      }
    }
  ],
  "done": false
}

**Example 4: Extract Data and Scroll**
{
  "steps": [
    {
      "actionType": "extract",
      "extract": {
        "extract_selector": ".headline"
      }
    },
    {
      "actionType": "scroll",
      "scroll": {
        "direction": "down",
        "amount": 500
      }
    }
  ],
  "done": false
}

**Example 5: Wait and Take a Screenshot**
{
  "steps": [
    {
      "actionType": "wait",
      "wait": {
        "wait_selector": "#loading",
        "wait_timeout": 3000
      }
    },
    {
      "actionType": "screenshot",
      "screenshot": {
        "screenshot_path": "/screenshots/page.png"
      }
    }
  ],
  "done": true
}

**Example 6: Complete Sequence**
{
  "steps": [
    {
      "actionType": "navigate",
      "navigate": {
        "url": "http://example.com"
      }
    },
    {
      "actionType": "fill",
      "fill": {
        "selector": "#login",
        "fill_value": "user123"
      }
    },
    {
      "actionType": "click",
      "click": {
        "selector": "#submit-button"
      }
    }
  ],
  "done": true
}

**Example 7: Scroll and Wait**
{
  "steps": [
    {
      "actionType": "scroll",
      "scroll": {
        "direction": "up",
        "amount": 300
      }
    },
    {
      "actionType": "wait",
      "wait": {
        "wait_selector": ".content-loaded",
        "wait_timeout": 2000
      }
    }
  ],
  "done": false
}

# Notes

- Adjust steps based on runtime information and page content.
- Ensure JSON steps accurately reflect actions without placeholders.
- Maintain flexibility for user guidance and feedback.
- Avoid sequencing beyond attainable steps; focus on immediate actions.
- Include the "done" property set to true only when the entire sequence is complete.
- Each example demonstrates different action types to cover all possible patterns.
- Before every extract, fill or click action, you MUST include a page_request action to ensure you have the latest page content but this is the only action you will take in that chain, you will NOT include any other actions in the same response, the extract, fill or click action, will be pushed to the next response.
- Steps that includes selectors should be preceded by a page_request to ensure you know which selectors are available.
- Do not exceed three steps in any single response.
- If you don't know where to find the URL, you can go on google.com and search for the website you are looking for, then copy the URL from the search results.
`

func OnStart(
	ctx context.Context,
	selector tools.Selector,
	logger tools.Logger,
	inputArea tools.InputArea,
	ttjBackend tools.TextToJSONBackend,
	conversation chat.Conversation,
) error {
	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleSystem),
			browserPrompt,
		),
	)

	firstGoal, err := inputArea.Read(">>> ")
	if err != nil {
		return fmt.Errorf("could not read input: %w", err)
	}

	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"User Goal:\n"+firstGoal,
		),
	)

	pw, err := playwright.Run()
	if err != nil {
		if strings.Contains(err.Error(), "please install the driver") {
			err = playwright.Install()
			if err != nil {
				return fmt.Errorf("could not install playwright: %w", err)
			}
			pw, err = playwright.Run()
			if err != nil {
				return fmt.Errorf("could not start playwright: %w", err)
			}
		}
		return fmt.Errorf("could not start playwright: %w", err)
	}
	defer func() {
		if err = pw.Stop(); err != nil {
			logger.Error("could not stop playwright: " + err.Error())
		}
	}()

	browser, err := pw.Chromium.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false), // TODO(nullswan): Add to memory
		},
	)
	if err != nil {
		return fmt.Errorf("could not launch browser: %w", err)
	}
	defer func() {
		if err = browser.Close(); err != nil {
			logger.Error("could not close browser: " + err.Error())
		}
	}()

	page, err := browser.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			stepsResp, err := ttjBackend.Do(ctx, conversation)
			if err != nil {
				return fmt.Errorf("could not generate completion: %w", err)
			}

			conversation.AddMessage(
				chat.NewMessage(
					chat.Role(chat.RoleAssistant),
					stepsResp,
				),
			)

			logger.Debug("Raw Steps: " + stepsResp)

			var stepsRespData stepsResponse
			if err = json.Unmarshal([]byte(stepsResp), &stepsRespData); err != nil {
				return fmt.Errorf("could not unmarshal response: %w", err)
			}

			fmt.Printf("Steps: %+v\n", stepsRespData)

			if err = executeSteps(
				page,
				stepsRespData.Steps,
				conversation,
				logger,
				inputArea,
			); err != nil {
				return fmt.Errorf("could not execute steps: %w", err)
			}

			if stepsRespData.Done {
				logger.Info("No steps left, continue?")
				if !selector.SelectBool("No steps left, continue?", true) {
					return nil
				}

				nextInstruction, err := inputArea.Read(">>> ")
				if err != nil {
					return fmt.Errorf("could not read input: %w", err)
				}

				conversation.AddMessage(
					chat.NewMessage(
						chat.Role(chat.RoleUser),
						nextInstruction,
					),
				)
			}
		}
	}
}

func executeSteps(
	page playwright.Page,
	steps []step,
	conversation chat.Conversation,
	logger tools.Logger,
	inputArea tools.InputArea,
) error {
	for _, step := range steps {
		logger.Debug("Executing step" + string(step.Action))
		err := executeStep(page, step, conversation, logger, inputArea)
		if err != nil {
			return fmt.Errorf("could not execute step: %w", err)
		}

		logger.Debug("Waiting for load state")
		if step.Action == ActionClick || step.Action == ActionFill ||
			step.Action == ActionNavigate {
			err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State: playwright.LoadStateDomcontentloaded,
			})
			if err != nil {
				return fmt.Errorf("could not wait for load state: %w",
					err)
			}
		}
	}

	return nil
}

func executeStep(
	page playwright.Page,
	step step,
	conversation chat.Conversation,
	logger tools.Logger,
	inputArea tools.InputArea,
) error {
	switch step.Action {
	case ActionNavigate:
		if step.Navigate != nil {
			_, err := page.Goto(step.Navigate.URL)
			if err != nil {
				return fmt.Errorf("could not navigate to URL: %w", err)
			}
		}
	case ActionClick:
		if step.Click != nil {
			err := page.Click(step.Click.Selector)
			if err != nil {
				return fmt.Errorf("could not click on selector: %w", err)
			}

			return nil
		}
	case ActionFill:
		if step.Fill != nil {
			err := page.Fill(step.Fill.Selector, step.Fill.FillValue)
			if err != nil {
				return fmt.Errorf("could not fill selector: %w", err)
			}
			return nil
		}
	case ActionExtract:
		if step.Extract != nil {
			elements, err := page.QuerySelectorAll(step.Extract.Selector)
			if err != nil {
				return err
			}

			for _, el := range elements {
				text, err := el.TextContent()
				if err != nil {
					return err
				}
				log.Println("Extracted Text:", text)
			}
			return nil
		}
	case ActionScroll:
		if step.Scroll != nil {
			var scrollScript string
			switch step.Scroll.Direction {
			case ScrollUp:
				scrollScript = "window.scrollBy(0, -arguments[0]);"
			case ScrollDown:
				scrollScript = "window.scrollBy(0, arguments[0]);"
			case ScrollLeft:
				scrollScript = "window.scrollBy(-arguments[0], 0);"
			case ScrollRight:
				scrollScript = "window.scrollBy(arguments[0], 0);"
			default:
				return nil
			}
			_, err := page.Evaluate(scrollScript, step.Scroll.Amount)
			return err
		}
	case ActionWait:
		if step.Wait != nil {
			_, err := page.WaitForSelector(
				step.Wait.Selector,
				playwright.PageWaitForSelectorOptions{
					Timeout: playwright.Float(float64(step.Wait.Timeout)),
				},
			)
			if err != nil {
				return fmt.Errorf("could not wait for selector: %w", err)
			}
			return nil
		}
	case ActionScreenshot: // TODO(nullswan): add file manager
		if step.Screenshot != nil {
			path := fmt.Sprintf(
				"browser-screenshot-%s",
				time.Now().Format("2006-01-02T150405"),
			)
			_, err := page.Screenshot(playwright.PageScreenshotOptions{
				Path: &path,
			})
			if err != nil {
				return fmt.Errorf("could not take screenshot: %w", err)
			}

			return nil
		}
	case ActionQuestion:
		logger.Info(step.Question.Question)
		content, err := inputArea.Read(">>> ")
		if err != nil {
			return fmt.Errorf("could not read input: %w", err)
		}

		conversation.AddMessage(
			chat.NewMessage(
				chat.Role(chat.RoleUser),
				content,
			),
		)

		return nil
	case ActionPageRequest:
		content, err := page.Content()
		if err != nil {
			return fmt.Errorf("could not retrieve page content: %w", err)
		}

		// remove previous page content requests to optimize token usage
		for i := len(conversation.GetMessages()) - 1; i >= 0; i-- {
			message := conversation.GetMessages()[i]
			if message.Role != chat.RoleAssistant {
				break
			}
			logger.Debug("Removed message: " + message.ID.String())
			conversation.RemoveMessage(message.ID)
		}

		conversation.AddMessage(
			chat.NewMessage(
				chat.Role(chat.RoleUser),
				"Current page content\nURL:"+page.URL()+"\n"+content,
			),
		)

		return nil
	default:
		log.Printf("Unknown action: `%s`", step.Action)
	}

	return nil
}
