package scaffold

import (
	"strings"
)

// indentPatterns indents each newline-separated pattern by two spaces so the
// result can be embedded directly after a YAML block-scalar key ("key: |").
func indentPatterns(patterns string) string {
	lines := strings.Split(strings.TrimRight(patterns, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "        " + l
	}
	return strings.Join(lines, "\n")
}

const skillMappingsBlock = `      # skill-mappings routes changed files to the appropriate review skill.
      # The catch-all "**" sends every PR to the ci-review skill, which lives
      # at .agent/skills/ci-review/SKILL.md (symlinked from .claude/skills/).
      # Add path-specific entries above "**" to use a different skill for
      # certain file patterns, e.g.:
      #   packages/migrations/** -> .claude/skills/migrations/SKILL.md
      skill-mappings: |
        ** -> .claude/skills/ci-review/SKILL.md`

// AiReviewWorkflow returns the event-driven AI code review workflow populated
// with language-appropriate file and excluded patterns.
func AiReviewWorkflow(filePatterns, excludedPatterns string) string {
	return `name: AI Code Review

# Runs automatically on every non-draft PR (opened / ready_for_review) AND
# on-demand when a collaborator posts a comment containing "/ai-review-requested".
#
# Keep automatic reviews enabled — they establish a consistent baseline for all
# PRs, improve auditability, and give human reviewers confidence that basics are
# covered. The on-demand trigger is for re-reviews after changes, not a
# replacement for automatic reviews.

on:
  pull_request:
    types: [opened, ready_for_review]
  issue_comment:
    types: [created]

permissions:
  contents: read
  pull-requests: write
  id-token: write

jobs:
  ai-review:
    # Run on non-draft PRs opened/marked ready, or when a collaborator/member
    # posts a comment containing "/ai-review-requested".
    # The author_association guard is a security requirement: issue_comment
    # always runs in the base repo's context with full permissions (including
    # pull-requests: write), even for fork PRs.
    if: >
      (github.event_name == 'pull_request' && github.event.pull_request.draft == false) ||
      (github.event_name == 'issue_comment' &&
       github.event.issue.pull_request != null &&
       contains(github.event.comment.body, '/ai-review-requested') &&
       contains(fromJSON('["COLLABORATOR", "MEMBER", "OWNER"]'), github.event.comment.author_association))
    uses: Workiva/gha-ai-code-review/.github/workflows/ai-code-review.yml@0.1.80
    secrets: inherit
    with:
      file-patterns: |
` + indentPatterns(filePatterns) + `
      excluded-patterns: |
` + indentPatterns(excludedPatterns) + `
` + skillMappingsBlock + `
`
}

// AiReviewScheduledWorkflow returns the scheduled catch-up AI code review
// workflow populated with language-appropriate file and excluded patterns.
func AiReviewScheduledWorkflow(filePatterns, excludedPatterns string) string {
	return `name: AI Code Review (Scheduled Catch-up)

# Backstop for PRs that slipped through the event-driven workflow —
# e.g. PRs opened while GHA was down, or that hit GitHub's search index lag.
# Runs hourly and reviews any open PR that hasn't received an AI review yet.

on:
  schedule:
    - cron: '0 * * * *'  # top of every hour
  workflow_dispatch:       # allow manual trigger for testing or catch-up

permissions:
  contents: read
  pull-requests: write
  id-token: write

jobs:
  scheduled-review:
    uses: Workiva/gha-ai-code-review/.github/workflows/ai-code-review-scheduled.yml@0.1.80
    secrets: inherit
    with:
      file-patterns: |
` + indentPatterns(filePatterns) + `
      excluded-patterns: |
` + indentPatterns(excludedPatterns) + `
` + skillMappingsBlock + `
      # Optional overrides (defaults shown):
      # max-reviews: 20    # circuit breaker — max PRs to review per run
      # max-parallel: 2    # review this many PRs concurrently
`
}
