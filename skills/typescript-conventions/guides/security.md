# Security Guide

## XSS Prevention

Never use `innerHTML`, `outerHTML`, or `document.write` with untrusted content. In React, avoid `dangerouslySetInnerHTML`. If you must render HTML, always sanitize first.

```typescript
// WRONG: XSS vector — attacker controls content
element.innerHTML = userContent;
element.innerHTML = `<div>${comment.body}</div>`;
document.write(serverData);

// WRONG: React XSS via dangerouslySetInnerHTML without sanitization
function Comment({ body }: { body: string }) {
  return <div dangerouslySetInnerHTML={{ __html: body }} />; // DO NOT DO THIS
}

// RIGHT: text content only — React escapes automatically
function Comment({ body }: { body: string }) {
  return <div>{body}</div>; // safe — React renders as text node
}

// RIGHT: when HTML rendering is unavoidable — sanitize with DOMPurify
import DOMPurify from "dompurify";

const SANITIZE_CONFIG = {
  ALLOWED_TAGS: ["b", "i", "em", "strong", "a", "p", "ul", "ol", "li"],
  ALLOWED_ATTR: ["href", "target", "rel"],
  ALLOW_DATA_ATTR: false,
};

function SafeHtml({ html }: { html: string }): React.ReactElement {
  const sanitized = DOMPurify.sanitize(html, SANITIZE_CONFIG);
  // Force noopener/noreferrer on all links after sanitization
  return <div dangerouslySetInnerHTML={{ __html: sanitized }} />;
}

// RIGHT: use textContent for dynamic DOM manipulation
element.textContent = userContent; // safe — treated as text
```

## Input Sanitization

Validate and sanitize at every trust boundary. Use Zod for structure validation; sanitize free-form text inputs for downstream use.

```typescript
// WRONG: raw user input used directly in query/HTML
const query = req.query.search; // could be anything
const result = await db.execute(`SELECT * FROM products WHERE name LIKE '%${query}%'`);

// RIGHT: validate with Zod, use parameterized queries
const searchSchema = z.string().trim().min(1).max(200);
const query = searchSchema.parse(req.query.search);
const result = await db.execute(
  "SELECT * FROM products WHERE name LIKE $1",
  [`%${query}%`] // parameterized — SQL injection impossible
);

// URL validation — never redirect to arbitrary URLs
function safeRedirect(url: string, allowedOrigins: string[]): string {
  try {
    const parsed = new URL(url);
    if (!allowedOrigins.includes(parsed.origin)) {
      return "/"; // fallback to safe default
    }
    return url;
  } catch {
    return "/"; // invalid URL — redirect to home
  }
}

// Filename sanitization for file uploads
function sanitizeFilename(name: string): string {
  return name
    .replace(/[^a-zA-Z0-9._-]/g, "_") // allow only safe chars
    .replace(/\.{2,}/g, "_")           // no path traversal (..)
    .slice(0, 255);                    // max filename length
}
```

## CSRF Protection

For cookie-based sessions, require CSRF tokens on state-mutating requests. For JWT-based APIs (Authorization header), CSRF is not applicable.

```typescript
// Server: set CSRF token in cookie + validate on POST/PUT/DELETE/PATCH
import csrf from "csrf";

const tokens = new csrf();
const secret = await tokens.secret();

// Set in cookie (httpOnly: false so JS can read it)
res.cookie("csrf-token", tokens.create(secret), {
  sameSite: "strict",
  secure: true,
});
// Store secret server-side (session or DB)
req.session.csrfSecret = secret;

// Validate on mutations
function csrfMiddleware(req: Request, res: Response, next: NextFunction) {
  if (["GET", "HEAD", "OPTIONS"].includes(req.method)) return next();

  const token = req.headers["x-csrf-token"] as string | undefined;
  if (!token || !tokens.verify(req.session.csrfSecret, token)) {
    return res.status(403).json({ error: "Invalid CSRF token" });
  }
  next();
}

// Client: attach token from cookie to every mutating request
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

Set CSP headers on every HTML response. Never use `unsafe-inline` or `unsafe-eval` in production.

```typescript
// Next.js — next.config.ts
const cspHeader = [
  "default-src 'self'",
  "script-src 'self' 'nonce-{NONCE}'",       // use nonces, not unsafe-inline
  "style-src 'self' 'nonce-{NONCE}'",
  "img-src 'self' data: https://cdn.example.com",
  "font-src 'self'",
  "connect-src 'self' https://api.example.com",
  "frame-ancestors 'none'",                   // prevents clickjacking
  "base-uri 'self'",
  "form-action 'self'",
  "upgrade-insecure-requests",
].join("; ");

// Nonce-based approach for inline scripts (when truly needed)
function generateNonce(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return btoa(String.fromCharCode(...bytes));
}

// Express middleware
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

## Secrets and Sensitive Data

```typescript
// WRONG: secrets in source code
const API_KEY = "sk-proj-abc123"; // committed to git

// WRONG: secrets in localStorage (accessible by JS, not encrypted)
localStorage.setItem("authToken", token);

// RIGHT: load secrets from environment at startup (validated with Zod)
const env = loadEnv(); // throws if missing — see rules/architecture.md
const apiKey = env.STRIPE_SECRET_KEY;

// RIGHT: auth tokens in httpOnly cookies — not accessible via JS
res.cookie("auth-token", jwt, {
  httpOnly: true,      // not accessible via document.cookie
  secure: true,        // HTTPS only
  sameSite: "lax",     // CSRF mitigation
  maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
});

// WRONG: logging sensitive fields
console.log("User logged in", { email, password, token }); // never log credentials

// RIGHT: log only safe identifiers
logger.info("User logged in", { userId: user.id, email: user.email });
```
