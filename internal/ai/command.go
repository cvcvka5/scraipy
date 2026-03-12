package ai

import (
	"fmt"
	"strconv"

	"github.com/cvcvka5/scraipy/internal/bridge"
	"github.com/playwright-community/playwright-go"
)

// Command represents a single instruction from the AI.
type Command struct {
	Action    bridge.BrowserAction `json:"action"`
	Arguments []any                `json:"arguments"`
}

// Handle executes the command against the browser and returns a message for the AI.
// It also automatically adds the observation to the agent's memory.
func (c Command) Handle(ctx playwright.BrowserContext, agent *Agent, currentPage *playwright.Page) (ChatMessage, error) {
	cm := ChatMessage{
		Role: "tool",
	}

	if !c.Action.Validate() {
		return cm, fmt.Errorf("invalid action: %s", c.Action)
	}

	var result string
	var err error

	// Helper to safely convert interface to int (JSON numbers are float64)
	toInt := func(val any) int {
		if f, ok := val.(float64); ok {
			return int(f)
		}
		if s, ok := val.(string); ok {
			i, _ := strconv.Atoi(s)
			return i
		}
		return 0
	}

	switch c.Action {
	case "think":
		if len(c.Arguments) > 0 {
			result = fmt.Sprintf("Thought Logged: %v", c.Arguments[0])
		} else {
			result = "Think action executed with no content."
		}

	case "get_html":
		result, err = bridge.HandleGetHTML(*currentPage)

	case bridge.NavigateAction:
		arg := c.Arguments[0].(string)
		result, err = bridge.HandleNavigate(*currentPage, arg)

	case bridge.ClickAction:
		arg := c.Arguments[0].(string)
		result, err = bridge.HandleClick(*currentPage, arg)

	case bridge.InputAction:
		selector := c.Arguments[0].(string)
		text := c.Arguments[1].(string)
		result, err = bridge.HandleInput(*currentPage, selector, text)

	case bridge.LocatorAction:
		arg := c.Arguments[0].(string)
		result, err = bridge.HandleLocator(*currentPage, arg)

	case "scroll":
		if len(c.Arguments) < 2 {
			err = fmt.Errorf("scroll requires x and y arguments")
		} else {
			x := toInt(c.Arguments[0])
			y := toInt(c.Arguments[1])
			result, err = bridge.HandleScroll(*currentPage, x, y)
		}

	case "refresh":
		result, err = bridge.HandleRefresh(*currentPage)

	case bridge.PagesInfoAction:
		count := bridge.HandlePagesInfo(ctx)
		result = fmt.Sprintf("Total open pages: %d", count)

	case bridge.NewPageAction:
		newPage, nerr := bridge.HandleNewPage(ctx)
		if nerr == nil {
			*currentPage = newPage
			result = "Opened new page and switched focus to it."
		}
		err = nerr

	case bridge.SwitchPageAction:
		idx := toInt(c.Arguments[0])
		newPage, serr := bridge.HandleSwitchPage(ctx, idx)
		if serr == nil {
			*currentPage = newPage
			result = fmt.Sprintf("Switched to page index %d", idx)
		}
		err = serr

	case bridge.ClosePageAction:
		err = bridge.HandleClosePage(*currentPage)
		if err == nil {
			result = "Closed the current page."
			pages := ctx.Pages()
			if len(pages) > 0 {
				*currentPage = pages[len(pages)-1]
			}
		}

	case bridge.SleepAction:
		secs := toInt(c.Arguments[0])
		result = bridge.HandleSleep(secs)

	default:
		err = fmt.Errorf("action %s is recognized but not yet implemented in handler", c.Action)
	}

	// Handle errors by reporting them back to the AI as an observation
	if err != nil {
		errorText := fmt.Sprintf("Error executing %s: %v", c.Action, err)
		agent.AddMessage("tool", MessagePart{Type: "text", Text: errorText})
		cm.Content = []MessagePart{{Type: "text", Text: errorText}}
		return cm, err
	}

	// Add successful observation to agent memory
	agent.AddMessage("tool", MessagePart{
		Type: "text",
		Text: result,
	})

	cm.Content = []MessagePart{{Type: "text", Text: result}}
	return cm, nil
}
