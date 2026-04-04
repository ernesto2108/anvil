---
name: read-files
description: Safely read project files to gather context before making decisions. Use when needing to understand existing code before modifying it, gathering context from multiple files, or summarizing large codebases.
user-invocable: false
---

Read only what is necessary to complete the task.

Capabilities:
- read one or multiple files
- read folders recursively
- summarize large files
- extract only relevant sections
- search by keyword or symbol

Usage rules:
- avoid loading the whole repo if not needed
- prefer targeted reads (specific files first)
- summarize long outputs
- highlight only relevant parts

Output format:
- file path
- short summary
- relevant snippets only

Never:
- hallucinate file contents
- assume code without reading it
