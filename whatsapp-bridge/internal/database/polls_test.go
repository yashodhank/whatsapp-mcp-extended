package database

import "testing"

func TestStoreAndGetPollOptions(t *testing.T) {
	tests := []struct {
		name      string
		messageID string
		chatJID   string
		options   []string
	}{
		{
			name:      "typical poll",
			messageID: "3EB0POLL1",
			chatJID:   "918668575753@s.whatsapp.net",
			options:   []string{"Expected — dev/test, no action needed", "Not expected — please escalate/fix", "Need to discuss first"},
		},
		{
			name:      "single option",
			messageID: "3EB0POLL2",
			chatJID:   "918668575753@s.whatsapp.net",
			options:   []string{"Yes"},
		},
		{
			name:      "empty options",
			messageID: "3EB0POLL3",
			chatJID:   "918668575753@s.whatsapp.net",
			options:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestMessageStore(t)

			if err := store.StorePollOptions(tt.messageID, tt.chatJID, tt.options); err != nil {
				t.Fatalf("StorePollOptions: %v", err)
			}

			got, err := store.GetPollOptions(tt.messageID, tt.chatJID)
			if err != nil {
				t.Fatalf("GetPollOptions: %v", err)
			}
			if len(got) != len(tt.options) {
				t.Fatalf("got %d options, want %d (%v vs %v)", len(got), len(tt.options), got, tt.options)
			}
			for i := range tt.options {
				if got[i] != tt.options[i] {
					t.Errorf("option[%d] = %q, want %q", i, got[i], tt.options[i])
				}
			}
		})
	}
}

func TestStorePollOptionsOverwritesPreviousSet(t *testing.T) {
	store := newTestMessageStore(t)
	const msgID, chatJID = "3EB0POLL", "918668575753@s.whatsapp.net"

	if err := store.StorePollOptions(msgID, chatJID, []string{"A", "B", "C"}); err != nil {
		t.Fatalf("first StorePollOptions: %v", err)
	}
	if err := store.StorePollOptions(msgID, chatJID, []string{"X", "Y"}); err != nil {
		t.Fatalf("second StorePollOptions: %v", err)
	}

	got, err := store.GetPollOptions(msgID, chatJID)
	if err != nil {
		t.Fatalf("GetPollOptions: %v", err)
	}
	want := []string{"X", "Y"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("option[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestGetPollOptionsUnknownPollReturnsEmptyNotError(t *testing.T) {
	store := newTestMessageStore(t)

	got, err := store.GetPollOptions("never-seen-this-poll", "918668575753@s.whatsapp.net")
	if err != nil {
		t.Fatalf("GetPollOptions on unknown poll should not error, got: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no options for an unknown poll, got %v", got)
	}
}

func TestPollOptionsAreScopedByChat(t *testing.T) {
	store := newTestMessageStore(t)
	const msgID = "3EB0POLL"

	if err := store.StorePollOptions(msgID, "chatA@s.whatsapp.net", []string{"A1", "A2"}); err != nil {
		t.Fatalf("store chatA: %v", err)
	}
	if err := store.StorePollOptions(msgID, "chatB@s.whatsapp.net", []string{"B1"}); err != nil {
		t.Fatalf("store chatB: %v", err)
	}

	gotA, err := store.GetPollOptions(msgID, "chatA@s.whatsapp.net")
	if err != nil {
		t.Fatalf("GetPollOptions chatA: %v", err)
	}
	if len(gotA) != 2 || gotA[0] != "A1" || gotA[1] != "A2" {
		t.Errorf("chatA options = %v, want [A1 A2]", gotA)
	}

	gotB, err := store.GetPollOptions(msgID, "chatB@s.whatsapp.net")
	if err != nil {
		t.Fatalf("GetPollOptions chatB: %v", err)
	}
	if len(gotB) != 1 || gotB[0] != "B1" {
		t.Errorf("chatB options = %v, want [B1]", gotB)
	}
}
