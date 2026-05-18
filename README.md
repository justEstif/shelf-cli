# shelf-cli

CLI client for [Shelf](https://github.com/justEstif/shelf), a self-hosted file sharing app for HTML artifacts.

## Install

```
go install github.com/justEstif/shelf-cli@latest
```

## Quick Start

```
export SHELF_API_TOKEN=your-token

# Upload a file (URL copied to clipboard)
shelf upload dist/index.html

# Upload into a folder with public visibility
shelf upload -f my-project -v public dist/index.html dist/app.js

# List files
shelf ls

# Delete a file
shelf rm my-project/old.html

# Open in browser
shelf open my-project/index.html
```

## Commands

### `upload <file...>`

Upload one or more files. Prints the URL for each and copies the first to clipboard.

| Flag | Description |
|------|-------------|
| `-f, --folder` | Upload into a subfolder |
| `-v, --visibility` | Set visibility: `public`, `private`, or `protected` |

Alias: `up`

### `rm <path...>`

Delete files from the server.

### `ls`

List all files on the server.

### `open [path]`

Open a file in your default browser.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SHELF_API_TOKEN` | Yes | — | API token from your Shelf instance |
| `SHELF_URL` | No | `https://shelf.estifanos.cc` | Base URL of your Shelf server |

Both can also be passed as flags: `--token` and `--url`.

## Related

- [justEstif/shelf](https://github.com/justEstif/shelf) — the Shelf server
