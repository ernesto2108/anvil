# Patrones de Consumo de APIs

Patrones para proyectos Astro que consumen APIs externas — estructura, transformación de datos, manejo de errores y testing.

## Arquitectura de Pods

Cuando un proyecto obtiene datos de APIs externas, organiza por **feature (pod)** en lugar de por tipo de archivo. Cada pod agrupa todo lo relacionado con un concepto de dominio:

```
src/
├── pods/
│   ├── hero/
│   │   ├── hero.api.ts          # fetch calls
│   │   ├── hero.mapper.ts       # API → ViewModel
│   │   ├── hero.model.ts        # raw API types
│   │   ├── hero.vm.ts           # view model types
│   │   └── hero.test.ts         # mapper tests
│   ├── speakers/
│   │   ├── speakers.api.ts
│   │   ├── speakers.mapper.ts
│   │   ├── speakers.model.ts
│   │   ├── speakers.vm.ts
│   │   └── speakers.test.ts
│   └── schedule/
│       ├── schedule.api.ts
│       ├── schedule.mapper.ts
│       ├── schedule.model.ts
│       ├── schedule.vm.ts
│       └── schedule.test.ts
├── components/
│   ├── common/
│   └── features/
├── layouts/
├── pages/
└── styles/
```

**Reglas:**
- Un pod por concepto de dominio — nunca mezcles datos de hero con lógica de speakers
- Las páginas importan desde pods, los pods nunca importan desde páginas
- Los componentes permanecen en `components/` — los pods manejan datos, no UI
- Las utilidades de API compartidas (base URL, headers, auth) van en `src/lib/api.ts`

**Cuándo usar Pods vs estructura estándar:**

| Usar Pods | Usar estructura estándar |
|---|---|
| El proyecto consume APIs externas | Basado en contenido con markdown/MDX |
| Múltiples entidades de dominio desde la API | Usando Content Collections |
| El equipo necesita límites claros de datos | Sitio estático simple |
| Las respuestas de la API necesitan transformación | Los datos vienen pre-formateados |

## Mappers (API → ViewModel)

Los mappers desacoplan las respuestas crudas de la API de lo que necesita la UI. La API puede cambiar de forma sin romper los componentes.

### Tipos

```typescript
// src/pods/speakers/speakers.model.ts
// Raw shape from the API — matches exactly what the endpoint returns
export interface SpeakerApiModel {
  id: number;
  full_name: string;
  bio_text: string | null;
  avatar_url: string | null;
  social_links: { platform: string; url: string }[] | null;
  talk_title: string;
  talk_abstract: string | null;
}
```

```typescript
// src/pods/speakers/speakers.vm.ts
// Clean shape for the UI — only what components actually need
export interface SpeakerVm {
  id: string;
  name: string;
  bio: string;
  avatarUrl: string;
  twitterUrl: string;
  talkTitle: string;
  talkSummary: string;
}
```

### Implementación del mapper

```typescript
// src/pods/speakers/speakers.mapper.ts
import type { SpeakerApiModel } from "./speakers.model";
import type { SpeakerVm } from "./speakers.vm";

const DEFAULT_AVATAR = "/images/placeholder-avatar.png";
const DEFAULT_BIO = "Speaker bio coming soon.";

export function mapSpeakerFromApiToVm(api: SpeakerApiModel): SpeakerVm {
  const twitter = api.social_links?.find((l) => l.platform === "twitter");

  return {
    id: String(api.id),
    name: api.full_name ?? "TBA",
    bio: api.bio_text ?? DEFAULT_BIO,
    avatarUrl: api.avatar_url ?? DEFAULT_AVATAR,
    twitterUrl: twitter?.url ?? "",
    talkTitle: api.talk_title,
    talkSummary: api.talk_abstract ?? "",
  };
}

export function mapSpeakersFromApiToVm(
  apiList: SpeakerApiModel[]
): SpeakerVm[] {
  return apiList.map(mapSpeakerFromApiToVm);
}
```

**Reglas:**
- Un archivo mapper por pod — `{pod}.mapper.ts`
- Solo funciones puras — sin efectos secundarios, sin fetch calls
- Maneja cada campo nullable con un valor por defecto sensato
- Nunca pases tipos crudos de la API a los componentes — siempre mapea primero
- Los nombres de mappers siguen la convención `map{Entity}FromApiToVm`

## Fetch Defensivo

Las llamadas a la API en el frontmatter deben manejar fallos con gracia. Una API rota no debe crashear toda la página.

### Patrón

