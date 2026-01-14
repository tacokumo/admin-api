package session

import (
	"encoding/hex"
	"testing"
)

func TestGenerateSessionID(t *testing.T) {
	t.Parallel()

	t.Run("generates unique session IDs", func(t *testing.T) {
		t.Parallel()

		// Generate multiple session IDs
		ids := make([]string, 100)
		for i := 0; i < 100; i++ {
			id, err := GenerateSessionID()
			if err != nil {
				t.Fatalf("GenerateSessionID() failed: %v", err)
			}
			ids[i] = id
		}

		// Check all IDs are unique
		seen := make(map[string]bool)
		for _, id := range ids {
			if seen[id] {
				t.Errorf("GenerateSessionID() generated duplicate ID: %s", id)
			}
			seen[id] = true
		}
	})

	t.Run("generates session ID with correct format", func(t *testing.T) {
		t.Parallel()

		id, err := GenerateSessionID()
		if err != nil {
			t.Fatalf("GenerateSessionID() failed: %v", err)
		}

		// Check length (32 bytes encoded as hex = 64 characters)
		if len(id) != 64 {
			t.Errorf("GenerateSessionID() returned ID with wrong length: got %d, want 64", len(id))
		}

		// Check that it's valid hex
		_, err = hex.DecodeString(id)
		if err != nil {
			t.Errorf("GenerateSessionID() returned non-hex string: %v", err)
		}
	})
}
