---
name: test-api
description: Probar y validar endpoints de API para cumplimiento de contratos. Úsalo cuando el usuario diga "probar la API", "verificar endpoint", "validar respuesta", "contrato de API", "curl este endpoint", o al verificar códigos de estado HTTP y esquemas de respuesta.
---

Prueba y valida endpoints de API externos o internos para asegurar que se ajusten a los contratos esperados.

Capacidades:
- Usar `curl` para solicitudes básicas
- Validar estructura JSON y tipos de campos
- Verificar códigos de estado HTTP y headers
- Probar diferentes escenarios de autenticación (token válido/inválido)

Reglas:
- NUNCA enviar datos sensibles de producción (keys/PII) en texto plano
- Usar placeholders para las API keys
- Preferir métodos no destructivos (GET/HEAD) a menos que sea necesario
- Para POST/PUT/PATCH, usar un entorno de desarrollo/mock