```typescript
// src/pods/hero/hero.api.ts
import type { HeroApiModel } from "./hero.model";

const API_BASE = import.meta.env.API_BASE_URL;

export async function getHero(): Promise<HeroApiModel | null> {
  try {
    const response = await fetch(`${API_BASE}/hero`);

    if (!response.ok) {
      console.error(`Hero API failed: ${response.status}`);
      return null;
    }

    return await response.json();
  } catch (error) {
    console.error("Hero API unreachable:", error);
    return null;
  }
}
```

### Uso en frontmatter

```astro
---
import { getHero } from "@pods/hero/hero.api";
import { mapHeroFromApiToVm } from "@pods/hero/hero.mapper";

const heroApi = await getHero();
const hero = heroApi ? mapHeroFromApiToVm(heroApi) : null;
---

{hero ? (
  <section class="hero">
    <h1>{hero.title}</h1>
    <p>{hero.subtitle}</p>
  </section>
) : (
  <section class="hero hero--fallback">
    <h1>Welcome</h1>
  </section>
)}
```

**Reglas:**
- Cada fetch envuelto en `try/catch` — sin rechazos sin manejar en el frontmatter
- Retorna `null` en caso de fallo, nunca lances desde funciones de API
- Verifica `response.ok` antes de parsear JSON
- Loguea errores con contexto (qué endpoint, qué status)
- Las páginas deben renderizar un fallback válido cuando los datos son `null`
- Usa `import.meta.env` para las URLs base de la API — nunca las hardcodees

## Path Aliases para Pods

Configura imports limpios para evitar rutas relativas profundas:

```json
// tsconfig.json
{
  "extends": "astro/tsconfigs/strict",
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@pods/*": ["src/pods/*"],
      "@components/*": ["src/components/*"],
      "@layouts/*": ["src/layouts/*"],
      "@lib/*": ["src/lib/*"]
    }
  }
}
```

```astro
---
// Clean imports with aliases
import { getSpeakers } from "@pods/speakers/speakers.api";
import { mapSpeakersFromApiToVm } from "@pods/speakers/speakers.mapper";
import SpeakerCard from "@components/features/SpeakerCard.astro";

// NOT this:
// import { getSpeakers } from "../../../pods/speakers/speakers.api";
---
```

## Testing de Mappers

Los mappers son funciones puras — ideales para tests unitarios. Enfócate en casos límite: campos null, arrays vacíos, tipos inesperados.

```typescript
// src/pods/speakers/speakers.test.ts
import { describe, test, expect } from "vitest";
import { mapSpeakerFromApiToVm } from "./speakers.mapper";
import type { SpeakerApiModel } from "./speakers.model";

const baseSpeaker: SpeakerApiModel = {
  id: 1,
  full_name: "Ada Lovelace",
  bio_text: "Pioneer of computing.",
  avatar_url: "https://example.com/ada.jpg",
  social_links: [{ platform: "twitter", url: "https://twitter.com/ada" }],
  talk_title: "The Analytical Engine",
  talk_abstract: "A deep dive into early computation.",
};

describe("speakers mapper", () => {
  test("maps complete API response to view model", () => {
    const result = mapSpeakerFromApiToVm(baseSpeaker);

    expect(result.id).toBe("1");
    expect(result.name).toBe("Ada Lovelace");
    expect(result.bio).toBe("Pioneer of computing.");
    expect(result.twitterUrl).toBe("https://twitter.com/ada");
  });

  test("handles null bio with default", () => {
    const result = mapSpeakerFromApiToVm({ ...baseSpeaker, bio_text: null });
    expect(result.bio).toBe("Speaker bio coming soon.");
  });

  test("handles null avatar with placeholder", () => {
    const result = mapSpeakerFromApiToVm({ ...baseSpeaker, avatar_url: null });
    expect(result.avatarUrl).toBe("/images/placeholder-avatar.png");
  });

  test("handles null social links", () => {
    const result = mapSpeakerFromApiToVm({
      ...baseSpeaker,
      social_links: null,
    });
    expect(result.twitterUrl).toBe("");
  });

  test("handles empty social links array", () => {
    const result = mapSpeakerFromApiToVm({
      ...baseSpeaker,
      social_links: [],
    });
    expect(result.twitterUrl).toBe("");
  });
});
```

**Reglas:**
- Testea cada campo nullable con `null` y `undefined`
- Testea arrays vacíos y propiedades anidadas faltantes
- Usa un objeto base completo y sobreescribe un campo por test
- Los tests de mappers deben correr sin red — sin llamadas a la API
