---
name: dependency-check
description: Analiza las dependencias del proyecto en busca de vulnerabilidades, problemas de licencia y versiones desactualizadas. Usar cuando el usuario diga "check dependencies", "audit packages", "outdated deps", "npm audit", "go mod tidy", "security vulnerabilities", o antes de actualizar librerías.
user-invocable: true
---

Analiza las dependencias del proyecto en busca de vulnerabilidades, problemas de licencia y actualizaciones.

Acciones:
1. Ejecutar `go list -m all` para ver el árbol completo de dependencias
2. Ejecutar `go mod tidy` para limpiar
3. Si se detecta Node.js, auditar via el gestor de paquetes del proyecto (detectar desde lockfile según CLAUDE.md): `pnpm audit` (preferido), `npm audit`, o `yarn audit`. NO leer `node_modules/` directamente — está denegado por `permissions.deny` y `audit` muestra la misma información.
4. Verificar CVEs conocidos en las versiones listadas

Reglas:
- Reportar solo vulnerabilidades high y critical salvo que se indique lo contrario
- Proponer actualizaciones de versión específicas
