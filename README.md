# shelf-cli

CLI client for [Shelf](https://github.com/justEstif/shelf), a self-hosted file sharing app.

## Install

```
go install github.com/justEstif/shelf-cli@latest
```

## Quick Start

```bash
export SHELF_API_TOKEN=your-token
export SHELF_URL=http://localhost:3000

# Upload a file (URL copied to clipboard)
shelf upload report.html

# Upload into a folder with protected visibility
shelf upload -f docs -v protected report.html

# List files
shelf ls

# Delete a file
shelf rm docs/old.html

# Open in browser
shelf open docs/report.html
```

## Commands

### `upload <file...>`

Upload one or more files. Prints the URL for each and copies the first to clipboard.

| Flag | Description |
|------|-------------|
| `-f, --folder` | Upload into a subfolder |
| `-v, --visibility` | Set visibility: `public` (default), `private`, or `protected` |

Alias: `up`

### `rm <path...>`

Delete files from the server.

### `ls`

List all files on the server.

### `open [path]`

Open a file in your default browser.

## Visibility

Shelf supports per-file access control. Visibility is set with `-v` on upload:

| Level | Meaning |
|-------|---------|
| `public` | Anyone with the link can view (default) |
| `private` | Requires admin login |
| `protected` | Requires a viewer password (configured via `SHELF_VIEWER_PASSWORD` on the server) |

Visibility inherits from parent directories — setting a folder to `private` makes all files inside private unless overridden.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SHELF_API_TOKEN` | Yes | — | API token from your Shelf admin dashboard |
| `SHELF_URL` | No | `http://localhost:3000` | Base URL of your Shelf server |

Both can also be passed as flags: `--token` and `--url`.

## Related

- [justEstif/shelf](https://github.com/justEstif/shelf) — the Shelf server
