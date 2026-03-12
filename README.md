# Scraipy: The Autonomous Browser Operator
Scraipy is a stateful, autonomous agent that lives in your terminal and navigates the web like a human. Built with Go and Playwright, it leverages advanced LLMs to "reason" through complex web workflows—from solving deep-research puzzles to automating high-friction digital labor.

## The Vision
Traditional scrapers break the second a CSS class changes. Scraipy doesn't care. By looking at the semantic structure of a page rather than just raw code, Scraipy plans its moves dynamically. It executes clicks, hovers, drags, and keyboard shortcuts until the goal is met. It’s not just a scraper; it’s the bridge between static data and truly autonomous digital labor.

## Modernized Capabilities
### V2
- Autonomous Reasoning: Provide a high-level goal like "Find the total length of the last 3 videos on this channel," and Scraipy handles the navigation, searching, and calculation.

- Human-Like Interaction: Full support for hover, drag_and_drop, press (keyboard combos), and multi-tab orchestration.

- Stateful Memory: A context-aware "Skeptical Observer" loop ensures the agent remembers what it saw. No more redundant navigations or context window overflows.

- Stealth by Default: Hardened browser configurations and human-like input delays to bypass modern bot detection.

- Semantic Filtering: Aggressively strips HTML noise to keep LLM tokens low and reasoning accuracy high.

- Strict JSON Enforcement: Zero "AI chatter." The agent communicates exclusively in actionable, schema-validated JSON.

### V1
- Session-based navigation (back/forward).
- Visual verification via high-res screenshots.
- Exponential backoff for resilient API interactions.

## Roadmap to V3
- Action Recording: Record manual browser sessions to generate AI-assisted scripts instantly.
- Optimization 2.0: Faster DOM processing and even leaner token usage.
- Advanced Interactables: Support for complex Shadow DOMs and Canvas-based elements.
- Collaborative Mode: Real-time human-in-the-loop for CAPTCHA solving or 2FA entry.

## Quick Start
1. Clone the project:
```sh
git clone https://github.com/cvcvka5/scraipy.git
```

2. Create '.env' in the project root.

3. Set the environment variables in .env:
```.env
OPENROUTER_API_KEY=sk-...
OPENROUTER_MODEL=openrouter/hunter-alpha
```

4. Create a file containing your goal in any format.

5. Run the script:
```sh
go run ./cmd/scraipy_goal.go "path/to/yourfile.txt"
```


## The "Skeptical Observer" Workflow
Scraipy operates on a strict Feedback Loop logic:
1. Observe: Get the current minified HTML state.
2. Think: Process the data and commit "CRUCIAL DATA" to its internal ledger.
3. Plan: Chain multiple browser commands (Input + Click + Wait).
4. Execute: Run the commands and return to Step 1 to verify the outcome.

## Operational DNA
Scraipy follows a set of "Hard Rules" to ensure reliability:
1. Never Assume Success: Every action is a hypothesis until the next get_html proves it.
2. Memory is Ledger: If a price is found, it's saved in a think block. Scraipy will never waste your tokens re-navigating to a page it has already processed.
3. Tab Mastery: Uses new_page and switch_page to cross-reference data across multiple sites simultaneously.


## Contributing
Scraipy is for the builders who are tired of brittle selectors. If you want to contribute to the future of autonomous browsing, feel free to open a PR or an issue.

## License
[MIT License](./LICENSE)