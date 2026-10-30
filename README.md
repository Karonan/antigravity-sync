<div align="center">

# Antigravity Sync

*Encrypted with [age](https://github.com/FiloSottile/age) • Cloudflare R2 (10GB Free Tier)*

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[Install](#install) • [Quick Start](#quick-start) • [Setup Guide](#setup-guide) • [What Gets Synced](#what-gets-synced) • [Commands](#commands) • [Architecture](#architecture)

</div>

---

## Features

- **Cross-device sync**: Seamlessly continue Antigravity conversations, artifacts, and project workspaces on any device.
- **Solves "Outside of Project"**: Automatically synchronizes `~/.gemini/config/projects/` definitions so conversations remain tied to their proper project workspaces across all machines.
- **Cloudflare R2 storage**: Built for Cloudflare R2 (10GB free tier, zero egress fees) with high throughput and S3 API compatibility.
- **End-to-end encryption**: All files encrypted client-side using `age` (X25519 + ChaCha20-Poly1305) before upload.
- **Passphrase-based keys**: Same passphrase = identical encryption key derived via Argon2id on any machine (no manual key transfer required).
- **Selective sync**: Choose `--scope sessions` for conversation databases, summaries, artifacts, and project registry; or `--scope full` to include custom skills and settings.
- **Interactive wizard**: Arrow-key driven setup wizard (`antigravity-sync init`) with connection and encryption testing.
- **Smart conflict handling**: Three-way conflict detection that safely creates `.conflict.<timestamp>` files rather than overwriting changes.
- **Gzip compression**: Transparent compression before encryption for rapid transfers.
- **Process & SQLite safety**: Detects active Antigravity instances to safeguard against SQLite WAL locks and memory cache overwrites.

---

## Install

Requires [Go 1.24+](https://go.dev/dl/):

```bash
go install github.com/Karonan/antigravity-sync/cmd/antigravity-sync@main
```

This builds the latest commit and puts `antigravity-sync` in Go's bin folder (`%USERPROFILE%\go\bin` on Windows, `~/go/bin` elsewhere), which is usually on your `PATH`. Later, `antigravity-sync update` rebuilds from the latest commit and replaces itself.

For a device without Go, run `make build-all` on one that has it and copy the matching binary from `bin/`.

---

## Quick Start

### First Device

```bash
# Run setup wizard
antigravity-sync init

# Push your conversations and projects
antigravity-sync push
```

### Second Device

```bash
# Run setup wizard with SAME credentials and passphrase
antigravity-sync init

# Preview changes safely
antigravity-sync pull --dry-run

# Pull state
antigravity-sync pull
```

---

## What Gets Synced

| Path | Content | Purpose |
|------|---------|---------|
| `~/.gemini/antigravity/conversations/` | SQLite conversation databases (`<id>.db`) | Full chat trajectory & message history |
| `~/.gemini/antigravity/conversation_summaries.db` | SQLite catalog table | Titles, project IDs, and metadata |
| `~/.gemini/antigravity/agyhub_summaries_proto.pb` | Summary cache | Fast sidebar loading |
| `~/.gemini/antigravity/brain/` | Artifacts, logs, transcripts | Markdown documents, task plans, logs |
| `~/.gemini/config/projects/` | Project JSON definitions (`<id>.json`) | **Prevents "Outside of Project" demotion** |
| `~/.gemini/config/skills/` | Custom agent skills | User skills & tools |
| `~/.gemini/config/plugins/` | Custom plugins | Extensible workflows |
| `~/.gemini/config/config.json` | Configuration file | Global preferences |
| `~/.gemini/config/mcp_config.json` | MCP configuration (opt-in: `antigravity-sync mcp enable`) | Merged per server, home paths translated |
| `~/.gemini/GEMINI.md` | Global prompt instructions | System-wide custom rules |

### Sync Scopes

`init` allows selecting your sync scope:
- **`full` (default)**: Syncs everything in the table above.
- **`sessions`**: Syncs `conversations/`, `conversation_summaries.db`, `agyhub_summaries_proto.pb`, `brain/`, and `config/projects/`.

---

## Commands

| Command | Description |
|---------|-------------|
| `antigravity-sync init` | Interactive setup wizard (storage, credentials, passphrase) |
| `antigravity-sync push` | Encrypt & upload local changes |
| `antigravity-sync pull` | Download & decrypt remote changes |
| `antigravity-sync status` | Check pending changes between local and remote |
| `antigravity-sync diff` | View file-by-file diffs |
| `antigravity-sync conflicts` | Review and resolve sync conflicts interactively |
| `antigravity-sync paths` | Manage synced paths and exclude filters |
| `antigravity-sync mcp` | Manage and sync MCP server configurations |
| `antigravity-sync update` | Rebuild from the latest commit on `main` (needs Go) |
| `antigravity-sync reset` | Reset local state or clear remote bucket |

---

## License

MIT, see [LICENSE](LICENSE).
