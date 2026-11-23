/*
  PANTALLAS (diseño) — única parte que se edita al agregar pantallas.
  Contiene: mocks de datos, notas, GROUPS (pantallas por journey), ALL y el router (sección "NO EDITAR").
  Script clásico: requiere que components.js ya esté cargado (usa btn, field, card, badge, page).

  Reglas:
    - Solo se compone con helpers de components.js. Prohibido markup o estilos de componente inline aquí:
      si falta un componente, se agrega primero a components.js (+ components.css) y luego se usa.
    - Cada pantalla = una entrada de GROUPS con id estable (kebab-case). Los CTAs navegan con data-goto="<id>"
      (helper btn(..., { goto })). Estados/variantes (dark, vacío, error, expandido) = otra entrada o parámetro de r().
    - Contenido real del Design Spec; sin dominio inventado. Los mocks de datos viven en MOCK.
*/

/* ================== Mocks de datos ================== */
/* Reemplazar por datos reales del Design Spec. */
const MOCK = {
  items: ['Elemento 1', 'Elemento 2', 'Elemento 3'],
};

/* ================== Registro de pantallas ================== */
/* n = notas del panel: c (qué cambia), a (supuestos), p (pendientes). r = render. */
const N = (c = [], a = [], p = []) => ({ c, a, p });
const GROUPS = [
  { id: 'ejemplo', title: '1 · Ejemplo', pill: 'Starter', screens: [
    { id: 'ej-login', label: 'Login',
      r: () => page('Iniciar sesión', card(`<div class="stack">${field('Correo', { ph: 'nombre@dominio.com' })}${field('Contraseña')}${btn('primary full', 'Entrar', { goto: 'ej-lista' })}${btn('outlined full', 'Crear cuenta', { goto: 'ej-registro' })}</div>`)),
      n: N(['Pantalla de ejemplo: reemplazar.'], [], ['Definir copy final.']) },
    { id: 'ej-registro', label: 'Registro',
      r: () => page('Crear cuenta', card(`<div class="stack">${field('Nombre')}${field('Correo', { err: 'Correo inválido' })}${btn('primary full', 'Registrarme', { goto: 'ej-lista' })}</div>`)) },
    { id: 'ej-lista', label: 'Lista',
      r: () => page('Elementos', `<div class="grid cols-3">${MOCK.items.map(t => card(`<div class="stack"><b>${t}</b>${badge('success', 'Activo')}${btn('outlined', 'Ver detalle', { goto: 'ej-login' })}</div>`)).join('')}</div>`) },
  ] },
];
const ALL = GROUPS.flatMap(g => g.screens.map(s => ({ ...s, g })));

/* ================== Router — NO EDITAR ================== */
/* Requiere los elementos de index.html (#side, #canvas, #notes, #scrTitle, #scrBadge, #scroll, #prev, #next, #themeBtn, #modeBtn). */
const $ = (s) => document.querySelector(s);
let current = ALL[0];
function buildSide() {
  $('#side').innerHTML = `<h1>Maqueta</h1>` + GROUPS.map(g => `<div class="grp"><h2>${g.title}</h2>${g.screens.map(s => `<button type="button" class="scr" data-id="${s.id}">${s.label}</button>`).join('')}</div>`).join('');
}
function renderNotes(n = N()) {
  const list = (a) => a.length ? `<ul>${a.map(x => `<li>${x}</li>`).join('')}</ul>` : '<ul><li>—</li></ul>';
  $('#notes').innerHTML = `<summary>Notas de la pantalla</summary><div class="cols"><div><h4>Qué cambia</h4>${list(n.c)}</div><div><h4>Supuestos</h4>${list(n.a)}</div><div><h4>Pendientes</h4>${list(n.p)}</div></div>`;
}
function show(id) {
  const s = ALL.find(x => x.id === id) || ALL[0];
  current = s;
  $('#canvas').innerHTML = s.r();
  $('#scrTitle').textContent = s.label;
  $('#scrBadge').textContent = s.g.pill || '';
  renderNotes(s.n);
  document.title = `${s.label} · Maqueta`;
  document.querySelectorAll('.scr').forEach(b => b.dataset.id === s.id ? b.setAttribute('aria-current', 'page') : b.removeAttribute('aria-current'));
  if (location.hash.slice(1) !== s.id) history.replaceState(null, '', '#' + s.id);
}
const go = (d) => { show(ALL[(ALL.indexOf(current) + d + ALL.length) % ALL.length].id); $('#scroll').scrollTop = 0; };

document.addEventListener('click', (e) => {
  const scr = e.target.closest('.scr'); if (scr) return show(scr.dataset.id);
  const to = e.target.closest('[data-goto]'); if (to) show(to.dataset.goto);
});
$('#prev').addEventListener('click', () => go(-1));
$('#next').addEventListener('click', () => go(1));
document.addEventListener('keydown', (e) => {
  if (e.target.matches('input, textarea')) return;
  if (e.key === 'ArrowLeft') go(-1);
  if (e.key === 'ArrowRight') go(1);
});
window.addEventListener('hashchange', () => show(location.hash.slice(1)));

/* Tema (color primario alterno) y modo claro/oscuro */
(function () {
  const t = $('#themeBtn'), m = $('#modeBtn');
  t.addEventListener('click', () => { const alt = document.body.dataset.theme === 'alt'; document.body.dataset.theme = alt ? '' : 'alt'; t.textContent = 'Tema: ' + (alt ? 'A' : 'B'); });
  m.addEventListener('click', () => { const dark = document.body.dataset.mode === 'dark'; document.body.dataset.mode = dark ? '' : 'dark'; m.textContent = 'Modo: ' + (dark ? 'claro' : 'oscuro'); });
})();

/* Auditoría de CTAs (consola): lista destinos data-goto que no existen en ALL */
window.auditCtas = () => {
  const ids = new Set(ALL.map(s => s.id)), orphans = [];
  ALL.forEach(s => { const d = document.createElement('div'); d.innerHTML = s.r(); d.querySelectorAll('[data-goto]').forEach(el => { if (!ids.has(el.dataset.goto)) orphans.push(`${s.id} → ${el.dataset.goto}`); }); });
  return orphans;
};

buildSide();
show(location.hash.slice(1) || ALL[0].id);
