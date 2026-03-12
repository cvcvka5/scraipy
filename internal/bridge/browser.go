package bridge

import (
	"fmt"
	"math/rand"

	"github.com/playwright-community/playwright-go"
)

type BrowserAction string

const (
	ThinkAction        BrowserAction = "think"
	GetHTMLAction      BrowserAction = "get_html"
	NavigateAction     BrowserAction = "navigate"
	ClickAction        BrowserAction = "click"
	InputAction        BrowserAction = "input"
	ClearInputAction   BrowserAction = "clear_input"
	HoverAction        BrowserAction = "hover"
	PressAction        BrowserAction = "press"
	BackAction         BrowserAction = "back"
	ForwardAction      BrowserAction = "forward"
	DragAndDropAction  BrowserAction = "drag_and_drop"
	EvaluateJSAction   BrowserAction = "evaluate_js"
	LocatorAction      BrowserAction = "locator"
	ScrollAction       BrowserAction = "scroll"
	RefreshAction      BrowserAction = "refresh"
	PagesInfoAction    BrowserAction = "pages_info"
	NewPageAction      BrowserAction = "new_page"
	ClosePageAction    BrowserAction = "close_page"
	SwitchPageAction   BrowserAction = "switch_page"
	SleepAction        BrowserAction = "sleep"
	SelectOptionAction BrowserAction = "select_option"
	WaitForAction      BrowserAction = "wait_for_selector"
	GetCookiesAction   BrowserAction = "get_cookies"
	TerminateAction    BrowserAction = "terminate"
)

func (ba BrowserAction) Validate() bool {
	switch ba {
	case ThinkAction, GetHTMLAction, NavigateAction, ClickAction, InputAction,
		ClearInputAction, HoverAction, PressAction, BackAction,
		ForwardAction, DragAndDropAction, EvaluateJSAction, LocatorAction,
		ScrollAction, RefreshAction, PagesInfoAction, NewPageAction,
		ClosePageAction, SwitchPageAction, SleepAction,
		SelectOptionAction, WaitForAction, GetCookiesAction, TerminateAction:
		return true
	}
	return false
}

func InitBrowser(headless bool) (playwright.BrowserContext, playwright.Browser, *playwright.Playwright, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not start playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--no-sandbox",
			"--disable-infobars",
		},
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not launch browser: %w", err)
	}

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1280 + rand.Intn(50),
			Height: 800 + rand.Intn(50),
		},
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"),
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not create context: %w", err)
	}

	err = ApplyStealth(context)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("stealth failed: %w", err)
	}

	return context, browser, pw, nil
}

func ApplyStealth(context playwright.BrowserContext) error {
	// Full human spoofing script provided in previous turn
	stealthScript := `(() => {
    try { delete Object.getPrototypeOf(navigator).webdriver; } catch (e) {}
    Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
    // ... rest of the stealth script ...
    })();`

	return context.AddInitScript(playwright.Script{
		Content: playwright.String(stealthScript),
	})
}

func Cleanup(pw *playwright.Playwright, browser playwright.Browser) {
	_ = browser.Close()
	_ = pw.Stop()
}
