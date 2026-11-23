# Patrones de Diseño — <ProjectName>

last_updated: <YYYY-MM-DD>

<!-- Este archivo se construye por inferencia estructural, no por nombres.
     Un patrón puede llamarse de cualquier manera o no tener nombre explícito.
     La firma del código es lo que importa. -->

## Creacionales

<!-- Factory, Builder, Singleton, Provider, Functional Options -->

### <NombreInferido o "Factory sin nombre"> — <archivo principal>
- **Archivo:** `<path>:<line>`
- **Qué construye:** <descripción>
- **Firma detectada:** `func New<X>(kind string) <Interface>` con switch interno
- **Cuándo usar:** <contexto — al agregar nueva variante de X>
- **Anti-pattern:** NO instanciar `<ConcreteType>` directamente fuera de esta función

## Estructurales

<!-- Repository, Adapter, Decorator, Facade, Middleware -->

### <NombreInferido> — <archivo>
- **Archivo:** `<path>:<line>`
- **Qué encapsula:** <descripción>
- **Firma detectada:** <descripción de la firma>
- **Cuándo usar:** <contexto>

## De comportamiento

<!-- Strategy, Observer, Command, Pipeline, Chain of Responsibility -->

### <NombreInferido> — <archivo>
- **Archivo:** `<path>:<line>`
- **Qué varía:** <qué comportamiento es intercambiable>
- **Implementaciones detectadas:** `<impl1>`, `<impl2>`
- **Cuándo agregar nueva impl:** <pasos>

## Go-específicos

<!-- Functional options, middleware chain, table-driven tests, context propagation -->

### Functional Options — <archivo>
- **Archivo:** `<path>`
- **Tipo option:** `type Option func(*<Config>)`
- **Cuándo usar:** al configurar `<Struct>` con parámetros opcionales

## TypeScript-específicos

<!-- HOC, hooks personalizados, context providers, compound components -->

<!-- Eliminar sección si el proyecto no tiene TypeScript -->

## Patrones a evitar en este proyecto

<!-- Patrones que se consideraron y descartaron, con razón -->
- <patrón>: <por qué no aplica aquí>
