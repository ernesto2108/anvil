---
name: bundle-analyzer
description: Analyze frontend bundle size impact of new dependencies or components. Use when user says "bundle size", "check bundle", "too heavy", "tree shaking", "webpack analyze", or after adding a new npm dependency to evaluate its size impact.
---

Monitor the size impact of adding new dependencies or complex components on the final bundle size.

Steps:
1. Run bundle analysis (e.g., webpack-bundle-analyzer, vite-bundle-visualizer)
2. Compare the current bundle size against the baseline
3. Identify "heavy" modules and suggest alternatives

Goal: Keep initial load times low and avoid bloat.
