# Starter Template Setup Notes

## What lives where

The starter template source tree is committed to the main `tuikit-go` repo under
`templates/starter/`. This keeps it versioned alongside the library so examples
stay in sync with the API.

A **separate GitHub repository** (`tuikit-go-starter`) will be created to host the
template as a proper GitHub Template Repository. Users can then run:

```bash
gh repo create my-app --template moneycaringcoder/tuikit-go-starter
```

## Flipping the Template flag

When the `tuikit-go-starter` repo is created, enable the Template Repository flag:

1. Go to **Settings → General** on the repo page.
2. Check **Template repository**.
3. Save.

Or via the GitHub CLI (requires admin token):

```bash
gh api repos/moneycaringcoder/tuikit-go-starter \
  --method PATCH \
  --field is_template=true
```

## Sync workflow

When `templates/starter/` is updated in `tuikit-go`, mirror the changes to
`tuikit-go-starter`. The recommended approach is a simple copy script or a GitHub
Actions workflow in `tuikit-go` that pushes `templates/starter/` contents to the
separate repo on merge to `main`.

Example workflow trigger (add to `.github/workflows/sync-starter.yml`):

```yaml
on:
  push:
    branches: [main]
    paths:
      - templates/starter/**
```

## Placeholder substitutions

The template uses the following placeholders that consumers must replace:

| Placeholder | Replace with |
|---|---|
| `OWNER` | GitHub username / org |
| `REPO` / `myapp` | Repository / binary name |
| `YEAR AUTHOR` | Current year and your name (in LICENSE) |

A `sed` one-liner for the common case:

```bash
find . -type f | xargs sed -i \
  -e 's/OWNER/yourname/g' \
  -e 's/myapp/yourapp/g'
```
