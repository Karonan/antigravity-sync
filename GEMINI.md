# GEMINI.md

This file provides architectural guidance and reference for Antigravity when working with code in this repository.

## Repository Purpose

`antigravity-sync` is a Go CLI that syncs `~/.gemini/` (Antigravity conversations, summaries, artifacts, project definitions, custom skills, and configs) across devices via encrypted cloud storage with Cloudflare R2 (10 GB free tier, zero egress fees). Files are gzip-compressed, then age-encrypted before upload. There are no prebuilt releases: it is installed with `go install ...@main`, and `antigravity-sync update` does the same and swaps the running binary in place.

## Common Commands

```bash
make build            # build ./bin/antigravity-sync for the host platform
make test             # go test -v ./...
make fmt              # go fmt ./...
make build-all        # cross-compile darwin/linux/windows arm64/amd64 into ./bin/
```

## Architecture

Layered with a pluggable storage abstraction:

- **CLI layer** — `cmd/antigravity-sync/main.go`. Cobra commands (`init`, `push`, `pull`, `status`, `diff`, `conflicts`, `reset`, `update`, `mcp`, `paths`) plus Survey-driven interactive wizards.
- **Sync layer** — `internal/sync/`. `Syncer` orchestrates push/pull; `SyncState` (`state.json`) tracks per-file SHA256 hash + size + mtime + last-uploaded time. `DetectChanges` compares local files against state to produce `add/modify/delete` work items.
- **Crypto layer** — `internal/crypto/encrypt.go`. Wraps `filippo.io/age` (X25519 + ChaCha20-Poly1305). Supports two key modes: random (`GenerateKey`) or passphrase-derived (`GenerateKeyFromPassphrase`, Argon2id with a **fixed salt** `sha256("antigravity-sync-v1")` so the same passphrase yields the same key on any device).
- **Storage layer** — `internal/storage/`. `Storage` interface (`Upload`/`Download`/`Delete`/`DeleteBatch`/`List`/`Head`/`BucketExists`) implemented for Cloudflare R2 via AWS SDK v2 (`r2.go`).
- **Config layer** — `internal/config/config.go`. YAML at `~/.antigravity-sync/config.yaml` (perms 0600). `SyncPaths` defines what gets synced under `~/.gemini/`.

### On-disk Layout

```
~/.antigravity-sync/    # tool's own state
├── config.yaml         # storage + encryption config
├── age-key.txt         # encryption identity (derived or random)
└── state.json          # per-file hash/size/mtime + last push/pull times

~/.gemini/              # what gets synced
├── antigravity/
│   ├── conversations/               # SQLite <id>.db files
│   ├── conversation_summaries.db    # SQLite summary metadata table
│   ├── agyhub_summaries_proto.pb    # Summary cache
│   └── brain/                       # Artifacts, transcripts, logs
├── config/
│   ├── projects/                    # <project_id>.json files (prevents "Outside of Project")
│   ├── skills/                      # Custom agent skills
│   ├── plugins/                     # Custom plugins
│   ├── config.json                  # Global preferences
│   └── mcp_config.json              # MCP servers configuration
└── GEMINI.md                        # Global instructions
```
