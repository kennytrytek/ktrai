.DEFAULT_GOAL := help

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"}; {printf "  %-20s %s\n", $$1, $$2}'

.agent/context/symbols.md: ## Regenerate AI context index (requires universal-ctags with JSON support)
	@ctags --list-output-formats 2>/dev/null | grep -q json || { echo "ctags found but lacks JSON output support — install universal-ctags (e.g. brew install universal-ctags)"; exit 1; }
	ctags --output-format=json --fields=+KSZnte --languages=Go \
	  -R . \
	  | ktrai gen > .agent/context/symbols.md

gen: .agent/context/symbols.md ## Regenerate AI context index

.PHONY: help gen prep
prep: gen ## Prepare for a commit (regenerate AI context)
