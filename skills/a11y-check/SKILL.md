---
name: a11y-check
description: Audit UI for WCAG 2.1 accessibility compliance and report violations. Use when user says "accessibility check", "a11y audit", "WCAG compliance", "screen reader", "ARIA labels", "keyboard navigation", or reviewing UI components for inclusivity.
---

Audit the UI for accessibility compliance (WCAG 2.1).

Steps:
1. Use automated tools (axe-core, cypress-axe, playwright-axe)
2. Scan components or pages for accessibility violations (e.g., missing alt text, low color contrast, missing ARIA labels)
3. Report found issues and propose fixes

Output:
- Compliance report
- List of accessibility violations (Critical, Moderate, Low)
