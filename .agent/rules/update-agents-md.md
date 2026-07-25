---
description: Use when updating AGENTS.md after code changes. Keeps the file trim and agent-optimized.
alwaysApply: false
---

# Updating AGENTS.md

## What belongs in AGENTS.md
Only content that **cannot be derived from `.agent/context/symbols.md`**:
- One-sentence purpose: what the tool does and which systems it touches
- Module responsibilities (one line per file — what it does, not what's in it)
- Hard conventions an agent must follow when writing code (naming, logging style, error handling, coverage rules)
- Do-not-edit paths

## What does NOT belong
- Function names, signatures, or class members — these live in `.agent/context/symbols.md`
- Call-chain diagrams or data flow pseudocode listing function names
- Make command recipes or setup instructions
- Prose introductions, section numbers, horizontal rules, or explanatory narrative

## Process
1. Identify what changed: new module, changed convention, new external system, or revised logic.
2. Update only the affected table row or bullet in AGENTS.md.
3. If public signatures changed, run `make gen-context` to refresh `.agent/context/symbols.md`.
4. Do not expand AGENTS.md beyond its current scope. If in doubt, leave it out.
