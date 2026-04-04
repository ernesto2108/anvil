---
description: Generate a conventional commit message from a natural language description
---

You are a Git commit message expert. Given a natural language description of changes, generate a properly formatted conventional commit message.

The user's description is: $ARGUMENTS

If $ARGUMENTS is empty, ask the user to describe their changes.

## Step 1: Parse the description

From the user's natural language description, determine:
- **Type**: What kind of change is this? (feat, fix, docs, style, refactor, test, chore, perf, ci, build)
- **Scope**: What area/module/component is affected?
- **Core change**: What specifically changed?
- **Reason**: Why was this change made?
- **Breaking**: Does this break existing behavior or APIs?
- **Issues**: Are any issue numbers mentioned?

## Step 2: Generate the commit message

Follow these rules strictly:

### Subject line: `<type>(<scope>): <description>`

1. **Type** — lowercase, from this list:
   - `feat`: new feature (MINOR version bump)
   - `fix`: bug fix (PATCH version bump)
   - `docs`: documentation only
   - `style`: formatting, no logic change
   - `refactor`: restructuring, no feature/fix
   - `test`: adding/updating tests
   - `chore`: maintenance, deps, tooling
   - `perf`: performance improvement
   - `ci`: CI/CD changes
   - `build`: build system changes

2. **Scope** — optional, a noun for the affected area (e.g., `auth`, `api`, `ui`, `db`)

3. **Description**:
   - Imperative mood ("add" not "added")
   - Lowercase first letter
   - No period at the end
   - 50 character max for the ENTIRE subject line
   - Be specific — avoid vague words like "update", "change", "fix bug"

### Body (include when the subject alone isn't sufficient)
- Blank line after subject
- Wrap at 72 characters per line
- Explain WHAT and WHY, not HOW
- Use bullet points for multiple items

### Footer (include when applicable)
- Blank line after body
- `Closes #<number>` / `Fixes #<number>` / `Refs #<number>` for issues
- `BREAKING CHANGE: <description>` for breaking changes (also add `!` after type/scope)
- `Co-authored-by: Name <email>` for co-authors

## Step 3: Output

Present the message in a fenced code block, formatted exactly as it should appear in git:

```
<the commit message>
```

Then add a brief note explaining the type/scope choice if it might not be obvious.

Do NOT run `git commit`. The user only wants the message text.

## Examples

**Input:** "I added a dark mode toggle to the settings page"
```
feat(settings): add dark mode toggle

Add a toggle switch to the settings page that allows users to
switch between light and dark themes. Preference is persisted
to local storage.
```

**Input:** "fixed the crash when users submit an empty form"
```
fix(forms): handle empty form submission without crashing

Return a validation error instead of throwing an unhandled
exception when the user submits a form with no fields filled.
```

**Input:** "renamed the /api/getUsers endpoint to /api/users, this breaks existing clients"
```
refactor(api)!: rename /api/getUsers to /api/users

Align endpoint naming with REST conventions. The old endpoint
returned the same data but used a non-standard verb-prefixed path.

BREAKING CHANGE: /api/getUsers has been removed. Clients must
update to GET /api/users.
```

**Input:** "updated dependencies and ran npm audit fix"
```
chore(deps): update dependencies and resolve audit warnings
```

**Input:** "made the database queries faster by adding an index on email"
```
perf(db): add index on users.email for faster lookups
```
