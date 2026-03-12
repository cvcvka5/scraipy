package bridge

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/playwright-community/playwright-go"
)

func HandleNavigate(page playwright.Page, url string) (string, error) {
	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Navigated to %s", url), nil
}

func sanitizeForJSON(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, str)
}

func HandleGetHTML(page playwright.Page) (string, error) {
	result, err := page.Evaluate(`() => {
		const result = {
			url: window.location.href,
			title: document.title,
			interactiveElements: [],
			headings: [],
			frames: [],
			bodyPreview: "",
		};

		document.querySelectorAll('h1, h2, h3, h4').forEach(h => {
			const t = h.innerText.trim();
			if (t) result.headings.push(t);
		});

		document.querySelectorAll('iframe').forEach(f => {
			result.frames.push({ id: f.id, src: f.src });
		});

		document.querySelectorAll('input, button, a, [role="button"], select, textarea').forEach(el => {
			const rect = el.getBoundingClientRect();
			if (rect.width === 0 && rect.height === 0) return;
			const text = (el.innerText || el.value || el.placeholder || "").trim().substring(0, 80);
			const entry = {
				tag:       el.tagName.toLowerCase(),
				text:      text,
				id:        el.id         || undefined,
				name:      el.getAttribute('name') || undefined,
				type:      el.getAttribute('type') || undefined,
				ariaLabel: el.getAttribute('aria-label') || undefined,
				href:      el.tagName === 'A' ? el.getAttribute('href') : undefined,
				class:     el.className  || undefined,
			};
			// Drop undefined keys to keep payload lean.
			Object.keys(entry).forEach(k => entry[k] === undefined && delete entry[k]);
			result.interactiveElements.push(entry);
		});

		// Richer body: strip script/style but keep readable text and structure.
		const clone = document.body.cloneNode(true);
		clone.querySelectorAll('script, style, noscript, svg').forEach(el => el.remove());
		result.bodyPreview = clone.innerText.replace(/\s+/g, ' ').trim().substring(0, 3000);

		return JSON.stringify(result);
	}`)
	if err != nil {
		return "", err
	}

	raw, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("page.Evaluate returned unexpected type %T", result)
	}

	// Only sanitize unprintable unicode — do NOT strip JSON structure.
	return sanitizeForJSON(raw), nil
}
func HandleClick(page playwright.Page, selector string) (string, error) {
	err := page.Click(selector)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Clicked %s", selector), nil
}

func HandleInput(page playwright.Page, selector, text string) (string, error) {
	// Using Type instead of Fill to simulate human keystrokes for bot detection
	err := page.Type(selector, text, playwright.PageTypeOptions{
		Delay: playwright.Float(100.0),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Typed '%s' into %s", text, selector), nil
}

func HandleSelectOption(page playwright.Page, selector, value string) (string, error) {
	_, err := page.SelectOption(selector, playwright.SelectOptionValues{
		Values: playwright.StringSlice(value),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Selected option %s in %s", value, selector), nil
}

func HandleWaitForSelector(page playwright.Page, selector string) (string, error) {
	_, err := page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(10000), // 10s timeout
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Element %s is now visible", selector), nil
}

func HandleGetCookies(ctx playwright.BrowserContext) (string, error) {
	cookies, err := ctx.Cookies()
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(cookies)
	return string(data), nil
}

func HandleClear(page playwright.Page, selector string) (string, error) {
	err := page.Fill(selector, "")
	return fmt.Sprintf("Cleared input at %s", selector), err
}

func HandleHover(page playwright.Page, selector string) (string, error) {
	err := page.Locator(selector).Hover()
	return fmt.Sprintf("Hovered over %s", selector), err
}

func HandlePress(page playwright.Page, key string) (string, error) {
	err := page.Keyboard().Press(key)
	return fmt.Sprintf("Pressed key: %s", key), err
}

func HandleBack(page playwright.Page) (string, error) {
	_, err := page.GoBack()
	return "Went back in history", err
}

func HandleForward(page playwright.Page) (string, error) {
	_, err := page.GoForward()
	return "Went forward in history", err
}

func HandleUploadFile(page playwright.Page, selector string, paths []string) (string, error) {
	err := page.Locator(selector).SetInputFiles(paths)
	return "Files uploaded successfully", err
}

func HandleDragAndDrop(page playwright.Page, source, target string) (string, error) {
	err := page.DragAndDrop(source, target)
	return fmt.Sprintf("Dragged %s to %s", source, target), err
}

func HandleEval(page playwright.Page, script string) (string, error) {
	res, err := page.Evaluate(script)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("JS Result: %v", res), nil
}

func HandleLocator(page playwright.Page, selector string) (string, error) {
	content, err := page.Locator(selector).First().InnerHTML()
	if err != nil {
		return "", err
	}
	return content, nil
}

func HandleScroll(page playwright.Page, x, y int) (string, error) {
	_, err := page.Evaluate(fmt.Sprintf("window.scrollBy(%d, %d)", x, y))
	return fmt.Sprintf("Scrolled by x:%d, y:%d", x, y), err
}

func HandleRefresh(page playwright.Page) (string, error) {
	_, err := page.Reload()
	return "Page refreshed", err
}

func HandleSleep(seconds int) string {
	time.Sleep(time.Duration(seconds) * time.Second)
	return fmt.Sprintf("Slept for %d seconds", seconds)
}

func HandlePagesInfo(ctx playwright.BrowserContext) int {
	return len(ctx.Pages())
}

func HandleNewPage(ctx playwright.BrowserContext) (playwright.Page, error) {
	return ctx.NewPage()
}

func HandleSwitchPage(ctx playwright.BrowserContext, index int) (playwright.Page, error) {
	pages := ctx.Pages()
	if index < 0 || index >= len(pages) {
		return nil, fmt.Errorf("page index %d out of bounds", index)
	}
	return pages[index], nil
}

func HandleClosePage(page playwright.Page) error {
	return page.Close()
}

func HandleTerminate(ctx playwright.BrowserContext, reason string, success bool) (string, bool) {
	for _, p := range ctx.Pages() {
		HandleClosePage(p)
	}

	return reason, success
}

// HandleSetCookie handles the "set_cookie" action from the AI.
// Expected Argument Order:
// 0: name, 1: value, 2: domain, 3: path
func HandleSetCookie(context playwright.BrowserContext, args []interface{}) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("set_cookie requires at least 'name' and 'value'")
	}

	name := fmt.Sprintf("%v", args[0])
	value := fmt.Sprintf("%v", args[1])

	var domain string
	if len(args) > 2 && args[2] != nil && fmt.Sprintf("%v", args[2]) != "" {
		domain = fmt.Sprintf("%v", args[2])
	}

	// Use arguments if provided, otherwise fallback to defaults
	path := "/"
	if len(args) > 3 && args[3] != nil && fmt.Sprintf("%v", args[3]) != "" {
		path = fmt.Sprintf("%v", args[3])
	}

	// Standard security settings required for most modern sites
	sameSite := playwright.SameSiteAttributeLax
	secure := true
	httpOnly := false

	cookie := playwright.OptionalCookie{
		Name:     name,
		Value:    value,
		Domain:   &domain,
		Path:     &path,
		SameSite: sameSite,
		Secure:   &secure,
		HttpOnly: &httpOnly,
	}

	err := context.AddCookies([]playwright.OptionalCookie{cookie})
	if err != nil {
		return "", fmt.Errorf("failed to inject: %w", err)
	}

	return fmt.Sprintf("Injected cookie '%s' for domain '%s' at path '%s'", name, domain, path), nil
}
