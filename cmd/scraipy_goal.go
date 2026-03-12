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
	"github.com/cvcvka5/scraipy/internal/config"
)

var DefaultMaxSteps = 100

func main() {
	config.LoadEnv()
	APIKey := config.GetEnv("OPENROUTER_API_KEY", "")
	Model := config.GetEnv("OPENROUTER_MODEL", "")

	// 1. Initialize Configuration
	goal, maxSteps := parseArgs()
	systemPrompt := getSystemPrompt("cmd/prompts/goal_prompt.txt")

	// Create Agent with modernized settings
	agent := ai.NewAgent(APIKey, Model, systemPrompt)

	// 2. Initialize Browser Environment
	// Headless is set to false so you can watch the agent work
	ctx, browser, pw, err := bridge.InitBrowser(false)
	exitOnErr(err)
	defer bridge.Cleanup(pw, browser)

	activePage, err := ctx.NewPage()
	exitOnErr(err)

	fmt.Printf("\n%s[SCRAIPY]%s 🚀 Starting Goal: %s\n", "\033[32m", "\033[0m", goal)

	// 3. Execution Loop
	// We send the goal once; subsequent turns are driven by observations
	currentInput := fmt.Sprintf("USER GOAL: %s\nMAX STEPS: %d steps.\nCURRENT STEP: %d", goal, maxSteps, 1)

	terminated := false
	for step := 1; step <= maxSteps; step++ {
		fmt.Printf("\n%s--- STEP %d / %d ---%s\n", "\033[35m", step, maxSteps, "\033[0m")

		// Get structured instructions from AI
		resp, err := agent.SendText("user", currentInput)
		if err != nil {
			log.Printf("🛑 Critical AI Error: %v", err)
			break
		}

		if len(resp.Choices) == 0 {
			log.Println("⚠️ AI returned no choices.")
			break
		}

		// Parse the JSON Step
		aiJSON := resp.Choices[0].Message.Content
		stepData := parseAIResponse(aiJSON)

		// Print Reasoning
		if stepData.Plan != "" {
			fmt.Printf("🧠 %sPlan:%s %s\n", "\033[33m", "\033[0m", stepData.Plan)
		}

		// Check if AI is finished
		if len(stepData.Commands) == 0 {
			fmt.Printf("\n✅ %s[TERMINATED]%s %s\n", "\033[32m", "\033[0m", stepData.Observation)
			break
		}

		// 4. Command Execution Phase

		// Force a getHTML at the end of every turn to ensure the AI has the latest page state for its next reasoning step.
		if stepData.Commands[len(stepData.Commands)-1].Action != bridge.GetHTMLAction {
			stepData.Commands = append(stepData.Commands, ai.Command{
				Action: bridge.GetHTMLAction,
			})
		}

		var turnObservations []string
		for _, cmd := range stepData.Commands {
			fmt.Printf("⚙️  %sAction:%s %-12s | Args: %v\n", "\033[36m", "\033[0m", cmd.Action, cmd.Arguments)

			// Handle the command via bridge
			// Note: cmd.Handle internally adds the observation to agent history via AddMessage("tool", ...)
			_, err := cmd.Handle(ctx, agent, &activePage)

			if err != nil {
				errMsg := fmt.Sprintf("Action '%s' failed: %v", cmd.Action, err)
				turnObservations = append(turnObservations, errMsg)
				continue
			}

			if cmd.Action == bridge.TerminateAction {
				terminated = true
				break
			}

			// We don't need to manually append successful observations here because
			// cmd.Handle(..., agent, ...) already logs them to the Agent's stateful history.
		}
		if terminated {
			break
		}

		// Prepare prompt for next turn if still running
		// We tell the AI to look at its own history for the results of the tool calls.
		currentInput = fmt.Sprintf("COMMANDS EXECUTED. CHECK YOUR HISTORY.\nUSER GOAL: %s\nMAX STEPS: %d steps.\nCURRENT STEP: %d", goal, maxSteps, step)
	}
}

// parseArgs handles CLI inputs
func parseArgs() (string, int) {
	if len(os.Args) < 2 {
		fmt.Println("Usage: scraipy \"Find the latest news on Go\" [max_steps]")
		os.Exit(1)
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

// parseAIResponse cleans and unmarshals the JSON output from the AI
func parseAIResponse(content string) ai.AgentStep {
	// Cleanup markdown blocks if the AI ignored system instructions
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var step ai.AgentStep
	if err := json.Unmarshal([]byte(content), &step); err != nil {
		log.Printf("⚠️ JSON Parse Warning: %v\nRaw Content: %s", err, content)
	}
	return step
}

func getSystemPrompt(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("⚠️ Warning: System prompt file not found at %s. Using default internal prompt.", path)
		return "You are a browser agent. Output JSON only."
	}
	return string(data)
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Printf("\n%s[FATAL]%s %v\n", "\033[31m", "\033[0m", err)
		os.Exit(1)
	}
}
