package browser

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
	playwright "github.com/playwright-community/playwright-go"
)

func OnStart(
	ctx context.Context,
	selector tools.Selector,
	logger tools.Logger,
	inputArea tools.InputArea,
	ttjBackend tools.TextToJSONBackend,
	conversation chat.Conversation,
) error {
	_, err := inputArea.Read(">>> ")
	if err != nil {
		return fmt.Errorf("could not read input: %w", err)
	}

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
			Headless: playwright.Bool(false),
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

	if _, err = page.Goto("https://news.ycombinator.com"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	entries, err := page.Locator(".athing").All()
	if err != nil {
		log.Fatalf("could not get entries: %v", err)
	}

	for i, entry := range entries {
		title, err := entry.Locator("td.title > span > a").TextContent()
		if err != nil {
			log.Fatalf("could not get text content: %v", err)
		}
		fmt.Printf("%d: %s\n", i+1, title)
	}

	return nil
}
