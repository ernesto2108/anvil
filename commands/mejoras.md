---
name: mejoras
description: Muestra mejoras detectadas durante sesiones de trabajo — optimizaciones, patrones a corregir, ineficiencias del workflow
---

Lee el archivo de mejoras detectadas y dame un resumen ejecutivo en español:

1. Lee `$MEMORY_DIR/improvements.md`
2. Lee `$MEMORY_DIR/session_costs.md`

Presenta:
- **Resumen rápido:** Cuántas mejoras pendientes hay y cuántas resueltas
- **Top 3 por impacto:** Las que más tokens/tiempo ahorrarían
- **Tendencia:** Si los costos por sesión están bajando o subiendo (comparar últimas sesiones)
- **Próxima acción sugerida:** La mejora más fácil de implementar ahora

Sé conciso. No repitas el contenido completo — solo lo actionable.
