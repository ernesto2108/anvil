---
name: social-content
description: Complete social media content knowledge — copywriting frameworks, visual design, platform-specific rules (LinkedIn, Instagram, Twitter/X), algorithm mechanics, engagement strategy, color psychology, typography, and image creation. Used by mkt-content agent. Use when creating any social media content.
---

# Social Content

Everything the mkt-content agent needs to create text + visual for any social platform.

---

## 1. Copywriting Frameworks

### PAS (Problem-Agitate-Solution)
Best for cold audiences and pain-point content.
```
[Problem] Name the pain.
[Agitate] Intensify — consequences, frustration, wasted time/money.
[Solution] Your product/service as the bridge out.
```

### BAB (Before-After-Bridge)
Best for transformation stories and feature highlights.
```
[Before] The painful status quo.
[After] The better world with your solution.
[Bridge] How your product gets them there.
```

### AIDA (Attention-Interest-Desire-Action)
Best for longer posts and launch announcements.
```
[Attention] Hook that stops the scroll.
[Interest] Build with facts, story, or insight.
[Desire] Show the transformation / result.
[Action] Clear CTA.
```

### Story Arc
Best for building-in-public and personal content.
```
[Situation] Set the scene.
[Complication] What went wrong / the struggle.
[Resolution] What you learned / built / changed.
```

---

## 2. Hook Writing

The first line decides everything. Every platform truncates early — everything after is hidden.

### The 7-Word Rule

Create a **curiosity gap** in 7 words or fewer.

### Five Proven Patterns

| Pattern | Example | Why It Works |
|---------|---------|-------------|
| **Opposite of expected** | "Stop writing unit tests." | Disrupts assumptions |
| **Challenge the norm** | "Your marketing agency is wasting your money." | Tension with status quo |
| **Shared frustration** | "Every restaurant owner has done this at midnight." | Instant identification |
| **Personal/raw** | "I almost closed my business last month." | Vulnerability = connection |
| **Lead with result** | "One image replaced our 20-slide deck." | Concrete outcome |

### Hook Anti-Patterns (NEVER)

- "Excited to announce..." — self-centered, no curiosity
- "I'm thrilled to share..." — corporate fluff
- "Check out our new..." — pure promotion
- "Big news!" — vague
- Engagement bait ("Comment YES if...") — algorithmically penalized

---

## 3. Visual Design

### Philosophy: Less Is More

1 powerful image > 10 mediocre slides. Every element must earn its place.

### Color Psychology

| Color family | Communicates | Best for |
|-------------|-------------|----------|
| **Black/dark gray** | Premium, sophistication, tech | Developer tools, luxury, editorial |
| **White/light** | Clean, simple, modern | Minimalist brands, healthcare, SaaS |
| **Dark blue/navy** | Trust, authority | Finance, enterprise, B2B |
| **Red/coral** | Urgency, passion, energy | Sales, food, entertainment, CTAs |
| **Green** | Growth, health, success | Finance, health, sustainability |
| **Yellow/amber** | Optimism, warmth, attention | Creative, education |
| **Purple** | Creativity, premium | Design tools, creative agencies |
| **Blue/cyan** | Technology, trust, calm | Tech, social media, corporate |
| **Orange** | Energy, playfulness | Startups, food, community |

### Color Contrast Rules

- **Minimum 4.5:1** ratio for text over backgrounds (accessibility)
- **Dark bg + light text** = premium/tech feel
- **Light bg + dark text** = clean/professional feel
- **One accent color** for emphasis — not rainbow
- **Brand colors first** — use the user's palette if they have one

### Typography as Design

Text IS the visual. A well-set phrase with the right font, weight, and size is a complete image.

**When text alone works:** bold statements, numbers/stats, single-line hooks, brand messages.

**Rules:**
- One font family, two weights (regular + bold/black)
- Size hierarchy: title 48-72px, subtitle 24-36px, body 18-22px
- Line height: 1.2 titles, 1.5 body
- Letter spacing: tight for large titles (-0.02em), normal for body
- Max 2 fonts. Don't mix 3+
- Recommended: Inter, Plus Jakarta Sans, Satoshi (modern), JetBrains Mono (code), Playfair Display (editorial)

### Carousel Rules

Only when content genuinely needs sequence:
- **Max 3-5 slides** — every slide must justify existing
- **Slide 1 = the only one visible in feed** — spend 50% effort here
- **One idea per slide**
- **Last slide = CTA**

**When carousel is NOT the answer:** 1-2 ideas (use single image), single statement (use typography), slides just repeat the text (image should ADD, not repeat).

---

## 4. Image Creation

### HTML Template (primary — free, full control)

```html
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    width: 1080px;
    height: 1080px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px;
    font-family: 'Inter', sans-serif;
    background: var(--bg);
    color: var(--text);
    --bg: #0f172a;
    --text: #f8fafc;
    --accent: #38bdf8;
  }
  .headline {
    font-size: 64px;
    font-weight: 900;
    line-height: 1.1;
    text-align: center;
    letter-spacing: -0.02em;
  }
  .accent { color: var(--accent); }
  .subtitle {
    font-size: 24px;
    font-weight: 400;
    margin-top: 24px;
    opacity: 0.7;
    text-align: center;
  }
  .brand {
    position: absolute;
    bottom: 40px;
    font-size: 18px;
    font-weight: 700;
    opacity: 0.5;
  }
</style>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;700;900&display=swap" rel="stylesheet">
</head>
<body>
  <div class="headline">Your <span class="accent">headline</span> here.</div>
  <div class="subtitle">Supporting text</div>
  <div class="brand">brandname</div>
</body>
</html>
```

