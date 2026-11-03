package sync

import (
	"context"
	"testing"
)

// protoBytes encodes a length-delimited protobuf field (short values only).
func protoBytes(field int, val []byte) []byte {
	return append([]byte{byte(field<<3 | 2), byte(len(val))}, val...)
}

func testCache(id, title string) []byte {
	summary := append(protoBytes(1, []byte(title)), 0x10, 78) // title, then a varint field
	return append(protoBytes(1, []byte(id)), protoBytes(2, summary)...)
}

func TestChatTitles(t *testing.T) {
	const id = "11111111-2222-3333-4444-555555555555"
	cache := append(protoBytes(1, testCache(id, "Weekly Planning")), protoBytes(2, []byte("x"))...)

	if got := chatTitles(cache)[id]; got != "Weekly Planning" {
		t.Errorf("title = %q", got)
	}
	// Truncated input must not panic
	for i := range cache {
		chatTitles(cache[:i])
	}
}

func TestChatID(t *testing.T) {
	const id = "11111111-2222-3333-4444-555555555555"
	cases := []struct{ path, rest string }{
		{"antigravity/conversations/" + id + ".db", ""},
		{"antigravity/conversations/" + id + ".db.bak", ".db.bak"},
		{"antigravity/brain/" + id + "/.system_generated/logs/transcript.jsonl", ".system_generated/logs/transcript.jsonl"},
	}
	for _, c := range cases {
		gotID, rest, ok := ChatID(c.path)
		if !ok || gotID != id || rest != c.rest {
			t.Errorf("ChatID(%q) = %q, %q, %v", c.path, gotID, rest, ok)
		}
	}
	for _, p := range []string{"config/projects/" + id + ".json", "antigravity/conversation_summaries.db", "antigravity/brain/notes.md"} {
		if _, _, ok := ChatID(p); ok {
			t.Errorf("ChatID(%q) should not be a chat", p)
		}
	}
}

// A chat that only the other device has is named from the remote sidebar cache.
func TestDiffNamesRemoteOnlyChats(t *testing.T) {
	env := setupTestEnv(t)
	env.syncer.cfg.SyncPaths = []string{"antigravity"}
	ctx := context.Background()
	const id = "11111111-2222-3333-4444-555555555555"

	for key, data := range map[string][]byte{
		summariesCachePath + ".age":                   protoBytes(1, testCache(id, "From the other device")),
		"antigravity/conversations/" + id + ".db.age": []byte("db"),
	} {
		encrypted, _ := env.syncer.encryptor.Encrypt(data)
		if err := env.store.Upload(ctx, key, encrypted); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := env.syncer.Diff(ctx)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.Path == "antigravity/conversations/"+id+".db" {
			found = true
			if e.Title != "From the other device" {
				t.Errorf("title = %q", e.Title)
			}
		}
	}
	if !found {
		t.Error("remote-only chat missing from diff")
	}
}
