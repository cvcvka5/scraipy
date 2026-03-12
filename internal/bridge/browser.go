package bridge

import (
	"fmt"
	"log"

	"github.com/playwright-community/playwright-go"
)

type BrowserAction string

const (
	// Think action for internal reasoning
	ThinkAction BrowserAction = "think"

	// Returns the current page's full html.
	GetHTMLAction BrowserAction = "get_html"

	// Changes the current page's url.
	NavigateAction BrowserAction = "navigate"

	// Clicks on an element.
	ClickAction BrowserAction = "click"

	// Types text into an input/text-area element.
	InputAction BrowserAction = "input"

	// Returns an element's inner HTML by its CSS Query
	LocatorAction BrowserAction = "locator"

	// Scrolls the page by x, y
	ScrollAction BrowserAction = "scroll"

	// Reloads the current page
	RefreshAction BrowserAction = "refresh"

	// Pages info
	PagesInfoAction BrowserAction = "pages_info"

	// Opens a new page
	NewPageAction BrowserAction = "new_page"

	// Closes the current page
	ClosePageAction BrowserAction = "close_page"

	// Switches to page at index
	SwitchPageAction BrowserAction = "switch_page"

	// Sleeps for x seconds
	SleepAction BrowserAction = "sleep"
)

// Validate checks if the action string matches our manifest
func (ba BrowserAction) Validate() bool {
	switch ba {
	case ThinkAction, GetHTMLAction, NavigateAction, ClickAction,
		InputAction, LocatorAction, ScrollAction, RefreshAction,
		PagesInfoAction, NewPageAction, ClosePageAction, SwitchPageAction,
		SleepAction:
		return true
	}
	return false
}

// InitBrowser starts the Playwright driver and launches a browser session
func InitBrowser(headless bool) (playwright.BrowserContext, playwright.Browser, *playwright.Playwright, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not start playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not launch browser: %w", err)
	}

	// It's often helpful to set a viewport size for consistency
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1280,
			Height: 800,
		},
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not create context: %w", err)
	}

	return context, browser, pw, nil
}

// Cleanup gracefully shuts down the browser and driver
func Cleanup(pw *playwright.Playwright, browser playwright.Browser) {
	if err := browser.Close(); err != nil {
		log.Printf("Error closing browser: %v", err)
	}
	if err := pw.Stop(); err != nil {
		log.Printf("Error stopping playwright: %v", err)
	}
}
