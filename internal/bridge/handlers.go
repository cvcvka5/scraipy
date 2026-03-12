package bridge

import (
	"fmt"
	"log"
	"time"

	"github.com/playwright-community/playwright-go"
)

// ANSI Color codes for prettier terminal logs
const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[34m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorRed    = "\033[31m"
)

// Global configuration for stability
const (
	settleDelay = 750 * time.Millisecond
	longSettle  = 2 * time.Second
)

// logAction prints a formatted action log to the terminal
func logAction(action, detail string) {
	log.Printf("%s[BRIDGE]%s %s%-12s%s | %s", colorBlue, colorReset, colorCyan, action, colorReset, detail)
}

// Wait for the page to be useful
func stabilize(page playwright.Page, duration time.Duration) {
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	time.Sleep(duration)
}

// HandleGetHTML returns a heavily stripped version of the page content.
// It removes scripts, styles, SVGs, and hidden elements to keep the context small.
func HandleGetHTML(page playwright.Page) (string, error) {
	logAction("GET_HTML", "Extracting interactive elements...")
	stabilize(page, settleDelay)

	// This script extracts only text and interactive element signatures.
	// It avoids backslash escaping issues by using a clean template literal.
	const script = `() => {
		const results = [];
		// We target interactive elements and YouTube-specific duration badges
		const elements = document.querySelectorAll('button, a, input, [role="button"], h1, h2, h3, #video-title, span.ytd-thumbnail-overlay-time-status-renderer');
		
		elements.forEach(el => {
			// Clean up the text content
			let text = (el.innerText || el.textContent || "").trim().replace(/\s+/g, ' ');
			
			// Skip elements that are effectively empty and not inputs
			if (!text && el.tagName !== 'INPUT') return;

			// Limit text length per element to avoid bloat
			if (text.length > 200) text = text.substring(0, 200) + "...";

			const id = el.id ? '#' + el.id : '';
			// Only take the first two classes to keep selectors readable and short
			const cls = (typeof el.className === 'string' && el.className) 
				? '.' + el.className.split(/\s+/).filter(c => c && !c.includes('style-scope')).slice(0, 2).join('.') 
				: '';

			results.push(el.tagName + id + cls + " : " + JSON.stringify(text));
		});
		
		return results.join('\n');
	}`

	cleaned, err := page.Evaluate(script)
	if err != nil {
		log.Printf("%s[ERROR]%s Script execution failed: %v", colorRed, colorReset, err)
		return "", err
	}

	res := fmt.Sprintf("%v", cleaned)

	// Final safety truncation
	if len(res) > 5000 {
		res = res[:5000] + "\n... [Truncated]"
	}

	return res, nil
}

// HandleNavigate changes the URL and waits for load.
func HandleNavigate(page playwright.Page, url string) (string, error) {
	logAction("NAVIGATE", url)
	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		log.Printf("%s[ERROR]%s Navigation failed: %v", colorRed, colorReset, err)
		return "", err
	}
	stabilize(page, longSettle)
	return "Navigated successfully. Content loaded.", nil
}

// HandleClick interacts with a selector and waits for potential changes.
func HandleClick(page playwright.Page, selector string) (string, error) {
	logAction("CLICK", selector)
	err := page.Locator(selector).Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		log.Printf("%s[ERROR]%s Click failed: %v", colorRed, colorReset, err)
		return "", err
	}

	stabilize(page, settleDelay)
	return "Successfully clicked: " + selector, nil
}

// HandleInput types into a field.
func HandleInput(page playwright.Page, selector, text string) (string, error) {
	logAction("INPUT", fmt.Sprintf("[%s] -> %s", selector, text))
	err := page.Locator(selector).Fill(text)
	if err != nil {
		log.Printf("%s[ERROR]%s Input failed: %v", colorRed, colorReset, err)
		return "", err
	}

	stabilize(page, settleDelay)
	return fmt.Sprintf("Typed text into %s", selector), nil
}

// HandleLocator returns specific element data.
func HandleLocator(page playwright.Page, selector string) (string, error) {
	logAction("LOCATOR", selector)
	stabilize(page, settleDelay)
	return page.Locator(selector).InnerHTML()
}

// HandleScroll performs a scroll action via JS execution.
func HandleScroll(page playwright.Page, x, y int) (string, error) {
	logAction("SCROLL", fmt.Sprintf("X: %d, Y: %d", x, y))
	_, err := page.Evaluate(fmt.Sprintf("window.scrollBy(%d, %d)", x, y))
	if err != nil {
		return "", err
	}
	stabilize(page, settleDelay)
	return fmt.Sprintf("Scrolled by X:%d, Y:%d", x, y), nil
}

// HandleRefresh reloads the current page.
func HandleRefresh(page playwright.Page) (string, error) {
	logAction("REFRESH", "Reloading page...")
	_, err := page.Reload(playwright.PageReloadOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return "", err
	}
	stabilize(page, settleDelay)
	return "Page refreshed.", nil
}

// HandleSleep handles waiting.
func HandleSleep(seconds int) string {
	logAction("SLEEP", fmt.Sprintf("%d seconds", seconds))
	time.Sleep(time.Duration(seconds) * time.Second)
	return fmt.Sprintf("Slept for %d seconds", seconds)
}

// HandleNewPage opens a new tab.
func HandleNewPage(ctx playwright.BrowserContext) (playwright.Page, error) {
	logAction("NEW_PAGE", "Opening new tab...")
	page, err := ctx.NewPage()
	if err == nil {
		stabilize(page, settleDelay)
	}
	return page, err
}

// HandleSwitchPage focuses on a tab index.
func HandleSwitchPage(ctx playwright.BrowserContext, index int) (playwright.Page, error) {
	logAction("SWITCH_PAGE", fmt.Sprintf("Index: %d", index))
	pages := ctx.Pages()
	if index < 0 || index >= len(pages) {
		return nil, fmt.Errorf("invalid page index: %d", index)
	}
	err := pages[index].BringToFront()
	stabilize(pages[index], settleDelay)
	return pages[index], err
}

// HandleClosePage closes current page.
func HandleClosePage(page playwright.Page) error {
	logAction("CLOSE_PAGE", "Closing active tab...")
	return page.Close()
}

// HandlePagesInfo returns count of open tabs.
func HandlePagesInfo(ctx playwright.BrowserContext) int {
	count := len(ctx.Pages())
	logAction("PAGES_INFO", fmt.Sprintf("Active tabs: %d", count))
	return count
}
