package storage

import (
	"errors"
	"testing"
)

func TestValidateObjectID(t *testing.T) {
	tests := []struct {
		objectID string
		wantErr  bool
	}{
		{objectID: "abc123", wantErr: false},
		{objectID: "A1B2C3", wantErr: false},
		{objectID: "", wantErr: true},
		{objectID: "with-dash", wantErr: true},
		{objectID: "with_underscore", wantErr: true},
		{objectID: "abcdefghijklmnopqrstuvwxyz1234567", wantErr: true},
	}

	for _, tt := range tests {
		err := ValidateObjectID(tt.objectID)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("ValidateObjectID(%q) expected error", tt.objectID)
			}
			if !errors.Is(err, ErrInvalidObjectID) {
				t.Fatalf("ValidateObjectID(%q) expected ErrInvalidObjectID, got %v", tt.objectID, err)
			}
			continue
		}

		if err != nil {
			t.Fatalf("ValidateObjectID(%q) unexpected error: %v", tt.objectID, err)
		}
	}
}
