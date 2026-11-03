package sync

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/Karonan/antigravity-sync/internal/storage"
)

// summariesCachePath is Antigravity's sidebar cache, which holds every chat's title.
const summariesCachePath = "antigravity/agyhub_summaries_proto.pb"

// chatLists are rewritten by Antigravity every time it opens, so they nearly
// always differ on both devices even when no chat changed.
var chatLists = map[string]bool{
	"antigravity/conversation_summaries.db": true,
	summariesCachePath:                      true,
}

// remoteWinsFunc returns a check for files changed on both sides: true means
// take the remote copy instead of saving it as a .conflict. That holds for the
// chat lists while this device has no chat changes to push, because the remote
// copies then already list every chat.
func (s *Syncer) remoteWinsFunc() func(relPath string) bool {
	var once sync.Once
	pending := true
	return func(relPath string) bool {
		if !chatLists[relPath] {
			return false
		}
		once.Do(func() {
			changes, err := s.state.DetectChanges(s.claudeDir, s.syncPaths(), s.isExcluded)
			pending = err != nil || slices.ContainsFunc(changes, func(c FileChange) bool {
				_, _, isChat := ChatID(c.Path)
				return isChat
			})
		})
		return !pending
	}
}

// deletedElsewhere returns local files another device deleted: synced before,
// gone from remote now, and unchanged here since. A file changed here stays,
// and push uploads it again. Nothing counts as deleted while some remote keys
// can't be mapped to local paths, since those files only look missing.
func (s *Syncer) deletedElsewhere(localFiles map[string]os.FileInfo, remoteFiles map[string]storage.ObjectInfo, skipped []string) map[string]bool {
	deleted := map[string]bool{}
	if len(skipped) > 0 || len(remoteFiles) == 0 {
		return deleted
	}
	for relPath := range localFiles {
		stateFile := s.state.GetFile(relPath)
		if _, exists := remoteFiles[relPath]; exists || stateFile == nil {
			continue
		}
		if hash, err := HashFile(filepath.Join(s.claudeDir, relPath)); err == nil && hash == stateFile.Hash {
			deleted[relPath] = true
		}
	}
	return deleted
}

// removeLocal moves a file deleted on another device into trashDir, keeping
// its relative path so it can be put back.
func (s *Syncer) removeLocal(relPath, trashDir string) error {
	src := filepath.Join(s.claudeDir, relPath)
	dst := filepath.Join(trashDir, relPath)
	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err != nil {
		return err
	}
	if strings.HasSuffix(relPath, ".db") {
		// SQLite side files belong to the database just moved
		for _, suffix := range []string{"-wal", "-shm", "-journal"} {
			_ = os.Rename(src+suffix, dst+suffix)
		}
	}
	// Drop folders left empty, such as a removed chat's brain/<id>/
	root := filepath.Clean(s.claudeDir)
	for dir := filepath.Dir(src); dir != root && strings.HasPrefix(dir, root) && os.Remove(dir) == nil; dir = filepath.Dir(dir) {
	}
	s.state.RemoveFile(relPath)
	return nil
}

// ChatID returns the chat a synced path belongs to, and the part of the path
// after the chat's id ("" for the chat's own .db file).
func ChatID(relPath string) (id, rest string, ok bool) {
	switch {
	case strings.HasPrefix(relPath, "antigravity/brain/"):
		id, rest, _ = strings.Cut(strings.TrimPrefix(relPath, "antigravity/brain/"), "/")
	case strings.HasPrefix(relPath, "antigravity/conversations/"):
		name := strings.TrimPrefix(relPath, "antigravity/conversations/")
		id, rest, _ = strings.Cut(name, ".")
		if rest = "." + rest; rest == ".db" {
			rest = ""
		}
	default:
		return "", "", false
	}
	if len(id) != 36 || strings.Count(id, "-") != 4 {
		return "", "", false
	}
	return id, rest, true
}

// chatTitles reads chat titles from the sidebar cache: repeated field 1
// entries of {1: chat id, 2: summary {1: title}}.
func chatTitles(cache []byte) map[string]string {
	titles := map[string]string{}
	protoFields(cache, func(field int, entry []byte) {
		if field != 1 {
			return
		}
		var id, title string
		protoFields(entry, func(field int, val []byte) {
			switch field {
			case 1:
				id = string(val)
			case 2:
				protoFields(val, func(field int, val []byte) {
					if field == 1 {
						title = string(val)
					}
				})
			}
		})
		if id != "" && title != "" {
			titles[id] = title
		}
	})
	return titles
}

// protoFields calls fn for each length-delimited field of a protobuf message,
// skipping other wire types. It stops quietly at malformed input.
func protoFields(b []byte, fn func(field int, val []byte)) {
	for len(b) > 0 {
		key, n := binary.Uvarint(b)
		if n <= 0 {
			return
		}
		b = b[n:]
		switch key & 7 {
		case 0:
			if _, n = binary.Uvarint(b); n <= 0 {
				return
			}
		case 1:
			n = 8
		case 5:
			n = 4
		case 2:
			size, m := binary.Uvarint(b)
			if m <= 0 || size > uint64(len(b)-m) {
				return
			}
			fn(int(key>>3), b[m:m+int(size)])
			n = m + int(size)
		default:
			return
		}
		if n > len(b) {
			return
		}
		b = b[n:]
	}
}

// addChatTitles names the chat each diff entry belongs to. Titles come from
// the local sidebar cache, then from the remote copy for chats not known here.
func (s *Syncer) addChatTitles(ctx context.Context, entries []DiffEntry, remoteFiles map[string]storage.ObjectInfo) {
	titles := map[string]string{}
	if data, err := os.ReadFile(filepath.Join(s.claudeDir, summariesCachePath)); err == nil {
		titles = chatTitles(data)
	}
	missing := false
	for _, e := range entries {
		if id, _, ok := ChatID(e.Path); ok && titles[id] == "" {
			missing = true
		}
	}
	if obj, ok := remoteFiles[summariesCachePath]; ok && missing {
		// Best-effort: without it, those chats show by id
		if data, err := s.fetch(ctx, obj.Key); err == nil {
			for id, title := range chatTitles(data) {
				if titles[id] == "" {
					titles[id] = title
				}
			}
		}
	}
	for i := range entries {
		if id, _, ok := ChatID(entries[i].Path); ok {
			entries[i].Title = titles[id]
		}
	}
}
