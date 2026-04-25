---
name: e2e-test-run
disable-model-invocation: true
description: Ejecutar pruebas end-to-end que validan flujos completos de usuario a través de todo el stack. Usar cuando el usuario diga "run e2e tests", "end to end", "Playwright", "Cypress", "test the full flow", o al validar que una funcionalidad trabaja desde la UI hasta la base de datos.
---

Valida flujos completos de usuario y funcionalidad end-to-end usando navegadores/herramientas especializadas.

Pasos:
1. Ejecutar pruebas basadas en browser (ej., Playwright, Cypress)
2. Ejecutar flujos comunes como "Login -> Create Resource -> Logout"
3. Verificar fallos, timeouts y errores inesperados

Objetivo: Asegurar que las funcionalidades de alto nivel funcionen correctamente a través de todo el stack.
