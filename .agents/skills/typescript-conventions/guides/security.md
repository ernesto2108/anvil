# Guía de Seguridad

## Prevención de XSS

Nunca usar `innerHTML`, `outerHTML` o `document.write` con contenido no confiable. En React, evitar `dangerouslySetInnerHTML`. Si es imprescindible renderizar HTML, siempre sanitizar primero.

```typescript
// INCORRECTO: vector de XSS — el atacante controla el contenido
element.innerHTML = userContent;
element.innerHTML = `<div>${comment.body}</div>`;
document.write(serverData);

// INCORRECTO: XSS en React via dangerouslySetInnerHTML sin sanitización
function Comment({ body }: { body: string }) {
  return <div dangerouslySetInnerHTML={{ __html: body }} />; // NO HACER ESTO
}

// CORRECTO: solo text content — React escapa automáticamente
function Comment({ body }: { body: string }) {
  return <div>{body}</div>; // seguro — React renderiza como nodo de texto
}

// CORRECTO: cuando el renderizado de HTML es inevitable — sanitizar con DOMPurify
import DOMPurify from "dompurify";

const SANITIZE_CONFIG = {
  ALLOWED_TAGS: ["b", "i", "em", "strong", "a", "p", "ul", "ol", "li"],
  ALLOWED_ATTR: ["href", "target", "rel"],
  ALLOW_DATA_ATTR: false,
};

function SafeHtml({ html }: { html: string }): React.ReactElement {
  const sanitized = DOMPurify.sanitize(html, SANITIZE_CONFIG);
  // Forzar noopener/noreferrer en todos los links después de la sanitización
  return <div dangerouslySetInnerHTML={{ __html: sanitized }} />;
}

// CORRECTO: usar textContent para manipulación dinámica del DOM
element.textContent = userContent; // seguro — tratado como texto
```

## Sanitización de Entradas

Validar y sanitizar en cada frontera de confianza. Usar Zod para validación de estructura; sanitizar entradas de texto libre para uso posterior.

```typescript
// INCORRECTO: entrada del usuario usada directamente en query/HTML
const query = req.query.search; // puede ser cualquier cosa
const result = await db.execute(`SELECT * FROM products WHERE name LIKE '%${query}%'`);

// CORRECTO: validar con Zod, usar queries parametrizadas
const searchSchema = z.string().trim().min(1).max(200);
const query = searchSchema.parse(req.query.search);
const result = await db.execute(
  "SELECT * FROM products WHERE name LIKE $1",
  [`%${query}%`] // parametrizado — inyección SQL imposible
);

// Validación de URL — nunca redirigir a URLs arbitrarias
function safeRedirect(url: string, allowedOrigins: string[]): string {
  try {
    const parsed = new URL(url);
    if (!allowedOrigins.includes(parsed.origin)) {
      return "/"; // fallback al default seguro
    }
    return url;
  } catch {
    return "/"; // URL inválida — redirigir al inicio
  }
}

// Sanitización de nombre de archivo para uploads
function sanitizeFilename(name: string): string {
  return name
    .replace(/[^a-zA-Z0-9._-]/g, "_") // permitir solo caracteres seguros
    .replace(/\.{2,}/g, "_")           // sin traversal de ruta (..)
    .slice(0, 255);                    // longitud máxima de nombre de archivo
}
```

## Protección CSRF

Para sesiones basadas en cookies, requerir tokens CSRF en solicitudes que mutan estado. Para APIs basadas en JWT (cabecera Authorization), CSRF no aplica.

```typescript
// Servidor: establecer token CSRF en cookie + validar en POST/PUT/DELETE/PATCH
import csrf from "csrf";

const tokens = new csrf();
const secret = await tokens.secret();

// Establecer en cookie (httpOnly: false para que JS pueda leerlo)
res.cookie("csrf-token", tokens.create(secret), {
  sameSite: "strict",
  secure: true,
});
// Almacenar secret en el servidor (sesión o DB)
req.session.csrfSecret = secret;

// Validar en mutaciones
function csrfMiddleware(req: Request, res: Response, next: NextFunction) {
  if (["GET", "HEAD", "OPTIONS"].includes(req.method)) return next();

  const token = req.headers["x-csrf-token"] as string | undefined;
  if (!token || !tokens.verify(req.session.csrfSecret, token)) {
    return res.status(403).json({ error: "Invalid CSRF token" });
  }
  next();
}

// Cliente: adjuntar token de la cookie en cada solicitud mutante
function getCsrfToken(): string {
  const match = document.cookie.match(/csrf-token=([^;]+)/);
  return match?.[1] ?? "";
}

async function apiPost(url: string, body: unknown): Promise<Response> {
  return fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": getCsrfToken(),
    },
    body: JSON.stringify(body),
    credentials: "same-origin",
  });
}
```

## Content Security Policy (CSP)

Establecer cabeceras CSP en cada respuesta HTML. Nunca usar `unsafe-inline` o `unsafe-eval` en producción.

```typescript
// Next.js — next.config.ts
const cspHeader = [
  "default-src 'self'",
  "script-src 'self' 'nonce-{NONCE}'",       // usar nonces, no unsafe-inline
  "style-src 'self' 'nonce-{NONCE}'",
  "img-src 'self' data: https://cdn.example.com",
  "font-src 'self'",
  "connect-src 'self' https://api.example.com",
  "frame-ancestors 'none'",                   // previene clickjacking
  "base-uri 'self'",
  "form-action 'self'",
  "upgrade-insecure-requests",
].join("; ");

// Enfoque basado en nonces para scripts inline (cuando sea realmente necesario)
function generateNonce(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return btoa(String.fromCharCode(...bytes));
}

// Middleware de Express
function cspMiddleware(req: Request, res: Response, next: NextFunction) {
  const nonce = generateNonce();
  res.locals.nonce = nonce;
  res.setHeader(
    "Content-Security-Policy",
    cspHeader.replaceAll("{NONCE}", nonce)
  );
  res.setHeader("X-Frame-Options", "DENY");
  res.setHeader("X-Content-Type-Options", "nosniff");
  res.setHeader("Referrer-Policy", "strict-origin-when-cross-origin");
  next();
}
```

## Secretos y Datos Sensibles

```typescript
// INCORRECTO: secretos en el código fuente
const API_KEY = "sk-proj-abc123"; // commiteado en git

// INCORRECTO: secretos en localStorage (accesible por JS, no encriptado)
localStorage.setItem("authToken", token);

// CORRECTO: cargar secretos desde el entorno al inicio (validados con Zod)
const env = loadEnv(); // lanza si falta — ver rules/architecture.md
const apiKey = env.STRIPE_SECRET_KEY;

// CORRECTO: tokens de auth en cookies httpOnly — no accesibles via JS
res.cookie("auth-token", jwt, {
  httpOnly: true,      // no accesible via document.cookie
  secure: true,        // solo HTTPS
  sameSite: "lax",     // mitigación CSRF
  maxAge: 7 * 24 * 60 * 60 * 1000, // 7 días
});

// INCORRECTO: registrar campos sensibles en logs
console.log("User logged in", { email, password, token }); // nunca loguear credenciales

// CORRECTO: loguear solo identificadores seguros
logger.info("User logged in", { userId: user.id, email: user.email });
```
