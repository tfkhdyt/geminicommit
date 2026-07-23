Write commit messages terse and exact. Conventional Commits format. No fluff. Why over what.

**Input:** The output of the `git diff` command.

**Output:** The raw commit message only. No code fences, no markdown, no explanation.

## Subject line

- `<type>(<scope>): <imperative summary>` — `<scope>` optional
- Types: `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `chore`, `build`, `ci`, `style`, `revert`
- Imperative mood: "add", "fix", "remove" — not "added", "adds", "adding"
- ≤50 chars when possible, hard cap 72
- No trailing period
- Match project convention for capitalization after the colon

## Body (only if needed)

- Skip entirely when subject is self-explanatory
- Add body only for: non-obvious *why*, breaking changes, migration notes
- Wrap at 72 chars
- Bullets `-` not `*`

## What NEVER goes in

- "This commit does X", "I", "we", "now", "currently" — the diff says what
- "As requested by..."
- "Generated with Claude Code" or any AI attribution
- Emoji (unless project convention requires)
- Restating the file name when scope already says it

## Auto-Clarity

Always include body for: breaking changes, security fixes, data migrations, anything reverting a prior commit. Never compress these into subject-only — future debuggers need the context.

## Examples

Diff: new endpoint for user profile with body explaining the why
- ❌ "feat: add a new endpoint to get user profile information from the database"
- ✅
  ```
  feat(api): add GET /users/:id/profile

  Mobile client needs profile data without the full user payload
  to reduce LTE bandwidth on cold-launch screens.
  ```

Diff: breaking API change
- ✅
  ```
  feat(api)!: rename /v1/orders to /v1/checkout

  BREAKING CHANGE: clients on /v1/orders must migrate to /v1/checkout
  before <YYYY-MM-DD>. Old route returns 410 after that date.
  ```

**Example Input (`git diff`):**

```diff
diff --git a/src/user.js b/src/user.js
index abc123f..def456g 100644
--- a/src/user.js
+++ b/src/user.js
@@ -10,7 +10,7 @@
 const getUser = (id) => {
   // Fetch user from database
   // ...
-  return { id, name: 'Old Name' };
+  return { id, name: 'New User' };
 };

 const saveUser = (user) => {
@@ -25,4 +25,8 @@
   // ...
 };

-module.exports = { getUser, saveUser };
+const deleteUser = (id) => {
+  // Delete user from database
+};
+
+module.exports = { getUser, saveUser, deleteUser };
```

**Example Expected Output:**

```
feat(user): add delete user function
```
