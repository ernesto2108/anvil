---
name: visual-diff
disable-model-invocation: true
description: Detect visual regressions by comparing before/after screenshots of UI changes. Use when user says "visual regression", "screenshot comparison", "UI looks different", "CSS broke something", or after modifying styles, themes, or shared components.
---

Detect unexpected visual regressions when modifying CSS, themes, or core components.

Steps:
1. Capture screenshots of target pages/components before and after changes
2. Compare the two versions and highlight pixel differences (e.g., pixelmatch, reg-viz)
3. Review differences for regressions or intended changes

Output:
- Visual diff report showing "before" and "after" images with highlighting

Note: Requires screenshot tooling — not available without browser automation setup.
