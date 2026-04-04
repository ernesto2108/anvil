---
name: dependency-check
description: Analyze project dependencies for vulnerabilities, license issues, and outdated versions. Use when user says "check dependencies", "audit packages", "outdated deps", "npm audit", "go mod tidy", "security vulnerabilities", or before upgrading libraries.
---

Analyze project dependencies for vulnerabilities, license issues, and updates.

Actions:
1. Run `go list -m all` to see full dependency tree
2. Run `go mod tidy` to clean up
3. Use `npm audit` if Node.js is detected
4. Check for known CVEs in the listed versions

Rules:
- Report only high and critical vulnerabilities unless asked otherwise
- Propose specific version updates