### Rendering

```bash
# Render to PNG
npx puppeteer screenshot image.html --viewport 1080x1080 --output image.png

# Batch render
for f in *.html; do
  npx puppeteer screenshot "$f" --viewport 1080x1080 --output "${f%.html}.png"
done
```

### AI Image Prompts (for editorial/abstract imagery)

```markdown
**Tool:** DALL-E 3 / Stable Diffusion XL
**Prompt:** "[composition, lighting, colors, style, mood]"
**Negative prompt:** "text, watermark, blurry, low quality" (SD only)
**Size:** [platform dimensions]
```

Rules: no text in AI images (unreliable), specify brand colors, describe composition not just subjects.

---

## 5. Platform References

### LinkedIn

**Dimensions:**

| Format | Size | When |
|--------|------|------|
| Single image | 1200 x 627 px | Standard posts |
| Square | 1080 x 1080 px | Bold statements |
| Carousel | 1080 x 1350 px (4:5) | Sequential content (max 3-5 slides) |
| Banner | 1584 x 396 px | Profile header |

**Algorithm (2025-2026):**

Ranking signals by weight:
1. **Saves** — #1 signal. Write reference-value content worth bookmarking
2. **Thoughtful comments** (3+ sentences) — 15x heavier than likes
3. **Dwell time** — carousels and long-form text win
4. **Shares** — secondary but strong
5. **Clicks** — count but links in body get deprioritized

Distribution:
- Tests with 2-5% of network first
- **First hour is critical** — reply to every comment
- Only 5% of underperforming posts recover
- Extended window: 3-8 hours

Penalized:
- External links in post body → say "Link in comments"
- Engagement bait → NLP detection penalizes
- Editing within 10 min → resets distribution
- Tagging non-engagers
- More than 1 post/day → cannibalization

**Post formats ranked:**

| Format | Engagement | Best for |
|--------|-----------|----------|
| Document/Carousel | 3x | Tutorials, lists, comparisons |
| Multi-image | 2.5x | Before/after, showcases |
| Long-form text | 2x | Stories, thought leadership |
| Native video (30-90s) | 1.4x | Demos, walkthroughs |
| Single image + text | 1x | Announcements |
| Text only (short) | 0.7x | Quick thoughts |
| Link post | 0.5x | Avoid |

**Post length:**
- Text: 1,000-1,300 characters
- With image: 500-800 characters
- Carousel caption: 300-500 characters

**Engagement strategy:**
1. Post → immediately add link as first comment
2. Reply to every comment in first 60 minutes
3. Ask follow-up questions in replies
4. 3-5 posts/week, never >1/day
5. Best times: Tue-Thu, 8-10 AM audience timezone

**Hashtags:** nearly irrelevant. Max 3-5 at end if used. Algorithm reads text via NLP.

### Instagram

**Dimensions:**

| Format | Size | When |
|--------|------|------|
| Square | 1080 x 1080 px | Feed posts |
| Portrait | 1080 x 1350 px | Feed posts (more real estate) |
| Story/Reel | 1080 x 1920 px (9:16) | Stories and Reels |
| Carousel | 1080 x 1080 px | Multi-slide (prefer 3-5) |

**Key differences from LinkedIn:**
- Visual-first — image quality matters more than text
- Hashtags still relevant (15-20, mix of sizes)
- Reels > static for reach
- Caption can be longer but front-load value
- Link in bio (no clickable links in posts)

### Twitter/X

**Dimensions:**

| Format | Size | When |
|--------|------|------|
| Single image | 1200 x 675 px (16:9) | Standard tweets |
| Square | 1200 x 1200 px | Bold statements |

**Key differences:**
- Short and punchy — 280 chars constraint shapes style
- Threads for depth (but each tweet must stand alone)
- Engagement happens fast — first 30 min matter most
- Retweets > likes for distribution
- Quote tweets spark discussion

### General Rules (all platforms)

- **RGB**, 72 DPI
- **Min text:** 22px body, 36px headings
- **Padding:** 80px all sides
- **Safe zone:** keep key content within central area
- **PNG** for quality, **JPG** for photos, **PDF** for LinkedIn carousels
- **Max 10 MB** for optimal mobile loading

---

## 6. Content Calendar Template

| Day | Type | Goal |
|-----|------|------|
| Monday | Thought leadership / contrarian take | Discussion, comments |
| Tuesday | Tutorial / how-to | Saves, reference value |
| Thursday | Behind the scenes / personal story | Connection, shares |
| Friday | Milestone / community spotlight | Reach, amplification |

---

## 7. Anti-Patterns

- **10 slides when 1 works** — more ≠ better
- **Text wall on image** — paragraphs go in the post body, not the image
- **Rainbow colors** — one accent, not a palette explosion
- **Decoration without purpose** — every element must communicate
- **Stock photo generic** — better no image than a generic stock photo
- **Tiny text** — if not readable on phone at arm's length, too small
- **Multiple focal points** — one dominant element, rest supports
- **Repeating the post in the image** — image should ADD information, not echo
- **Corporate fluff in copy** — no "excited to announce", no "revolutionary"
- **Link in post body** (LinkedIn) — always in comments
