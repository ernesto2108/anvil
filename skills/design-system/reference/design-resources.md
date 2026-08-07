# Recursos de Referencia para Diseño

Fuentes curadas por categoría para investigar patrones, fuentes y paletas antes de especificar un diseño.

## Patrones UI — productos reales
- [Mobbin](https://mobbin.com/) — flujos y pantallas reales de apps mobile y web
- [Page Flows](https://pageflows.com/) — grabaciones en video de flujos completos (onboarding, checkout, etc.)
- [Refero](https://refero.design/) — 30K+ screenshots de apps reales, búsqueda por patrón UX
- [Screenlane](https://screenlane.com/) — screenshots organizados por tipo de pantalla (login, empty state, etc.)

## SaaS
- [SaaSFrame](https://www.saasframe.io) — 5,000+ ejemplos de UI SaaS con archivos Figma
- [SaaS Interface](https://saasinterface.com/) — UI de apps SaaS por tipo de flujo
- [SaaSUI](https://www.saasui.design/) — patrones de dashboard SaaS reales
- [Saaspo](https://saaspo.com/) — sitios web SaaS curados, filtrado por industria

## Landing pages y web
- [Lapa Ninja](https://www.lapa.ninja/) — 7,300+ diseños de landing pages desde 2015
- [Godly](https://godly.website/) — diseño de vanguardia con motion e interacciones avanzadas
- [Awwwards](https://www.awwwards.com/) — estándar de la industria para web design de excelencia
- [Land-book](https://land-book.com/) — galería diaria de diseños web curados
- [One Page Love](https://onepagelove.com/) — single-page sites y landing pages

## Mobile
- [Scrnshts.club](https://scrnshts.club/) — screenshots curados de App Store
- [Pttrns](https://pttrns.com/) — patrones de UI mobile por categoría

## Design systems
- [Design Systems Repo](https://designsystemsrepo.com/) — colección de design systems reales (Atlassian, GitHub, Shopify)
- [The Component Gallery](https://component.gallery/) — mismo componente comparado entre múltiples design systems

## Dashboards y data-viz
- [Muzli](https://muz.li/) — tendencias curadas de diseño, sección de dashboards dedicada
- [Tableau Public](https://public.tableau.com/app/discover) — millones de dashboards interactivos reales

## Inspiración general
- [Dribbble](https://dribbble.com/) — inspiración de componentes UI y pantallas
- [SiteInspire](https://www.siteinspire.com/) — web design curado por estética y tipo

## Guías canónicas de plataforma
- [Apple Human Interface Guidelines](https://developer.apple.com/design/human-interface-guidelines) — estándar oficial iOS/iPadOS (navegación, sheets, layout, controles)
- [Material Design 3](https://m3.material.io) — estándar oficial Android (navigation bar, bottom sheets, tonal palettes, componentes)

> Para composición de pantallas móviles nativas, el banco condensado vive en `reference/mobile-patterns.md`; estas dos son la fuente primaria cuando se necesita más detalle.

## Uso

El `explorer` usa estas fuentes para investigar patrones del dominio y pasa los hallazgos inline en el prompt del diseñador. Nunca decirle al humano "busca en Dribbble" — el explorer investiga y entrega resultados.

### Formato estándar de entrega de la investigación del explorer

Cuando el explorer investiga para un diseño, entrega inline (no como paths ni "busca tú"):

1. **3-5 productos de referencia del dominio** — cada uno con qué imitar de él (ej. "Linear: densidad compacta y neutrales teñidos"; "Wise: cifras con tabular figures y verde solo para saldos")
2. **2-3 pairings de Google Fonts** (titular + cuerpo) con justificación de por qué encaja con el tono del dominio
3. **Paleta con hex concretos** — al menos primario y acento (más neutral/soporte si aplica), lista para generar la rampa con `reference/color-craft.md`

Sin estos tres bloques, la investigación está incompleta. Si el explorer no puede investigar, el fallback es `reference/domain-styles.md` (banco local por dominio).
