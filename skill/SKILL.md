---
name: antigravity-sync
description: Synchronize Antigravity conversation databases, project registries, artifacts, custom skills, and MCP configs across devices via encrypted cloud storage (R2, S3, GCS, WebDAV) with age encryption.
---

# Antigravity Sync

`antigravity-sync` is an end-to-end encrypted tool designed to sync Google Antigravity state across multiple machines.

## Storage Layout Managed
- **Conversations**: `~/.gemini/antigravity/conversations/*.db` and `conversation_summaries.db`
- **Protobuf Cache**: `~/.gemini/antigravity/agyhub_summaries_proto.pb`
- **Brain Artifacts & Transcripts**: `~/.gemini/antigravity/brain/`
- **Project Registry**: `~/.gemini/config/projects/*.json` (prevents conversations demoting to "Outside of Project")
- **Custom Skills**: `~/.gemini/config/skills/`
- **Custom Plugins**: `~/.gemini/config/plugins/`
- **Configuration & Rules**: `~/.gemini/config/config.json` and `~/.gemini/GEMINI.md`

## CLI Commands

### 1. Initialize
Run `antigravity-sync init` to launch the interactive configuration wizard:
- Select storage provider (Cloudflare R2, AWS S3, Google Cloud Storage, S3-compatible, or WebDAV)
- Enter credentials and bucket name
- Choose encryption (Passphrase-derived or age key)
- Test connection and key verification

### 2. Push Local State
```bash
antigravity-sync push
```
Scans for local changes under `~/.gemini/`, compresses with gzip, encrypts with `age`, and uploads to remote storage.

### 3. Pull Remote State
```bash
antigravity-sync pull
```
Downloads remote changes, decrypts, and safely unpacks them into `~/.gemini/`. Detects if Antigravity is actively running and prompts to avoid SQLite WAL locking or memory overwrite issues.

### 4. Check Status & Diff
```bash
antigravity-sync status
antigravity-sync diff
```
Shows added, modified, deleted, or conflicted files without applying changes.

### 5. Conflict Resolution
```bash
antigravity-sync conflicts
```
Interactively resolves conflicts (keeping local or remote versions).
