# API Consumption Patterns

Patterns for Astro projects that consume external APIs — structure, data transformation, error handling, and testing.

## Pods Architecture

When a project fetches data from external APIs, organize by **feature (pod)** instead of by file type. Each pod groups everything related to a single domain concept:

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

**Rules:**
- One pod per domain concept — never mix hero data with speakers logic
- Pages import from pods, pods never import from pages
- Components remain in `components/` — pods handle data, not UI
- Shared API utilities (base URL, headers, auth) go in `src/lib/api.ts`

**When to use Pods vs standard structure:**

| Use Pods | Use standard structure |
|---|---|
| Project consumes external APIs | Content-driven with markdown/MDX |
| Multiple domain entities from API | Using Content Collections |
| Team needs clear data boundaries | Simple static site |
| API responses need transformation | Data comes pre-formatted |

## Mappers (API → ViewModel)

Mappers decouple raw API responses from what the UI needs. The API can change shape without breaking components.

### Types

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

### Mapper implementation

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

**Rules:**
- One mapper file per pod — `{pod}.mapper.ts`
- Pure functions only — no side effects, no fetch calls
- Handle every nullable field with a sensible default
- Never pass raw API types to components — always map first
- Mapper names follow `map{Entity}FromApiToVm` convention

## Defensive Fetching

API calls in frontmatter must handle failures gracefully. A broken API should not crash the entire page.

### Pattern

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

### Using in frontmatter

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

**Rules:**
- Every fetch wrapped in `try/catch` — no unhandled rejections in frontmatter
- Return `null` on failure, never throw from API functions
- Check `response.ok` before parsing JSON
- Log errors with context (which endpoint, what status)
- Pages must render a valid fallback when data is `null`
- Use `import.meta.env` for API base URLs — never hardcode

## Path Aliases for Pods

Configure clean imports to avoid deep relative paths:

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

## Testing Mappers

Mappers are pure functions — ideal for unit tests. Focus on edge cases: null fields, missing arrays, unexpected types.

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

**Rules:**
- Test every nullable field with `null` and `undefined`
- Test empty arrays and missing nested properties
- Use a complete base object and override one field per test
- Mapper tests must run without network — no API calls
