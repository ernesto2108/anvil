# Report Format

El reporte se imprime directamente en consola. Sin archivos, sin preamble.

## Formato

### Header — Modo local

```
═══════════════════════════════════════════
  POST-DEVELOPMENT REVIEW
═══════════════════════════════════════════
  Branch:            {nombre de branch}
  Stack detectado:   {Go, React, etc.}
  Archivos revisados: {N}
  Archivos con hallazgos: {N}
═══════════════════════════════════════════
```

### Header — Modo PR

```
═══════════════════════════════════════════
  POST-DEVELOPMENT REVIEW — PR #{number}
═══════════════════════════════════════════
  PR:                {title}
  Branch:            {head} → {base}
  Autor:             {author}
  Stack detectado:   {Go, React, etc.}
  Archivos revisados: {N}
  Archivos con hallazgos: {N}
═══════════════════════════════════════════
```

### Body (igual para ambos modos)

── LINT ───────────────────────────────────

Linter:          {nombre del linter o "No configurado"}
Config:          {path al config file o "No encontrado"}
Estado:          {OK (0 issues) | {N} issues | No ejecutado | No configurado}

{Si hay lint issues, listarlos como hallazgos normales en CRITICOS/MEJORAS}
{Si no hay linter configurado, se reporta como CRITICO en la seccion correspondiente}

── CRITICOS ({N}) ─────────────────────────

{numero}. [{categoria}] {archivo}:{linea}
   {descripcion del problema}
   
   Por que es problema:
   {explicacion de que puede salir mal}
   
   Como reproducir:
   {pasos concretos para provocar el bug}
   
   Fix sugerido:
   {codigo o descripcion del fix}

── MEJORAS ({N}) ──────────────────────────

{numero}. [{categoria}] {archivo}:{linea}
   {descripcion}
   
   Por que:
   {razon concreta}
   
   Fix sugerido:
   {sugerencia}

── NOTAS ({N}) ────────────────────────────

{numero}. [{categoria}] {archivo}:{linea}
   {observacion}

═══════════════════════════════════════════
  RESUMEN
═══════════════════════════════════════════
  Score:          {N}/10
  Criticos:       {N}
  Mejoras:        {N}
  Notas:          {N}
  
  Lint:           {OK | N issues | No configurado}
  Recomendacion:  {texto segun score}
═══════════════════════════════════════════
```

## Recomendaciones segun score

| Score | Recomendacion |
|---|---|
| 9-10 | Listo para mergear. Sin observaciones bloqueantes. |
| 7-8 | Seguro para mergear. Considera las mejoras sugeridas. |
| 5-6 | Recomiendo resolver las mejoras antes de mergear. |
| 3-4 | No recomiendo mergear. Resolver los criticos primero. |
| 1-2 | Bloqueo recomendado. Hay riesgos graves que resolver. |

## Reglas de formato

- Si no hay hallazgos de una severidad, omitir esa seccion completa
- Cada hallazgo CRITICO debe tener las 3 subsecciones: Por que, Como reproducir, Fix sugerido
- Cada hallazgo MEJORA debe tener: Por que, Fix sugerido
- Las NOTAS solo llevan la observacion, sin subsecciones
- Paths y codigo en English. Descripciones en Spanish.
- Usar line numbers reales del archivo, no del diff
