package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cvcvka5/scraipy/internal/ai"
	"github.com/cvcvka5/scraipy/internal/bridge"
)

const (
	DefaultAPIKey   = "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"
	DefaultModel    = "stepfun/step-3.5-flash:free"
	DefaultMaxSteps = 10
)

func main() {
	// 1. Setup Configuration & Agent
	goal, maxSteps := parseArgs()
	systemPrompt := getSystemPrompt("cmd/prompts/goal_prompt.txt")
	agent := ai.NewAgent(DefaultAPIKey, DefaultModel, systemPrompt)

	// 2. Initialize Browser
	ctx, browser, pw, err := bridge.InitBrowser(false) // Set to true for headless
	exitOnErr(err)
	defer bridge.Cleanup(pw, browser)

	activePage, err := ctx.NewPage()
	exitOnErr(err)

	// 3. Main Execution Loop
	fmt.Printf("🚀 Goal: %s\n", goal)

	// Initial message to get the ball rolling
	currentInput := fmt.Sprintf("Goal: %s\nMax Steps: %d", goal, maxSteps)

	for step := 1; step <= maxSteps; step++ {
		fmt.Printf("\n--- [Step %d/%d] ---\n", step, maxSteps)

		// Get AI instructions
		resp, err := agent.SendText("user", currentInput)
		if err != nil {
			log.Printf("❌ AI Error: %v", err)
			break
		}

		aiJSON := resp.Choices[0].Message.Content

		// Parse commands
		instructions := parseAIResponse(aiJSON)

		// If no commands, the AI likely thinks it's finished
		if len(instructions.Commands) == 0 {
			fmt.Println("✅ Goal reached or AI stopped.")
			break
		}

		// Execute commands and collect observations
		var observations []string
		for _, cmd := range instructions.Commands {
			fmt.Printf("⚙️ Executing: %s\n", cmd.Action)

			obsMsg, err := cmd.Handle(ctx, agent, &activePage)
			if err != nil {
				obsStr := fmt.Sprintf("Action %s failed: %v", cmd.Action, err)
				observations = append(observations, obsStr)
				// Also add the error to history so the AI knows it failed
				agent.AddMessage("user", ai.MessagePart{Type: "text", Text: obsStr})
				continue
			}

			// Extract observation text and add to agent memory
			obsText := obsMsg.Content[0].Text
			observations = append(observations, obsText)
			agent.AddMessage("user", ai.MessagePart{Type: "text", Text: obsText})
		}

		// Prepare input for next turn
		currentInput = strings.Join(observations, "\n")
		if len(currentInput) > 2000 {
			currentInput = currentInput[:2000] + "... [truncated]"
		}
	}
}

// --- Helpers ---

func parseArgs() (string, int) {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run . \"your goal\" [max_steps]")
	}
	goal := os.Args[1]
	steps := DefaultMaxSteps
	if len(os.Args) > 2 {
		if s, err := strconv.Atoi(os.Args[2]); err == nil {
			steps = s
		}
	}
	return goal, steps
}

func parseAIResponse(content string) ai.AgentStep {
	// Clean markdown formatting if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var step ai.AgentStep
	if err := json.Unmarshal([]byte(content), &step); err != nil {
		log.Printf("⚠️ Failed to parse JSON: %v", err)
	}
	return step
}

func getSystemPrompt(path string) string {
	data, err := os.ReadFile(path)
	exitOnErr(err)
	return string(data)
}

func exitOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
