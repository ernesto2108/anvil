---
name: write-files
description: Create or modify files with minimal, focused changes respecting permissions. Use when writing code changes to ensure minimal diffs, preserved formatting, and no unrelated modifications.
user-invocable: false
---

Create or modify files in the repository.

General rules:
- modify the minimum number of files
- small, focused changes
- preserve formatting
- never rewrite unrelated code
- explain changes clearly

Safe writing strategy:
1. Read file first
2. Understand context
3. Apply minimal patch
4. Show diff
5. Then write

Output format per change:

File: <path>
Reason: <why>

Diff:
+ <added lines>
- <removed lines>

Never:
- mass rewrite files
- change architecture without design
- modify files outside scope

## Design Source Validation

When modifying UI code that has a corresponding design file (.pen or Figma):

1. **Before writing**: Read the design to understand exact spacing, colors, layout, and typography
2. **After writing**: Compare the implementation visually against the design
3. **Never assume**: If you don't have the design open, open it first. Memory of "what the design looked like" is not reliable
4. **Report discrepancies**: If the design and current code already differ, flag it to the user before making changes
