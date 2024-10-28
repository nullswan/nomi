package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
	playwright "github.com/playwright-community/playwright-go"
)

// TODO(nullswan): Ability to export as memory and load from memory
// TODO(nullswan): Amplify and rework that prompt
// TODO(nullswan): Leverage bounding boxes for better element selection and shorter prompts
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
  ],
  "done": false
}

{
	"steps": [
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
      "actionType": "page_request",
      "page_request": {}
    },
	],
	"done": false
}

{
 "steps": [
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

**Example 8: Handle Press**
{
	"steps": [
		{
			"actionType": "press",
			"press": {
				"key": "Enter"
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
- If you don't know where to find the URL, you can go on https://google.com/search?q={your search} and search for the website you are looking for, then copy the URL from the search results.
- If there is alert or popup, you MUST provide a step to handle it before proceeding with other steps. Element that can hide are preferences elements, like cookie consent, login, etc. For example, on the google search page, you will be asked to accept cookies, you must provide a step to accept the cookies before proceeding with other steps.
`

func OnStart(
	ctx context.Context,
	selector tools.Selector,
	logger tools.Logger,
	inputHandler tools.InputHandler,
	ttjBackend tools.TextToJSONBackend,
	conversation chat.Conversation,
) error {
	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleSystem),
			browserPrompt,
		),
	)

	firstGoal, err := inputHandler.Read(ctx, ">>> ")
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

	var consecutiveErrors int
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
				ctx,
				page,
				stepsRespData.Steps,
				conversation,
				logger,
				inputHandler,
			); err != nil {
				consecutiveErrors++
				conversation.AddMessage(
					chat.NewMessage(
						chat.Role(chat.RoleAssistant),
						"Error: "+err.Error(),
					),
				)

				if consecutiveErrors >= 3 {
					logger.Error("Too many consecutive errors")
					return fmt.Errorf("too many consecutive errors: %w", err)
				}
				continue
			}

			consecutiveErrors = 0

			if stepsRespData.Done {
				logger.Info("No steps left, continue?")
				if !selector.SelectBool("No steps left, continue?", true) {
					return nil
				}

				nextInstruction, err := inputHandler.Read(ctx, ">>> ")
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
	ctx context.Context,
	page playwright.Page,
	steps []step,
	conversation chat.Conversation,
	logger tools.Logger,
	inputHandler tools.InputHandler,
) error {
	for _, step := range steps {
		logger.Debug("Executing step: " + string(step.Action))
		err := executeStep(ctx, page, step, conversation, logger, inputHandler)
		if err != nil {
			return fmt.Errorf("could not execute step: %w", err)
		}

		if step.Action == ActionClick || step.Action == ActionFill ||
			step.Action == ActionNavigate {
			logger.Debug("Waiting for load state")
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

const (
	fillTimeout  = 1000
	clickTimeout = 1000
)

func executeStep(
	ctx context.Context,
	page playwright.Page,
	step step,
	conversation chat.Conversation,
	logger tools.Logger,
	inputHandler tools.InputHandler,
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
			err := page.Click(step.Click.Selector, playwright.PageClickOptions{
				Button:  playwright.MouseButtonLeft,
				Timeout: playwright.Float(float64(clickTimeout)),
			})
			if err != nil {
				return fmt.Errorf("could not click on selector: %w", err)
			}

			return nil
		}
	case ActionFill:
		if step.Fill != nil {
			err := page.Fill(
				step.Fill.Selector,
				step.Fill.FillValue,
				playwright.PageFillOptions{
					Timeout: playwright.Float(float64(fillTimeout)),
				},
			)
			if err != nil {
				return fmt.Errorf("could not fill selector: %w", err)
			}
			return nil
		}
	case ActionExtract:
		if step.Extract != nil {
			elements, err := page.QuerySelectorAll(step.Extract.Selector)
			if err != nil {
				return fmt.Errorf("could not query selector: %w", err)
			}

			for _, el := range elements {
				text, err := el.TextContent()
				if err != nil {
					return fmt.Errorf("could not extract text: %w", err)
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
			return fmt.Errorf("could not scroll: %w", err)
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
	case ActionPress:
		err := page.Keyboard().Press(step.Press.Key)
		if err != nil {
			return fmt.Errorf("could not press key: %w", err)
		}
	case ActionQuestion:
		logger.Info(step.Question.Question)
		content, err := inputHandler.Read(ctx, ">>> ")
		if err != nil {
			return fmt.Errorf("could not read input: %w", err)
		}

		conversation.AddMessage(
			chat.NewMessage(
				chat.RoleUser,
				content,
			),
		)

		return nil
	case ActionPageRequest:
		// fetch only interesting content
		elements, err := fetchContent(page)
		if err != nil {
			return fmt.Errorf("could not retrieve page content: %w", err)
		}

		// Remove previous page content requests to optimize token usage
		for _, msg := range conversation.GetMessages() {
			if msg.Role != chat.RoleAssistant {
				break
			}

			logger.Debug("Removed message: " + msg.ID.String())
			conversation.RemoveMessage(msg.ID)
		}

		content := ""
		for i, el := range elements {
			content += fmt.Sprintf("Element %d: %s\n", i+1, el.String())
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

type ElementInfo struct {
	Selector    string `json:"selector"`
	IsVisible   bool   `json:"is_visible"`
	IsClickable bool   `json:"is_clickable"`
	IsInputable bool   `json:"is_inputable"`
	Text        string `json:"text"`
}

func (e ElementInfo) String() string {
	return fmt.Sprintf(
		"Selector: %s, Visible: %t, Clickable: %t, Inputable: %t, Text: %s",
		e.Selector,
		e.IsVisible,
		e.IsClickable,
		e.IsInputable,
		e.Text,
	)
}

func fetchContent(page playwright.Page) ([]ElementInfo, error) {
	elements, err := page.QuerySelectorAll("body *")
	if err != nil {
		return nil, fmt.Errorf("could not query selector: %w", err)
	}

	processedTexts := make(map[string]ElementInfo)
	var result []ElementInfo

	for _, el := range elements {
		tagName, err := el.Evaluate("el => el.tagName.toLowerCase()")
		if err != nil || tagName == "" {
			continue
		}

		tagNameStr, ok := tagName.(string)
		if !ok {
			continue
		}
		tagNameStr = strings.ToLower(tagNameStr)

		if tagNameStr == "style" || tagNameStr == "script" {
			continue
		}

		visible, err := el.IsVisible()
		if err != nil {
			continue
		}

		// For now, we only care about visible elements
		if !visible {
			continue
		}

		textContent, err := el.TextContent()
		if err != nil {
			textContent = ""
		}

		if textContent == "" {
			inputValue, err := el.InputValue()
			if err == nil && inputValue != "" {
				textContent = inputValue
			}
		}

		cleanedText := cleanText(strings.TrimSpace(textContent))
		isClickable := false
		if tagNameStr == "a" {
			href, _ := el.GetAttribute("href")
			if href != "" {
				isClickable = true
			}
		}
		if tagNameStr == "button" || tagNameStr == "input" {
			isClickable = true
		}
		onclick, _ := el.GetAttribute("onclick")
		if onclick != "" {
			isClickable = true
		}
		role, _ := el.GetAttribute("role")
		if strings.ToLower(role) == "button" {
			isClickable = true
		}

		isInputable := false
		if tagNameStr == "input" || tagNameStr == "textarea" ||
			tagNameStr == "select" {
			isInputable = true
		}

		if cleanedText == "" && !isClickable && !isInputable {
			continue
		}

		selector, err := generateSelector(el)
		if err != nil {
			continue
		}

		if existingElem, exists := processedTexts[cleanedText]; exists {
			if (!existingElem.IsClickable || !existingElem.IsVisible) &&
				(isClickable && visible) {
				processedTexts[cleanedText] = ElementInfo{
					Selector:    selector,
					IsVisible:   visible,
					IsClickable: isClickable,
					IsInputable: isInputable,
					Text:        cleanedText,
				}
			}
			continue
		}

		elementInfo := ElementInfo{
			Selector:    selector,
			IsVisible:   visible,
			IsClickable: isClickable,
			IsInputable: isInputable,
			Text:        cleanedText,
		}

		processedTexts[cleanedText] = elementInfo
	}

	for _, elem := range processedTexts {
		result = append(result, elem)
	}

	for i, elem := range result {
		log.Printf(
			"Element %d: Selector=%s, Visible=%t, Clickable=%t, Inputable=%t, Text=%s",
			i+1,
			elem.Selector,
			elem.IsVisible,
			elem.IsClickable,
			elem.IsInputable,
			elem.Text,
		)
	}

	return result, nil
}

func cleanText(text string) string {
	cssRegex := regexp.MustCompile(`\.?[\w-]+\s*\{[^}]*\}`)
	text = cssRegex.ReplaceAllString(text, "")

	jsFunctionRegex := regexp.MustCompile(
		`\b(function|return|var|const|let|if|for|while|else|ajax|new|class)\b`,
	)
	text = jsFunctionRegex.ReplaceAllString(text, "")

	jsCodeRegex := regexp.MustCompile(`\([^)]*\)\s*\{[^}]*\}`)
	text = jsCodeRegex.ReplaceAllString(text, "")

	htmlTagRegex := regexp.MustCompile(`<.*?>`)
	text = htmlTagRegex.ReplaceAllString(text, "")

	whitespaceRegex := regexp.MustCompile(`\s+`)
	text = whitespaceRegex.ReplaceAllString(text, " ")

	// Remove embedded JavaScript snippets and encoded styles
	embeddedJSRegex := regexp.MustCompile(
		`\(\s*function\s*\(\)\s*\{[^}]*\}\s*\)\s*\(\s*\);`,
	)
	text = embeddedJSRegex.ReplaceAllString(text, "")

	// Remove data URIs and encoded SVGs
	dataURICodeRegex := regexp.MustCompile(
		`data:image\/[^;]+;base64,[A-Za-z0-9+/=]+`,
	)
	text = dataURICodeRegex.ReplaceAllString(text, "")

	// Remove any JavaScript events or attributes left
	jsAttributesRegex := regexp.MustCompile(`\[(?:on\w+|data-\w+)="[^"]*"\]`)
	text = jsAttributesRegex.ReplaceAllString(text, "")

	return strings.TrimSpace(text)
}

func generateSelector(element playwright.ElementHandle) (string, error) {
	id, err := element.GetAttribute("id")
	if err == nil && id != "" {
		return "#" + id, nil
	}

	var selectorParts []string
	currentElement := element

	for currentElement != nil {
		tagName, err := currentElement.Evaluate(
			"el => el.tagName.toLowerCase()",
		)
		if err != nil || tagName == "" {
			break
		}
		tagNameStr, ok := tagName.(string)
		if !ok {
			break
		}

		classes, err := currentElement.GetAttribute("class")
		classSelector := ""
		if err == nil && classes != "" {
			classList := strings.Fields(classes)
			var escapedClasses []string
			for _, class := range classList {
				escapedClass := regexp.QuoteMeta(class)
				escapedClasses = append(escapedClasses, "."+escapedClass)
			}
			classSelector = strings.Join(escapedClasses, "")
		}

		tagWithClass := tagNameStr + classSelector

		index, err := currentElement.Evaluate(`el => {
			const siblings = Array.from(el.parentElement.children).filter(e => e.tagName.toLowerCase() === el.tagName.toLowerCase());
			return siblings.indexOf(el) + 1;
		}`)
		if err != nil {
			break
		}

		indexInt := 1
		if idx, ok := index.(float64); ok {
			indexInt = int(idx)
		}

		tagWithClass = tagWithClass + ":nth-of-type(" + strconv.Itoa(
			indexInt,
		) + ")"

		selectorParts = append([]string{tagWithClass}, selectorParts...)

		parentHandle, err := currentElement.Evaluate(
			"el => el.parentElement",
		)
		if err != nil || parentHandle == nil {
			break
		}

		parentElement, ok := parentHandle.(playwright.ElementHandle)
		if !ok {
			break
		}

		currentElement = parentElement
	}

	fullSelector := strings.Join(selectorParts, " > ")
	return fullSelector, nil
}
