---
name: generate-diagram
description: Create Mermaid.js diagrams for architecture, flows, sequences, ERDs, and C4. Use when user says "draw a diagram", "create a flowchart", "sequence diagram", "ERD", "architecture diagram", "visualize", or needs visual documentation.
---

Create or update visual representations of code or architecture using Mermaid.js syntax within Markdown files.

Capabilities:
- Flowcharts (process flows)
- Sequence diagrams (component interaction)
- Class/Module diagrams (structure)
- Entity Relationship Diagrams (ERD — database schema)
- C4 diagrams (system context and containers)

Rules:
- Always use standard Mermaid.js syntax
- Place diagrams inside ```mermaid ... ``` blocks
- Keep diagrams simple enough to be readable on a standard screen
- Prefer sequence diagrams for explaining complex asynchronous flows
- Use C4 for high-level architectural overviews
