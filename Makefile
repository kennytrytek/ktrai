gen-context: ## Regenerate AI context index (requires universal-ctags with JSON support)
	@ctags --list-output-formats 2>/dev/null | grep -q json || { echo "ctags found but lacks JSON output support — install universal-ctags (e.g. brew install universal-ctags)"; exit 1; }
	ctags --output-format=json --fields=+KSZnte --languages=Go \
	  -R . \
	  | ktrai gen-symbols > .agent/context/symbols.md

prep: gen-context ## Prepare for a commit (regenerate AI context)
