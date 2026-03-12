package ai

import (
	"fmt"
	"strconv"

	"github.com/cvcvka5/scraipy/internal/bridge"
	"github.com/playwright-community/playwright-go"
)

const ToolCallRole = "user"

// Handle executes an AI command against the browser and updates the Agent's memory.
func (c Command) Handle(ctx playwright.BrowserContext, agent *Agent, currentPage *playwright.Page) (ChatMessage, error) {
	cm := ChatMessage{Role: ToolCallRole}

	if !c.Action.Validate() {
		errText := fmt.Sprintf("Error: Action '%s' is not supported.", c.Action)
		agent.AddMessage(ToolCallRole, MessagePart{Type: "text", Text: errText})
		cm.Content = []MessagePart{{Type: "text", Text: errText}}
		return cm, fmt.Errorf("invalid action: %s", c.Action)
	}

	var result string
	var err error

	// Helper to safely convert interface (from JSON) to int.
	toInt := func(val any) int {
		switch v := val.(type) {
		case float64:
			return int(v)
		case string:
			i, _ := strconv.Atoi(v)
			return i
		default:
			return 0
		}
	}

	// Helper to safely get string arguments.
	getStr := func(idx int) string {
		if idx < len(c.Arguments) {
			if s, ok := c.Arguments[idx].(string); ok {
				return s
			}
		}
		return ""
	}
	getBool := func(idx int) bool {
		if idx < len(c.Arguments) {
			if b, ok := c.Arguments[idx].(bool); ok {
				return b
			}
		}
		return false
	}
	switch c.Action {
	case bridge.ThinkAction:
		result = fmt.Sprintf("Thought Logged: %s", getStr(0))

	case bridge.GetHTMLAction:
		result, err = bridge.HandleGetHTML(*currentPage)

	case bridge.NavigateAction:
		result, err = bridge.HandleNavigate(*currentPage, getStr(0))

	case bridge.ClickAction:
		result, err = bridge.HandleClick(*currentPage, getStr(0))

	case bridge.InputAction:
		result, err = bridge.HandleInput(*currentPage, getStr(0), getStr(1))

	case bridge.ClearInputAction:
		result, err = bridge.HandleClear(*currentPage, getStr(0))

	case bridge.HoverAction:
		result, err = bridge.HandleHover(*currentPage, getStr(0))

	case bridge.PressAction:
		result, err = bridge.HandlePress(*currentPage, getStr(0))

	case bridge.BackAction:
		result, err = bridge.HandleBack(*currentPage)

	case bridge.ForwardAction:
		result, err = bridge.HandleForward(*currentPage)

	case bridge.DragAndDropAction:
		result, err = bridge.HandleDragAndDrop(*currentPage, getStr(0), getStr(1))

	case bridge.EvaluateJSAction:
		result, err = bridge.HandleEval(*currentPage, getStr(0))

	case bridge.LocatorAction:
		result, err = bridge.HandleLocator(*currentPage, getStr(0))

	case bridge.ScrollAction:
		if len(c.Arguments) < 2 {
			err = fmt.Errorf("scroll requires x and y arguments")
		} else {
			result, err = bridge.HandleScroll(*currentPage, toInt(c.Arguments[0]), toInt(c.Arguments[1]))
		}

	case bridge.RefreshAction:
		result, err = bridge.HandleRefresh(*currentPage)

	case bridge.SleepAction:
		result = bridge.HandleSleep(toInt(c.Arguments[0]))

	case bridge.PagesInfoAction:
		result = fmt.Sprintf("Total open pages: %d", bridge.HandlePagesInfo(ctx))

	case bridge.NewPageAction:
		var newPage playwright.Page
		newPage, err = bridge.HandleNewPage(ctx)
		if err == nil {
			*currentPage = newPage
			result = "Opened new tab and switched focus."
		}

	case bridge.SwitchPageAction:
		idx := toInt(c.Arguments[0])
		var newPage playwright.Page
		newPage, err = bridge.HandleSwitchPage(ctx, idx)
		if err == nil {
			*currentPage = newPage
			result = fmt.Sprintf("Switched focus to tab index %d.", idx)
		}

	case bridge.ClosePageAction:
		err = bridge.HandleClosePage(*currentPage)
		if err == nil {
			result = "Closed active tab."
			pages := ctx.Pages()
			if len(pages) > 0 {
				*currentPage = pages[len(pages)-1]
			}
		}

	case bridge.GetCookiesAction:
		result, err = bridge.HandleGetCookies(ctx)

	case bridge.SelectOptionAction:
		result, err = bridge.HandleSelectOption(*currentPage, getStr(0), getStr(1))

	case bridge.SetCookieAction:
		result, err = bridge.HandleSetCookie(ctx, c.Arguments)

	case bridge.TerminateAction:
		reason, success := bridge.HandleTerminate(ctx, getStr(0), getBool(1))
		result = fmt.Sprintf("Terminating session. Reason: %s | Success: %t", reason, success)

	default:
		err = fmt.Errorf("action %s validated but missing handler mapping", c.Action)
	}

	// Reporting results back to the Agent.
	output := fmt.Sprintf("Executed action '%s': %s", c.Action, result)
	if err != nil {
		output = fmt.Sprintf("Error executing '%s': %v", c.Action, err)
	}
	cm.Content = []MessagePart{{Type: "text", Text: output}}

	agent.AddMessage(ToolCallRole, cm.Content...)
	fmt.Println("- " + output[:min(len(output), 200)] + "...")

	return cm, err
}
