package sync

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"

	"github.com/Karonan/antigravity-sync/internal/storage"
)

// summariesCachePath is Antigravity's sidebar cache, which holds every chat's title.
const summariesCachePath = "antigravity/agyhub_summaries_proto.pb"

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
