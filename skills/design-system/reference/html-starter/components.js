/*
  COMPONENTES (helpers JS) — único lugar de componentes reutilizables de la maqueta.
  Script clásico (sin módulos ES, para que funcione abriendo file:// sin servidor ni build):
  las declaraciones top-level quedan disponibles para screens.js, que se carga DESPUÉS.

  Reglas:
    - Una función por componente; devuelve HTML. Variantes = parámetros. Estilos en components.css.
    - Componente nuevo (shell, tabla, stepper, modal, íconos nuevos…) = se agrega AQUÍ, nunca inline en screens.js.
    - Sin pantallas, sin datos de dominio, sin router (eso es screens.js).
    - Si el proyecto ya tiene librería de componentes (existing_design_system: yes): cada helper se
      mapea a la librería existente (mismos nombres/props/clases) en vez de duplicar markup propio.
*/

/* Íconos: set mínimo. Ampliar con el set del proyecto (documentar su nombre para los devs). */
const ICONS = {
  check: '<path d="M5 12.5l4.5 4.5L19 7.5"/>',
  x: '<path d="M6 6l12 12M18 6L6 18"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  chev: '<path d="M9 6l6 6-6 6"/>',
};
const ic = (n, s) => `<svg class="i" viewBox="0 0 24 24" aria-hidden="true"${s ? ` style="font-size:${s}px"` : ''}>${ICONS[n] || ''}</svg>`;

/* goto: id de la pantalla destino del CTA (se valida contra ALL en auditCtas(), definido en screens.js) */
const btn = (cls, text, { icon, goto, disabled } = {}) =>
  `<button type="button" class="btn ${cls}"${goto ? ` data-goto="${goto}"` : ''}${disabled ? ' aria-disabled="true"' : ''}>${icon ? ic(icon, 18) : ''}${text}</button>`;
const badge = (kind, text) => `<span class="badge ${kind}">${text}</span>`;
const field = (label, { value = '', err = '', ph = '' } = {}) =>
  `<div class="field"><label>${label}</label><input class="input${err ? ' err' : ''}" value="${value}" placeholder="${ph}">${err ? `<span class="help err">${err}</span>` : ''}</div>`;
const card = (inner) => `<section class="card">${inner}</section>`;
const page = (title, inner) => `<div class="page"><h1 class="pg">${title}</h1>${inner}</div>`;
